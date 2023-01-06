use casper_types::ApiError;
use num_enum::{FromPrimitive, IntoPrimitive};

#[cfg(std)]
pub trait ContractError: std::error::Error + Into<u16> + From<u16> {}

#[cfg(not(std))]
pub trait ContractError: Into<u16> + From<u16> {}

pub const ERROR_CONTRACT_START: u16 = 256;

#[cfg_attr(std, derive(Debug, thiserror::Error))]
#[derive(Clone, Copy, PartialEq, Eq, FromPrimitive, IntoPrimitive)]
#[repr(u16)]
pub enum UtilError {
    #[cfg_attr(std, error("current context is not a contract"))]
    CurrentContextNotContract = 0,

    #[cfg_attr(std, error("current context is not a session"))]
    CurrentContextNotSession = 1,

    #[cfg_attr(std, error("invalid stack depth specified"))]
    InvalidStackDepth = 2,

    #[cfg_attr(std, error("current context is not a contract"))]
    #[default]
    Unknown = 255,
}

#[cfg_attr(std, derive(Debug, thiserror::Error))]
#[derive(Clone, Copy, PartialEq, Eq)]
pub enum Error<T: ContractError> {
    #[cfg_attr(std, error("util error"))]
    Util(#[cfg_attr(std, source)] UtilError),

    #[cfg_attr(std, error("contract error"))]
    Contract(#[cfg_attr(std, source)] T),
}

impl<T: ContractError> Into<u16> for Error<T> {
    fn into(self) -> u16 {
        match self {
            Error::Util(error) => error.into(),
            Error::Contract(error) => error.into() + ERROR_CONTRACT_START,
        }
    }
}

impl<T: ContractError> From<u16> for Error<T> {
    fn from(code: u16) -> Self {
        if code >= ERROR_CONTRACT_START {
            Self::Contract(T::from(code - ERROR_CONTRACT_START))
        } else {
            Self::Util(UtilError::from(code))
        }
    }
}

impl Into<ApiError> for UtilError {
    fn into(self) -> ApiError {
        ApiError::User(self.into())
    }
}

impl From<ApiError> for UtilError {
    fn from(api: ApiError) -> Self {
        match api {
            ApiError::User(code) => Self::from(code),
            _ => Self::Unknown,
        }
    }
}

impl<T: ContractError> From<ApiError> for Error<T> {
    fn from(api: ApiError) -> Self {
        match api {
            ApiError::User(code) => Self::from(code),
            _ => Error::Util(UtilError::Unknown),
        }
    }
}

impl<T: ContractError> Into<ApiError> for Error<T> {
    fn into(self) -> ApiError {
        ApiError::User(self.into())
    }
}
