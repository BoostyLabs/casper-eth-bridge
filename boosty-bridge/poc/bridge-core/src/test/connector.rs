use std::{
    iter::repeat_with,
    sync::{
        atomic::{AtomicBool, AtomicU64, Ordering},
        Arc,
    },
};

use async_trait::async_trait;
use parking_lot::Mutex;
use primitive_types::U256;
use tokio::sync::mpsc::Sender;
use tonic_codegen::StringNetworkAddress;

use crate::{
    bridge::Bridge,
    error::BridgeError,
    registry::NetworkMetadata,
    types::{
        Address, BridgeEvent, BridgeTokenTransferIn, BridgeTokenTransferOut, ConfirmedTx,
        NetworkId, NetworkType, StringAddress, TokenId, TxHash,
    },
    NetworkConnector,
};

pub const CASPER_NID: NetworkId = NetworkId::new(0);
pub const EVM_NID: NetworkId = NetworkId::new(1);

pub struct MockConnector {
    metadata: NetworkMetadata,
    event_tx: Sender<BridgeEvent>,
    block_counter: AtomicU64,
    hash_counter: AtomicU64,
    recorded_events: Mutex<Vec<BridgeEvent>>,
    failing: AtomicBool,
    hash_prefix: [u8; 8],
}

impl MockConnector {
    pub async fn start(
        metadata: NetworkMetadata,
        bridge: Arc<Bridge>,
    ) -> Result<Arc<Self>, BridgeError> {
        let event_tx = bridge.event_tx().clone();
        let hash_prefix = repeat_with(|| fastrand::u8(..))
            .take(8)
            .collect::<Vec<_>>()
            .try_into()
            .unwrap();

        let connector = Arc::new(Self {
            metadata,
            event_tx,
            block_counter: AtomicU64::new(1),
            hash_counter: AtomicU64::new(1),
            recorded_events: Default::default(),
            hash_prefix,
            failing: false.into(),
        });

        bridge.register_connector(connector.clone()).await?;

        Ok(connector.clone())
    }

    pub async fn start_casper(bridge: Arc<Bridge>) -> Result<Arc<Self>, BridgeError> {
        let metadata = NetworkMetadata::new(
            NetworkType::Casper,
            CASPER_NID,
            "CASPER-TEST".into(),
            "void".into(),
            true,
        );

        Self::start(metadata, bridge).await
    }

    pub async fn start_evm(bridge: Arc<Bridge>) -> Result<Arc<Self>, BridgeError> {
        let metadata = NetworkMetadata::new(
            NetworkType::Evm,
            EVM_NID,
            "GOERLI".into(),
            "void".into(),
            true,
        );

        Self::start(metadata, bridge).await
    }

    pub fn switch_failing(&self) {
        let value = self.failing.load(Ordering::SeqCst);
        self.failing.store(!value, Ordering::SeqCst);
    }

    pub async fn bridge_in(
        &self,
        from: &Address,
        to: &StringAddress,
        token: &Address,
        amount: U256,
    ) {
        let from = Address::new(self.metadata.id(), from.data().to_vec());

        let event = BridgeEvent::TokenTransferIn(BridgeTokenTransferIn {
            from,
            to: to.clone(),
            amount,
            token: token.clone(),
            tx: self.generate_tx(),
        });

        self.recorded_events.lock().push(event.clone());
        self.event_tx.send(event).await.unwrap();
    }

    pub fn generate_tx(&self) -> ConfirmedTx {
        let sender = match self.metadata.ty() {
            NetworkType::Casper => {
                let mut v = [0u8].repeat(31);
                v.push(42);
                v
            }
            NetworkType::Evm => {
                let mut v = [0u8].repeat(19);
                v.push(42);
                v
            }
            NetworkType::Solana => {
                let mut v = [0u8].repeat(31);
                v.push(42);
                v
            }
        };

        let block_number = self.block_counter.fetch_add(1, Ordering::SeqCst);
        let hash_number = self.hash_counter.fetch_add(1, Ordering::SeqCst);

        let mut hash = hash_number.to_be_bytes().repeat(4);
        hash[..8].copy_from_slice(&self.hash_prefix);
        let hash = TxHash::new(self.metadata.id(), hash);
        let sender = Address::new(self.metadata.id(), sender);

        ConfirmedTx::new(hash, sender, block_number)
    }

    pub fn generate_address(&self) -> Address {
        Address::random(self.metadata.id(), self.metadata.ty())
    }

    pub fn view_events(&self) -> Vec<BridgeEvent> {
        self.recorded_events.lock().clone()
    }
}

#[async_trait]
impl NetworkConnector for MockConnector {
    fn metadata(&self) -> NetworkMetadata {
        self.metadata.clone()
    }

    async fn bridge_out(
        &self,
        to: Address,
        token: Address,
        amount: U256,
        from: StringAddress,
        _: u64,
    ) -> Result<TxHash, BridgeError> {
        if self.failing.load(Ordering::SeqCst) {
            return Err(BridgeError::Connector(
                crate::error::ConnectorError::NetworkIdMismatch {
                    expected: TokenId::new(255),
                    actual: TokenId::new(255),
                },
            ));
        }

        let tx = self.generate_tx();

        let tx_hash = tx.hash().clone();

        let event = BridgeEvent::TokenTransferOut(BridgeTokenTransferOut {
            from,
            to,
            token,
            amount,
            tx,
        });

        self.recorded_events.lock().push(event.clone());
        self.event_tx.send(event).await.unwrap();

        Ok(tx_hash)
    }

    async fn bridge_in_signature(
        &self,
        _sender: Address,
        _token: Address,
        _nonce: u64,
        _amount: U256,
        _destination: StringNetworkAddress,
        _gas_comission: U256,
    ) -> Result<tonic_codegen::BridgeInSignatureResponse, BridgeError> {
        unimplemented!("bridge_in_signature: for now, we don't need this method")
    }
    async fn cancel_signature(
        &self,
        _token: Address,
        _recipient: Address,
        _nonce: u64,
        _comission: U256,
        _amount: U256,
    ) -> Result<tonic_codegen::CancelTransferResponse, BridgeError> {
        unimplemented!("cancel_signature: for now, we don't need this method")
    }

    async fn estimate_transfer(
        &self,
        _amount: U256,
        _network_name: String,
    ) -> Result<tonic_codegen::EstimateTransferResponse, BridgeError> {
        unimplemented!("estimate_transfer: for now, we don't need this method")
    }
}
