extern crate casper_util;

use std::fs::File;
use std::io::Write;

use anyhow::{anyhow, Context};
use casper_execution_engine::core::engine_state::ExecutableDeployItem;
use casper_node::{crypto::AsymmetricKeyExt, rpcs::state::GlobalStateIdentifier};
use casper_types::{U128, CLValue};
use casper_types::{
    account::AccountHash,
    bytesrepr::{Bytes, FromBytes},
    ContractHash, DeployHash, Key, PublicKey, RuntimeArgs, SecretKey, StoredValue, U256, U512,
};
use casper_util::util::{BridgeEnv, CommonEnv};
use clap::Parser;
use connectors_common::connector_config::ConnectorConfig;
use reqwest::Url;
use serde_json::json;

#[derive(Parser)]
enum Command {
    DeployContract {
        #[clap(short = 'c')]
        session_code_path: String,
    },
    DeployErc20 {
        #[clap(short = 'c')]
        session_code_path: String,

        #[clap(short = 's')]
        symbol: String,
        #[clap(short = 'n')]
        name: String,
        #[clap(short = 'd')]
        decimals: String,
        #[clap(short = 't')]
        total_supply: String,
    },
    GetDeploy {
        #[clap(short = 'd')]
        deploy_hash: String,
    },
    Erc20Balance {
        #[clap(short = 't')]
        token_contract: String,
        #[clap(short = 'o')]
        owner: String,
    },
    Erc20GetBalanceDeploy {
        #[clap(short = 't')]
        token_contract: String,
        #[clap(short = 'o')]
        owner: String,
    },
    Erc20Transfer {
        #[clap(short = 't')]
        token_contract: String,
        #[clap(short = 'd')]
        destination: String,
        #[clap(short = 'a')]
        amount: String,
    },

    DeployBridgeContract {
        #[clap(short = 'c')]
        session_code_path: String,
    },

    BridgeIn {
        #[clap(short = 't')]
        token_contract: String,
        #[clap(short = 'a')]
        amount: String,
        gas_commission: String,
        deadline: String,
        nonce: String,
        destination_chain: String,
        destination_address: String,
        signature: String,
    },
    BridgeOut {
        #[clap(short = 't')]
        token_contract: String,
        #[clap(short = 'a')]
        amount: String,
        transaction_id: String,
        source_chain: String,
        source_address: String,

        #[clap(short = 'r')]
        recipieint: String,
    },
    SetStableCommissionPercent {
        #[clap(short = 'a')]
        stable_commission_percent: String,
    },
    SetSigner {
        #[clap(short = 'a')]
        signer: String,
    },
    // GetStableCommissionPercent {},
    DeployEverythingOnLocalnet {
        #[clap(short = 'b')]
        bridge_contract_path: String,
        #[clap(short = 't')]
        token_contract_path: String,
        #[clap(short = 's')]
        secret_key_path: String,
        #[clap(short = 'f')]
        config_output_path: Option<String>,
    },
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    dotenv::dotenv().context("couldn't load .env file")?;

    let command = Command::parse();
    let env: CommonEnv = envy::from_env().context("couldn't parse environment")?;

