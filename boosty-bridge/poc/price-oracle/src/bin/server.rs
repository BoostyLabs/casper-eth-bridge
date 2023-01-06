use std::{net::SocketAddr, path::PathBuf};

use anyhow::Context;
use clap::Parser;
use price_oracle::oracle::{PriceOracle, Subscription};
use reqwest::Url;
use serde::Deserialize;
use tonic_codegen::bridge_oracle_server::BridgeOracleServer;
use tracing_subscriber::{fmt, EnvFilter};

#[derive(Parser, Debug)]
#[clap(about, long_about = None)]
struct Args {
    #[clap(long, short)]
    addr: SocketAddr,
    #[clap(short, long, value_parser, value_name = "oracle config file")]
    pub toml_config: PathBuf,
}

#[derive(Deserialize, Debug)]
struct Config {
    node_url: String,
    subscriptions: Vec<Subscription>,
}

async fn run() -> anyhow::Result<()> {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .event_format(fmt::format().pretty())
        .init();

    let args = Args::parse();

    let file = std::fs::read_to_string(args.toml_config).context("failed to read config file")?;
    let config: Config = toml::from_str(&file)?;

    tracing::debug!("Starting server...");
    let node_url = Url::parse(&config.node_url).context("couldn't parse node url")?;
    let oracle = PriceOracle::new_with_subscription(node_url, config.subscriptions).await?;

    tonic::transport::Server::builder()
        .add_service(BridgeOracleServer::new(oracle))
        .serve(args.addr)
        .await?;
    Ok(())
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    run().await?;
    Ok(())
}
