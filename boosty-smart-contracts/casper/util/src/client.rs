use std::sync::Arc;

use anyhow::Context;
use arc_swap::ArcSwap;
use casper_hashing::Digest;
use casper_node::{
    event_stream_server::SseData,
    rpcs::{
        account::PutDeployParams,
        info::{GetDeployParams, JsonExecutionResult},
        state::{
            DictionaryIdentifier, GetDictionaryItemParams, GlobalStateIdentifier,
            QueryGlobalStateParams,
        },
    },
    types::{json_compatibility, Deploy},
};
use casper_types::{
    account::AccountHash, DeployHash, ExecutionResult, Key, ProtocolVersion, PublicKey, SecretKey,
    StoredValue,
};
use eventsource_stream::{EventStreamError, Eventsource};
use futures::{Stream, StreamExt};
use jsonrpc_lite::{JsonRpc, Params};
use serde::{de::DeserializeOwned, Serialize};
use serde_json::{json, Map, Value};

use crate::error::Error;

#[derive(Debug)]
pub struct ClientConfig {
    pub chain_name: String,
    pub main_account_secret: Option<SecretKey>,
    pub main_account_public: Option<PublicKey>,
    pub event_port: Option<u16>,
}

impl ClientConfig {
    pub fn validate(&mut self) -> Result<(), Error> {
        match (&self.main_account_secret, &self.main_account_public) {
            (Some(secret), None) => self.main_account_public = Some(PublicKey::from(secret)),
            (Some(secret), Some(public)) => {
                let secret_pk = PublicKey::from(secret);
                if &secret_pk != public {
                    return Err(Error::MainSecretPublicMismatch {
                        secret: format!("{secret_pk}"),
                        public: format!("{public}"),
                    });
                }
            }
            _ => {}
        }

        match self.chain_name.as_str() {
            "casper-test" | "casper" | "casper-net-1" => {}
            other => {
                return Err(Error::UnexpectedChainName {
                    name: other.to_string(),
                })
            }
        }

        Ok(())
    }

    pub fn main_secret(&self) -> Result<&SecretKey, Error> {
        self.main_account_secret
            .as_ref()
            .ok_or_else(|| Error::MissingConfigSetting {
                name: "main_account_secret".into(),
            })
    }

    pub fn main_public(&self) -> Result<PublicKey, Error> {
        self.main_account_public
            .clone()
            .ok_or_else(|| Error::MissingConfigSetting {
                name: "main_account_public".into(),
            })
    }

    pub fn main_account_hash(&self) -> Result<AccountHash, Error> {
        self.main_public().map(|public| AccountHash::from(&public))
    }

    pub fn main_key(&self) -> Result<Key, Error> {
        self.main_account_hash().map(|hash| Key::Account(hash))
    }
}

#[derive(Debug)]
pub struct CasperClient {
    pub(crate) http_client: reqwest::Client,
    pub(crate) node_url: reqwest::Url,

    pub(crate) config: ArcSwap<ClientConfig>,
}

impl CasperClient {
    pub fn new(node_url: reqwest::Url, config: ClientConfig) -> Self {
        let http_client = reqwest::Client::new();
        let config = ArcSwap::new(Arc::new(config));

        Self {
            node_url,
            config,
            http_client,
        }
    }

    pub(crate) fn config(&self) -> arc_swap::Guard<Arc<ClientConfig>> {
        self.config.load()
    }

    async fn make_request(&self, method: &str, params: Params) -> Result<Value, Error> {
        const RPC_API_PATH: &str = "rpc";
        let url = self
            .node_url
            .join(RPC_API_PATH)
            .expect("failed to construct url");

        let request = JsonRpc::request_with_params(0, method, params);
        let response = self
            .http_client
            .post(url)
            .json(&request)
            .send()
            .await?
            .error_for_status()?
            .json::<JsonRpc>()
            .await?;

        match &response {
            JsonRpc::Request(_) => Err(Error::UnexpectedRpcResponse {
                kind: "request".into(),
            }),
            JsonRpc::Notification(_) => Err(Error::UnexpectedRpcResponse {
                kind: "notification".into(),
            }),
            JsonRpc::Success(_) => Ok(response.get_result().expect("infallible").clone()),
            JsonRpc::Error(_) => Err(Error::RpcError(
                response.get_error().expect("infallible").clone(),
            )),
        }
    }

    pub async fn get_deploy(
        &self,
        deploy_hash: DeployHash,
    ) -> Result<(Deploy, Vec<ExecutionResult>), Error> {
        let request = GetDeployParams {
            deploy_hash: casper_node::types::DeployHash::new(Digest::from(deploy_hash.value())),
            finalized_approvals: false,
        };

        let response = self
            .make_request("info_get_deploy", Params::Map(request.into_json_map()))
            .await?;

        let deploy = response.field_parse::<_, Deploy>("deploy")?;
        let execution_results = response
            .field_parse::<_, Vec<JsonExecutionResult>>("execution_results")?
            .into_iter()
            .map(|s| s.result)
            .collect();

        Ok((deploy, execution_results))
    }

