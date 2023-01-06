use std::{
    collections::HashMap,
    sync::{
        atomic::{AtomicBool, Ordering},
        Arc,
    },
    time::Duration,
};

use anyhow::{anyhow, Context};
use futures::channel::oneshot;
use parking_lot::{Mutex, RwLock};
use primitive_types::U256;
use serde::Deserialize;
use slab::Slab;
use tokio::{
    sync::mpsc::{channel, Receiver, Sender},
    task::JoinHandle,
};
use tracing::{info, instrument, Instrument};

use crate::{
    db::{self, rows::Transfer, BridgeReadQueries},
    error::BridgeError,
};
use crate::{
    db::{BridgeWriteQueries, Database},
    NetworkConnector,
};
use crate::{
    math::Decimal,
    types::{BridgeEvent, BridgeTokenTransferIn, NetworkId},
};
use crate::{
    registry::{NetworkRegistry, TokenMetadata, TokenNetworkMetadata, TokenRegistry},
    types::Address,
};
use crate::{
    types::{BridgeTokenTransferOut, ConfirmedTx, StringAddress, StringTxHash, TokenId, TxHash},
    TimeSource,
};

#[derive(Deserialize, Clone, Default)]
pub struct Config {
    pub tx_pending_time: Option<u32>,
}

impl Config {
    pub fn from_env() -> Result<Self, anyhow::Error> {
        envy::prefixed("BRIDGE_")
            .from_env::<Config>()
            .context("could not load bridge config")
    }

    pub fn tx_pending_time(&self) -> Duration {
        let time = self.tx_pending_time.unwrap_or(10);
        Duration::from_secs(time as u64)
    }
}

#[derive(Clone)]
pub struct Bridge {
    inner: Arc<BridgeInner>,
}

struct BridgeInner {
    config: Config,
    network_registry: RwLock<NetworkRegistry>,
    token_registry: RwLock<TokenRegistry>,

    connectors: RwLock<HashMap<NetworkId, Arc<dyn NetworkConnector>>>,
    transfer_cancel_handles: RwLock<HashMap<u64, oneshot::Sender<()>>>,
    event_tx: Sender<BridgeEvent>,
    db: Database,
    time_source: Box<dyn TimeSource>,

    is_shutting_down: AtomicBool,
    active_transfers_tasks: RwLock<Slab<oneshot::Receiver<()>>>,
    event_loop_task: Mutex<Option<JoinHandle<()>>>,
}

impl Bridge {
    /// Starts the bridge service using provided database config, bridge config and time source
    #[instrument(skip(time_source, bridge_config))]
    pub async fn start(
        db_config: db::Config,
        bridge_config: Config,
        time_source: Box<dyn TimeSource>,
    ) -> Result<Bridge, BridgeError> {
        let network_registry = RwLock::new(NetworkRegistry::default());
        let token_registry = RwLock::new(TokenRegistry::default());
        let connectors = RwLock::new(HashMap::new());

        let (event_tx, event_rx) = channel(256);
        let db = db::Database::connect(db_config).await?;
        let transfer_cancel_handles = Default::default();
        let active_transfers_tasks = Default::default();
        let event_loop_task = Default::default();

        let bridge = BridgeInner {
            config: bridge_config,
            network_registry,
            token_registry,
            event_tx,
            connectors,
            db,
            transfer_cancel_handles,
            time_source,

            is_shutting_down: AtomicBool::new(false),
            active_transfers_tasks,
            event_loop_task,
        };

        let bridge = Bridge {
            inner: Arc::new(bridge),
        };

        let handle = {
            let bridge = bridge.clone();
            tokio::spawn(bridge.event_loop(event_rx))
        };

        *bridge.inner.event_loop_task.lock() = Some(handle);

        Ok(bridge)
    }

