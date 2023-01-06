use std::collections::HashMap;

use slab::Slab;

use crate::{
    error::RegistryError,
    types::{
        address_from_casper_string, address_from_evm_string, address_from_solana_string,
        address_to_casper_string, address_to_evm_string, address_to_solana_string,
        solana_txhash_from_string, solana_txhash_to_string, txhash_from_string, txhash_to_string,
        Address, NetworkId, NetworkType, StringAddress, StringTxHash, TokenId, TxHash,
    },
};

/// Network metadata.
#[derive(Clone, Debug)]
pub struct NetworkMetadata {
    ty: NetworkType,
    id: NetworkId,
    name: String,
    node: String,
    is_testnet: bool,
}

impl NetworkMetadata {
    pub fn new(
        ty: NetworkType,
        id: NetworkId,
        name: String,
        node: String,
        is_testnet: bool,
    ) -> Self {
        Self {
            ty,
            id,
            name,
            node,
            is_testnet,
        }
    }

    pub fn ty(&self) -> NetworkType {
        self.ty
    }

    pub fn id(&self) -> NetworkId {
        self.id
    }

    pub fn name(&self) -> &str {
        &self.name
    }

    pub fn node(&self) -> &str {
        &self.node
    }

    pub fn is_testnet(&self) -> bool {
        self.is_testnet
    }
}

/// Token metadata.
#[derive(Clone, Debug)]

pub struct TokenMetadata {
    id: TokenId,
    short_name: String,
    long_name: String,
}

impl TokenMetadata {
    pub fn new(id: TokenId, short_name: String, long_name: String) -> Self {
        Self {
            id,
            short_name,
            long_name,
        }
    }

    pub fn id(&self) -> TokenId {
        self.id
    }

    pub fn short_name(&self) -> &str {
        &self.short_name
    }

    pub fn long_name(&self) -> &str {
        &self.long_name
    }
}

/// Network specific token metadata.
#[derive(Clone, Debug)]
pub struct TokenNetworkMetadata {
    contract: Address,
    decimals: u8,
}

impl TokenNetworkMetadata {
    pub fn new(contract: Address, decimals: u8) -> Self {
        Self { contract, decimals }
    }

    pub fn contract(&self) -> &Address {
        &self.contract
    }

    pub fn decimals(&self) -> u8 {
        self.decimals
    }
}

/// A registry of networks.
#[derive(Default)]
pub struct NetworkRegistry {
    networks: Slab<NetworkMetadata>,
    by_id: HashMap<NetworkId, usize>,
    by_name: HashMap<String, usize>,
}

impl NetworkRegistry {
    // Registers a new network.
    pub fn register(&mut self, metadata: NetworkMetadata) {
        let network_id = metadata.id;
        let network_name = metadata.name.clone();
        let slab_id = self.networks.insert(metadata);

        self.by_id.insert(network_id, slab_id);
        self.by_name.insert(network_name, slab_id);
    }

    /// Returns the metadata for a network by its ID.
    pub fn by_id(&self, id: NetworkId) -> Result<&NetworkMetadata, RegistryError> {
        self.by_id
            .get(&id)
            .and_then(|slab_id| self.networks.get(*slab_id))
            .ok_or(RegistryError::UnknownNetworkId(id))
    }

    /// Returns the metadata for a network by its name.
    pub fn by_name(&self, name: &str) -> Result<&NetworkMetadata, RegistryError> {
        self.by_name
            .get(name)
            .and_then(|slab_id| self.networks.get(*slab_id))
            .ok_or_else(|| RegistryError::UnknownNetworkName(name.into()))
    }

    /// Returns an iterator over all registered networks.
    pub fn all(&self) -> impl Iterator<Item = &NetworkMetadata> {
        self.networks.iter().map(|(_, item)| item)
    }

    /// network specific stringification of the address to string representation
    pub fn stringify_address(&self, address: &Address) -> Result<StringAddress, RegistryError> {
        let metadata = self.by_id(address.network_id())?;

        let ty = metadata.ty;

        let address = match ty {
            NetworkType::Casper => address_to_casper_string(address.data())?,
            NetworkType::Evm => address_to_evm_string(address.data())?,
            NetworkType::Solana => address_to_solana_string(address.data())?,
        };

        Ok(StringAddress::new(metadata.name.clone(), address))
    }

