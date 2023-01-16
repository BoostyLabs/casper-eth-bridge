use std::{str::FromStr, sync::Arc};

use anyhow::{anyhow, Context};
use async_trait::async_trait;
use futures::StreamExt;
use primitive_types::U256;
use tonic::transport::{Channel, Endpoint};
use tonic_codegen::{
    connector_client::ConnectorClient, event, CancelSignatureResponse, EventFundsIn, EventFundsOut,
    EventsRequest, StringNetworkAddress, TokenOutRequest,
};
use tracing::{error, info, warn, Instrument};

use crate::{
    bridge::Bridge,
    error::BridgeError,
    registry::NetworkMetadata,
    types::{
        Address, BridgeEvent, BridgeTokenTransferIn, BridgeTokenTransferOut, ConfirmedTx,
        StringAddress, TxHash,
    },
    NetworkConnector,
};

pub struct Config {
    endpoint: String,
}

impl Config {
    pub fn new(endpoint: String) -> Self {
        Self { endpoint }
    }
}

pub struct Connector {
    client: ConnectorClient<Channel>,
    metadata: NetworkMetadata,
    bridge: Arc<Bridge>,
}

impl Connector {
    /// Starts a new connector in a separate task.
    #[tracing::instrument(skip(config, bridge), err)]
    pub async fn start(config: Config, bridge: Arc<Bridge>) -> Result<(), BridgeError> {
        let Config { endpoint } = config;
        let channel = Endpoint::from_str(&endpoint)
            .context("couldn't parse endpoint")?
            .connect_lazy();

        let span = tracing::info_span!("network_connector", endpoint = %endpoint);
        let client = ConnectorClient::new(channel);

        tokio::spawn(
            async move {
                loop {
                    match Self::try_start(&bridge, &client).await {
                        Ok(()) => {
                            warn!("connector terminated unexpectedly without an explicit error")
                        }
                        Err(error) => error!("{:#}", anyhow::Error::new(error)),
                    }

                    tokio::time::sleep(crate::consts::RETRY_TIMEOUT).await;
                    warn!("retrying connector start");
                }
            }
            .instrument(span),
        );

        Ok(())
    }

    /// Tries to connect to the connector and register it with the bridge. It will retry until it is successful.
    /// The returned future will never return.
    async fn try_start(
        bridge: &Arc<Bridge>,
        client: &ConnectorClient<Channel>,
    ) -> Result<(), BridgeError> {
        let metadata = crate::grpc::retry_request(
            move || {
                let mut client = client.clone();
                async move { client.metadata(()).await }
            },
            |error| warn!(error=%error, "error when trying to request metadata, retrying"),
        )
        .await
        .map_err(|error| anyhow::anyhow!("failed to send request with: {error}"))?
        .into_inner();

        let metadata: NetworkMetadata = metadata.into();

        let connector = Arc::new(Self {
            client: client.clone(),
            metadata,
            bridge: bridge.clone(),
        });

        bridge
            .register_connector(connector.clone())
            .await
            .context("couldn't register connector with bridge")?;

        connector.event_loop().await;

        unreachable!("event loop must never terminate once started")
    }

    /// Processes events from the connected connector. It will never return.
    #[tracing::instrument(skip(self))]
    async fn event_loop(&self) {
        let mut client = self.client.clone();

        let metadata = &self.metadata;

        loop {
            tokio::time::sleep(crate::consts::RETRY_TIMEOUT).await;

            let block_number = match self.bridge.last_seen_network_block(metadata.id()).await {
                Ok(v) => v,
                Err(error) => {
                    warn!(network_id = %metadata.id(), network_name = %metadata.name(), error=%error, "error fetching block number from connector");

                    None
                }
            };

            let stream = match client.event_stream(EventsRequest { block_number }).await {
                Ok(stream) => stream.into_inner(),
                Err(error) => {
                    error!(network_id = %metadata.id(), network_name = %metadata.name(), error=%error, "error establishing stream");
                    continue;
                }
            };

            stream
                .for_each_concurrent(16, |result| async {
                    match result {
                        Ok(event) => {
                            info!(network_id = %metadata.id(), network_name = %metadata.name(), "event from grpc connector");

                            if let Some(variant) = event.variant {
                                if let Err(error) = self.handle_gprc_event(variant).await {
                                    error!(network_id = %metadata.id(), network_name = %metadata.name(), error=%error, "error sending event to bridge");
                                }
                            }
                        }
                        Err(error) => {
                            error!(network_id = %metadata.id(), network_name = %metadata.name(), error=%error, "error from grpc connector");
                        }
                    }
                })
            .await;

            warn!(network_id = %metadata.id(), network_name = %metadata.name(), "event stream terminated, retrying");
        }
    }

