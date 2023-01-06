use std::{collections::HashMap, net::SocketAddr, str::FromStr, sync::Arc};

use async_trait::async_trait;
use primitive_types::U256;
use reqwest::Url;
use tokio::sync::oneshot;
use tonic::{Request, Response, Status};
use tonic_codegen::{
    connected_networks_response::Network,
    connector_bridge_server::ConnectorBridgeServer,
    gateway_bridge_server::{GatewayBridge, GatewayBridgeServer},
    *,
};

use futures::FutureExt;

use crate::{
    bridge::Bridge,
    crypto::verify_auth_signature,
    db::{rows::TransferDetails, BridgeReadQueries, BridgeWriteQueries},
    grpc::signer::SignerProxy,
    types::{NetworkId, TokenId, TxHash},
};

struct GatewayBridgeImpl {
    bridge: Arc<Bridge>,
}

pub async fn start(
    addr: SocketAddr,
    signer_url: Url,
    bridge: Arc<Bridge>,
    termination: oneshot::Receiver<()>,
) -> anyhow::Result<()> {
    let span = tracing::info_span!("gprc_server_start");

    {
        let span = span.clone();
        tonic::transport::Server::builder()
            .trace_fn(move |request| {
                let span =
                    tracing::info_span!(parent: &span, "grpc_server_request", request = ?request);
                tracing::info!(parent: &span, "intercepted inbound request");
                span
            })
            .add_service(GatewayBridgeServer::new(GatewayBridgeImpl { bridge }))
            .add_service(ConnectorBridgeServer::new(
                SignerProxy::new(signer_url).await?,
            ))
            .serve_with_shutdown(addr, termination.map(drop))
            .await?;
    }

    tracing::info!(parent: &span, "grpc server terminated");

    Ok(())
}

impl GatewayBridgeImpl {
    /// Maps data from the database into a `tonic_codegen::transfer_response::Transfer`.
    fn map_transfer(
        &self,
        transfer: crate::db::rows::TransferWithHashes,
    ) -> Result<tonic_codegen::transfer_response::Transfer, Status> {
        let amount = U256::from_little_endian(&transfer.amount);
        let sender_address = self
            .bridge
            .stringify_address(&crate::types::Address::new(
                NetworkId::new(transfer.sender_network_id as u32),
                transfer.sender_address,
            ))
            .map_err(|_| Status::internal("could not stringify address"))?;

        let recipient_address = self
            .bridge
            .stringify_address(&crate::types::Address::new(
                NetworkId::new(transfer.recipient_network_id as u32),
                transfer.recipient_address,
            ))
            .map_err(|_| Status::internal("could not stringify address"))?;

        let triggering_tx = self
            .bridge
            .stringify_tx_hash(&TxHash::new(
                NetworkId::new(transfer.triggering_tx_nid as u32),
                transfer.triggering_tx_hash,
            ))
            .map_err(|_| Status::internal("could not stringify hash"))?;

        let outbound_tx = transfer
            .outbound_tx_nid
            .and_then(|nid| transfer.outbound_tx_hash.map(|hash| (nid, hash)))
            .map(|(nid, hash)| TxHash::new(NetworkId::new(nid as u32), hash))
            .map(|hash| self.bridge.stringify_tx_hash(&hash))
            .transpose()
            .map_err(|_| Status::internal("could not stringify hash"))?;

        let created_at = prost_types::Timestamp {
            seconds: transfer.seen_at.timestamp(),
            nanos: transfer.seen_at.timestamp_subsec_nanos() as i32,
        };

        let mut response = transfer_response::Transfer {
            id: transfer.id as u64,
            amount: format!("{amount}"),
            sender: Some((&sender_address).into()),
            recipient: Some((&recipient_address).into()),
            status: 0,
            triggering_tx: Some((&triggering_tx).into()),
            outbound_tx: outbound_tx.map(|tx| (&tx).into()),
            created_at: Some(created_at),
        };

        let status = crate::types::TransferStatus::from_str(&transfer.status)
            .expect("invalid status message");
        response.set_status(status.into());

        Result::<_, Status>::Ok(response)
    }
}

