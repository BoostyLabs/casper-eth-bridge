use std::{collections::HashMap, fmt::Debug, hash::Hash};

use serde::{Deserialize, Serialize};

#[derive(PartialEq, Eq, Hash, Clone, Copy, Debug, Deserialize, Serialize)]
pub struct TokenId(u32);

impl TokenId {
    pub fn new(id: u32) -> Self {
        TokenId(id)
    }

    pub fn id(&self) -> u32 {
        self.0
    }
}

#[derive(PartialEq, Eq, Hash, Clone, Copy, Debug, Deserialize, Serialize)]
pub struct TokenContractAddress<T> {
    pub id: TokenId,
    pub address: T,
}

#[derive(Debug, Clone)]
pub struct TokenContractRegistry<T>
where
    T: Hash + Eq + Copy + Clone + Debug,
{
    by_id: HashMap<TokenId, T>,
    by_contract: HashMap<T, TokenId>,
}

impl<T> TokenContractRegistry<T>
where
    T: Hash + Eq + Copy + Clone + Debug,
{
    pub fn new(tokens: &[TokenContractAddress<T>]) -> Self {
        let by_id = tokens
            .iter()
            .map(|token_info| (token_info.id, token_info.address))
            .collect();
        let by_contract = tokens
            .iter()
            .map(|token_info| (token_info.address, token_info.id))
            .collect();

        Self { by_id, by_contract }
    }

    pub fn contract_by_id(&self, id: TokenId) -> Option<T> {
        self.by_id.get(&id).copied()
    }

    pub fn id_by_contract(&self, hash: &T) -> Option<TokenId> {
        self.by_contract.get(hash).copied()
    }

    pub fn tokens(&self) -> impl Iterator<Item = (&TokenId, &T)> {
        self.by_id.iter()
    }
}