    /// The function stops event loop processing and waits until tasks finalization.
    pub async fn shutdown(&self) {
        info!("Starting shut down");
        self.inner.is_shutting_down.swap(true, Ordering::SeqCst);

        info!("Waiting for event loop to shutdown");
        let event_loop_handle = self
            .inner
            .event_loop_task
            .lock()
            .take()
            .expect("expected join handle");

        event_loop_handle.await.ok();

        let tasks = self
            .inner
            .active_transfers_tasks
            .write()
            .drain()
            .collect::<Vec<_>>();

        info!("Waiting for tasks to shutdown");
        for task in tasks {
            if let Err(err) = task.await {
                tracing::warn!("Couldn't finish task during shutdown: {err}")
            }
        }
    }

    pub fn event_tx(&self) -> &Sender<BridgeEvent> {
        &self.inner.event_tx
    }

    pub fn network_registry(&self) -> &RwLock<NetworkRegistry> {
        &self.inner.network_registry
    }

    pub fn token_registry(&self) -> &RwLock<TokenRegistry> {
        &self.inner.token_registry
    }

    pub fn connectors(&self) -> &RwLock<HashMap<NetworkId, Arc<dyn NetworkConnector>>> {
        &self.inner.connectors
    }

    pub fn db(&self) -> &Database {
        &self.inner.db
    }

    /// Loads supported token from the database to the token and network registries.
    #[instrument(skip(self), err)]
    pub async fn load_tokens(&self) -> Result<(), BridgeError> {
        let mut dtx = self.db().read_tx().await?;

        let tokens = dtx.all_tokens().await.context("couldn't load tokens")?;

        let network_tokens = dtx
            .all_network_tokens()
            .await
            .context("couldn't load network tokens")?;

        let mut token_registry = self.token_registry().write();

        for token in tokens {
            let metadata = TokenMetadata::new(
                TokenId::new(token.id as u32),
                token.short_name,
                token.long_name,
            );
            token_registry.register(metadata);
        }

        for network_token in network_tokens {
            let contract = Address::new(
                NetworkId::new(network_token.network_id as u32),
                network_token.contract_key,
            );
            let metadata = TokenNetworkMetadata::new(contract, network_token.decimals as u8);
            let token_id = TokenId::new(network_token.token_id as u32);
            token_registry.register_token_network(token_id, metadata)?;
        }

        Ok(())
    }

    /// Loads last seen block for given network from the database.
    #[instrument(skip(self), err)]
    pub async fn last_seen_network_block(
        &self,
        network_id: NetworkId,
    ) -> Result<Option<u64>, BridgeError> {
        let mut dtx = self.db().read_tx().await?;

        Ok(dtx.last_seen_network_block(network_id).await?)
    }

    /// Update last seen block for given network in the database
    #[instrument(skip(self), err)]
    pub async fn update_last_seen_network_block(
        &self,
        network_id: NetworkId,
        block: u64,
    ) -> Result<(), BridgeError> {
        let mut dtx = self.db().write_tx().await?;

        dtx.update_seen_network_block(network_id, block).await?;
        dtx.commit().await?;

        Ok(())
    }

    /// Registers connector in the bridge.
    /// Registering is mandatory for connector to give bridge a way to communicate with it.
    #[instrument(skip(self, connector), err)]
    pub async fn register_connector(
        &self,
        connector: Arc<dyn NetworkConnector>,
    ) -> Result<(), BridgeError> {
        let metadata = connector.metadata();
        tracing::debug!("registering connector: {:?}", &metadata);

        let network_id = metadata.id();
        let mut registry = self.inner.network_registry.write();
        let mut connectors = self.inner.connectors.write();
        registry.register(metadata);
        connectors.insert(network_id, connector);

        Ok(())
    }

