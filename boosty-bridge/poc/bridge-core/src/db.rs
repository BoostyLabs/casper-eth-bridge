use std::fmt::Debug;

use anyhow::Context;

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use primitive_types::U256;
use sea_query::{Cond, Expr, OnConflict, PostgresQueryBuilder, Query};
use serde::Deserialize;
use sqlx::{
    postgres::{PgConnectOptions, PgSslMode},
    Executor, FromRow, PgPool, Postgres,
};

use crate::{
    error::DbError,
    types::{Address, NetworkId, TokenId, TransferStatus, TxHash},
};

pub mod rows;
mod schema;

use sea_query_binder::SqlxBinder;

use schema::*;

use self::rows::{
    Id32, Id64, LastSeenBlock, NetworkToken, Token, Transaction, Transfer, TransferWithHashes,
};

#[derive(Deserialize, Clone)]
pub struct Config {
    pub host: String,
    pub port: u16,
    pub user: String,
    pub pass: String,
    pub dbname: String,
}

impl Config {
    pub fn from_env() -> Result<Self, anyhow::Error> {
        envy::prefixed("PG_")
            .from_env::<Config>()
            .context("could not load db config")
    }

    fn descriptor(&self) -> String {
        let Self {
            host,
            port,
            user,
            dbname,
            ..
        } = &self;

        format!("{user}@{host}:{port}/{dbname}")
    }
}

pub struct Database {
    pool: sqlx::PgPool,
    descriptor: String,
}

impl std::fmt::Debug for Database {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_tuple("Database").field(&self.descriptor).finish()
    }
}

impl std::fmt::Debug for Config {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_tuple("Config").field(&self.descriptor()).finish()
    }
}