#[async_trait]
impl GatewayBridge for GatewayBridgeImpl {
    /// Returns the list of connected networks to the bridge.
    async fn connected_networks(
        &self,
        _request: Request<()>,
    ) -> Result<Response<ConnectedNetworksResponse>, Status> {
        let connected_networks = {
            let networks = self.bridge.network_registry().read();

            networks
                .all()
                .map(|network| {
                    let mut connected_network = Network {
                        id: network.id().value(),
                        name: network.name().to_string(),
                        r#type: 0,
                        is_testnet: network.is_testnet(),
                    };
                    connected_network.set_type(network.ty().into());
                    connected_network
                })
                .collect::<Vec<_>>()
        };

        let response = ConnectedNetworksResponse {
            networks: connected_networks,
        };

        Ok(Response::new(response))
    }

    /// Returns list of supported tokens for a given network.
    async fn supported_tokens(
        &self,
        request: Request<SupportedTokensRequest>,
    ) -> Result<Response<TokensResponse>, Status> {
        let network_id = NetworkId::new(request.into_inner().network_id);

        let (tokens, network_tokens) = {
            let tokens = self.bridge.token_registry().read();
            let network_tokens = tokens
                .all_token_networks()
                .filter_map(|(tid, nid, meta)| {
                    if *nid == network_id {
                        Some((*tid, meta.clone()))
                    } else {
                        None
                    }
                })
                .collect::<HashMap<_, _>>();

            let tokens = tokens
                .all_tokens()
                .filter_map(|meta| {
                    if network_tokens.contains_key(&meta.id()) {
                        Some((meta.id(), meta.clone()))
                    } else {
                        None
                    }
                })
                .collect::<HashMap<_, _>>();

            (tokens, network_tokens)
        };

        let tokens = tokens
            .into_iter()
            .filter_map(|(id, token)| {
                network_tokens
                    .get(&id)
                    .map(|network_token| (id, token, network_token.clone()))
            })
            .map(|(id, token, network_token)| {
                self.bridge
                    .stringify_address(network_token.contract())
                    .map(|string_address| tonic_codegen::tokens_response::Token {
                        id: id.value(),
                        short_name: token.short_name().to_string(),
                        long_name: token.long_name().to_string(),
                        addresses: vec![tonic_codegen::tokens_response::TokenAddress {
                            network_id: network_id.value(),
                            address: string_address.address().to_string(),
                            decimals: network_token.decimals() as u32,
                        }],
                    })
            })
            .collect::<Result<Vec<_>, _>>()
            .map_err(|_| Status::internal("could not stringify address"))?;

        Ok(Response::new(TokensResponse { tokens }))
    }

    /// Returns the comission fee for a given transfer.
    async fn estimate_transfer(
        &self,
        request: Request<EstimateTransferRequest>,
    ) -> Result<Response<EstimateTransferResponse>, Status> {
        let EstimateTransferRequest {
            sender_network,
            recipient_network,
            token_id: _,
            amount: _,
        } = request.into_inner();
        let (recipient_network_id, sender_network_id) = {
            let network_registry = self.bridge.network_registry().read();
            let recipient_network_id = network_registry
                .by_name(&recipient_network)
                .map_err(|_| Status::invalid_argument("invalid recipient network name"))?;
            let sender_network_id = network_registry
                .by_name(&sender_network)
                .map_err(|_| Status::invalid_argument("invalid sender network name"))?;
            (recipient_network_id.id(), sender_network_id.id())
        };

        let (recipient_connector, sender_connector) = {
            let connector_registry = self.bridge.connectors().read();
            let recipient_connector = connector_registry
                .get(&recipient_network_id)
                .ok_or_else(|| Status::invalid_argument("invalid recipient network id"))?;
            let sender_connector = connector_registry
                .get(&sender_network_id)
                .ok_or_else(|| Status::invalid_argument("invalid sender network id"))?;
            (recipient_connector.clone(), sender_connector.clone())
        };

        // TODO: for now, we probably don't care about amount for fee estimation.
        let recipient_network_fee = recipient_connector
            .estimate_transfer(U256::zero(), recipient_network)
            .await
            .map_err(|_| Status::internal("could not estimate recipient network fee"))?;

        let sender_network_fee = sender_connector
            .estimate_transfer(U256::zero(), sender_network)
            .await
            .map_err(|_| Status::internal("could not estimate sender network fee"))?;

        Ok(Response::new(EstimateTransferResponse {
            fee: recipient_network_fee.fee,
            fee_percentage: sender_network_fee.fee_percentage,
            estimated_confirmation: 60,
        }))
    }

