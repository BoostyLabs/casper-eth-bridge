CREATE SCHEMA bridge;

-- Holds information about known network-specific transactions
CREATE TABLE transactions
(
    id bigserial PRIMARY KEY,
    network_id integer NOT NULL,
    txhash bytea UNIQUE NOT NULL,
    sender bytea NOT NULL,
    blocknumber bigint NOT NULL,
    seen_at timestamptz NOT NULL,

    UNIQUE (network_id, txhash)
);

-- Holds supported token types
CREATE TABLE tokens
(
    id serial PRIMARY KEY,
    short_name text NOT NULL,
    long_name text NOT NULL
);

-- Holds information about a cross-chain token transfer
CREATE TABLE token_transfers
(
    id bigserial PRIMARY KEY,
    triggering_tx bigint REFERENCES transactions (id) NOT NULL,
    outbound_tx bigint REFERENCES transactions (id),
    token_id integer REFERENCES tokens (id) NOT NULL,
    amount bytea NOT NULL,
    status text NOT NULL,

    sender_network_id integer NOT NULL,
    sender_address bytea NOT NULL,
    recipient_network_id integer NOT NULL,
    recipient_address bytea NOT NULL
);

-- Association between token type and specific on-chain representation of that token
CREATE TABLE network_tokens
(
    network_id integer,
    token_id integer REFERENCES tokens(id) NOT NULL,
    contract_key bytea NOT NULL,
    decimals smallint NOT NULL,

    PRIMARY KEY(network_id, token_id)
);

-- Current network block height for each network
CREATE TABLE network_blocks
(
    network_id integer PRIMARY KEY,
    last_seen_block bigint
);

-- Current network nonce for each network
CREATE TABLE network_nonces
(
    network_id integer PRIMARY KEY,
    nonce bigint
);