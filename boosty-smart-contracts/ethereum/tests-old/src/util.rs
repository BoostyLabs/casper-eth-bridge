use std::sync::Arc;

use ethers::{
    contract::Contract,
    prelude::*,
    utils::{Anvil, AnvilInstance},
};

use crate::abi::{BridgeContract, BridgeContractEvents, ERC20Contract};

pub type TestProvider = SignerMiddleware<Provider<Http>, LocalWallet>;
pub type TestContract = Contract<TestProvider>;

pub struct Contracts {
    bridge_contract: BridgeContract<TestProvider>,
    token_contract: ERC20Contract<TestProvider>,
}

impl Contracts {
    pub fn bridge_contract(&self) -> &BridgeContract<TestProvider> {
        &self.bridge_contract
    }

    pub fn token_contract(&self) -> &ERC20Contract<TestProvider> {
        &self.token_contract
    }

    pub fn token_contract_by(&self, provider: Arc<TestProvider>) -> ERC20Contract<TestProvider> {
        self.token_contract.connect(provider).into()
    }

    pub fn bridge_contract_by(&self, provider: Arc<TestProvider>) -> BridgeContract<TestProvider> {
        self.bridge_contract.connect(provider).into()
    }
}

pub struct AnvilContext {
    pub instance: AnvilInstance,

    pub wallet: LocalWallet,
    pub provider: Arc<TestProvider>,
}

impl Default for AnvilContext {
    fn default() -> Self {
        let anvil = Anvil::new();
        let instance = anvil.spawn();
        let main_key = instance.keys()[0].clone();
        let wallet = LocalWallet::from(main_key);

        let provider =
            Provider::<Http>::try_from(instance.endpoint()).expect("failed to get client");
        let provider = SignerMiddleware::new(provider, wallet.clone());
        let provider = Arc::new(provider);

        Self {
            instance,
            wallet,
            provider,
        }
    }
}

pub async fn guest_provider(context: &AnvilContext) -> Arc<TestProvider> {
    let main_key = context.instance.keys()[1].clone();
    let wallet = LocalWallet::from(main_key);

    let provider =
        Provider::<Http>::try_from(context.instance.endpoint()).expect("failed to get client");
    let provider = SignerMiddleware::new(provider, wallet.clone());
    let provider = Arc::new(provider);
    provider
}

pub async fn deploy_contracts(provider: &Arc<TestProvider>) -> Contracts {
    let compiled = Solc::default()
        .compile_source("./contract")
        .expect("compilation failed");

    let bridge_contract = compiled
        .get("./contract/Bridge.sol", "Bridge")
        .expect("couldn't find Bridge contract");

    let bridge_factory = ContractFactory::new(
        bridge_contract.abi.unwrap().clone(),
        bridge_contract.bytecode().unwrap().clone(),
        provider.clone(),
    );

    let bridge_contract = bridge_factory
        .deploy(())
        .expect("couldn't prepare deploy")
        .confirmations(0usize)
        .send()
        .await
        .expect("couldn't deploy contract");

    let token_contract = compiled
        .get("./contract/TestERC20.sol", "TestToken")
        .expect("couldn't find TestToken contract");

    let token_factory = ContractFactory::new(
        token_contract.abi.unwrap().clone(),
        token_contract.bytecode().unwrap().clone(),
        provider.clone(),
    );

    let token_contract = token_factory
        .deploy(U256::from(1_000_000_000_000u64))
        .expect("couldn't prepare deploy")
        .confirmations(0usize)
        .send()
        .await
        .expect("couldn't deploy contract");

    Contracts {
        bridge_contract: bridge_contract.into(),
        token_contract: token_contract.into(),
    }
}

pub async fn get_bridge_event(contracts: &Contracts) -> BridgeContractEvents {
    let events = contracts.bridge_contract.events();
    let mut events = events.query().await.unwrap();
    assert_eq!(events.len(), 1);
    events.pop().unwrap()
}

pub async fn fill_purse(contracts: &Contracts, amount: U256, recipient: H160) {
    let balance_before = contracts
        .token_contract
        .balance_of(recipient)
        .call()
        .await
        .expect("balance_of");
    contracts
        .token_contract
        .transfer(recipient, amount)
        .send()
        .await
        .unwrap();

    let balance_after = contracts
        .token_contract
        .balance_of(recipient)
        .call()
        .await
        .expect("balance_of");
    assert_eq!(balance_after, balance_before + amount);
}