/// A trait that defines the read-only queries that can be performed on the database.
#[async_trait]
pub trait BridgeReadQueries<'a>
where
    Self: Debug,
{
    fn tx(&mut self) -> &mut sqlx::Transaction<'a, Postgres>;

    #[tracing::instrument(err)]
    async fn find_transfer_details_by_transfer_id(
        &mut self,
        transfer_id: i64,
    ) -> Result<Option<rows::TransferDetails>, DbError> {
        let (query, values) = Query::select()
            .from(TokenTransfers::Table)
            .columns([
                TokenTransfers::SenderAddress,
                TokenTransfers::TokenId,
                TokenTransfers::Amount,
            ])
            .cond_where(Cond::all().add(Expr::col(TokenTransfers::Id).eq(transfer_id)))
            .build_sqlx(PostgresQueryBuilder);

        let details: Option<rows::TransferDetails> = sqlx::query_as_with(&query, values)
            .fetch_optional(self.tx())
            .await?;

        Ok(details)
    }

    /// Returns transaction by hash from the `transactions` table.
    #[tracing::instrument(err)]
    async fn find_transaction_by_hash(
        &mut self,
        hash: &TxHash,
    ) -> Result<Option<Transaction>, DbError> {
        let (query, values) = Query::select()
            .from(Transactions::Table)
            .columns([
                Transactions::Id,
                Transactions::NetworkId,
                Transactions::Txhash,
                Transactions::Blocknumber,
                Transactions::SeenAt,
            ])
            .cond_where(
                Cond::all()
                    .add(Expr::col(Transactions::NetworkId).eq(hash.network_id().value() as i32))
                    .add(Expr::col(Transactions::Txhash).eq(hash.data())),
            )
            .build_sqlx(PostgresQueryBuilder);

        let tx: Option<Transaction> = sqlx::query_as_with(&query, values)
            .fetch_optional(self.tx())
            .await?;

        Ok(tx)
    }

    /// Returns transfers data by hash from the `token_transfers` table joined with `transactions` table.
    #[tracing::instrument(err)]
    async fn find_transfers_by_hash(
        &mut self,
        hash: &TxHash,
    ) -> Result<Vec<TransferWithHashes>, DbError> {
        const QUERY: &str = include_str!("sql/select_transfers_by_txhash.sql");

        let transfers = sqlx::query_as(QUERY)
            .bind(hash.network_id().value() as i32)
            .bind(hash.data())
            .fetch_all(self.tx())
            .await?;

        Ok(transfers)
    }

    /// Returns transfers for given user address with pagination support from the `token_transfers` table joined with `transactions` table.
    #[tracing::instrument(err)]
    async fn find_transfers_by_sender_paged(
        &mut self,
        sender: &Address,
        limit: u64,
        offset: u64,
    ) -> Result<Vec<TransferWithHashes>, DbError> {
        const QUERY: &str = include_str!("sql/select_transfers_by_sender_paged.sql");

        let transfers = sqlx::query_as(QUERY)
            .bind(sender.network_id().value() as i32)
            .bind(sender.data())
            .bind(limit as i64)
            .bind(offset as i64)
            .fetch_all(self.tx())
            .await?;

        Ok(transfers)
    }

    /// Returns total number of transfers for given user address from the `token_transfers` table.
    #[tracing::instrument(err)]
    async fn count_transfer_for_sender(&mut self, sender: &Address) -> Result<i64, DbError> {
        const QUERY: &str = include_str!("sql/count_transfers_by_sender.sql");

        #[derive(FromRow)]
        struct Count {
            total_value: i64,
        }

        let count: Count = sqlx::query_as(QUERY)
            .bind(sender.network_id().value() as i32)
            .bind(sender.data())
            .fetch_one(self.tx())
            .await?;

        Ok(count.total_value)
    }

    /// Returns all supported tokens by the bridge.
    /// This table contains token metadata.
    #[tracing::instrument(err)]
    async fn all_tokens(&mut self) -> Result<Vec<Token>, DbError> {
        let (query, values) = Query::select()
            .columns([Tokens::Id, Tokens::ShortName, Tokens::LongName])
            .from(Tokens::Table)
            .build_sqlx(PostgresQueryBuilder);

        let result: Vec<Token> = sqlx::query_as_with(&query, values)
            .fetch_all(self.tx())
            .await?;

        Ok(result)
    }

    /// Returns tokens network specific metadata for all supported tokens from the `network_tokens` table.
    /// This table contains network specific token metadata.
    #[tracing::instrument(err)]
    async fn all_network_tokens(&mut self) -> Result<Vec<NetworkToken>, DbError> {
        let (query, values) = Query::select()
            .columns([
                NetworkTokens::NetworkId,
                NetworkTokens::TokenId,
                NetworkTokens::ContractKey,
                NetworkTokens::Decimals,
            ])
            .from(NetworkTokens::Table)
            .build_sqlx(PostgresQueryBuilder);

        let result: Vec<NetworkToken> = sqlx::query_as_with(&query, values)
            .fetch_all(self.tx())
            .await?;

        Ok(result)
    }

    /// Returns last seen block number for given network from the `network_blocks` table.
    async fn last_seen_network_block(
        &mut self,
        network_id: NetworkId,
    ) -> Result<Option<u64>, DbError> {
        let (query, values) = Query::select()
            .columns([NetworkBlocks::LastSeenBlock])
            .from(NetworkBlocks::Table)
            .cond_where(Expr::col(NetworkBlocks::NetworkId).eq(network_id.value() as i32))
            .build_sqlx(PostgresQueryBuilder);

        let result: Option<LastSeenBlock> = sqlx::query_as_with(&query, values)
            .fetch_optional(self.tx())
            .await?;

        Ok(result.map(|row| row.last_seen_block as u64))
    }

    /// This function is purely designed for loading transactions in progress transactions after crash.
    /// The only case when we need to restore processing is in WAITING state, all other states should be processed by event system just fine.
    async fn get_transactions_in_waiting(&mut self) -> Result<Vec<Transfer>, DbError> {
        const QUERY: &str = include_str!("sql/transactions_in_waiting.sql");

        let result: Vec<Transfer> = sqlx::query_as(QUERY).fetch_all(self.tx()).await?;
        Ok(result)
    }
}

/// Write queries for the bridge.
#[async_trait]
pub trait BridgeWriteQueries<'a>: BridgeReadQueries<'a> {
    /// Creates all tables required for the bridge. This function should be called only once.
    /// If tables already exist, this function will return an error.
    #[tracing::instrument(err)]
    async fn create_tables(&mut self) -> Result<(), DbError> {
        const QUERY: &str = include_str!("./sql/create_tables.sql");

        for sub_query in QUERY.trim().split(';') {
            sqlx::query(sub_query).execute(self.tx()).await?;
        }

        Ok(())
    }

    /// Inserts new token header metadata into the `tokens` table.
    #[tracing::instrument(err)]
    async fn insert_token(
        &mut self,
        short_name: &str,
        long_name: &str,
    ) -> Result<TokenId, DbError> {
        let (query, values) = Query::insert()
            .into_table(Tokens::Table)
            .columns([Tokens::LongName, Tokens::ShortName])
            .values(vec![long_name.into(), short_name.into()])?
            .returning_col(Tokens::Id)
            .build_sqlx(PostgresQueryBuilder);

        let result: Id32 = sqlx::query_as_with(&query, values)
            .fetch_one(self.tx())
            .await?;

        let id = result.id as u32;
        Ok(TokenId::new(id))
    }