    /// network specific parsing of the address from string representation
    pub fn parse_address(&self, address: &StringAddress) -> Result<Address, RegistryError> {
        let metadata = self.by_name(address.network_name())?;

        let ty = metadata.ty;

        let address = match ty {
            NetworkType::Casper => address_from_casper_string(address.address())?,
            NetworkType::Evm => address_from_evm_string(address.address())?,
            NetworkType::Solana => address_from_solana_string(address.address())?,
        };

        Ok(Address::new(metadata.id, address))
    }

    /// network specific parsing of the transaction hash from string representation
    pub fn parse_tx_hash(&self, hash: &StringTxHash) -> Result<TxHash, RegistryError> {
        let network_name = hash.network_name();
        let hash = hash.hash();
        let metadata = self.by_name(network_name)?;

        let ty = metadata.ty;

        let data = match ty {
            NetworkType::Casper | NetworkType::Evm => txhash_from_string(hash)?,
            NetworkType::Solana => solana_txhash_from_string(hash)?,
        };

        Ok(TxHash::new(metadata.id, data))
    }

    /// network specific stringification of the transaction hash to string representation
    pub fn stringify_tx_hash(&self, hash: &TxHash) -> Result<StringTxHash, RegistryError> {
        let metadata = self.by_id(hash.network_id())?;
        let ty = metadata.ty;

        let hash = match ty {
            NetworkType::Casper | NetworkType::Evm => txhash_to_string(hash.data())?,
            NetworkType::Solana => solana_txhash_to_string(hash.data())?,
        };

        Ok(StringTxHash::new(metadata.name.clone(), hash))
    }
}

/// Token registry containing all known tokens and their metadata by bridge.
/// This registry is used to map between token IDs and addresses.
#[derive(Default)]
pub struct TokenRegistry {
    tokens: Slab<TokenMetadata>,
    token_networks: HashMap<(NetworkId, TokenId), TokenNetworkMetadata>,
    by_id: HashMap<TokenId, usize>,
    by_address: HashMap<Address, usize>,
}

impl TokenRegistry {
    /// Registers a new token.
    pub fn register(&mut self, metadata: TokenMetadata) {
        let token_id = metadata.id;
        let slab_id = self.tokens.insert(metadata);
        self.by_id.insert(token_id, slab_id);
    }

    /// Register a new network for a previously registered token.
    pub fn register_token_network(
        &mut self,
        token_id: TokenId,
        metadata: TokenNetworkMetadata,
    ) -> Result<(), RegistryError> {
        let slab_id = self
            .by_id
            .get(&token_id)
            .ok_or(RegistryError::UnknownTokenId(token_id))?;

        let address = metadata.contract.clone();

        self.token_networks
            .insert((address.network_id(), token_id), metadata);
        self.by_address.insert(address, *slab_id);
        Ok(())
    }

    /// Returns the network specific metadata for a token by its ID.
    pub fn token_network_by_ids(
        &self,
        token_id: TokenId,
        network_id: NetworkId,
    ) -> Result<&TokenNetworkMetadata, RegistryError> {
        self.token_networks
            .get(&(network_id, token_id))
            .ok_or(RegistryError::UnknownNetworkOrToken(token_id, network_id))
    }

    /// Returns token metadata by its ID.
    pub fn token_by_id(&self, id: TokenId) -> Result<&TokenMetadata, RegistryError> {
        self.by_id
            .get(&id)
            .and_then(|slab_id| self.tokens.get(*slab_id))
            .ok_or(RegistryError::UnknownTokenId(id))
    }

    /// Returns token metadata by its address.
    pub fn token_by_address(&self, address: &Address) -> Result<&TokenMetadata, RegistryError> {
        self.by_address
            .get(address)
            .and_then(|slab_id| self.tokens.get(*slab_id))
            .ok_or_else(|| RegistryError::UnknownTokenAddress(address.clone()))
    }

    /// Returns an iterator over all registered tokens.
    pub fn all_tokens(&self) -> impl Iterator<Item = &TokenMetadata> {
        self.tokens.iter().map(|(_, item)| item)
    }

    /// Returns an iterator over all registered tokens and their networks.
    pub fn all_token_networks(
        &self,
    ) -> impl Iterator<Item = (&TokenId, &NetworkId, &TokenNetworkMetadata)> {
        self.token_networks
            .iter()
            .map(|((nid, tid), meta)| (tid, nid, meta))
    }
}
