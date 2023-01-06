use thiserror::Error;

#[derive(Debug, Error)]
pub enum Error {
    #[error("main secret key ({secret}) does not match with main public key ({public})")]
    MainSecretPublicMismatch { secret: String, public: String },

    #[error("unexpected chain name ({name})")]
    UnexpectedChainName { name: String },

    #[error("config setting `{name}` needs to be set")]
    MissingConfigSetting { name: String },

    #[error("unexpected rpc response kind ({kind})")]
    UnexpectedRpcResponse { kind: String },

    #[error("json deserialization error")]
    JsonDeserError(#[from] serde_json::Error),

    #[error("missing rpc response field `{field}`")]
    MissingResponseField { field: String },

    #[error("unexpected StoredValue type (expected: {expected}, got: {got}")]
    UnexpectedStoredValueType { expected: String, got: String },

    #[error("invalid key format ({given})")]
    InvalidKeyFormat { given: String },

    #[error("rpc error: {0}")]
    RpcError(#[from] jsonrpc_lite::Error),

    #[error("transport error: {0}")]
    Transport(#[from] reqwest::Error),

    #[error("{0}")]
    Anyhow(#[from] anyhow::Error),
}
