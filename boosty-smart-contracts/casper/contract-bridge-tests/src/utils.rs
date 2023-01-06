use casper_engine_test_support::{
    DeployItemBuilder, ExecuteRequestBuilder, InMemoryWasmTestBuilder, WasmTestBuilder, ARG_AMOUNT,
    DEFAULT_ACCOUNT_INITIAL_BALANCE, DEFAULT_GENESIS_CONFIG, DEFAULT_GENESIS_CONFIG_HASH,
    DEFAULT_PAYMENT,
};

use casper_execution_engine::{
    core::engine_state::{
        run_genesis_request::RunGenesisRequest, Error as EngineError, ExecutionResult,
        GenesisAccount,
    },
    shared::transform::Transform,
    storage::global_state::{CommitProvider, StateProvider},
};
use casper_types::{
    account::AccountHash,
    bytesrepr::{Bytes, ToBytes},
    runtime_args, ContractHash, ContractPackageHash, Key, Motes, PublicKey, RuntimeArgs, SecretKey,
    StoredValue, U256, U512,
};

const CONTRACT_ERC20_BYTES: &[u8] = include_bytes!("contract_erc20.wasm");
const CONTRACT_BRIDGE_BYTES: &[u8] = include_bytes!("contract_bridge.wasm");

pub const TEST_ACCOUNT: [u8; 32] = [255u8; 32];

pub struct TestContext {
    pub account: UserAccount,
    pub builder: InMemoryWasmTestBuilder,
}

pub const ACCOUNT_BALANCE: u64 = 10_000_000_000_000u64;

pub struct UserAccount {
    pub secret_key: SecretKey,
    pub public_key: PublicKey,
    pub address: AccountHash,
}

impl UserAccount {
    fn new(secret_key: SecretKey) -> Self {
        let public_key = PublicKey::from(&secret_key);
        let address = AccountHash::from(&public_key);
        Self {
            secret_key,
            public_key,
            address,
        }
    }

    pub fn unique_account(context: &mut TestContext, unique_id: u8) -> Self {
        if unique_id == 255 {
            panic!("Account with id 255 booked for genesis account");
        }
        // Create a key using unique_id
        let secret_key = SecretKey::ed25519_from_bytes([unique_id; 32]).unwrap();
        let account = UserAccount::new(secret_key);

        // We need to transfer some funds to the account so it become active
        let deploy = simple_deploy_builder(context.account.address)
            .with_transfer_args(runtime_args![
                ARG_AMOUNT => U512::one() * ACCOUNT_BALANCE,
                "target" => account.public_key.clone(),
                "id" => Some(u64::from(unique_id))
            ])
            .build();
        context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy).build())
            .commit()
            .expect_success();
        account
    }

    pub fn key(&self) -> Key {
        Key::from(self.address)
    }
}

use casper_execution_engine::core::execution::Error as ExecError;
use contract_bridge::entry_points::{PARAM_AMOUNT, PARAM_RECIPIENT};
use contract_util::event::ContractEvent;
pub fn deploy_contract<S>(
    builder: &mut WasmTestBuilder<S>,
    account: AccountHash,
    wasm_bytes: &[u8],
    deploy_args: RuntimeArgs,
    contract_key: &str,
) -> (ContractHash, ContractPackageHash)
where
    S: StateProvider + CommitProvider,
    EngineError: From<S::Error>,
    <S as StateProvider>::Error: Into<ExecError>,
{
    let deploy_item = DeployItemBuilder::new()
        .with_empty_payment_bytes(runtime_args! {
            ARG_AMOUNT => *DEFAULT_PAYMENT
        })
        .with_session_bytes(wasm_bytes.into(), deploy_args)
        .with_authorization_keys(&[account])
        .with_address(account)
        .build();

    let execute_request = ExecuteRequestBuilder::from_deploy_item(deploy_item).build();
    builder.exec(execute_request).commit();

    let stored_account = builder.query(None, Key::Account(account), &[]).unwrap();

    let contract_hash = stored_account
        .as_account()
        .unwrap()
        .named_keys()
        .get(contract_key)
        .unwrap()
        .into_hash()
        .unwrap();

    let contract_package_hash = builder
        .query(None, Key::Hash(contract_hash), &[])
        .unwrap()
        .as_contract()
        .unwrap()
        .contract_package_hash();

    (ContractHash::new(contract_hash), contract_package_hash)
}

pub fn deploy_erc20<S>(
    builder: &mut WasmTestBuilder<S>,
    account: AccountHash,
) -> (ContractHash, ContractPackageHash)
where
    S: StateProvider + CommitProvider,
    EngineError: From<S::Error>,
    <S as StateProvider>::Error: Into<ExecError>,
{
    let deploy_args = runtime_args! {
        "name" => "test token".to_string(),
        "symbol" => "TTKN",
        "decimals" => 9u8,
        "total_supply" => U256::max_value(),
    };

    deploy_contract(
        builder,
        account,
        CONTRACT_ERC20_BYTES,
        deploy_args,
        "erc20_token_contract",
    )
}