    /// Converts a gRPC event into a bridge event and sends it to the bridge.
    async fn handle_gprc_event(&self, event: event::Variant) -> Result<(), BridgeError> {
        match event {
            event::Variant::FundsIn(event) => {
                let event = self
                    .parse_funds_in(event)
                    .await
                    .context("couldn't parse funds in event")?;
                let tx = self.bridge.event_tx();
                tx.send(BridgeEvent::TokenTransferIn(event))
                    .await
                    .context("couldn't send event to bridge")?;
            }
            event::Variant::FundsOut(event) => {
                let event = self
                    .parse_funds_out(event)
                    .await
                    .context("couldn't parse funds in event")?;
                let tx = self.bridge.event_tx();
                tx.send(BridgeEvent::TokenTransferOut(event))
                    .await
                    .context("couldn't send event to bridge")?;
            }
        }

        Ok(())
    }

    /// Parses a gRPC funds in event into a bridge event.
    async fn parse_funds_in(
        &self,
        event: EventFundsIn,
    ) -> Result<BridgeTokenTransferIn, BridgeError> {
        let EventFundsIn {
            from,
            to,
            amount,
            token,
            tx,
        } = event;

        let from = from.ok_or_else(|| anyhow!("missing `from` field"))?;
        let to = to.ok_or_else(|| anyhow!("missing `to` field"))?.into();
        let amount: U256 = U256::from_dec_str(&amount).context("couldn't parse `amount` field")?;
        let token = token.ok_or_else(|| anyhow!("misstoken` field"))?;
        let tx = tx.ok_or_else(|| anyhow!("missing `tx` field"))?;
        let tx_hash = TxHash::new(self.metadata.id(), tx.hash);
        let tx = ConfirmedTx::new(
            tx_hash,
            Address::new(self.metadata.id(), tx.sender),
            tx.blocknumber,
        );

        if let Err(error) = self
            .bridge
            .update_last_seen_network_block(self.metadata.id(), tx.block_number())
            .await
        {
            let metadata = &self.metadata;
            warn!(network_id = %metadata.id(), network_name = %metadata.name(), error=%error, "couldn't update last seen network block");
        }

        let from = Address::new(self.metadata.id(), from.address);
        let token = Address::new(self.metadata.id(), token.address);

        Ok(BridgeTokenTransferIn {
            from,
            to,
            amount,
            token,
            tx,
        })
    }

    /// Parses a gRPC funds out event into a bridge event.
    async fn parse_funds_out(
        &self,
        event: EventFundsOut,
    ) -> Result<BridgeTokenTransferOut, BridgeError> {
        let EventFundsOut {
            from,
            to,
            amount,
            token,
            tx,
        } = event;

        let from = from.ok_or_else(|| anyhow!("missing `from` field"))?;
        let to = to.ok_or_else(|| anyhow!("missing `to` field"))?;
        let amount: U256 = U256::from_dec_str(&amount).context("couldn't parse `amount` field")?;
        let token = token.ok_or_else(|| anyhow!("missing `token` field"))?;
        let tx = tx.ok_or_else(|| anyhow!("missing `tx` field"))?;
        let tx_hash = TxHash::new(self.metadata.id(), tx.hash);
        let tx = ConfirmedTx::new(
            tx_hash,
            Address::new(self.metadata.id(), tx.sender),
            tx.blocknumber,
        );

        if let Err(error) = self
            .bridge
            .update_last_seen_network_block(self.metadata.id(), tx.block_number())
            .await
        {
            let metadata = &self.metadata;
            warn!(network_id = %metadata.id(), network_name = %metadata.name(), error=%error, "couldn't update last seen network block");
        }

        let from = StringAddress::new(from.network_name, from.address);
        let to = Address::new(self.metadata.id(), to.address);
        let token = Address::new(self.metadata.id(), token.address);

        Ok(BridgeTokenTransferOut {
            from,
            to,
            amount,
            token,
            tx,
        })
    }
}

