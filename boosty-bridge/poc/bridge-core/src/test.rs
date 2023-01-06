use std::{
    net::{Ipv4Addr, SocketAddrV4, TcpListener, ToSocketAddrs},
    sync::Arc,
    time::Duration,
};

use chrono::{TimeZone, Utc};
use pg_embed::postgres::PgEmbed;
use tracing::metadata::LevelFilter;
use tracing_subscriber::fmt;

use crate::{
    bridge::Bridge,
    db::BridgeWriteQueries,
    time::MockTimeSource,
    types::{Address, NetworkId, TokenId},
};

use self::{
    connector::{MockConnector, CASPER_NID, EVM_NID},
    db::{db_config, init_pg},
};

pub mod connector;
pub mod db;
pub mod transfers;

pub struct TestContext {
    pub pg: PgEmbed,
    pub bridge: Arc<Bridge>,
    pub casper: Arc<MockConnector>,
    pub evm: Arc<MockConnector>,
    pub time_source: MockTimeSource,
}

pub struct TestToken {
    pub id: TokenId,
    pub casper: Address,
    pub evm: Address,
}

fn test_bind_tcp<A: ToSocketAddrs>(addr: A) -> Option<u16> {
    Some(TcpListener::bind(addr).ok()?.local_addr().ok()?.port())
}

pub fn is_free_tcp(port: u16) -> bool {
    let ipv4 = SocketAddrV4::new(Ipv4Addr::UNSPECIFIED, port);
    test_bind_tcp(ipv4).is_some()
}

pub fn allocate_port() -> u16 {
    let mut port = 10000;

    loop {
        if is_free_tcp(port) {
            return port;
        }

        port += 1;
    }
}

impl TestContext {
    pub async fn create() -> Self {
        // For debug purpose
        // tracing_subscriber::fmt()
        //     .with_max_level(LevelFilter::DEBUG)
        //     .event_format(fmt::format().compact())
        //     .init();

        let default_time = Utc.with_ymd_and_hms(2020, 1, 1, 0, 0, 0).unwrap();

        let port = allocate_port();
        let pg = init_pg(port).await;
        let db_config = db_config(port);
        let time_source = MockTimeSource::new(default_time);

        let bridge = Bridge::start(
            db_config,
            crate::bridge::Config::default(),
            Box::new(time_source.clone()),
        )
        .await
        .unwrap();

        let bridge = Arc::new(bridge);

        let casper = MockConnector::start_casper(bridge.clone()).await.unwrap();
        let evm = MockConnector::start_evm(bridge.clone()).await.unwrap();
        let mut tx = bridge.db().write_tx().await.unwrap();

        tx.create_tables().await.unwrap();
        tx.commit().await.unwrap();

        // Initialization takes time...
        tokio::time::sleep(Duration::from_secs(5)).await;

        Self {
            bridge,
            casper,
            evm,
            pg,
            time_source,
        }
    }

    pub async fn restart_bridge(&mut self) {
        self.bridge.shutdown().await;
        let db_config = db_config(self.pg.pg_settings.port as u16);
        self.bridge = Arc::new(
            Bridge::start(
                db_config,
                crate::bridge::Config::default(),
                Box::new(self.time_source.clone()),
            )
            .await
            .unwrap(),
        );

        self.casper = MockConnector::start_casper(self.bridge.clone())
            .await
            .unwrap();
        self.evm = MockConnector::start_evm(self.bridge.clone()).await.unwrap();

        self.bridge.load_tokens().await.unwrap();

        // Initialization takes time
        tokio::time::sleep(Duration::from_secs(5)).await;
    }

    pub async fn create_token(&self, short_name: &str, long_name: &str) -> TokenId {
        let mut tx = self.bridge.db().write_tx().await.unwrap();

        let id = tx.insert_token(short_name, long_name).await.unwrap();
        tx.commit().await.unwrap();

        id
    }

    pub async fn create_network_token(
        &self,
        token_id: TokenId,
        network_id: NetworkId,
        decimals: u8,
    ) -> Address {
        let network_ty = self
            .bridge
            .network_registry()
            .read()
            .by_id(network_id)
            .unwrap()
            .ty();

        let address = Address::random(network_id, network_ty);

        let mut tx = self.bridge.db().write_tx().await.unwrap();
        tx.insert_network_token(network_id, token_id, address.clone(), decimals)
            .await
            .unwrap();
        tx.commit().await.unwrap();

        address
    }

    pub async fn create_test_token(&self) -> TestToken {
        let token_id = self.create_token("TEST", "Test Token").await;
        let casper = self.create_network_token(token_id, CASPER_NID, 9).await;
        let evm = self.create_network_token(token_id, EVM_NID, 18).await;

        TestToken {
            id: token_id,
            casper,
            evm,
        }
    }
}

impl Drop for TestContext {
    fn drop(&mut self) {
        self.pg.stop_db_sync().unwrap();
    }
}

#[tokio::test]
async fn init() {
    let ctx = TestContext::create().await;

    let networks = ctx
        .bridge
        .network_registry()
        .read()
        .all()
        .cloned()
        .collect::<Vec<_>>();

    assert!(networks.iter().any(|nm| nm.name() == "CASPER-TEST"));
    assert!(networks.iter().any(|nm| nm.name() == "GOERLI"));
}
