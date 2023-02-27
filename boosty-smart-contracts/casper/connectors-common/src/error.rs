use thiserror::Error;

#[macro_export]
macro_rules! get_argument {
    ($value:ident) => {
        match $value {
            Some(value) => value,
            None => return Err(ConnectorError::argument_not_found(stringify!($value)))?,
        }
    };
}

#[derive(Error, Debug)]
pub enum ConnectorError {
    #[error("Argument not found: {0}")]
    ArgumentNotFound(String),
    #[error("Invalid argument: `{0}`: {1}")]
    InvalidArgument(String, String),
    #[error("Token contract ({0}) is not registered")]
    TokenContractNotFound(String),
    #[error("Blockchain client error: {0}")]
    BlockchainClientError(anyhow::Error),
    #[error("Decoding failure: `{0}`: {1}")]
    DecodingError(String, anyhow::Error),
    #[error("Encoding failure: `{0}`: {1}")]
    EncodingError(String, anyhow::Error),
    #[error("Configuration error: `{0}`: {1}")]
    ConfigurationError(String, String),
}

impl ConnectorError {
    pub fn argument_not_found(name: &str) -> Self {
        ConnectorError::ArgumentNotFound(name.to_string())
    }

    pub fn invalid_argument(name: &str, message: &str) -> Self {
        ConnectorError::InvalidArgument(name.to_string(), message.to_string())
    }

    pub fn token_contract_not_found(address: &str) -> Self {
        ConnectorError::TokenContractNotFound(address.to_string())
    }

    pub fn blockchain_client_error(error: anyhow::Error) -> Self {
        ConnectorError::BlockchainClientError(error)
    }

    pub fn decoding_error(name: &str, error: anyhow::Error) -> Self {
        ConnectorError::DecodingError(name.to_string(), error)
    }

    pub fn encoding_error(name: &str, error: anyhow::Error) -> Self {
        ConnectorError::EncodingError(name.to_string(), error)
    }

    pub fn configuration_error(name: &str, message: &str) -> Self {
        ConnectorError::ConfigurationError(name.to_string(), message.to_string())
    }
}

impl From<ConnectorError> for tonic::Status {
    fn from(error: ConnectorError) -> Self {
        match error {
            ConnectorError::ArgumentNotFound(name) => tonic::Status::new(
                tonic::Code::InvalidArgument,
                format!("Argument not found: {name}"),
            ),
            ConnectorError::InvalidArgument(name, message) => tonic::Status::new(
                tonic::Code::InvalidArgument,
                format!("Invalid argument: `{name}`: {message}"),
            ),
            ConnectorError::TokenContractNotFound(address) => tonic::Status::new(
                tonic::Code::NotFound,
                format!("Token contract `{address}` is not registered"),
            ),
            ConnectorError::BlockchainClientError(error) => tonic::Status::new(
                tonic::Code::Internal,
                format!("Blockchain client error: {:#}", error),
            ),
            ConnectorError::DecodingError(name, error) => tonic::Status::new(
                tonic::Code::Internal,
                format!("Decoding failure: `{name}`: {error}"),
            ),
            ConnectorError::EncodingError(name, error) => tonic::Status::new(
                tonic::Code::Internal,
                format!("Encoding failure: `{name}`: {error}"),
            ),
            ConnectorError::ConfigurationError(name, message) => tonic::Status::new(
                tonic::Code::Internal,
                format!("Configuration error: `{name}`: {message}"),
            ),
        }
    }
}
