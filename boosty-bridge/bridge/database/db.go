// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package database

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/zeebo/errs"

	"tricorn/bridge"
	"tricorn/bridge/networks"
	"tricorn/bridge/transactions"
	"tricorn/bridge/transfers"
)

// ensures that database implements bridge.DB.
var _ bridge.DB = (*database)(nil)

var (
	// Error is the default bridge error class.
	Error = errs.Class("master database")
)

// database combines access to different database tables with a record
// of the db driver, db implementation, and db source URL.
//
// architecture: Master Database
type database struct {
	conn *sql.DB
}

// New returns bridge.DB postgresql implementation.
func New(databaseURL string) (bridge.DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return &database{conn: conn}, nil
}

// CreateSchema create schema for all tables and databases.
func (db *database) CreateSchema(ctx context.Context) error {
	createTableQuery :=
		`CREATE TABLE IF NOT EXISTS network_blocks (
            network_id      INTEGER PRIMARY KEY NOT NULL,
            last_seen_block INTEGER             NOT NULL
        );
        CREATE TABLE IF NOT EXISTS network_nonces (
            network_id INTEGER PRIMARY KEY NOT NULL,
            nonce      INTEGER             NOT NULL
        );
        CREATE TABLE IF NOT EXISTS network_tokens (
            network_id   INTEGER NOT NULL,
            token_id     INTEGER NOT NULL,
            contract_key BYTEA   NOT NULL,
            decimals     INTEGER NOT NULL,
            PRIMARY KEY(network_id,token_id)
        );
        CREATE TABLE IF NOT EXISTS token_transfers (
            id                   BIGSERIAL PRIMARY KEY NOT NULL,
            triggering_tx        INTEGER               NOT NULL,
            outbound_tx          INTEGER,
            token_id             INTEGER               NOT NULL,
            amount               BYTEA                 NOT NULL,
            status               VARCHAR               NOT NULL,
            sender_network_id    INTEGER               NOT NULL,
            sender_address       BYTEA                 NOT NULL,
            recipient_network_id INTEGER               NOT NULL,
            recipient_address    BYTEA                 NOT NULL
        );
        CREATE TABLE IF NOT EXISTS tokens (
            id         SERIAL  PRIMARY KEY NOT NULL,
            short_name VARCHAR             NOT NULL,
            long_name  VARCHAR             NOT NULL
        );
        CREATE TABLE IF NOT EXISTS transactions (
            id           BIGSERIAL PRIMARY KEY    NOT NULL,
            network_id   INTEGER                  NOT NULL,
            tx_hash      BYTEA                    NOT NULL,
            sender       BYTEA                    NOT NULL,
            block_number INTEGER                  NOT NULL,
            seen_at      TIMESTAMP WITH TIME ZONE NOT NULL
        );`

	_, err := db.conn.ExecContext(ctx, createTableQuery)
	return Error.Wrap(err)
}

// Close closes underlying db connection.
func (db *database) Close() error {
	return Error.Wrap(db.conn.Close())
}

// NetworkBlocks provides access to accounts db.
func (db *database) NetworkBlocks() networks.NetworkBlocks {
	return &networkBlocksDB{conn: db.conn}
}

// Nonces provides access to accounts db.
func (db *database) Nonces() networks.Nonces {
	return &networkNoncesDB{conn: db.conn}
}

// NetworkTokens provides access to accounts db.
func (db *database) NetworkTokens() networks.NetworkTokens {
	return &networkTokensDB{conn: db.conn}
}

// TokenTransfers provides access to accounts db.
func (db *database) TokenTransfers() transfers.TokenTransfers {
	return &tokenTransfersDB{conn: db.conn}
}

// Tokens provides access to accounts db.
func (db *database) Tokens() bridge.Tokens {
	return &tokensDB{conn: db.conn}
}

// Transactions provides access to accounts db.
func (db *database) Transactions() transactions.DB {
	return &transactionsDB{conn: db.conn}
}
