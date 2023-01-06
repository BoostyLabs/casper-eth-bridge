use bridge_core::db::{self, BridgeWriteQueries};
use clap::Parser;
mod bridge;

#[derive(Parser)]
pub enum Command {
    CreateTables,
    GrpcConnectedNetworks { endpoint: String },
    SetupTokens,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    dotenv::dotenv().ok();

    let command = Command::parse();
    match command {
        Command::CreateTables => create_tables().await?,
        Command::GrpcConnectedNetworks { endpoint } => grpc_connected_networks(endpoint).await?,
        Command::SetupTokens => setup_tokens().await?,
    }

    Ok(())
}

async fn create_tables() -> anyhow::Result<()> {
    let db_config = db::Config::from_env()?;
    let db = db::Database::connect(db_config).await?;
    let mut tx = db.write_tx().await?;

    tx.create_tables().await?;

    Ok(())
}

async fn grpc_connected_networks(endpoint: String) -> anyhow::Result<()> {
    let mut client =
        tonic_codegen::gateway_bridge_client::GatewayBridgeClient::connect(endpoint).await?;

    let response = client.connected_networks(()).await?;

    dbg!(&response);

    Ok(())
}

async fn setup_tokens() -> anyhow::Result<()> {
    let db_config = db::Config::from_env()?;
    let db = db::Database::connect(db_config).await?;

    bridge::setup_tokens(&db).await;
    Ok(())
}