    match command {
        Command::DeployContract { session_code_path } => {
            deploy_contract(&env, session_code_path).await?
        }
        Command::DeployErc20 {
            session_code_path,
            symbol,
            name,
            decimals,
            total_supply,
        } => {
            deploy_erc20(
                &env,
                session_code_path,
                name,
                symbol,
                decimals,
                total_supply,
            )
            .await?
        }
        Command::GetDeploy { deploy_hash } => get_deploy(&env, deploy_hash).await?,
        Command::Erc20Balance {
            token_contract,
            owner,
        } => erc20_balance(&env, token_contract, owner).await?,
        Command::Erc20GetBalanceDeploy {
            token_contract,
            owner,
        } => erc20_get_balance_deploy(&env, token_contract, owner).await?,
        Command::Erc20Transfer {
            token_contract,
            destination,
            amount,
        } => erc20_transfer(&env, token_contract, destination, amount).await?,

        Command::DeployBridgeContract { session_code_path } => {
            deploy_bridge_contract(&env, session_code_path).await?
        }

        Command::BridgeIn {
            token_contract,
            amount,
            gas_commission,
            nonce,
            deadline,
            destination_chain,
            destination_address,
            signature,
        } => {
            bridge_transfer_in(
                &env,
                destination_chain,
                destination_address,
                token_contract,
                amount,
                gas_commission,
                deadline,
                nonce,
                signature,
            )
            .await?
        }
        Command::BridgeOut {
            token_contract,
            amount,
            transaction_id,
            source_chain,
            source_address,
            recipieint: recipient,
        } => {
            bridge_transfer_out(
                &env,
                source_chain,
                source_address,
                token_contract,
                amount,
                transaction_id,
                recipient,
            )
            .await?
        }
        Command::SetStableCommissionPercent {
            stable_commission_percent,
        } => set_stable_commission_percent(&env, stable_commission_percent).await?,
        Command::SetSigner { signer } => set_signer(&env, signer).await?,
        // Command::GetStableCommissionPercent => {
        //     get_stable_commission_percent(
        //         &env
        //     ).await?
        // }
        // Command::GetSigner => {
        //     get_signer(
        //         &env
        //     ).await?
        // }
        Command::DeployEverythingOnLocalnet {
            bridge_contract_path,
            token_contract_path,
            secret_key_path,
            config_output_path,
        } => {
            deploy_everything_on_localnet(
                bridge_contract_path,
                token_contract_path,
                secret_key_path,
                config_output_path,
            )
            .await?
        }
    }

    Ok(())
}

async fn deploy_contract(env: &CommonEnv, session_code_path: String) -> anyhow::Result<()> {
    let client = env.make_client()?;

    let code = tokio::fs::read(session_code_path)
        .await
        .context("couldn't read code file")?;

    let result = client.put_contract(code, RuntimeArgs::new()).await?;

    println!("{}", json!(result));
    Ok(())
}

async fn deploy_erc20(
    env: &CommonEnv,
    session_code_path: String,
    name: String,
    symbol: String,
    decimals: String,
    total_supply: String,
) -> anyhow::Result<()> {
    let client = env.make_client()?;

    let code = tokio::fs::read(session_code_path)
        .await
        .context("couldn't read code file")?;

    let deploy_args = RuntimeArgs::try_new(|args| {
        args.insert("name", name)?;
        args.insert("symbol", symbol)?;
        args.insert(
            "decimals",
            u8::from_str_radix(&decimals, 10).expect("invalid decimals"),
        )?;
        args.insert(
            "total_supply",
            U256::from_dec_str(&total_supply).expect("invalid total supply"),
        )?;
        Ok(())
    })
    .expect("invalid args");

    let deploy_hash = client.put_contract(code, deploy_args).await?;

    println!("{}", json!(deploy_hash));
    eprintln!("waiting for deploy to be confirmed");
    client.confirm_deploy(deploy_hash).await?;
    eprintln!("deploy confirmed");

    Ok(())
}

async fn get_deploy(env: &CommonEnv, deploy_hash: String) -> anyhow::Result<()> {
    let client = env.make_client()?;

    let (deploy_hash, _) = DeployHash::from_bytes(
        &base16::decode(&deploy_hash).context("couldn't decode deploy hash")?,
    )
    .map_err(|_| anyhow!("couldn't decode deploy hash"))?;

    let (deploy, execution_results) = client.get_deploy(deploy_hash).await?;

    println!(
        "{:#}",
        json!({"deploy": deploy, "execution_results":  execution_results})
    );

    Ok(())
}

async fn erc20_balance(env: &CommonEnv, contract: String, owner: String) -> anyhow::Result<()> {
    let client = env.make_client()?;

    let contract = client.key_from_str(&contract)?;
    let who = client.key_from_str(&owner)?;
    let result = client.erc20_query_balance(contract, who).await?;

    println!("{:#}", json!(result));

    Ok(())
}

async fn erc20_get_balance_deploy(
    env: &CommonEnv,
    contract: String,
    who: String,
) -> anyhow::Result<()> {
    let client = env.make_client()?;

    let contract = client.key_from_str(&contract)?;
    let who = client.key_from_str(&who)?;

    let deploy_hash = client.erc20_deploy_get_balance(contract, who).await?;

    println!("{:#}", json!(deploy_hash));
    eprintln!("waiting for deploy to be confirmed");
    client.confirm_deploy(deploy_hash).await?;
    eprintln!("deploy confirmed");

    Ok(())
}