    /// Inserts new token network specific metadata into the `network_tokens` table.
    #[tracing::instrument(err)]
    async fn insert_network_token(
        &mut self,
        network_id: NetworkId,
        token_id: TokenId,
        contract_key: Address,
        decimals: u8,
    ) -> Result<(), DbError> {
        let (query, values) = Query::insert()
            .into_table(NetworkTokens::Table)
            .columns([
                NetworkTokens::NetworkId,
                NetworkTokens::TokenId,
                NetworkTokens::ContractKey,
                NetworkTokens::Decimals,
            ])
            .values(vec![
                (network_id.value() as i32).into(),
                (token_id.value() as i32).into(),
                contract_key.data().into(),
                (decimals as i16).into(),
            ])?
            .build_sqlx(PostgresQueryBuilder);

        sqlx::query_with(&query, values).execute(self.tx()).await?;

        Ok(())
    }

    /// Finalizes transfer by setting the status to `Finished` and updating the outgoing transaction hash.
    /// This function is used to finalize transfers that were initiated by the bridge early on.
    /// Please note that currently we can't distinguish between equivalent transfers, so we finalize only the first one.
    #[tracing::instrument(err)]
    async fn finalize_transfer(
        &mut self,
        from: Address,
        to: Address,
        amount: U256,
        token: TokenId,
        tx_id: u64,
    ) -> Result<(), DbError> {
        let mut amount_buf = [0u8; 32];
        amount.to_little_endian(&mut amount_buf);

        let status: &'static str = TransferStatus::Confirming.into();
        let result_status: &'static str = TransferStatus::Finished.into();

        let (query, values) = Query::select()
            .from(TokenTransfers::Table)
            .column(TokenTransfers::Id)
            .cond_where(
                Cond::all()
                    .add(Expr::col(TokenTransfers::SenderAddress).eq(from.data()))
                    .add(
                        Expr::col(TokenTransfers::SenderNetworkId)
                            .eq(from.network_id().value() as i32),
                    )
                    .add(Expr::col(TokenTransfers::RecipientAddress).eq(to.data()))
                    .add(
                        Expr::col(TokenTransfers::RecipientNetworkId)
                            .eq(to.network_id().value() as i32),
                    )
                    .add(
                        Expr::col(TokenTransfers::Amount).eq::<Vec<u8>>(amount_buf.as_ref().into()),
                    )
                    .add(Expr::col(TokenTransfers::TokenId).eq(token.value() as i32))
                    .add(Expr::col(TokenTransfers::Status).eq(status)),
            )
            .limit(1)
            .build_sqlx(PostgresQueryBuilder);

        #[derive(FromRow)]
        struct Id {
            id: i64,
        }

        let id: Id = sqlx::query_as_with(&query, values)
            .fetch_one(self.tx())
            .await?;

        let (query, values) = Query::update()
            .table(TokenTransfers::Table)
            .values(vec![
                (TokenTransfers::Status, result_status.into()),
                (TokenTransfers::OutboundTx, (tx_id as i32).into()),
            ])
            .cond_where(Cond::all().add(Expr::col(TokenTransfers::Id).eq(id.id)))
            .build_sqlx(PostgresQueryBuilder);

        sqlx::query_with(&query, values).execute(self.tx()).await?;

        Ok(())
    }

    /// Insert new transaction into the `transactions` table and return its id.
    #[tracing::instrument(err)]
    async fn insert_transaction(
        &mut self,
        hash: &TxHash,
        block_number: u64,
        seen_at: DateTime<Utc>,
        sender: &Address,
    ) -> Result<u64, DbError> {
        let (query, values) = Query::insert()
            .into_table(Transactions::Table)
            .columns([
                Transactions::NetworkId,
                Transactions::Txhash,
                Transactions::Blocknumber,
                Transactions::SeenAt,
                Transactions::Sender,
            ])
            .values(vec![
                (hash.network_id().value() as i32).into(),
                hash.data().into(),
                (block_number as i64).into(),
                seen_at.into(),
                sender.data().into(),
            ])?
            .returning_col(Transactions::Id)
            .build_sqlx(PostgresQueryBuilder);

        let result: Id64 = sqlx::query_as_with(&query, values)
            .fetch_one(self.tx())
            .await?;

        let id = result.id as u64;

        Ok(id)
    }