pub fn deploy_bridge<S>(
    builder: &mut WasmTestBuilder<S>,
    account: AccountHash,
) -> (ContractHash, ContractPackageHash)
where
    S: StateProvider + CommitProvider,
    EngineError: From<S::Error>,
    <S as StateProvider>::Error: Into<ExecError>,
{
    let deploy_args = runtime_args! {};

    deploy_contract(
        builder,
        account,
        CONTRACT_BRIDGE_BYTES,
        deploy_args,
        "bridge_contract",
    )
}

pub fn setup_context() -> TestContext {
    // Create keypair.
    let secret_key = SecretKey::ed25519_from_bytes(TEST_ACCOUNT).unwrap();
    let account_data = UserAccount::new(secret_key);

    // Create a GenesisAccount.
    let account = GenesisAccount::account(
        account_data.public_key.clone(),
        Motes::new(U512::from(DEFAULT_ACCOUNT_INITIAL_BALANCE)),
        None,
    );

    let mut genesis_config = DEFAULT_GENESIS_CONFIG.clone();
    genesis_config.ee_config_mut().push_account(account);

    let run_genesis_request = RunGenesisRequest::new(
        *DEFAULT_GENESIS_CONFIG_HASH,
        genesis_config.protocol_version(),
        genesis_config.take_ee_config(),
    );

    let mut builder = InMemoryWasmTestBuilder::default();
    builder.run_genesis(&run_genesis_request).commit();

    TestContext {
        account: account_data,
        builder,
    }
}

pub fn simple_deploy_builder(account: AccountHash) -> DeployItemBuilder {
    DeployItemBuilder::new()
        .with_empty_payment_bytes(runtime_args! {
            ARG_AMOUNT => *DEFAULT_PAYMENT
        })
        .with_authorization_keys(&[account])
        .with_address(account)
}

pub fn erc20_dictionary_key(owner: &Key) -> String {
    base64::encode(owner.to_bytes().expect("infallible"))
}

pub fn query_balance<S>(
    builder: &mut WasmTestBuilder<S>,
    contract: ContractHash,
    address: &Key,
) -> U256
where
    S: StateProvider + CommitProvider,
    EngineError: From<S::Error>,
    <S as StateProvider>::Error: Into<ExecError>,
{
    let contract = builder
        .query(None, Key::Hash(contract.value()), &[])
        .unwrap()
        .as_contract()
        .cloned()
        .unwrap();

    let balance_uref = contract
        .named_keys()
        .get("balances")
        .unwrap()
        .as_uref()
        .cloned()
        .unwrap();

    let balance = builder
        .query_dictionary_item(None, balance_uref, &erc20_dictionary_key(address))
        .unwrap()
        .as_cl_value()
        .cloned()
        .unwrap()
        .into_t::<U256>()
        .unwrap();

    balance
}

pub fn read_contract_event<S, E>(
    builder: &mut WasmTestBuilder<S>,
    contract: ContractHash,
    event_uref_name: &str,
) -> E
where
    S: StateProvider + CommitProvider,
    EngineError: From<S::Error>,
    <S as StateProvider>::Error: Into<ExecError>,
    E: ContractEvent,
{
    let contract = builder
        .query(None, Key::Hash(contract.value()), &[])
        .unwrap()
        .as_contract()
        .cloned()
        .unwrap();

    let event_uref = contract
        .named_keys()
        .get(event_uref_name)
        .unwrap()
        .as_uref()
        .cloned()
        .unwrap();

    let last_result = builder.last_exec_result();
    let journal = match last_result {
        ExecutionResult::Failure {
            execution_journal, ..
        } => execution_journal,
        ExecutionResult::Success {
            execution_journal, ..
        } => execution_journal,
    };

    let mut event: Vec<E> = journal
        .clone()
        .into_iter()
        .filter_map(|item| match item {
            (Key::URef(uref), Transform::Write(StoredValue::CLValue(value)))
                if uref.addr() == event_uref.addr() =>
            {
                let data: Bytes = value.into_t().unwrap();
                let (event, _) = E::from_bytes(&data).unwrap();
                Some(event)
            }
            _ => None,
        })
        .collect();
    assert_eq!(event.len(), 1);
    event.pop().unwrap()
}

pub fn fill_purse_on_token_contract(
    context: &mut TestContext,
    token_hash: ContractHash,
    amount: U256,
    recipient: Key,
) {
    // Transferings token on bridge token purse
    let deploy_item = simple_deploy_builder(context.account.address)
        .with_stored_session_hash(
            token_hash,
            "transfer",
            runtime_args! {
                PARAM_RECIPIENT => recipient,
                PARAM_AMOUNT => amount,
            },
        )
        .build();

    context
        .builder
        .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
        .commit()
        .expect_success();

    let balance = query_balance(&mut context.builder, token_hash, &recipient);

    assert_eq!(balance, amount);
}