async fn erc20_transfer(
    env: &CommonEnv,
    contract: String,
    to: String,
    amount: String,
) -> anyhow::Result<()> {
    let client = env.make_client()?;

    let contract_key = client.key_from_str(&contract)?;
    let to_key = client.key_from_str(&to)?;
    let amount = U256::from_dec_str(&amount).context("couldn't parse amount")?;

    let deploy_hash = client
        .erc20_deploy_transfer(contract_key, to_key, amount)
        .await?;

    println!("{:#}", json!(deploy_hash));
    eprintln!("waiting for deploy to be confirmed");
    client.confirm_deploy(deploy_hash).await?;
    eprintln!("deploy confirmed");

    Ok(())
}

async fn deploy_bridge_contract(env: &CommonEnv, session_code_path: String) -> anyhow::Result<()> {
    let client = env.make_client()?;
    let session_code = tokio::fs::read(session_code_path)
        .await
        .context("couldn't read session code file")?;
    let mut args = RuntimeArgs::new();

    args.insert_cl_value("signer", CLValue::from_t("").expect("infallible"));
    
    let deploy = client.make_simple_deploy(
        U512::one() * 200_000_000_000u64,
        ExecutableDeployItem::ModuleBytes {
            module_bytes: Bytes::from(session_code),
            args: args,
        },
    )?;

    let deploy_hash = client.put_deploy(deploy).await?;

    println!("{}", json!(deploy_hash));
    eprintln!("waiting for deploy to be confirmed");
    client.confirm_deploy(deploy_hash).await?;
    eprintln!("deploy confirmed");

    Ok(())
}

async fn bridge_transfer_in(
    env: &CommonEnv,
    destination_chain: String,
    destination_address: String,
    token_contract: String,
    amount: String,
    gas_commission: String,
    deadline: String,
    nonce: String,
    signature: String,
) -> anyhow::Result<()> {
    let bridge_env: BridgeEnv = envy::from_env().context("couldn't parse environment")?;

    let client = env.make_client()?;

    let bridge_contract = ContractHash::from_formatted_str(bridge_env.bridge_contract_hash())
        .expect("invalid bridge contract hash");
    let token_contract = client.key_from_str(&token_contract)?;
    let amount = U256::from_dec_str(&amount).context("couldn't parse amount")?;
    let gas_commission =
        U256::from_dec_str(&gas_commission).context("couldn't parse gas_commission")?;
    let deadline = U256::from_dec_str(&deadline).context("couldn't parse deadline")?;
    let nonce = U128::from_dec_str(&nonce).context("couldn't parse amount")?;

    let deploy_hash = client
        .bridge_in(
            bridge_contract,
            token_contract,
            amount,
            gas_commission,
            deadline,
            nonce,
            destination_chain,
            destination_address,
            signature,
        )
        .await?;

    println!("{}", json!(deploy_hash));
    eprintln!("waiting for deploy to be confirmed");
    client.confirm_deploy(deploy_hash).await?;
    eprintln!("deploy confirmed");

    Ok(())
}

async fn bridge_transfer_out(
    env: &CommonEnv,
    source_chain: String,
    source_address: String,
    token_contract: String,
    amount: String,
    transaction_id: String,
    recipient: String,
) -> anyhow::Result<()> {
    let bridge_env: BridgeEnv = envy::from_env().context("couldn't parse environment")?;

    let client = env.make_client()?;

    let bridge_contract = ContractHash::from_formatted_str(bridge_env.bridge_contract_hash())
        .expect("invalid bridge contract hash");
    let token_contract = client.key_from_str(&token_contract)?;
    let amount = U256::from_dec_str(&amount).context("couldn't parse amount")?;
    let transaction_id =
        U256::from_dec_str(&transaction_id).context("couldn't parse transaction_id")?;
    let recipient = client.key_from_str(&recipient)?;

    let deploy_hash = client
        .bridge_out(
            bridge_contract,
            token_contract,
            amount,
            transaction_id,
            recipient,
            source_chain,
            source_address,
        )
        .await?;

    println!("{}", json!(deploy_hash));
    eprintln!("waiting for deploy to be confirmed");
    client.confirm_deploy(deploy_hash).await?;
    eprintln!("deploy confirmed");

    Ok(())
}

