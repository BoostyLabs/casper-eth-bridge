use std::time::Duration;

use async_trait::async_trait;

use chrono::{DateTime, Utc};
use error::BridgeError;

use primitive_types::U256;
use registry::NetworkMetadata;

use tonic_codegen::StringNetworkAddress;
use types::{Address, StringAddress, TxHash};

pub mod bridge;
pub mod consts;
pub mod crypto;
pub mod db;
pub mod error;
pub mod grpc;
pub mod math;
pub mod registry;
pub mod time;
pub mod types;

#[cfg(test)]
pub mod test;

/// A trait for a connector to a network.
#[async_trait]
pub trait NetworkConnector: Send + Sync {
    fn metadata(&self) -> NetworkMetadata;

    async fn bridge_out(
        &self,
        recipient: Address,
        token: Address,
        amount: U256,
        source_address: StringAddress,
        transaction_id: u64,
    ) -> Result<TxHash, BridgeError>;

    async fn bridge_in_signature(
        &self,
        sender: Address,
        token: Address,
        nonce: u64,
        amount: U256,
        destination: StringNetworkAddress,
        gas_comission: U256,
    ) -> Result<tonic_codegen::BridgeInSignatureResponse, BridgeError>;
    async fn cancel_signature(
        &self,
        token: Address,
        recipient: Address,
        nonce: u64,
        comission: U256,
        amount: U256,
    ) -> Result<tonic_codegen::CancelTransferResponse, BridgeError>;

    async fn estimate_transfer(
        &self,
        amount: U256,
        recipient_network_name: String,
    ) -> Result<tonic_codegen::EstimateTransferResponse, BridgeError>;
}

/// A trait for a time source. We use it for mocking time in tests.
#[async_trait]
pub trait TimeSource: Send + Sync {
    fn now(&self) -> DateTime<Utc>;

    async fn sleep(&self, duration: Duration);
    async fn sleep_until(&self, until: DateTime<Utc>);
}
