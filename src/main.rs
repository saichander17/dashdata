mod store;
mod resp;
mod server;

use std::sync::Arc;
use server::Server;
use clap::Parser;
use log::{info, error};

#[derive(Parser)]
#[clap(name = "Dashdata Server")]
struct Opts {
    #[clap(short, long, default_value = "127.0.0.1:6379")]
    bind_address: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    env_logger::init();
    let opts = Opts::parse();
    let store = Arc::new(store::Store::new());
    let server = Server::new(store);

//     server.run(&opts.bind_address).await?;
    match server.run(&opts.bind_address).await {
        Ok(_) => info!("Server shutdown gracefully"),
        Err(e) => error!("Server error: {}", e),
    }

    Ok(())
}
