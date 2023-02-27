use std::{
    iter::repeat,
    sync::atomic::{AtomicUsize, Ordering},
};

use casper_engine_test_support::{
    DeployItemBuilder, ExecuteRequestBuilder, InMemoryWasmTestBuilder, WasmTestBuilder, ARG_AMOUNT,
    DEFAULT_ACCOUNT_INITIAL_BALANCE, DEFAULT_GENESIS_CONFIG, DEFAULT_GENESIS_CONFIG_HASH,
    DEFAULT_PAYMENT,
};
use contract_util::signatures::{cook_msg_bridge_in, cook_msg_transfer_out, get_signature_bytes};

use casper_execution_engine::{
    core::{
        engine_state::{
            run_genesis_request::RunGenesisRequest, DeployItem, Error as EngineError,
            ExecutionResult, GenesisAccount,
        },
        execution::Error as ExecError,
    },
    shared::transform::Transform,
    storage::global_state::{in_memory::InMemoryGlobalState, CommitProvider, StateProvider},
};

use casper_types::{
    account::AccountHash,
    bytesrepr::{Bytes, ToBytes},
    runtime_args, ContractHash, ContractPackageHash, Key, Motes, PublicKey, RuntimeArgs, SecretKey,
    StoredValue, U128, U256, U512,
};

use contract_bridge::entry_points::{
    EP_BRIDGE_IN, EP_BRIDGE_OUT, EP_SET_SIGNER, EP_TRANSFER_OUT, EP_WITHDRAW_COMMISSION,
    PARAM_AMOUNT, PARAM_COMMISSION, PARAM_DEADLINE, PARAM_DESTINATION_ADDRESS,
    PARAM_DESTINATION_CHAIN, PARAM_GAS_COMMISSION, PARAM_NONCE, PARAM_RECIPIENT, PARAM_SIGNATURE,
    PARAM_SIGNER, PARAM_SOURCE_ADDRESS, PARAM_SOURCE_CHAIN, PARAM_TOKEN_CONTRACT,
    PARAM_TRANSACTION_ID,
};
use contract_util::event::ContractEvent;

const CONTRACT_ERC20_BYTES: &[u8] = include_bytes!("contract_erc20.wasm");
const CONTRACT_BRIDGE_BYTES: &[u8] = include_bytes!("contract_bridge.wasm");

static DEPLOY_COUNTER: AtomicUsize = AtomicUsize::new(0);

use crate::constants::{
    TEST_ACCOUNT, TEST_ACCOUNT_BALANCE, TEST_AMOUNT, TEST_BLOCK_TIME, TEST_DESTINATION_ADDRESS,
    TEST_DESTINATION_CHAIN, TEST_GAS_COMMISSION, TEST_STABLE_COMMISSION_PERCENT,
};

pub fn test_public_key() -> &'static str {
    include_str!("config/public_key.in")
}

pub fn test_signer_secret_key() -> &'static str {
    include_str!("config/signer_secret_key.in")
}

pub fn new_deploy_hash() -> [u8; 32] {
    let counter = DEPLOY_COUNTER.fetch_add(1, Ordering::SeqCst);
    let hash = repeat(counter)
        .take(4)
        .flat_map(|i| i.to_le_bytes())
        .collect::<Vec<_>>();
    hash.try_into().unwrap()
}

pub fn deploy_builder() -> DeployItemBuilder {
    DeployItemBuilder::new().with_deploy_hash(new_deploy_hash())
}

pub struct TestContext {
    pub account: UserAccount,
    pub builder: InMemoryWasmTestBuilder,
}

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
                ARG_AMOUNT => U512::one() * TEST_ACCOUNT_BALANCE,
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
    let deploy_item = deploy_builder()
        .with_empty_payment_bytes(runtime_args! {
            ARG_AMOUNT => *DEFAULT_PAYMENT
        })
        .with_session_bytes(wasm_bytes.into(), deploy_args)
        .with_authorization_keys(&[account])
        .with_address(account)
        .build();

    let execute_request = ExecuteRequestBuilder::from_deploy_item(deploy_item).build();

    builder.exec(execute_request).commit().expect_success();

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
    let deploy_args = runtime_args! {
        PARAM_SIGNER => test_public_key(),
    };

    deploy_contract(
        builder,
        account,
        CONTRACT_BRIDGE_BYTES,
        deploy_args,
        "bridge_contract",
    )
}

