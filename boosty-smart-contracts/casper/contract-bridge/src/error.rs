use casper_types::ApiError;
use contract_util::error::ContractError;
use num_enum::{FromPrimitive, IntoPrimitive};

#[cfg_attr(std, derive(Debug, thiserror::Error))]
#[repr(u16)]
#[derive(IntoPrimitive, FromPrimitive)]
pub enum BridgeError {
    #[cfg_attr(std, error("this method is only callable by this contract"))]
    OnlyCallableBySelf = 0,

    #[cfg_attr(std, error("transferred amount did not match expected amount"))]
    UnexpectedTransferAmount = 1,

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
