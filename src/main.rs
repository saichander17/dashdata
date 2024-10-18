mod store;
mod resp;
mod server;

use std::sync::Arc;
use server::Server;
use clap::Parser;

#[derive(Parser)]
#[clap(name = "Dashdata Server")]
struct Opts {
    #[clap(short, long, default_value = "127.0.0.1:6379")]
    bind_address: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let opts = Opts::parse();
    let store = Arc::new(store::Store::new());
    let server = Server::new(store);

    server.run(&opts.bind_address).await?;

    Ok(())
}
