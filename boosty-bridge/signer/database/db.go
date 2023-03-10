// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package database

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/zeebo/errs"

	"tricorn/signer"
)

// ensures that database implements signer.DB.
var _ signer.DB = (*database)(nil)

var (
	// Error is the default signer error class.
	Error = errs.Class("master database")
)

// database combines access to different database tables with a record
// of the db driver, db implementation, and db source URL.
//
// architecture: Master Database
type database struct {
	conn *sql.DB
}

// New returns signer.DB postgresql implementation.
func New(databaseURL string) (signer.DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return &database{conn: conn}, nil
}

// CreateSchema create schema for all tables and databases.
func (db *database) CreateSchema(ctx context.Context) error {
	createTableQuery :=
		`CREATE TABLE IF NOT EXISTS private_keys (
            network_type VARCHAR NOT NULL,
            private_key  VARCHAR NOT NULL,
			type         VARCHAR NOT NULL,
			PRIMARY KEY(network_type, type)
        );`

	_, err := db.conn.ExecContext(ctx, createTableQuery)
	return Error.Wrap(err)
}

// Close closes underlying db connection.
func (db *database) Close() error {
	return Error.Wrap(db.conn.Close())
}

// KeyStore provides access to accounts db.
func (db *database) KeyStore() signer.KeyStore {
	return &privateKeysDB{conn: db.conn}
}