pub fn deploy_bridge_and_erc20<S>(
    builder: &mut WasmTestBuilder<S>,
    account_address: AccountHash,
) -> (
    ContractHash,
    ContractPackageHash,
    ContractHash,
    ContractPackageHash,
)
where
    S: StateProvider + CommitProvider,
    EngineError: From<S::Error>,
    <S as StateProvider>::Error: Into<ExecError>,
{
    let (token_hash, token_package_hash) = deploy_erc20(builder, account_address);

    let (bridge_hash, bridge_package_hash) = deploy_bridge(builder, account_address);

    (
        token_hash,
        token_package_hash,
        bridge_hash,
        bridge_package_hash,
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
    deploy_builder()
        .with_empty_payment_bytes(runtime_args! {
            ARG_AMOUNT => *DEFAULT_PAYMENT
        })
        .with_authorization_keys(&[account])
        .with_address(account)
}

pub fn dictionary_key<T: ToBytes>(value: &T) -> String {
    base64::encode(value.to_bytes().expect("infallible"))
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
        .query_dictionary_item(None, balance_uref, &dictionary_key(address))
        .unwrap()
        .as_cl_value()
        .cloned()
        .unwrap()
        .into_t::<U256>()
        .unwrap();

    balance
}

pub fn query_commission_pool<S>(
    builder: &mut WasmTestBuilder<S>,
    contract: ContractHash,
    address: ContractPackageHash,
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

    let uref = contract
        .named_keys()
        .get("commission_by_token")
        .unwrap()
        .as_uref()
        .cloned()
        .unwrap();

    let value = builder
        .query_dictionary_item(None, uref, &dictionary_key(&address))
        .unwrap()
        .as_cl_value()
        .cloned()
        .unwrap()
        .into_t::<U256>()
        .unwrap();

    value
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

pub fn bridge_in(
    bridge_hash: ContractHash,
    token_package_hash: ContractPackageHash,
    account_address: AccountHash,
    amount: U256,
    deadline: U256,
    nonce: U128,
    gas_commission: U256,
    mut signature_bytes: Vec<u8>,
) -> DeployItem {

    if signature_bytes.is_empty() {
        signature_bytes = cook_msg_bridge_in(
            bridge_hash,
            token_package_hash,
            account_address,
            amount,
            gas_commission,
            deadline,
            nonce,
            &TEST_DESTINATION_CHAIN(),
            &TEST_DESTINATION_ADDRESS(),
        );
    }
    let signature_bytes = get_signature_bytes(&signature_bytes, test_signer_secret_key());

    // println!("bridge_hash {bridge_hash}");
    // println!("token_package_hash {token_package_hash}");
    // println!("account_address {account_address}");
    // println!("PARAM_TOKEN_CONTRACT {token_package_hash}"); // +
    // println!("PARAM_AMOUNT {}", U256::one() * 1_000_000_000_000u64); // +
    // println!("PARAM_GAS_COMMISSION {}", gas_commission); // +
    // println!("PARAM_DEADLINE {deadline}"); // +
    // println!("PARAM_NONCE {nonce}"); // +
    // println!("PARAM_DESTINATION_CHAIN {}", TEST_DESTINATION_CHAIN()); // +
    // println!("PARAM_DESTINATION_ADDRESS {}", TEST_DESTINATION_ADDRESS()); // +
    // println!("PARAM_SIGNATURE1111 {:?}", signature_bytes.clone());

    simple_deploy_builder(account_address)
        .with_stored_session_hash(
            bridge_hash,
            EP_BRIDGE_IN,
            runtime_args! {
                PARAM_TOKEN_CONTRACT => token_package_hash,
                PARAM_AMOUNT => amount,
                PARAM_GAS_COMMISSION => gas_commission,
                PARAM_DEADLINE => deadline,
                PARAM_NONCE => nonce,
                PARAM_DESTINATION_CHAIN => TEST_DESTINATION_CHAIN(),
                PARAM_DESTINATION_ADDRESS => TEST_DESTINATION_ADDRESS(),
                PARAM_SIGNATURE => signature_bytes,
            },
        )
        .build()
}

pub fn transfer_out(
    bridge_hash: ContractHash,
    token_package_hash: ContractPackageHash,
    account_address: AccountHash,
    recipient: Key,
    amount_to_transfer: U256,
    commission: U256,
    nonce: U128,
    mut signature_bytes: Vec<u8>,
) -> DeployItem {
    if signature_bytes.is_empty() {
        signature_bytes = cook_msg_transfer_out(
            bridge_hash,
            token_package_hash,
            recipient,
            amount_to_transfer,
            commission,
            nonce,
        );
    }
    let signature_bytes = get_signature_bytes(&signature_bytes, test_signer_secret_key());

    simple_deploy_builder(account_address)
        .with_stored_session_hash(
            bridge_hash,
            EP_TRANSFER_OUT,
            runtime_args! {
                PARAM_TOKEN_CONTRACT => token_package_hash,
                PARAM_AMOUNT => amount_to_transfer,
                PARAM_COMMISSION => commission,
                PARAM_NONCE => nonce,
                PARAM_RECIPIENT => recipient,
                PARAM_SIGNATURE => signature_bytes,
            },
        )
        .build()
}

pub fn bridge_out(
    bridge_hash: ContractHash,
    token_package_hash: ContractPackageHash,
    account_address: AccountHash,
    recipient_key: Key,
    amount: U256,
) -> DeployItem {
    simple_deploy_builder(account_address)
        .with_stored_session_hash(
            bridge_hash,
            EP_BRIDGE_OUT,
            runtime_args! {
                PARAM_TOKEN_CONTRACT => token_package_hash,
                PARAM_AMOUNT => amount,
                PARAM_TRANSACTION_ID => U256::one(),
                PARAM_SOURCE_CHAIN => "SOUR".to_string(),
                PARAM_SOURCE_ADDRESS => "SOURADDR".to_string(),
                PARAM_RECIPIENT => recipient_key
            },
        )
        .build()
}

pub fn withdraw_commission(
    bridge_hash: ContractHash,
    token_package_hash: ContractPackageHash,
    account_address: AccountHash,
    recipient_key: Key,
    amount: U256,
) -> DeployItem {
    simple_deploy_builder(account_address)
        .with_stored_session_hash(
            bridge_hash,
            EP_WITHDRAW_COMMISSION,
            runtime_args! {
               PARAM_TOKEN_CONTRACT => token_package_hash,
               PARAM_AMOUNT => amount,
               PARAM_RECIPIENT => recipient_key,
            },
        )
        .build()
}

pub fn set_test_signer(
    bridge_hash: ContractHash,
    account_address: AccountHash,
    test_signer_public_key: &str,
) -> DeployItem {
    simple_deploy_builder(account_address)
        .with_stored_session_hash(
            bridge_hash,
            EP_SET_SIGNER,
            runtime_args! {
                PARAM_SIGNER => test_signer_public_key,
            },
        )
        .build()
}

pub fn arbitrary_user(context: &mut TestContext) -> UserAccount {
    UserAccount::unique_account(context, 0)
}

pub fn arbitrary_user_key(context: &mut TestContext) -> Key {
    arbitrary_user(context).key()
}

pub fn execution_context(
    context: &mut TestContext,
    deploy_item: DeployItem,
) -> &mut WasmTestBuilder<InMemoryGlobalState> {
    context
        .builder
        .exec(
            ExecuteRequestBuilder::from_deploy_item(deploy_item)
                .with_block_time(TEST_BLOCK_TIME)
                .build(),
        )
        .commit()
}

pub fn execution_error(context: &mut TestContext, deploy_item: DeployItem) -> EngineError {
    execution_context(context, deploy_item)
        .expect_failure()
        .get_error()
        .unwrap()
}

pub fn get_context(
    context: &mut TestContext,
    deploy_item: DeployItem,
) -> &mut WasmTestBuilder<InMemoryGlobalState> {
    context
        .builder
        .exec(
            ExecuteRequestBuilder::from_deploy_item(deploy_item)
                .with_block_time(TEST_BLOCK_TIME) // tim: << return value of runtime::get_blocktime() is set here per-deploy
                .build(),
        )
        .commit()
}

pub fn get_expected_total_commission() -> U256 {
    (TEST_AMOUNT() * TEST_STABLE_COMMISSION_PERCENT() / 100) + TEST_GAS_COMMISSION()
}
