use primitive_types::U256;
use std::{
    fmt::{Debug, Display},
    iter::{once, repeat_with},
};

use crate::error::RegistryError;

pub const CASPER_TAG_ACCOUNT: u8 = 0;
pub const CASPER_TAG_HASH: u8 = 1;

const CASPER_ADDRESS_LENGTH: usize = 33;
const SOLANA_ADDRESS_LENGTH: usize = 32;
const EVM_ADDRESS_LENGTH: usize = 20;

const CASPER_ACCOUNT_PREFIX: &str = "account-hash-";
const CASPER_HASH_PREFIX: &str = "hash-";

const TX_HASH_LENGTH: usize = 32;
const SOLANA_TX_HASH_LENGTH: usize = 64;

/// Network address. NetworkId is used to distinguish between networks.
#[derive(PartialEq, Eq, Hash, Clone)]
pub struct Address {
    network_id: NetworkId,
    data: Vec<u8>,
}

impl Display for Address {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let hex = base16::encode_lower(&self.data);
        write!(f, "{}:{hex}", self.network_id)
    }
}

impl Debug for Address {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let hex = base16::encode_lower(&self.data);
        write!(f, "Address({}:{hex})", self.network_id.value())
    }
}

#[derive(PartialEq, Eq, Hash, Clone, Debug)]
pub struct StringAddress {
    network_name: String,
    address: String,
}

impl Display for StringAddress {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}:{}", self.network_name, self.address)
    }
}

/// Transaction hash. NetworkId is used to distinguish between networks.
#[derive(PartialEq, Eq, Hash, Clone)]
pub struct TxHash {
    network_id: NetworkId,
    data: Vec<u8>,
}

impl Display for TxHash {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let hex = base16::encode_lower(&self.data);
        write!(f, "{}:{hex}", self.network_id)
    }
}

impl Debug for TxHash {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let hex = base16::encode_lower(&self.data);
        write!(f, "TxHash({}:{hex})", self.network_id.value())
    }
}

#[derive(PartialEq, Eq, Hash, Clone, Debug)]
pub struct StringTxHash {
    network_name: String,
    hash: String,
}

impl Display for StringTxHash {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let hex = base16::encode_lower(&self.hash);
        write!(f, "{}:{hex}", self.network_name)
    }
}

impl Address {
    pub fn random(network_id: NetworkId, ty: NetworkType) -> Self {
        let data = match ty {
            NetworkType::Casper => once(CASPER_TAG_ACCOUNT)
                .chain(repeat_with(|| fastrand::u8(..)))
                .take(CASPER_ADDRESS_LENGTH)
                .collect(),
            NetworkType::Evm => repeat_with(|| fastrand::u8(..))
                .take(EVM_ADDRESS_LENGTH)
                .collect(),
            NetworkType::Solana => repeat_with(|| fastrand::u8(..))
                .take(SOLANA_ADDRESS_LENGTH)
                .collect(),
        };

        Self::new(network_id, data)
    }

    pub fn new(network_id: NetworkId, data: Vec<u8>) -> Self {
        Self { network_id, data }
    }

    pub fn network_id(&self) -> NetworkId {
        self.network_id
    }

    pub fn data(&self) -> &[u8] {
        &self.data
    }
}

impl StringAddress {
    pub fn new(network_name: String, address: String) -> Self {
        Self {
            network_name,
            address,
        }
    }

    pub fn network_name(&self) -> &str {
        &self.network_name
    }

    pub fn address(&self) -> &str {
        &self.address
    }
}

impl TxHash {
    pub fn new(network_id: NetworkId, data: Vec<u8>) -> Self {
        Self { network_id, data }
    }

    pub fn network_id(&self) -> NetworkId {
        self.network_id
    }

    pub fn data(&self) -> &[u8] {
        &self.data
    }
}

impl StringTxHash {
    pub fn new(network_name: String, hash: String) -> Self {
        Self { network_name, hash }
    }

    pub fn network_name(&self) -> &str {
        &self.network_name
    }

    pub fn hash(&self) -> &str {
        &self.hash
    }
}

#[derive(PartialEq, Eq, Hash, Clone, Copy, Debug)]
pub struct NetworkId(u32);

impl NetworkId {
    pub const fn new(value: u32) -> Self {
        Self(value)
    }

    pub fn value(&self) -> u32 {
        self.0
    }
}

impl Display for NetworkId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

/// Confirmed transaction. Contains transaction hash, sender address and block number where it was confirmed.
#[derive(PartialEq, Eq, Hash, Clone, Debug)]
pub struct ConfirmedTx {
    hash: TxHash,
    sender: Address,
    block_number: u64,
}

impl ConfirmedTx {
    pub fn new(hash: TxHash, sender: Address, block_number: u64) -> Self {
        Self {
            hash,
            sender,
            block_number,
        }
    }

    pub fn sender(&self) -> &Address {
        &self.sender
    }