    /// Insert new transfer into the `transfers` table and return its id.
    #[tracing::instrument(err)]
    async fn insert_transfer(
        &mut self,
        tx_id: u64,
        token_id: TokenId,
        amount: U256,
        sender: &Address,
        recipient: &Address,
    ) -> Result<u64, DbError> {
        let status_string: &'static str = TransferStatus::Waiting.into();

        let mut amount_buf = [0u8; 32];
        amount.to_little_endian(&mut amount_buf);

        let (query, values) = Query::insert()
            .into_table(TokenTransfers::Table)
            .columns([
                TokenTransfers::TriggeringTx,
                TokenTransfers::TokenId,
                TokenTransfers::Amount,
                TokenTransfers::Status,
                TokenTransfers::SenderNetworkId,
                TokenTransfers::SenderAddress,
                TokenTransfers::RecipientNetworkId,
                TokenTransfers::RecipientAddress,
            ])
            .values(vec![
                (tx_id as i64).into(),
                (token_id.value() as i32).into(),
                amount_buf.as_ref().into(),
                status_string.into(),
                (sender.network_id().value() as i32).into(),
                sender.data().into(),
                (recipient.network_id().value() as i32).into(),
                recipient.data().into(),
            ])?
            .returning_col(TokenTransfers::Id)
            .build_sqlx(PostgresQueryBuilder);

        let result: Id64 = sqlx::query_as_with(&query, values)
            .fetch_one(self.tx())
            .await?;

        let id = result.id as u64;

        Ok(id)
    }

    /// Update transfer status for the given transfer id.
    #[tracing::instrument(err)]
    async fn update_transfer_status(
        &mut self,
        transfer_id: u64,
        status: TransferStatus,
    ) -> Result<(), DbError> {
        let status_string: &'static str = status.into();

        let (query, values) = Query::update()
            .table(TokenTransfers::Table)
            .values([(TokenTransfers::Status, status_string.into())])
            .cond_where(Expr::col(TokenTransfers::Id).eq(transfer_id as i64))
            .build_sqlx(PostgresQueryBuilder);

        sqlx::query_with(&query, values).execute(self.tx()).await?;

        Ok(())
    }

    /// Update last seen block for the given network in the `network_blocks` table.
    async fn update_seen_network_block(
        &mut self,
        network_id: NetworkId,
        block: u64,
    ) -> Result<(), DbError> {
        let (query, values) = Query::insert()
            .into_table(NetworkBlocks::Table)
            .columns([NetworkBlocks::NetworkId, NetworkBlocks::LastSeenBlock])
            .values([(network_id.value() as i32).into(), (block as i64).into()])?
            .on_conflict(
                OnConflict::column(NetworkBlocks::NetworkId)
                    .update_column(NetworkBlocks::LastSeenBlock)
                    .to_owned(),
            )
            .build_sqlx(PostgresQueryBuilder);

        sqlx::query_with(&query, values).execute(self.tx()).await?;

        Ok(())
    }

    /// Increment nonce value and retrieve for the given network in the `network_nonces` table.
    async fn increment_nonce(&mut self, network_id: NetworkId) -> Result<u64, DbError> {
        let (query, values) = Query::insert()
            .into_table(NetworkNonces::Table)
            .columns([NetworkNonces::NetworkId, NetworkNonces::Nonce])
            .values([(network_id.value() as i32).into(), 0i64.into()])?
            .on_conflict(
                OnConflict::column(NetworkNonces::NetworkId)
                    .value(
                        NetworkNonces::Nonce,
                        Expr::col((NetworkNonces::Table, NetworkNonces::Nonce)).add(1i64),
                    )
                    .to_owned(),
            )
            .returning_col(NetworkNonces::Nonce)
            .build_sqlx(PostgresQueryBuilder);

        #[derive(sqlx::FromRow)]
        struct Internal {
            nonce: i64,
        }

        let result: Internal = sqlx::query_as_with(&query, values)
            .fetch_one(self.tx())
            .await?;

        let nonce = result.nonce as u64;

        Ok(nonce)
    }
}

pub struct ReadTransaction<'c> {
    inner: &'c Database,
    tx: sqlx::Transaction<'c, Postgres>,
}

pub struct WriteTransaction<'c> {
    inner: &'c Database,
    tx: sqlx::Transaction<'c, Postgres>,
}