    /// Returns transfer status for the given transaction hash.
    async fn transfer(
        &self,
        request: Request<TransferRequest>,
    ) -> Result<Response<TransferResponse>, Status> {
        let string_tx_hash = request
            .into_inner()
            .tx_hash
            .ok_or_else(|| Status::invalid_argument("tx_hash field must be present"))?
            .into();

        let tx_hash = self
            .bridge
            .parse_tx_hash(&string_tx_hash)
            .map_err(|err| Status::invalid_argument(format!("invalid hash format: {err}")))?;

        let transfers = self
            .bridge
            .db()
            .read_tx()
            .await
            .map_err(|_| Status::internal("database error".to_string()))?
            .find_transfers_by_hash(&tx_hash)
            .await
            .map_err(|_| Status::internal("database error".to_string()))?
            .into_iter()
            .map(|transfer| self.map_transfer(transfer))
            .collect::<Result<Vec<_>, _>>()?;

        Ok(Response::new(TransferResponse {
            statuses: transfers,
        }))
    }

    /// Cancels a transfer id for a given network. It requires a signature from the user as an authorization.
    async fn cancel_transfer(
        &self,
        request: Request<CancelTransferRequest>,
    ) -> Result<Response<CancelTransferResponse>, Status> {
        let CancelTransferRequest {
            transfer_id,
            signature,
            network_id,
            public_key,
        } = request.into_inner();

        let network_id = NetworkId::new(network_id);

        let network_ty = self
            .bridge
            .network_registry()
            .read()
            .by_id(network_id)
            .map_err(|_| Status::invalid_argument("no such network id"))?
            .ty();

        let address = verify_auth_signature(network_ty, &signature, public_key.as_deref())
            .map_err(|err| {
                Status::invalid_argument(format!("signature verification failed: {err}"))
            })?;

        let TransferDetails {
            sender_address,
            amount,
            token_id,
        } = self
            .bridge
            .db()
            .read_tx()
            .await
            .map_err(|err| Status::internal(format!("db error: {err}")))?
            .find_transfer_details_by_transfer_id(transfer_id as i64)
            .await
            .map_err(|_| Status::invalid_argument("couldn't find transfer"))?
            .ok_or_else(|| Status::invalid_argument("couldn't find transfer"))?;

        if address != sender_address {
            return Err(Status::invalid_argument(
                "sender doesn't match to signature address",
            ));
        }

        let token = self
            .bridge
            .token_registry()
            .read()
            .token_network_by_ids(TokenId::new(token_id as u32), network_id)
            .map_err(|_| Status::invalid_argument("no such token id"))?
            .clone();

        let amount = U256::from_little_endian(&amount);
        let comission = amount * 4 / 1000;
        let amount = amount - comission;

        let result = self.bridge.cancel_transfer(transfer_id).await; // Send a signal to the bridge to cancel the transfer
        match result {
            Ok(()) => {
                let connector = {
                    let connectors = self.bridge.connectors().read();
                    connectors
                        .get(&network_id)
                        .ok_or_else(|| Status::invalid_argument("no such network id"))?
                        .clone()
                };

                let recipient = crate::types::Address::new(network_id, address);
                let mut dtx = self
                    .bridge
                    .db()
                    .write_tx()
                    .await
                    .map_err(|_| Status::internal("db error"))?;
                let nonce = dtx
                    .increment_nonce(network_id)
                    .await
                    .map_err(|_| Status::internal("db error"))?;

                // TODO: calculate gas comission once we have a way to estimate it
                let response = connector
                    .cancel_signature(
                        token.contract().to_owned(),
                        recipient,
                        nonce,
                        comission,
                        amount,
                    )
                    .await
                    .map_err(|_| Status::internal("could not generate cancel signature"))?;

                dtx.update_transfer_status(transfer_id, crate::types::TransferStatus::Cancelled)
                    .await
                    .map_err(|_| Status::internal("could not update transfer status"))?;
                dtx.commit()
                    .await
                    .map_err(|_| Status::internal("could not commit changes to database"))?;

                return Ok(Response::new(response));
            }
            _ => Err(Status::deadline_exceeded("cannot cancel")),
        }
    }

