use std::{sync::Arc, time::Duration};

use anyhow::Context;
use ethers::{
    prelude::Http,
    providers::{Middleware, Provider},
    types::H160,
};
use futures::{
    channel::mpsc::{Receiver, Sender},
    future::join_all,
    SinkExt,
};
use prost_types::Timestamp;
use reqwest::Url;
use tonic::{Request, Response, Status};
use tonic_codegen::{bridge_oracle_server::BridgeOracle, PriceUpdate};
use tracing::instrument;

use crate::error::OracleError;

pub type TokenName = String;
pub type AggregatorENSDomain = String;

pub type Subscription = (TokenName, AggregatorENSDomain);

const EVENT_SLEEP_DURATION: Duration = Duration::from_secs(1);

mod chain_link_abi {
    use ethers::prelude::abigen;

    abigen!(AggregatorV3Contract, "aggregatorV3InterfaceABI.json");
}

/// The price oracle. It fetches prices from the chainlink oracle contracts on the Ethereum network.
#[derive(Debug)]
pub struct PriceOracle {
    provider: Arc<Provider<Http>>,
    subscription_list: Vec<(TokenName, H160)>,
}

impl PriceOracle {
    /// Creates a new price oracle with the given subscription list.
    pub async fn new_with_subscription(
        ethereum_node: Url,
        subscription_list: Vec<Subscription>,
    ) -> Result<Self, OracleError> {
        let provider =
            Provider::<Http>::try_from(ethereum_node.to_string()).with_context(|| {
                format!("Couldn't connect to ethereum node under {ethereum_node} url",)
            })?;
        let provider = Arc::new(provider);
        let subscription_list = subscription_list.into_iter().map(|(name, ens_name)| async {
            let ens_name = ens_name;
            provider
                .resolve_name(&ens_name)
                .await
                .map(|address| (name, address))
        });
        let subscription_list: Result<Vec<(TokenName, ethers::types::H160)>, _> =
            join_all(subscription_list).await.into_iter().collect();
        let subscription_list = subscription_list?;

        Ok(PriceOracle {
            provider,
            subscription_list,
        })
    }

    /// Get the price for the given token name and address from the chainlink oracle contract.
    #[instrument(skip(self))]
    pub async fn get_price(
        &self,
        token_name: &str,
        address: H160,
    ) -> Result<PriceUpdate, OracleError> {
        let oracle_contract =
            chain_link_abi::AggregatorV3Contract::new(address, self.provider.clone());
        let decimals = oracle_contract
            .decimals()
            .call()
            .await
            .context("Couldn't fetch decimals data for {token_name}")?;
        let (_, latest_price, _, updated_at, _) =
            oracle_contract.latest_round_data().call().await?;

        if latest_price.is_negative() {
            return Err(OracleError::NegativeLatestPrice(token_name.to_string()));
        }
        Ok(PriceUpdate {
            token_name: token_name.to_string(),
            amount: latest_price.to_string(),
            decimals: decimals as u32,
            last_update: Some(Timestamp {
                seconds: updated_at.as_u64() as i64,
                nanos: 0,
            }),
        })
    }

    /// Get the prices for all the tokens in the subscription list.
    #[instrument(skip(self))]
    pub async fn get_prices(&self) -> Vec<Result<PriceUpdate, OracleError>> {
        let mut result = vec![];
        result.reserve(self.subscription_list.len());

        for (token_name, address) in &self.subscription_list {
            result.push(self.get_price(token_name, *address));
        }

        join_all(result).await
    }
}

#[tonic::async_trait]
impl BridgeOracle for PriceOracle {
    type PriceStreamStream = Receiver<Result<PriceUpdate, Status>>;

    /// Returns a stream of prices for the tokens in the subscription list.
    #[instrument(skip(self))]
    async fn price_stream(
        &self,
        _: Request<()>,
    ) -> Result<Response<Self::PriceStreamStream>, Status> {
        let (sender, receiver) = futures::channel::mpsc::channel(256);
        let oracle = PriceOracle {
            provider: self.provider.clone(),
            subscription_list: self.subscription_list.clone(),
        };
        tokio::spawn(price_processor(oracle, sender));
        Ok(Response::new(receiver))
    }
}

/// The price processor. It fetches prices from the chainlink oracle contracts on the Ethereum network
/// and sends to the given sender.
/// It sleeps for 1 second between each price fetch.
async fn price_processor(
    price_oracle: PriceOracle,
    mut sender: Sender<Result<PriceUpdate, tonic::Status>>,
) {
    loop {
        tokio::time::sleep(EVENT_SLEEP_DURATION).await;
        let prices = price_oracle.get_prices().await;
        for data in prices {
            match data {
                Ok(price) => {
                    if let Err(err) = sender.feed(Ok(price)).await {
                        tracing::error!("failed to send an price: {err}")
                    }
                }
                Err(err) => tracing::error!("Failed to fetch price for one of the streams: {err}"),
            }
        }

        if let Err(err) = sender.flush().await {
            tracing::error!("failed to flush prices: {err}")
        }
    }
}

#[cfg(test)]
mod tests {

    use super::PriceOracle;

    #[tokio::test]
    async fn oracle_test() {
        let subscription_list = vec![
            ("ETH".to_string(), "aggregator.eth-usd.data.eth".to_string()),
            (
                "CSPR".to_string(),
                "aggregator.cspr-usd.data.eth".to_string(),
            ),
        ];
        let oracle = PriceOracle::new_with_subscription(
            "https://eth-mainnet.nodereal.io/v1/1659dfb40aa24bbb8153a677b98064d7"
                .try_into()
                .unwrap(),
            subscription_list,
        )
        .await
        .unwrap();
        println!("ETH - {:?}$", oracle.get_prices().await);
    }
}