    /// Cancels transfer with given id. Currently, we suppoprt cancels only in the WAITING state.
    #[instrument(skip(self), err)]
    pub async fn cancel_transfer(&self, transfer_id: u64) -> Result<(), BridgeError> {
        let cancel_tx = self
            .inner
            .transfer_cancel_handles
            .write()
            .remove(&transfer_id)
            .ok_or_else(|| {
                anyhow!("unknown transfer id {transfer_id}, or it already sent/finished")
            })?;

        cancel_tx
            .send(())
            .ok()
            .context("too late to cancel transfer")?;

        Ok(())
    }

    /// Parses address from string representation.
    pub fn parse_address(&self, address: &StringAddress) -> Result<Address, BridgeError> {
        Ok(self.network_registry().read().parse_address(address)?)
    }

    /// Stringifies address to string representation.
    pub fn stringify_address(&self, address: &Address) -> Result<StringAddress, BridgeError> {
        Ok(self.network_registry().read().stringify_address(address)?)
    }

    /// Parse tx hash from string representation.
    pub fn parse_tx_hash(&self, hash: &StringTxHash) -> Result<TxHash, BridgeError> {
        Ok(self.network_registry().read().parse_tx_hash(hash)?)
    }

    /// Stringifies tx hash to string representation.
    pub fn stringify_tx_hash(&self, hash: &TxHash) -> Result<StringTxHash, BridgeError> {
        Ok(self.network_registry().read().stringify_tx_hash(hash)?)
    }

    /// Processes events from registered connectors.
    /// This method is blocking and should be run in a separate thread.
    /// It will return when bridge is shutting down.
    /// The event loop restarts old events processing on startup.
    #[instrument(skip(self, event_rx))]
    async fn event_loop(self, mut event_rx: Receiver<BridgeEvent>) {
        info!("processing old events");
        loop {
            if self.inner.is_shutting_down.load(Ordering::SeqCst) {
                return;
            }

            if let Err(err) = self.restore_processing().await {
                tracing::warn!("Couldn't restore processing, retrying after few seconds. {err}");
                tokio::time::sleep(crate::consts::RETRY_TIMEOUT).await;
            } else {
                break;
            }
        }

        info!("starting bridge event loop");
        loop {
            if self.inner.is_shutting_down.load(Ordering::SeqCst) {
                break;
            }

            match event_rx.try_recv() {
                Ok(BridgeEvent::TokenTransferIn(transfer)) => {
                    let bridge = self.clone();

                    let span = tracing::info_span!("handle_transfer_in_event", from = %transfer.from, to = %transfer.to);
                    let mut transfers = self.inner.active_transfers_tasks.write();
                    let id = transfers.vacant_key();
                    let (tx, rx) = oneshot::channel();
                    tokio::spawn(
                        async move {
                            if let Err(error) = bridge.handle_transfer_in_event(transfer).await {
                                tracing::error!("{:#}", anyhow::Error::new(error));
                            }

                            tx.send(()).ok();
                            bridge.inner.active_transfers_tasks.write().try_remove(id);
                        }
                        .instrument(span),
                    );
                    transfers.insert(rx);
                }

                Ok(BridgeEvent::TokenTransferOut(transfer)) => {
                    let bridge = self.clone();

                    let span = tracing::info_span!("handle_transfer_out_event", from = %transfer.from, to = %transfer.to);
                    let mut transfers = self.inner.active_transfers_tasks.write();
                    let id = transfers.vacant_key();
                    let (tx, rx) = oneshot::channel();
                    tokio::spawn(
                        async move {
                            if let Err(error) = bridge.handle_transfer_out_event(transfer).await {
                                tracing::error!("{:#}", anyhow::Error::new(error));
                            }

                            tx.send(()).ok();
                            bridge.inner.active_transfers_tasks.write().try_remove(id);
                        }
                        .instrument(span),
                    );
                    transfers.insert(rx);
                }
                _ => {
                    tracing::trace!("No events. Sleeping...");
                    tokio::time::sleep(Duration::from_millis(100)).await
                }
            }
        }
        info!("terminated bridge event loop");
    }

