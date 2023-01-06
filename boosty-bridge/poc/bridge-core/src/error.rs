use thiserror::Error;

use crate::types::{Address, NetworkId, TokenId};

/// Error type for the bridge. All errors are wrapped in this type.
#[derive(Error, Debug)]
pub enum BridgeError {
    #[error(transparent)]
    Registry(#[from] RegistryError),

    #[error(transparent)]
    Connector(#[from] ConnectorError),

    #[error(transparent)]
    Database(#[from] DbError),

    #[error(transparent)]
    CryptoError(#[from] CryptoError),

    #[error(transparent)]
    Other(#[from] anyhow::Error),
}

/// Error types that can be returned by the connector.
#[derive(Error, Debug)]
pub enum ConnectorError {
    #[error("incorrect network id passed to connector, expected {expected}, but got {actual}")]
    NetworkIdMismatch { expected: TokenId, actual: TokenId },
    #[error("passed event (block number {block_number}, transaction_id {transaction_id}) that were already reported")]
    EventDuplicate {
        block_number: u64,
        transaction_id: i64,
    },
}

/// Error types that can be returned by the registry.
#[derive(Error, Debug)]
pub enum RegistryError {
    #[error("unknown network id {0}")]
    UnknownNetworkId(NetworkId),
    #[error("unknown network name {0}")]
    UnknownNetworkName(String),

    #[error("unknown token id {0}")]
    UnknownTokenId(TokenId),
    #[error("unknown token name {0}")]
    UnknownTokenName(String),
    #[error("unknown token address {0}")]
    UnknownTokenAddress(Address),

    #[error("unknown token/network ({0}/{1}) combination")]
    UnknownNetworkOrToken(TokenId, NetworkId),

    #[error("invalid address length, expected {expected} got {actual}")]
    InvalidAddressLength { expected: usize, actual: usize },
    #[error("invalid address format: {reason}")]
    InvalidAddressFormat { reason: String },

    #[error("invalid txhash length, expected {expected} got {actual}")]
    InvalidTxHashLength { expected: usize, actual: usize },
    #[error("invalid txhash format: {reason}")]
    InvalidTxHashFormat { reason: String },

    #[error(transparent)]
    Other(#[from] anyhow::Error),
}

/// Error types that can be returned by the database.
#[derive(Error, Debug)]
pub enum DbError {
    #[error(transparent)]
    Sqlx(#[from] sqlx::Error),

    #[error(transparent)]
    SeaQuery(#[from] sea_query::error::Error),
}

/// Error types that can be returned by the crypto module.
#[derive(Error, Debug)]
pub enum CryptoError {
    #[error("invalid signature format")]
    InvalidSignatureFormat,

    #[error("invalid key format")]
    InvalidKeyFormat,

    #[error("key recovery failed")]
    KeyRecoveryFailed,

    #[error("message verification failed")]
    VerificationFailed,

    #[error("algorithm mismatch")]
    AlgorithmMismatch,

    #[error("missing public key")]
    MissingPublicKey,
}
