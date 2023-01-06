use anyhow::{anyhow, Context};
use casper_node::crypto::AsymmetricKeyExt;
use casper_types::{bytesrepr::ToBytes, AsymmetricType, Key, PublicKey, SecretKey};
use jsonrpc_lite::JsonRpc;
use once_cell::sync::Lazy;
use regex::Regex;
use reqwest::Url;
use serde::{de::DeserializeOwned, Deserialize};
use serde_json::Value;

use crate::{
    client::{CasperClient, ClientConfig},
    error::Error,
};

#[derive(Clone, Debug, Deserialize)]
pub struct CommonEnv {
    cspr_node: Option<String>,
    cspr_chain_name: Option<String>,
    cspr_secret: Option<String>,
    cspr_pk: Option<String>,
}

#[derive(Clone, Debug, Deserialize)]
pub struct BridgeEnv {
    cspr_bridge_contract_hash: String,
}

impl CommonEnv {
    pub fn local_node(secret: String) -> Self {
        Self {
            cspr_node: Some("http://localhost:11101".to_string()),
            cspr_chain_name: Some("casper-net-1".to_string()),
            cspr_secret: Some(secret),
            cspr_pk: None,
        }
    }

    pub fn node(&self) -> Result<&str, Error> {
        self.cspr_node
            .as_ref()
            .map(|s| s.as_str())
            .ok_or_else(|| Error::MissingConfigSetting {
                name: "CSPR_NODE".into(),
            })
    }

    pub fn chain_name(&self) -> Result<&str, Error> {
        self.cspr_chain_name
            .as_ref()
            .map(|s| s.as_str())
            .ok_or_else(|| Error::MissingConfigSetting {
                name: "CSPR_CHAIN_NAME".into(),
            })
    }

    pub fn secret(&self) -> Result<&str, Error> {
        self.cspr_secret
            .as_ref()
            .map(|s| s.as_str())
            .ok_or_else(|| Error::MissingConfigSetting {
                name: "CSPR_SECRET".into(),
            })
    }

    pub fn pk(&self) -> Result<&str, Error> {
        self.cspr_pk
            .as_ref()
            .map(|s| s.as_str())
            .ok_or_else(|| Error::MissingConfigSetting {
                name: "CSPR_PK".into(),
            })
    }

    pub fn make_client(&self) -> Result<CasperClient, Error> {
        let secret_key = self
            .cspr_secret
            .clone()
            .and_then(|s| SecretKey::from_file(&s).ok());

        let public_key = self
            .cspr_pk
            .clone()
            .and_then(|s| PublicKey::from_hex(s).ok());

        let mut config = ClientConfig {
            chain_name: self.chain_name()?.into(),
            main_account_secret: secret_key,
            main_account_public: public_key,
            event_port: None,
        };

        config.validate()?;
        let url = Url::parse(self.node()?).expect("invalid url");
        let client = CasperClient::new(url, config);

        Ok(client)
    }
}

impl BridgeEnv {
    pub fn bridge_contract_hash(&self) -> &str {
        &self.cspr_bridge_contract_hash
    }
}

pub trait JsonRpcExt {
    fn parse_as<'a, T: DeserializeOwned>(&self) -> anyhow::Result<T>;

    fn take_result(&self) -> anyhow::Result<Value>;
}

impl JsonRpcExt for JsonRpc {
    fn parse_as<'a, T: DeserializeOwned>(&self) -> anyhow::Result<T> {
        match self {
            JsonRpc::Request(_) => unimplemented!("jsonrpc request parsing unimplemented"),
            JsonRpc::Notification(_) => {
                unimplemented!("jsonrpc notification parsing unimplemented")
            }
            JsonRpc::Success(_) => {
                let result = self.get_result().expect("no result");
                serde_json::from_value(result.clone()).context("couldn't deserialize")
            }
            JsonRpc::Error(_) => {
                Err(self.get_error().expect("no error").clone()).context("jsonrpc request failed")
            }
        }
    }

    fn take_result(&self) -> anyhow::Result<Value> {
        match self {
            JsonRpc::Request(_) => unimplemented!("jsonrpc request parsing unimplemented"),
            JsonRpc::Notification(_) => {
                unimplemented!("jsonrpc notification parsing unimplemented")
            }
            JsonRpc::Success(_) => Ok(self.get_result().expect("no result").clone()),
            JsonRpc::Error(_) => {
                Err(self.get_error().expect("no error").clone()).context("jsonrpc request failed")
            }
        }
    }
}

pub fn process_simple_args(args: Vec<String>) -> anyhow::Result<Vec<String>> {
    static REGEX: Lazy<Regex> =
        Lazy::new(|| Regex::new(r"(.+?):(.+?)=(.+)").expect("invalid regex"));

    args.into_iter()
        .map(|s| {
            let matches = REGEX
                .captures(&s)
                .ok_or_else(|| anyhow!("no captures found in {s}"))?;

            let name = matches.get(1).expect("1-match").as_str();
            let ty = matches.get(2).expect("2-match").as_str();
            let value = matches.get(3).expect("3-match").as_str();

            if value == "null" || (value.starts_with('\'') && value.ends_with('\'')) {
                Ok(s)
            } else {
                Ok(format!("{name}:{ty}='{value}'"))
            }
        })
        .collect()
}

pub fn erc20_dictionary_key(owner: &Key) -> String {
    base64::encode(owner.to_bytes().expect("infallible"))
}