    pub async fn get_state_root_hash(&self) -> Result<Digest, Error> {
        let response = self
            .make_request("chain_get_state_root_hash", Params::None(()))
            .await?;

        let state_root_hash = response
            .field_parse::<_, Option<Digest>>("state_root_hash")?
            .unwrap_or_else(|| todo!("handle missing digest"));

        Ok(state_root_hash)
    }

    pub async fn query_global_state<'a, P: IntoIterator<Item = &'a str>>(
        &self,
        state_identifier: GlobalStateIdentifier,
        key: Key,
        path: P,
    ) -> Result<QueryGlobalStateResult, Error> {
        let request = QueryGlobalStateParams {
            state_identifier,
            key: key.to_formatted_string(),
            path: path.into_iter().map(|s| s.to_string()).collect(),
        };

        let response = self
            .make_request("query_global_state", Params::Map(request.into_json_map()))
            .await?;

        let stored_value = response
            .field_parse::<_, json_compatibility::StoredValue>("stored_value")?
            .into();

        let merkle_proof = response.field_parse::<_, String>("merkle_proof")?;

        Ok(QueryGlobalStateResult {
            stored_value,
            merkle_proof,
        })
    }

    pub async fn get_dictionary_item(
        &self,
        state_root_hash: Digest,
        identifier: DictionaryIdentifier,
    ) -> Result<GetDictionaryItemResult, Error> {
        let request = GetDictionaryItemParams {
            state_root_hash,
            dictionary_identifier: identifier,
        };

        let response = self
            .make_request(
                "state_get_dictionary_item",
                Params::Map(request.into_json_map()),
            )
            .await?;

        let stored_value = response
            .field_parse::<_, json_compatibility::StoredValue>("stored_value")?
            .into();
        let merkle_proof = response.field_parse::<_, String>("merkle_proof")?;
        let dictionary_key = response.field_parse::<_, String>("dictionary_key")?;

        Ok(GetDictionaryItemResult {
            dictionary_key,
            stored_value,
            merkle_proof,
        })
    }

    pub async fn put_deploy(&self, deploy: Deploy) -> Result<DeployHash, Error> {
        let request = PutDeployParams { deploy };

        let response = self
            .make_request("account_put_deploy", Params::Map(request.into_json_map()))
            .await?;

        let deploy_hash = response.field_parse::<_, DeployHash>("deploy_hash")?;

        Ok(deploy_hash)
    }

    async fn event_stream(
        &self,
        stream_kind: &str,
    ) -> Result<impl Stream<Item = Result<SseData, EventStreamError<reqwest::Error>>>, Error> {
        // let mut stream = self.http_client.
        let mut node_event_url = self.node_url.clone();
        node_event_url
            .set_port(self.config.load().event_port.or_else(|| Some(9999)))
            .expect("invalid url");
        node_event_url = node_event_url
            .join(&format!("events/{stream_kind}"))
            .unwrap();

        // TODO: proper error handling
        Ok(self
            .http_client
            .get(node_event_url)
            .send()
            .await
            .context("couldn't establish connection to event server")?
            .bytes_stream()
            .eventsource()
            .map(|result| {
                result.map(|event| {
                    let data = event.data;
                    match serde_json::from_str::<SseData>(&data) {
                        Ok(data) => data,
                        Err(err) => {
                            println!("ssedata deser error: {err:?}");
                            SseData::ApiVersion(ProtocolVersion::V1_0_0)
                        }
                    }
                })
            }))
    }

    pub async fn event_stream_main(
        &self,
    ) -> Result<impl Stream<Item = Result<SseData, EventStreamError<reqwest::Error>>>, Error> {
        self.event_stream("main").await
    }

    pub async fn event_stream_deploys(
        &self,
    ) -> Result<impl Stream<Item = Result<SseData, EventStreamError<reqwest::Error>>>, Error> {
        self.event_stream("deploys").await
    }
}

pub(crate) trait JsonValueExt {
    fn field<I>(&self, index: I) -> Result<&Value, Error>
    where
        I: serde_json::value::Index + ToString;

    fn field_parse<I, T>(&self, index: I) -> Result<T, Error>
    where
        I: serde_json::value::Index + ToString,
        T: DeserializeOwned,
    {
        let value = self.field(index)?.clone();
        serde_json::from_value(value).map_err(Error::JsonDeserError)
    }
}

impl JsonValueExt for Value {
    fn field<I>(&self, index: I) -> Result<&Value, Error>
    where
        I: serde_json::value::Index + ToString,
    {
        let value = &self[&index];

        if value.is_null() {
            Err(Error::MissingResponseField {
                field: index.to_string(),
            })
        } else {
            Ok(value)
        }
    }
}

pub(crate) trait IntoJsonMap: Serialize {
    fn into_json_map(self) -> Map<String, Value>
    where
        Self: Sized,
    {
        json!(self)
            .as_object()
            .unwrap_or_else(|| panic!("should be a JSON object"))
            .clone()
    }
}

impl<T: Serialize> IntoJsonMap for T {}

pub struct QueryGlobalStateResult {
    pub stored_value: StoredValue,
    pub merkle_proof: String,
}

pub struct GetDictionaryItemResult {
    pub dictionary_key: String,
    pub stored_value: StoredValue,
    pub merkle_proof: String,
}
