use crate::{
    registry::NetworkMetadata,
    types::{Address, NetworkId, NetworkType, StringAddress, StringTxHash, TransferStatus},
};

mod connector;
mod server;
mod signer;

pub use connector::{Config as ConnectorConfig, Connector};
use futures::Future;
pub use server::start as start_server;
use tonic::Status;
use tonic_codegen::transfer_response;

/// Retry a request until it succeeds while the server is unavailable.
/// If the server throws other error, it is returned.
async fn retry_request<T, C, F, E>(mut callback: C, mut err_callback: E) -> Result<T, Status>
where
    C: FnMut() -> F,
    F: Future<Output = Result<T, Status>>,
    E: FnMut(Status),
{
    loop {
        match callback().await {
            Ok(value) => return Ok(value),
            Err(error) if error.code() == tonic::Code::Unavailable => err_callback(error),
            err => return err,
        }

        tokio::time::sleep(crate::consts::RETRY_TIMEOUT).await;
    }
}

impl From<tonic_codegen::NetworkType> for NetworkType {
    fn from(ty: tonic_codegen::NetworkType) -> Self {
        match ty {
            tonic_codegen::NetworkType::NtEvm => NetworkType::Evm,
            tonic_codegen::NetworkType::NtCasper => NetworkType::Casper,
            tonic_codegen::NetworkType::NtSolana => NetworkType::Solana,
        }
    }
}

impl From<NetworkType> for tonic_codegen::NetworkType {
    fn from(ty: NetworkType) -> Self {
        match ty {
            NetworkType::Casper => tonic_codegen::NetworkType::NtCasper,
            NetworkType::Evm => tonic_codegen::NetworkType::NtEvm,
            NetworkType::Solana => tonic_codegen::NetworkType::NtSolana,
        }
    }
}

impl From<tonic_codegen::NetworkMetadata> for NetworkMetadata {
    fn from(meta: tonic_codegen::NetworkMetadata) -> Self {
        Self::new(
            meta.ty().into(),
            NetworkId::new(meta.id),
            meta.name,
            meta.node,
            meta.is_testnet,
        )
    }
}

impl From<tonic_codegen::StringNetworkAddress> for StringAddress {
    fn from(addr: tonic_codegen::StringNetworkAddress) -> Self {
        Self::new(addr.network_name, addr.address)
    }
}

impl From<&StringAddress> for tonic_codegen::StringNetworkAddress {
    fn from(addr: &StringAddress) -> Self {
        Self {
            network_name: addr.network_name().to_string(),
            address: addr.address().to_string(),
        }
    }
}

impl From<tonic_codegen::NetworkAddress> for Address {
    fn from(addr: tonic_codegen::NetworkAddress) -> Self {
        Self::new(NetworkId::new(addr.network_id), addr.address)
    }
}

impl From<tonic_codegen::StringTxHash> for StringTxHash {
    fn from(hash: tonic_codegen::StringTxHash) -> Self {
        Self::new(hash.network_name, hash.hash)
    }
}

impl From<&StringTxHash> for tonic_codegen::StringTxHash {
    fn from(hash: &StringTxHash) -> Self {
        Self {
            network_name: hash.network_name().to_string(),
            hash: hash.hash().to_string(),
        }
    }
}

impl From<TransferStatus> for transfer_response::Status {
    fn from(status: TransferStatus) -> Self {
        match status {
            TransferStatus::Waiting => transfer_response::Status::Waiting,
            TransferStatus::Confirming => transfer_response::Status::Confirming,
            TransferStatus::Cancelled => transfer_response::Status::Cancelled,
            TransferStatus::Finished => transfer_response::Status::Finished,
        }
    }
}

impl From<&Address> for tonic_codegen::Address {
    fn from(addr: &Address) -> Self {
        Self {
            address: addr.data().to_vec(),
        }
    }
}