#[async_trait]
impl NetworkConnector for Connector {
    /// Returns the network metadata that was cached when the connector was created.
    fn metadata(&self) -> NetworkMetadata {
        self.metadata.clone()
    }

    /// Calls the `bridge_out` method of the gRPC client and returns the transaction hash on the destination network.
    async fn bridge_out(
        &self,
        recipient: Address,
        token: Address,
        amount: U256,
        source_address: StringAddress,
        transaction_id: u64,
    ) -> Result<TxHash, BridgeError> {
        let client = self.client.clone();
        let network_id = self.metadata.id();
        let request = TokenOutRequest {
            amount: amount.to_string(),
            token: Some((&token).into()),
            to: Some((&recipient).into()),
            from: Some((&source_address).into()),
            transaction_id,
        };

        let hash = crate::grpc::retry_request(
            move || {
                let mut client = client.clone();
                let request = request.clone();
                async move { client.bridge_out(request).await }
            },
            |error| warn!(error=%error, "error when trying to bridge out, retrying"),
        )
        .await
        .map_err(|error| anyhow::anyhow!("failed to send request with: {error}"))?
        .into_inner()
        .txhash;

        Ok(TxHash::new(network_id, hash))
    }

    async fn bridge_in_signature(
        &self,
        sender: Address,
        token: Address,
        nonce: u64,
        amount: U256,
        destination: StringNetworkAddress,
        gas_commission: U256,
    ) -> Result<tonic_codegen::BridgeInSignatureResponse, BridgeError> {
        let client = self.client.clone();
        let request = tonic_codegen::BridgeInSignatureWithNonceRequest {
            sender: sender.data().to_owned(),
            token: token.data().to_owned(),
            nonce,
            amount: amount.to_string(),
            destination: Some(destination),
            gas_commission: gas_commission.to_string(),
        };

        let response = crate::grpc::retry_request(
            move || {
                let mut client = client.clone();
                let request = request.clone();
                async move { client.bridge_in_signature(request).await }
            },
            |error| warn!(error=%error, "error when trying to get bridge in signature, retrying"),
        )
        .await
        .map_err(|error| anyhow::anyhow!("failed to send request with: {error}"))?
        .into_inner();

        Ok(response)
    }

    async fn cancel_signature(
        &self,
        token: Address,
        recipient: Address,
        nonce: u64,
        comission: U256,
        amount: U256,
    ) -> Result<tonic_codegen::CancelTransferResponse, BridgeError> {
        let client = self.client.clone();
        let request = tonic_codegen::CancelSignatureRequest {
            token: token.data().to_owned(),
            recipient: recipient.data().to_owned(),
            nonce,
            commission: comission.to_string(),
            amount: amount.to_string(),
        };

        let CancelSignatureResponse { signature } = crate::grpc::retry_request(
            move || {
                let mut client = client.clone();
                let request = request.clone();
                async move { client.cancel_signature(request).await }
            },
            |error| warn!(error=%error, "error when trying to get cancel signature, retrying"),
        )
        .await
        .map_err(|error| anyhow::anyhow!("failed to send request with: {error}"))?
        .into_inner();

        Ok(tonic_codegen::CancelTransferResponse {
            status: "Success".to_string(),
            nonce,
            signature,
            token: token.data().to_owned(),
            recipient: recipient.data().to_owned(),
            commission: comission.to_string(),
            amount: amount.to_string(),
        })
    }

    async fn estimate_transfer(
        &self,
        amount: U256,
        recipient_network: String,
    ) -> Result<tonic_codegen::EstimateTransferResponse, BridgeError> {
        let client = self.client.clone();
        // Connector don't really care about sender network and token_id
        let request = tonic_codegen::EstimateTransferRequest {
            amount: amount.to_string(),
            recipient_network,
            ..Default::default()
        };

        let response = crate::grpc::retry_request(
            move || {
                let mut client = client.clone();
                let request = request.clone();
                async move { client.estimate_transfer(request).await }
            },
            |error| warn!(error=%error, "error when trying to get estimated transfer, retrying"),
        )
        .await
        .map_err(|error| anyhow::anyhow!("failed to send request with: {error}"))?
        .into_inner();

        Ok(response)
    }
}