    pub fn hash(&self) -> &TxHash {
        &self.hash
    }

    pub fn block_number(&self) -> u64 {
        self.block_number
    }
}

#[derive(PartialEq, Eq, Hash, Clone, Copy, Debug)]
pub struct TokenId(u32);

impl TokenId {
    pub fn new(value: u32) -> Self {
        Self(value)
    }

    pub fn value(&self) -> u32 {
        self.0
    }
}

impl Display for TokenId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

#[derive(PartialEq, Eq, Hash, Clone, Copy, Debug)]
pub enum NetworkType {
    Casper,
    Evm,
    Solana,
}

/// Bridge event that is emitted when a token transfer is initiated by the user.
#[derive(Debug, Clone)]
pub struct BridgeTokenTransferIn {
    pub from: Address,
    pub to: StringAddress,
    pub amount: U256,
    pub token: Address,
    pub tx: ConfirmedTx,
}

/// Bridge event that is emitted when a token transfer is finalized by bridge.
#[derive(Debug, Clone)]
pub struct BridgeTokenTransferOut {
    pub from: StringAddress,
    pub to: Address,
    pub amount: U256,
    pub token: Address,
    pub tx: ConfirmedTx,
}

#[derive(Debug, Clone)]
pub enum BridgeEvent {
    TokenTransferIn(BridgeTokenTransferIn),
    TokenTransferOut(BridgeTokenTransferOut),
}

/// Transfer status
/// Waiting - transfer is waiting for `delay` time to pass
/// Confirming - transfer is waiting for confirmation on destination chain
/// Cancelled - transfer is cancelled
/// Finished - transfer is finished and confirmed on destination chain
#[derive(Debug, Clone, Copy, PartialEq, Eq, strum::EnumString, strum::IntoStaticStr)]
#[strum(serialize_all = "SCREAMING_SNAKE_CASE")]
pub enum TransferStatus {
    Waiting,
    Confirming,
    Cancelled,
    Finished,
}

/// address stringification for casper network
pub fn address_to_casper_string(address: &[u8]) -> Result<String, RegistryError> {
    if address.len() != CASPER_ADDRESS_LENGTH {
        return Err(RegistryError::InvalidAddressLength {
            expected: CASPER_ADDRESS_LENGTH,
            actual: address.len(),
        });
    }

    let tag = address[0];
    let hash = &address[1..];

    let prefix = match tag {
        CASPER_TAG_ACCOUNT => CASPER_ACCOUNT_PREFIX,
        CASPER_TAG_HASH => CASPER_HASH_PREFIX,
        _ => {
            return Err(RegistryError::InvalidAddressFormat {
                reason: "invalid account tag".to_string(),
            })
        }
    };

    let hex = base16::encode_lower(&hash);
    Ok(format!("{prefix}{hex}"))
}

/// address stringification for evm network
pub fn address_to_evm_string(address: &[u8]) -> Result<String, RegistryError> {
    if address.len() != EVM_ADDRESS_LENGTH {
        return Err(RegistryError::InvalidAddressLength {
            expected: EVM_ADDRESS_LENGTH,
            actual: address.len(),
        });
    }

    Ok(base16::encode_lower(address))
}

/// address stringification for solana network
pub fn address_to_solana_string(address: &[u8]) -> Result<String, RegistryError> {
    if address.len() != SOLANA_ADDRESS_LENGTH {
        return Err(RegistryError::InvalidAddressLength {
            expected: CASPER_ADDRESS_LENGTH,
            actual: address.len(),
        });
    }

    Ok(bs58::encode(&address).into_string())
}

/// parsing address from string for casper network
pub fn address_from_casper_string(address: &str) -> Result<Vec<u8>, RegistryError> {
    let (hash, tag) = if let Some(hash) = address.strip_prefix(CASPER_ACCOUNT_PREFIX) {
        (hash, CASPER_TAG_ACCOUNT)
    } else if let Some(hash) = address.strip_prefix(CASPER_HASH_PREFIX) {
        (hash, CASPER_TAG_HASH)
    } else {
        return Err(RegistryError::InvalidAddressFormat {
            reason: "unknown address prefix".to_string(),
        });
    };

    let mut data = base16::decode(hash).map_err(|err| RegistryError::InvalidAddressFormat {
        reason: format!("invalid hex format: {err}"),
    })?;

    if data.len() != (CASPER_ADDRESS_LENGTH - 1) {
        return Err(RegistryError::InvalidAddressLength {
            expected: CASPER_ADDRESS_LENGTH,
            actual: data.len() + 1,
        });
    }

    data.insert(0, tag);

    Ok(data)
}

