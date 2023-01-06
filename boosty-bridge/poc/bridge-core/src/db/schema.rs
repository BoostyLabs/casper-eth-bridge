#![allow(unused)]

use sea_query::Iden;

#[derive(Iden)]
pub enum Transactions {
    Table,
    Id,
    NetworkId,
    Txhash,
    Blocknumber,
    SeenAt,
    Sender,
}

#[derive(Iden)]
pub enum Tokens {
    Table,
    Id,
    ShortName,
    LongName,
}

#[derive(Iden)]
pub enum TokenTransfers {
    Table,
    Id,
    TriggeringTx,
    OutboundTx,
    TokenId,
    Amount,
    Status,
    SenderNetworkId,
    SenderAddress,
    RecipientNetworkId,
    RecipientAddress,
}

#[derive(Iden)]
pub enum NetworkTokens {
    Table,
    NetworkId,
    TokenId,
    ContractKey,
    Decimals,
}

#[derive(Iden)]
pub enum NetworkBlocks {
    Table,
    NetworkId,
    LastSeenBlock,
}

#[derive(Iden)]
pub enum NetworkNonces {
    Table,
    NetworkId,
    Nonce,
}