    /// Insert transaction into database if it doesn't exist. Otherwise, return error.
    async fn insert_tx_if_not_exists(&self, tx: &ConfirmedTx) -> Result<u64, BridgeError> {
        let mut dtx = self.db().write_tx().await?;

        if let Some(transaction_id) = dtx.find_transaction_by_hash(tx.hash()).await? {
            Err(BridgeError::Connector(
                crate::error::ConnectorError::EventDuplicate {
                    block_number: tx.block_number(),
                    transaction_id: transaction_id.id,
                },
            ))
        } else {
            let tx_id = dtx
                .insert_transaction(
                    tx.hash(),
                    tx.block_number(),
                    self.inner.time_source.now(),
                    tx.sender(),
                )
                .await
                .context("couldn't insert transaction")?;

            dtx.commit().await.context("couldn't commit tx")?;

            Ok(tx_id)
        }
    }

    /// Processes transfer out event. Matches transaction by event data.
    /// Currently, it's hard to distinguish betweent equal transaction from the same user, so it would finalize random-one
    async fn handle_transfer_out_event(
        &self,
        event: BridgeTokenTransferOut,
    ) -> Result<(), BridgeError> {
        let BridgeTokenTransferOut {
            from,
            to,
            amount,
            token,
            tx,
        } = event;

        info!(from = %from, to = %to, "received token transfer out");

        let token_id = {
            let tokens = self.token_registry().read();
            tokens.token_by_address(&token)?.id()
        };
        let from = self.parse_address(&from)?;

        let tx_id = self.insert_tx_if_not_exists(&tx).await?;

        let mut dtx = self.db().write_tx().await?;
        // Find transaction and update status and outbound tx
        dtx.finalize_transfer(from, to, amount, token_id, tx_id)
            .await?;

        dtx.commit().await.context("couldn't commit tx")?;

        Ok(())
    }

    /// Processes transfer in event.
    /// Inserts transaction into database and sends it to the destination connector after waiting for delay period.
    async fn handle_transfer_in_event(
        &self,
        event: BridgeTokenTransferIn,
    ) -> Result<(), BridgeError> {
        let BridgeTokenTransferIn {
            from,
            to,
            amount,
            token,
            tx,
        } = event;

        info!(from = %from, to = %to, amount = %amount, "received token transfer in");

        let to_name = to.network_name();

        let to_metadata = {
            let networks = self.network_registry().read();
            networks.by_name(to_name)?.clone()
        };

        let to_connector = {
            let connectors = self.connectors().read();

            connectors
                .get(&to_metadata.id())
                .with_context(|| format!("unknown destination network id {}", to_metadata.id()))?
                .clone()
        };

        let (token_id, from_decimals, to_decimals, to_token) = {
            let tokens = self.token_registry().read();
            let token_id = tokens.token_by_address(&token)?.id();

            let from_decimals = tokens
                .token_network_by_ids(token_id, from.network_id())?
                .decimals();

            let to_token = tokens.token_network_by_ids(token_id, to_metadata.id())?;

            let to_decimals = to_token.decimals();
            let to_token = to_token.contract().clone();

            (token_id, from_decimals, to_decimals, to_token)
        };

        info!(amount = %amount, from_decimals = %from_decimals, to_decimals = %to_decimals, "converting amount to other network decimals");

        let amount = Decimal::from_raw_with_scale(amount, from_decimals)
            .with_context(|| {
                format!("couldn't convert amount ({amount}) to decimals ({from_decimals})")
            })?
            .to_raw_with_scale(to_decimals)
            .with_context(|| format!("couldn't decimal ({amount}) to decimals ({to_decimals})"))?;

        info!(amount = %amount, "converted amount to other network decimals");

        let to_address = self.parse_address(&to)?;

        let tx_id = self.insert_tx_if_not_exists(&tx).await?;

        let mut dtx = self.db().write_tx().await?;
        let transfer_id = dtx
            .insert_transfer(tx_id, token_id, amount, &from, &to_address)
            .await
            .context("couldn't insert transfer")?;
        dtx.commit().await?;

        let from_string_address = self.stringify_address(&from)?;

        self.process_transfer(
            transfer_id,
            to_connector,
            to_address,
            to_token,
            amount,
            from_string_address,
            self.inner.config.tx_pending_time(),
        )
        .await
    }

