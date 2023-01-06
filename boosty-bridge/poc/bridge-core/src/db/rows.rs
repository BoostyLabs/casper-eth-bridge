use chrono::{DateTime, Utc};
use sqlx::FromRow;

#[derive(FromRow)]
pub struct Id64 {
    pub id: i64,
}

#[derive(FromRow)]
pub struct Id32 {
    pub id: i32,
}

#[derive(FromRow)]
pub struct Token {
    pub id: i32,
    pub short_name: String,
    pub long_name: String,
}

#[derive(FromRow)]
pub struct NetworkToken {
    pub token_id: i32,
    pub network_id: i32,
    pub contract_key: Vec<u8>,
    pub decimals: i16,
}

#[derive(FromRow)]
pub struct Transaction {
    pub id: i64,
    pub network_id: i32,
    pub txhash: Vec<u8>,
    pub blocknumber: i64,
    pub seen_at: DateTime<Utc>,
}

#[derive(FromRow)]
pub struct TransferWithHashes {
    pub id: i64,
    pub triggering_tx_nid: i32,
    pub triggering_tx_hash: Vec<u8>,
    pub outbound_tx_nid: Option<i32>,
    pub outbound_tx_hash: Option<Vec<u8>>,
    pub token_id: i32,
    pub amount: Vec<u8>,
    pub status: String,
    pub seen_at: DateTime<Utc>,

    pub sender_network_id: i32,
    pub sender_address: Vec<u8>,
    pub recipient_network_id: i32,
    pub recipient_address: Vec<u8>,
}

#[derive(FromRow)]
pub struct Transfer {
    pub id: i64,
    pub token_id: i32,
    pub amount: Vec<u8>,
    pub sender_network_id: i32,
    pub sender_address: Vec<u8>,
    pub recipient_network_id: i32,
    pub recipient_address: Vec<u8>,
    pub seen_at: DateTime<Utc>,
}

#[derive(FromRow)]
pub struct LastSeenBlock {
    pub last_seen_block: i64,
}

#[derive(FromRow)]
pub struct TransferDetails {
    pub sender_address: Vec<u8>,
    pub token_id: i32,
    pub amount: Vec<u8>,
}
