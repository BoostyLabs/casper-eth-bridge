use ethers::{
    prelude::{ContractError, Http},
    providers::{Provider, ProviderError},
};

/// The error type for the oracle.
#[derive(thiserror::Error, Debug)]
pub enum OracleError {
    #[error(transparent)]
    Anyhow(#[from] anyhow::Error),

    #[error(transparent)]
    ProviderError(#[from] ProviderError),

    #[error(transparent)]
    ContractError(#[from] ContractError<Provider<Http>>),

    #[error("Latest price is negative for {0} token")]
    NegativeLatestPrice(String),
}