    /// Returns the transfer history for a given user. It requires a signature from the user as an authorization.
    async fn transfer_history(
        &self,
        request: Request<TransferHistoryRequest>,
    ) -> Result<Response<TransferHistoryResponse>, Status> {
        let tonic_codegen::TransferHistoryRequest {
            offset,
            limit,
            user_signature,
            network_id,
            public_key,
        } = request.into_inner();

        let network_id = NetworkId::new(network_id);

        let network_ty = self
            .bridge
            .network_registry()
            .read()
            .by_id(network_id)
            .map_err(|_| Status::invalid_argument("no such network id"))?
            .ty();

        let address = verify_auth_signature(network_ty, &user_signature, public_key.as_deref())
            .map_err(|err| {
                Status::invalid_argument(format!("signature verification failed: {err}"))
            })?;

        let address = crate::types::Address::new(network_id, address);

        let mut dtx = self
            .bridge
            .db()
            .read_tx()
            .await
            .map_err(|_| Status::internal("database error".to_string()))?;

        let total_size = dtx
            .count_transfer_for_sender(&address)
            .await
            .map_err(|_| Status::internal("database error".to_string()))?;

        let transfers = dtx
            .find_transfers_by_sender_paged(&address, limit, offset)
            .await
            .map_err(|_| Status::internal("database error".to_string()))?
            .into_iter()
            .map(|transfer| self.map_transfer(transfer))
            .collect::<Result<Vec<_>, _>>()?;

        Ok(Response::new(TransferHistoryResponse {
            statuses: transfers,
            total_size: total_size as u64,
        }))
    }

    /// Return signature for user to send bridgeIn transaction.
    async fn bridge_in_signature(
        &self,
        request: tonic::Request<tonic_codegen::BridgeInSignatureRequest>,
    ) -> Result<tonic::Response<tonic_codegen::BridgeInSignatureResponse>, tonic::Status> {
        let tonic_codegen::BridgeInSignatureRequest {
            sender,
            token_id,
            amount,
            destination,
        } = request.into_inner();
        let sender = sender.ok_or(tonic::Status::invalid_argument("sender is required"))?;
        let destination =
            destination.ok_or(tonic::Status::invalid_argument("destination is required"))?;

        let sender = self
            .bridge
            .parse_address(&sender.into())
            .map_err(|_| tonic::Status::invalid_argument("sender is invalid"))?;
        let token = {
            let registry = self.bridge.token_registry().read();
            registry
                .token_network_by_ids(TokenId::new(token_id), sender.network_id())
                .map_err(|_| tonic::Status::invalid_argument("no such token id"))?
                .clone()
        };
        let amount = U256::from_dec_str(&amount)
            .map_err(|_| tonic::Status::invalid_argument("amount is invalid"))?;

        let connector = {
            let connectors = self.bridge.connectors().read();

            connectors
                .get(&sender.network_id())
                .ok_or_else(|| {
                    tonic::Status::invalid_argument("connector is missing for the given network id")
                })?
                .clone()
        };

        let mut dbx = self
            .bridge
            .db()
            .write_tx()
            .await
            .map_err(|_| tonic::Status::internal("cannot get db"))?;
        let nonce = dbx
            .increment_nonce(sender.network_id())
            .await
            .map_err(|_| tonic::Status::internal("cannot get nonce"))?;
        dbx.commit()
            .await
            .map_err(|_| tonic::Status::internal("cannot commit db"))?;
        // TODO: handle gas comission properly
        let response = connector
            .bridge_in_signature(
                sender,
                token.contract().to_owned(),
                nonce,
                amount,
                destination,
                U256::zero(),
            )
            .await
            .map_err(|_| tonic::Status::internal("cannot get signature"))?;
        Ok(tonic::Response::new(response))
    }
}
