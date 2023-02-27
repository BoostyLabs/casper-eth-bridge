use casper_types::ApiError;
use contract_util::error::ContractError;
use num_enum::{FromPrimitive, IntoPrimitive};

#[cfg_attr(std, derive(Debug, thiserror::Error))]
#[repr(u16)]
#[derive(IntoPrimitive, FromPrimitive)]
pub enum BridgeError {
    #[cfg_attr(std, error("this method is only callable by this contract"))]
    OnlyCallableBySelf = 0,

    #[cfg_attr(std, error("Transferred amount did not match expected amount"))]
    UnexpectedTransferAmount = 1,

    #[cfg_attr(std, error("Expired signature deadline"))]
    ExpiredSignature = 2,

    #[cfg_attr(std, error("Signature nonce already used"))]
    AlreadyUsedSignature = 3,

    #[cfg_attr(std, error("Invalid commission percent, above 100%"))]
    InvalidCommissionPercent = 4,

    #[cfg_attr(std, error("Amount exceed available in the commission pool"))]
    AmountExceedCommissionPool = 5,

    #[cfg_attr(std, error("Amount exceed available in the bridge pool"))]
    AmountExceedBridgePool = 6,

    #[cfg_attr(std, error("Invalid Signature"))]
    InvalidSignature = 7,

    #[cfg_attr(std, error("Signer is not established"))]
    SignerIsNotEstablished = 8,

    #[cfg_attr(std, error("Total commission bigger than transferred amount"))]
    CommissionBiggerThanTransferredAmount = 9,

    #[cfg_attr(std, error("Integer Underflow"))]
    Underflow = 253,

    #[cfg_attr(std, error("Integer Overflow"))]
    Overflow = 254,

    #[num_enum(default)]
    #[cfg_attr(std, error("unknown error"))]
    Unknown = 255,
}

impl ContractError for BridgeError {}

impl Into<ApiError> for BridgeError {
    fn into(self) -> ApiError {
        ApiError::User(contract_util::error::Error::<BridgeError>::Contract(self).into())
    }
}

impl From<ApiError> for BridgeError {
    fn from(api: ApiError) -> Self {
        match api {
            ApiError::User(code) => match contract_util::error::Error::<BridgeError>::from(code) {
                contract_util::error::Error::Contract(error) => error,
                _ => BridgeError::Unknown,
            },
            _ => BridgeError::Unknown,
        }
    }
}

pub type Error = contract_util::error::Error<BridgeError>;