/// parsing address from string for evm network
pub fn address_from_evm_string(mut address: &str) -> Result<Vec<u8>, RegistryError> {
    if address.starts_with("0x") {
        address = &address[2..];
    }

    let data = base16::decode(address).map_err(|err| RegistryError::InvalidAddressFormat {
        reason: format!("invalid hex format: {err}"),
    })?;

    if data.len() != EVM_ADDRESS_LENGTH {
        return Err(RegistryError::InvalidAddressLength {
            expected: EVM_ADDRESS_LENGTH,
            actual: data.len(),
        });
    }

    Ok(data)
}

/// address parse for solana network
pub fn address_from_solana_string(address: &str) -> Result<Vec<u8>, RegistryError> {
    let data =
        bs58::decode(address)
            .into_vec()
            .map_err(|err| RegistryError::InvalidAddressFormat {
                reason: format!("invalid bs58 format: {err}"),
            })?;

    if data.len() != SOLANA_ADDRESS_LENGTH {
        return Err(RegistryError::InvalidAddressLength {
            expected: SOLANA_ADDRESS_LENGTH,
            actual: data.len(),
        });
    }

    Ok(data)
}

/// tx hash stringification
pub fn txhash_to_string(data: &[u8]) -> Result<String, RegistryError> {
    if data.len() != TX_HASH_LENGTH {
        return Err(RegistryError::InvalidTxHashLength {
            expected: TX_HASH_LENGTH,
            actual: data.len(),
        });
    }

    Ok(base16::encode_lower(data))
}

/// solana hash stringification
pub fn solana_txhash_to_string(data: &[u8]) -> Result<String, RegistryError> {
    if data.len() != SOLANA_TX_HASH_LENGTH {
        return Err(RegistryError::InvalidTxHashLength {
            expected: TX_HASH_LENGTH,
            actual: data.len(),
        });
    }

    Ok(bs58::encode(data).into_string())
}

/// parsing tx hash from string
pub fn txhash_from_string(mut hash: &str) -> Result<Vec<u8>, RegistryError> {
    if hash.starts_with("0x") {
        hash = &hash[2..];
    }

    let data = base16::decode(hash).map_err(|err| RegistryError::InvalidTxHashFormat {
        reason: format!("invalid hex format: {err}"),
    })?;

    if data.len() != TX_HASH_LENGTH {
        return Err(RegistryError::InvalidTxHashLength {
            expected: TX_HASH_LENGTH,
            actual: data.len(),
        });
    }

    Ok(data)
}

/// parsing solana tx hash from string
pub fn solana_txhash_from_string(hash: &str) -> Result<Vec<u8>, RegistryError> {
    let data = bs58::decode(hash)
        .into_vec()
        .map_err(|err| RegistryError::InvalidTxHashFormat {
            reason: format!("invalid hex format: {err}"),
        })?;

    if data.len() != SOLANA_TX_HASH_LENGTH {
        return Err(RegistryError::InvalidTxHashLength {
            expected: SOLANA_TX_HASH_LENGTH,
            actual: data.len(),
        });
    }

    Ok(data)
}

#[cfg(test)]
mod tests {
    use crate::types::{
        address_from_casper_string, address_from_evm_string, address_from_solana_string,
        address_to_casper_string, address_to_evm_string, address_to_solana_string,
    };

    #[test]
    fn test_decode_encode_casper() {
        let address =
            "account-hash-9060c0820b5156b1620c8e3344d17f9fad5108f5dc2672f2308439e84363c88e";
        let data = address_from_casper_string(address).unwrap();
        let address2 = address_to_casper_string(&data).unwrap();
        assert_eq!(address, address2);
    }

    #[test]
    fn test_decode_encode_eth() {
        let address = "3095F955Da700b96215CFfC9Bc64AB2e69eB7DAB".to_lowercase();
        let data = address_from_evm_string(&address).unwrap();
        let address2 = address_to_evm_string(&data).unwrap();
        assert_eq!(address, address2);
    }

    #[test]
    fn test_decode_encode_solana() {
        let address = "8HR5rCobbFMDe5EbgKdJLNDWVCeGG79w837BUxtsCngs";
        let data = address_from_solana_string(&address).unwrap();
        let address2 = address_to_solana_string(&data).unwrap();
        assert_eq!(address, address2);
    }

    #[test]
    fn tx_decode_default() {
        let hash = "df162c5198eb67014f14e1cf4be8d9b785940cf4fca7ecc592a20e142b928f5f";
        let data = super::txhash_from_string(hash).unwrap();
        let hash2 = super::txhash_to_string(&data).unwrap();
        assert_eq!(hash, hash2);
    }

    #[test]
    fn tx_decode_solana() {
        let hash = "5Q6YzXWReDpmLc2bHSKD11tUqZQD5XZj4Za4xwmstd1unrS7fhJEFwBUyzb5Ph9MyZQRgwiPbGULiKfps9GjR1QF";
        let data = super::solana_txhash_from_string(hash).unwrap();
        let hash2 = super::solana_txhash_to_string(&data).unwrap();
        assert_eq!(hash, hash2);
    }
}
