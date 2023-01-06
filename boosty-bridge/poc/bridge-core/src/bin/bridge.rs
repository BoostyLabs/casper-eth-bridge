use std::{net::SocketAddr, path::PathBuf, sync::Arc, time::Duration};

use anyhow::Context;
use bridge_core::db::BridgeWriteQueries;
use bridge_core::{
    bridge::Bridge,
    time::RealTimeSource,
    types::{address_from_casper_string, address_from_evm_string, Address, NetworkId},
};
use clap::Parser;
use pg_embed::{
    pg_fetch::{PgFetchSettings, PG_V13},
    postgres::{PgEmbed, PgSettings},
};
use reqwest::Url;
use tokio::sync::oneshot;
use tracing::{info, log::warn};
use tracing_subscriber::{fmt, EnvFilter};

#[derive(Parser, Debug)]
#[clap(about, long_about = None)]
struct Args {
    #[clap(long, short)]
    addr: SocketAddr,

    #[clap(long, short)]
    signer_addr: Url,

    #[clap(long)]
    embed_db: bool,

    #[clap(long)]
    init_tables: bool,

    #[clap(long)]
    connectors: Vec<Url>,
}

#[tokio::main]
#[tracing::instrument]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .event_format(fmt::format().pretty())
        .init();

    let args = Args::parse();
    dotenv::dotenv().ok();

    let pg = if args.embed_db {
        Some(init_pg().await)
    } else {
        None
    };

    let db_config = if args.embed_db {
        embed_config()
    } else {
        bridge_core::db::Config::from_env()?
    };

    let bridge_config = bridge_core::bridge::Config::from_env()?;

    let bridge = Bridge::start(db_config, bridge_config, Box::new(RealTimeSource)).await?;
    let bridge = Arc::new(bridge);

    let (grpc_shutdown_tx, grpc_handle) = {
        let (tx, rx) = oneshot::channel();
        let handle = tokio::spawn(bridge_core::grpc::start_server(
            args.addr,
            args.signer_addr,
            bridge.clone(),
            rx,
        ));
        (tx, handle)
    };

    if args.embed_db || args.init_tables {
        let mut tx = bridge.db().write_tx().await?;
        tx.create_tables().await?;
        tx.commit().await?;
        setup_tokens(bridge.db()).await;
    }

    bridge.load_tokens().await?;

    for connector in args.connectors {
        let bridge = bridge.clone();
        let config = bridge_core::grpc::ConnectorConfig::new(connector.to_string());
        bridge_core::grpc::Connector::start(config, bridge).await?;
    }

    tokio::signal::ctrl_c()
        .await
        .context("unable to wait for shutdown signal")?;

    warn!("received first shutdown signal. terminating gracefully");

    grpc_shutdown_tx.send(()).ok();

    tokio::select! {
        _ = tokio::signal::ctrl_c() => {
            warn!("received second shutdown signal. terminating forcefully");
        },
        _ = grpc_handle => {
            info!("grpc server shut down.");
        }
    }

    tokio::select! {
        _ = tokio::signal::ctrl_c() => {
            warn!("received second shutdown signal. terminating forcefully");
        },
        _ = bridge.shutdown() => {
            info!("bridge shut down.");
        }
    }

    if let Some(mut pg) = pg {
        info!("shutting down embedded db.");
        pg.stop_db().await.unwrap();
        info!("embedded db shut down.");
    }

    Ok(())
}

#[allow(dead_code)] // I dunno why it's dead code, cause it's used in main.
/// Returns a config for an embedded postgres instance used for local testing.
fn embed_config() -> bridge_core::db::Config {
    bridge_core::db::Config {
        host: "127.0.0.1".to_string(),
        port: 5499,
        user: "postgres".to_string(),
        pass: "password".to_string(),
        dbname: "golden_gate".to_string(),
    }
}

/// Initializes an embedded postgres instance used for local testing.
#[allow(dead_code)] // I dunno why it's dead code, cause it's used in main.
async fn init_pg() -> PgEmbed {
    let settings = PgSettings {
        database_dir: PathBuf::from("data/db"),
        port: 5499,
        user: "postgres".into(),
        password: "password".into(),
        auth_method: pg_embed::pg_enums::PgAuthMethod::Plain,
        persistent: false,
        timeout: Some(Duration::from_secs(15)),
        migration_dir: None,
    };

    let fetch_settings = PgFetchSettings {
        version: PG_V13,
        ..Default::default()
    };

    let mut pg = PgEmbed::new(settings, fetch_settings).await.unwrap();

    pg.setup().await.unwrap();
    pg.start_db().await.unwrap();
    pg.create_database("golden_gate").await.unwrap();

    pg
}

/// Sets up the testing tokens in the database.
pub async fn setup_tokens(db: &bridge_core::db::Database) {
    const CASPER_TEST_TOKEN: &str =
        "hash-3c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23";
    const EVM_TEST_TOKEN: &str = "9fF6D0788066982c95D26F4A74d6C700F3Dc29ec";

    let mut tx = db.write_tx().await.expect("failed to acquire tx");

    let token_id = tx
        .insert_token("TEST", "TEST TOKEN")
        .await
        .expect("failed to insert token");

    let casper_address = Address::new(
        NetworkId::new(0),
        address_from_casper_string(CASPER_TEST_TOKEN).unwrap(),
    );

    let evm_address = Address::new(
        NetworkId::new(1),
        address_from_evm_string(EVM_TEST_TOKEN).unwrap(),
    );

    tx.insert_network_token(NetworkId::new(0), token_id, casper_address, 9)
        .await
        .unwrap();
    tx.insert_network_token(NetworkId::new(1), token_id, evm_address, 18)
        .await
        .unwrap();

    tx.commit().await.expect("unable to commit tx");
}