async fn set_stable_commission_percent(
    env: &CommonEnv,
    stable_commission_percent: String,
) -> anyhow::Result<()> {
    let bridge_env: BridgeEnv = envy::from_env().context("couldn't parse environment")?;

    let client = env.make_client()?;

    let bridge_contract = ContractHash::from_formatted_str(bridge_env.bridge_contract_hash())
        .expect("invalid bridge contract hash");
    let stable_commission_percent = U256::from_dec_str(&stable_commission_percent)
        .context("couldn't parse stable_commission_percent")?;

    let deploy_hash = client
        .set_stable_commission_percent(bridge_contract, stable_commission_percent)
        .await?;

    println!("{}", json!(deploy_hash));
    eprintln!("waiting for deploy to be confirmed");
    client.confirm_deploy(deploy_hash).await?;
    eprintln!("deploy confirmed");

    Ok(())
}

async fn set_signer(env: &CommonEnv, signer: String) -> anyhow::Result<()> {
    let bridge_env: BridgeEnv = envy::from_env().context("couldn't parse environment")?;

    let client = env.make_client()?;

    let bridge_contract = ContractHash::from_formatted_str(bridge_env.bridge_contract_hash())
        .expect("invalid bridge contract hash");

    let deploy_hash = client.set_signer(bridge_contract, signer).await?;

    println!("{}", json!(deploy_hash));
    eprintln!("waiting for deploy to be confirmed");
    client.confirm_deploy(deploy_hash).await?;
    eprintln!("deploy confirmed");

    Ok(())
}
async fn get_stable_commission_percent(env: &CommonEnv) -> anyhow::Result<()> {
    //
    let bridge_env: BridgeEnv = envy::from_env().context("couldn't parse environment")?;

    let client = env.make_client()?;

    let bridge_contract = ContractHash::from_formatted_str(bridge_env.bridge_contract_hash())
        .expect("invalid bridge contract hash");

    let deploy_hash = client
        .get_stable_commission_percent(bridge_contract)
        .await?;

    println!("{}", json!(deploy_hash));
    eprintln!("waiting for deploy to be confirmed");
    client.confirm_deploy(deploy_hash).await?;
    eprintln!("deploy confirmed");

    Ok(())
}

async fn deploy_everything_on_localnet(
    bridge_path: String,
    token_path: String,
    secret: String,
    config_output: Option<String>,
) -> anyhow::Result<()> {
    let env = CommonEnv::local_node(secret.clone());
    deploy_bridge_contract(&env, bridge_path).await?;
    deploy_erc20(
        &env,
        token_path,
        "TEST".to_string(),
        "TTT".to_string(),
        "12".to_string(),
        U256::max_value().to_string(),
    )
    .await?;

    if let Some(config_output) = config_output {
        let client = env.make_client()?;
        let state = client.get_state_root_hash().await?;
        let secret_key = SecretKey::from_file(secret)?;
        let pk = PublicKey::from(&secret_key);
        let account_hash = AccountHash::from(&pk);
        println!("Contracts deployed, receiving user info");
        let result = client
            .query_global_state(
                GlobalStateIdentifier::StateRootHash(state),
                Key::from(account_hash),
                vec![],
            )
            .await?;
        if let StoredValue::Account(account) = result.stored_value {
            let keys = account.named_keys();
            println!("User info received, parsing named keys");

            let bridge_hash: Key = *keys.get("bridge_contract").context("no bridge contract")?;

            let token_key = *keys
                .get("erc20_token_contract")
                .context("no erc20 token contract")?;

            let bridge_hash = ContractHash::from(
                bridge_hash
                    .into_hash()
                    .context("couldn't convert bridge hash to contract hash")?,
            );

            println!("Converting token contrach hash to contract package hash");

            client
                .query_global_state(
                    GlobalStateIdentifier::StateRootHash(state),
                    token_key,
                    vec![],
                )
                .await?;

            println!("Token package hash received, parsing");

            let config = ConnectorConfig {
                url: Url::parse("http://localhost:11101")?,
                network_id: 0,
                network_name: "casper-net-1".to_string(),
                is_testnet: true,
                bridge_contract_hash: bridge_hash,
            };
            println!("Writing config");
            let config_string = toml::to_string(&config)?;
            let mut config_file = File::create(config_output)?;
            config_file.write_all(config_string.as_bytes())?;
        } else {
            anyhow::bail!("couldn't find bridge account")
        }
    }
    Ok(())
}
