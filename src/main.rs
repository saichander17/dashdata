mod store;
mod resp;

use tokio::net::TcpListener;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::sync::Semaphore;
use tokio::time::{timeout, Duration};
use std::sync::Arc;
use resp::RespValue;

const MAX_CONNECTIONS: usize = 1000;
const CONNECTION_TIMEOUT: Duration = Duration::from_secs(5);

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let store = Arc::new(store::Store::new());
    let listener = TcpListener::bind("0.0.0.0:6379").await?;
    let semaphore = Arc::new(Semaphore::new(MAX_CONNECTIONS));

    println!("Server listening on port 6379");

    loop {
        let (socket, _) = listener.accept().await?;
        let store = Arc::clone(&store);
        let sem = Arc::clone(&semaphore);

        tokio::spawn(async move {
            match timeout(CONNECTION_TIMEOUT, sem.acquire()).await {
                Ok(Ok(permit)) => {
                    handle_connection(socket, store).await;
                    drop(permit);
                },
                _ => {
                    let error_response = RespValue::Error("Server is busy. Try again later.".to_string());
                    if let Err(e) = socket.writable().await.and_then(|_| {
                        socket.try_write(&error_response.serialize())
                    }) {
                        eprintln!("Failed to send error response: {:?}", e);
                    }
                }
            }
        });
    }
}

async fn handle_connection(mut socket: tokio::net::TcpStream, store: Arc<store::Store>) {
    let mut buffer = Vec::new();

    loop {
        let mut temp_buffer = [0; 1024];
        match socket.read(&mut temp_buffer).await {
            Ok(0) => break,
            Ok(n) => {
                buffer.extend_from_slice(&temp_buffer[..n]);
                println!("Received: {:?}", String::from_utf8_lossy(&buffer));
            },
            Err(e) => {
                eprintln!("Error reading from socket: {:?}", e);
                break;
            },
        }

        while !buffer.is_empty() {
            match RespValue::parse(&buffer) {
                Ok((value, consumed)) => {
                    buffer.drain(..consumed);
                    let response = handle_command(value, &store);
                    let serialized = response.serialize();
                    println!("Sending response: {:?}", String::from_utf8_lossy(&serialized));
                    if let Err(e) = socket.write_all(&serialized).await {
                        eprintln!("Error writing to socket: {:?}", e);
                        break;
                    }
                }
                Err(e) => {
                    eprintln!("Error parsing RESP: {:?}", e);
                    break;
                },
            }
        }
    }
}

fn handle_command(command: RespValue, store: &store::Store) -> RespValue {
    match command {
        RespValue::Array(parts) => {
            if parts.len() < 1 {
                return RespValue::Error("Invalid command".to_string());
            }
            match parts[0] {
                RespValue::BulkString(Some(ref cmd)) if cmd == b"GET" => {
                    if let RespValue::BulkString(Some(key)) = &parts[1] {
                        match store.get(&String::from_utf8_lossy(key)) {
                            Some(value) => RespValue::BulkString(Some(value.into_bytes())),
                            None => RespValue::BulkString(None),
                        }
                    } else {
                        RespValue::Error("Invalid GET command".to_string())
                    }
                }
                RespValue::BulkString(Some(ref cmd)) if cmd == b"SET" => {
                    if let (RespValue::BulkString(Some(key)), RespValue::BulkString(Some(value))) = (&parts[1], &parts[2]) {
                        store.set(String::from_utf8_lossy(key).to_string(), String::from_utf8_lossy(value).to_string());
                        RespValue::SimpleString("OK".to_string())
                    } else {
                        RespValue::Error("Invalid SET command".to_string())
                    }
                }
                RespValue::BulkString(Some(ref cmd)) if cmd == b"DEL" => {
                    if let RespValue::BulkString(Some(key)) = &parts[1] {
                        store.delete(&String::from_utf8_lossy(key));
                        RespValue::SimpleString("OK".to_string())
                    } else {
                        RespValue::Error("Invalid DEL command".to_string())
                    }
                }
                _ => RespValue::Error("Unknown command".to_string()),
            }
        }
        _ => RespValue::Error("Invalid command format".to_string()),
    }
}
