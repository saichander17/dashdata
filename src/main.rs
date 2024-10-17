mod store;
mod resp;

use std::sync::Arc;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::TcpListener;
use resp::RespValue;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let store = Arc::new(store::Store::new());
    let listener = TcpListener::bind("127.0.0.1:8080").await?;

    println!("Server listening on port 8080");

    loop {
        let (mut socket, _) = listener.accept().await?;
        let store = Arc::clone(&store);

        tokio::spawn(async move {
            let mut buffer = Vec::new();

            loop {
                let mut chunk = [0; 1024];
                let n = match socket.read(&mut chunk).await {
                    Ok(n) if n == 0 => return,
                    Ok(n) => n,
                    Err(_) => return,
                };

                buffer.extend_from_slice(&chunk[..n]);

                match RespValue::parse(&buffer) {
                    Ok(command) => {
                        let response = handle_command(&store, command);
                        if let Err(_) = socket.write_all(&response.serialize()).await {
                            return;
                        }
                        buffer.clear();
                    },
                    Err(parse_error) => {
                        let error_response = RespValue::Error(format!("ERR parsing failed: {}", parse_error));
                        if let Err(_) = socket.write_all(&error_response.serialize()).await {
                            return;
                        }
                        buffer.clear();
                    }
                }
            }
        });
    }
}

fn handle_command(store: &Arc<store::Store>, command: RespValue) -> RespValue {
    match command {
        RespValue::Array(args) => {
            if let Some(RespValue::BulkString(cmd)) = args.get(0) {
                match cmd.to_uppercase().as_str() {
                    "GET" => {
                        if let Some(RespValue::BulkString(key)) = args.get(1) {
                            match store.get(key) {
                                Some(value) => RespValue::BulkString(value),
                                None => RespValue::Null,
                            }
                        } else {
                            RespValue::Error("ERR wrong number of arguments for 'get' command".to_string())
                        }
                    },
                    "SET" => {
                        if let (Some(RespValue::BulkString(key)), Some(RespValue::BulkString(value))) = (args.get(1), args.get(2)) {
                            store.set(key.clone(), value.clone());
                            RespValue::SimpleString("OK".to_string())
                        } else {
                            RespValue::Error("ERR wrong number of arguments for 'set' command".to_string())
                        }
                    },
                    "DEL" => {
                        if let Some(RespValue::BulkString(key)) = args.get(1) {
                            store.delete(key);
                            RespValue::SimpleString("OK".to_string())
                        } else {
                            RespValue::Error("ERR wrong number of arguments for 'del' command".to_string())
                        }
                    },
                    _ => RespValue::Error("ERR unknown command".to_string()),
                }
            } else {
                RespValue::Error("ERR invalid command".to_string())
            }
        },
        _ => RespValue::Error("ERR invalid command".to_string()),
    }
}