    /// Restore processing transfers that were missed in waiting state (Calling only on start up)
    pub async fn restore_processing(&self) -> Result<(), BridgeError> {
        let transfers_to_restore = self
            .db()
            .read_tx()
            .await?
            .get_transactions_in_waiting()
            .await?;

        for transfer in transfers_to_restore {
            let Transfer {
                id,
                token_id,
                amount,
                sender_network_id,
                sender_address,
                recipient_network_id,
                recipient_address,
                seen_at,
            } = transfer;

            let recipient = Address::new(
                NetworkId::new(recipient_network_id as u32),
                recipient_address,
            );

            let source_address =
                Address::new(NetworkId::new(sender_network_id as u32), sender_address);
            let source_address = self.stringify_address(&source_address)?;

            let to_connector = {
                let connectors = self.connectors().read();

                connectors
                    .get(&recipient.network_id())
                    .with_context(|| {
                        format!("unknown destination network id {recipient_network_id}")
                    })?
                    .clone()
            };

            let token_address = {
                let tokens = self.token_registry().read();
                tokens
                    .token_network_by_ids(TokenId::new(token_id as u32), recipient.network_id())?
                    .contract()
                    .clone()
            };

            // Maybe while we crashed passed enough time, so we calculate this value.
            let sleep_time = self.inner.config.tx_pending_time();
            let passed_time = (self.inner.time_source.now() - seen_at)
                .to_std()
                .context("couldn't process time difference")?;

            let duration = if passed_time >= sleep_time {
                Duration::ZERO
            } else {
                sleep_time - passed_time
            };

            self.process_transfer(
                id as u64,
                to_connector,
                recipient,
                token_address,
                U256::from_little_endian(&amount),
                source_address,
                duration,
            )
            .await?;
        }
        Ok(())
    }

    /// Process transfer by sending it to the recipient network after waiting for `time_await`.
    /// It's possible to cancel the transfer by calling `cancel_transfer` in the meantime.
    #[allow(clippy::too_many_arguments)]
    async fn process_transfer(
        &self,
        transfer_id: u64,
        to_connector: Arc<dyn NetworkConnector>,
        recipient: Address,
        token_address: Address,
        amount: U256,
        source_address: StringAddress,
        time_await: Duration,
    ) -> Result<(), BridgeError> {
        let (cancel_tx, mut cancel_rx) = oneshot::channel();
        self.inner
            .transfer_cancel_handles
            .write()
            .insert(transfer_id, cancel_tx);
        let timeout = self.inner.time_source.sleep(time_await);

        let mut cancelled = false;
        info!("Waiting for {time_await:?} time pass for {transfer_id}");
        tokio::select! {
            _ = &mut cancel_rx => {
                cancelled = true;
            }

            _ = timeout => {
                cancel_rx.close();

                if let Ok(Some(())) = cancel_rx.try_recv() {
                    cancelled = true;
                }
            }
        }
        self.inner
            .transfer_cancel_handles
            .write()
            .remove(&transfer_id);

        if cancelled {
            info!("Received cancelled signal for {transfer_id} transfer_id");
        } else {
            let mut dtx = self.db().write_tx().await?;

            to_connector
                .bridge_out(
                    recipient,
                    token_address,
                    amount,
                    source_address,
                    transfer_id,
                )
                .await
                .context("could not bridge out funds")?;

            dtx.update_transfer_status(transfer_id, crate::types::TransferStatus::Confirming)
                .await
                .context("could not update transfer status")?;

            dtx.commit().await?;
        }

        Ok(())
    }
}
