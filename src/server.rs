use crate::store::Store;
use crate::resp::RespValue;
use tokio::net::TcpListener;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::sync::{Semaphore, mpsc};
use tokio::time::{timeout, Duration};
use std::sync::Arc;
use log::{info, warn, error, debug};

const MAX_CONNECTIONS: usize = 1000;
const CONNECTION_TIMEOUT: Duration = Duration::from_secs(5);
const QUEUE_CAPACITY: usize = 10000;

pub struct Server {
    store: Arc<Store>,
    connection_semaphore: Arc<Semaphore>,
    connection_queue: mpsc::Sender<tokio::net::TcpStream>,
}

impl Server {
    pub fn new(store: Arc<Store>) -> Self {
        let (tx, rx) = mpsc::channel(QUEUE_CAPACITY);
        let server = Server {
            store: Arc::clone(&store),
            connection_semaphore: Arc::new(Semaphore::new(MAX_CONNECTIONS)),
            connection_queue: tx,
        };

        let store_clone = Arc::clone(&store);
        let sem_clone = Arc::clone(&server.connection_semaphore);
        tokio::spawn(async move {
            Server::process_queue(store_clone, sem_clone, rx).await;
        });
        info!("Server initialized with {} max connections", MAX_CONNECTIONS);

        server
    }

    pub async fn run(&self, addr: &str) -> Result<(), Box<dyn std::error::Error>> {
        let listener = TcpListener::bind(addr).await?;
        info!("Server listening on {}", addr);

        loop {
            let (socket, _) = listener.accept().await?;
            debug!("New connection accepted");
            if let Err(_) = self.connection_queue.send(socket).await {
                error!("Connection queue full, rejecting connection");
            }
        }
    }

    async fn process_queue(
        store: Arc<Store>,
        sem: Arc<Semaphore>,
        mut rx: mpsc::Receiver<tokio::net::TcpStream>
    ) {
        debug!("Starting to process connection queue");
        while let Some(socket) = rx.recv().await {
            debug!("Processing new connection from queue");
            let store = Arc::clone(&store);
            let sem = Arc::clone(&sem);

            tokio::spawn(async move {
                match timeout(CONNECTION_TIMEOUT, sem.acquire()).await {
                    Ok(Ok(permit)) => {
                        Self::handle_connection(socket, store).await;
                        drop(permit);
                    },
                    _ => {
                        warn!("Connection timed out while waiting for processing");
                        let error_response = RespValue::Error("Server is busy. Try again later.".to_string());
                        if let Err(e) = socket.writable().await.and_then(|_| {
                            socket.try_write(&error_response.serialize())
                        }) {
                            error!("Failed to send error response: {:?}", e);
                        }
                    }
                }
            });
        }
    }

    async fn handle_connection(mut socket: tokio::net::TcpStream, store: Arc<Store>) {
        debug!("Handling new connection");
        let mut buffer = Vec::new();

        loop {
            let mut temp_buffer = [0; 1024];
            match socket.read(&mut temp_buffer).await {
                Ok(0) => break,
                Ok(n) => {
                    buffer.extend_from_slice(&temp_buffer[..n]);
                    debug!("Received data: {:?}", String::from_utf8_lossy(&buffer));
                },
                Err(e) => {
                    error!("Error reading from socket: {:?}", e);
                    break;
                },
            }

            while !buffer.is_empty() {
                match RespValue::parse(&buffer) {
                    Ok((value, consumed)) => {
                        buffer.drain(..consumed);
                        let response = Self::handle_command(value, &store);
                        let serialized = response.serialize();
                        debug!("Sent response: {:?}", String::from_utf8_lossy(&serialized));
                        if let Err(e) = socket.write_all(&serialized).await {
                            error!("Error writing to socket: {:?}", e);
                            break;
                        }
                    }
                    Err(e) => {
                        error!("Error parsing RESP: {:?}", e);
                        break;
                    },
                }
            }
        }
    }

    fn handle_command(command: RespValue, store: &Store) -> RespValue {
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
}