impl<'c> WriteTransaction<'c> {
    pub async fn commit(self) -> Result<(), DbError> {
        self.tx.commit().await?;

        Ok(())
    }
}

impl<'c> Debug for ReadTransaction<'c> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_tuple("ReadTransaction").field(&self.inner).finish()
    }
}

impl<'c> Debug for WriteTransaction<'c> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_tuple("WriteTransaction")
            .field(&self.inner)
            .finish()
    }
}

impl<'a> BridgeReadQueries<'a> for ReadTransaction<'a> {
    fn tx(&mut self) -> &mut sqlx::Transaction<'a, Postgres> {
        &mut self.tx
    }
}

impl<'a> BridgeReadQueries<'a> for WriteTransaction<'a> {
    fn tx(&mut self) -> &mut sqlx::Transaction<'a, Postgres> {
        &mut self.tx
    }
}

impl<'a> BridgeWriteQueries<'a> for WriteTransaction<'a> {}

impl Database {
    #[tracing::instrument(err)]
    pub async fn connect(config: Config) -> Result<Self, DbError> {
        let descriptor = config.descriptor();

        let Config {
            host,
            port,
            user,
            pass,
            dbname,
        } = config;

        let config = PgConnectOptions::new()
            .host(&host)
            .port(port)
            .username(&user)
            .password(&pass)
            .database(&dbname)
            .ssl_mode(PgSslMode::Disable);
        let pool = PgPool::connect_with(config).await?;

        Ok(Self { pool, descriptor })
    }

    pub async fn read_tx(&self) -> Result<ReadTransaction, DbError> {
        let mut tx = self.pool.begin().await?;

        tx.execute("SET TRANSACTION READ ONLY");

        Ok(ReadTransaction { inner: self, tx })
    }

    pub async fn write_tx(&self) -> Result<WriteTransaction, DbError> {
        let mut tx = self.pool.begin().await?;

        tx.execute("SET TRANSACTION READ WRITE");

        Ok(WriteTransaction { inner: self, tx })
    }
}

#[cfg(test)]
mod tests {
    use chrono::Utc;
    use primitive_types::U256;

    use crate::{
        test::{
            allocate_port,
            db::{init_db, init_pg},
        },
        types::{Address, NetworkId, TxHash},
    };

    use super::*;

    #[tokio::test]
    async fn tables() {
        let port = allocate_port();
        let mut pg = init_pg(port).await;
        let db = init_db(port).await;
        let mut dtx = db.write_tx().await.unwrap();

        dtx.create_tables().await.unwrap();

        pg.stop_db().await.unwrap();
    }

    #[tokio::test]
    async fn inserts() {
        let port = allocate_port();
        let mut pg = init_pg(port).await;
        let db = init_db(port).await;
        let mut dtx = db.write_tx().await.unwrap();

        dtx.create_tables().await.unwrap();

        let token_id = dtx.insert_token("TEST", "Test Token").await.unwrap();
        let tx_id = dtx
            .insert_transaction(
                &TxHash::new(NetworkId::new(0), vec![0, 1, 2, 3, 4, 5, 6, 7, 8]),
                10000,
                Utc::now(),
                &Address::new(NetworkId::new(0), vec![0, 1, 2, 3, 4, 5, 6, 7, 8]),
            )
            .await
            .unwrap();
        dtx.insert_transfer(
            tx_id,
            token_id,
            U256::from(1_000u64),
            &Address::new(NetworkId::new(0), vec![0, 1, 2, 3, 4, 5, 6, 7, 8]),
            &Address::new(NetworkId::new(0), vec![0, 1, 2, 3, 4, 5, 6, 7, 9]),
        )
        .await
        .unwrap();

        pg.stop_db().await.unwrap();
    }

    #[tokio::test]
    async fn increment_nonce() {
        let port = allocate_port();
        let mut pg = init_pg(port).await;
        let db = init_db(port).await;
        let mut dtx = db.write_tx().await.unwrap();

        dtx.create_tables().await.unwrap();
        assert_eq!(dtx.increment_nonce(NetworkId::new(0)).await.unwrap(), 0);
        assert_eq!(dtx.increment_nonce(NetworkId::new(0)).await.unwrap(), 1);
        assert_eq!(dtx.increment_nonce(NetworkId::new(0)).await.unwrap(), 2);
        assert_eq!(dtx.increment_nonce(NetworkId::new(0)).await.unwrap(), 3);
        assert_eq!(dtx.increment_nonce(NetworkId::new(1)).await.unwrap(), 0);
        pg.stop_db().await.unwrap();
    }
}
