// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zeebo/errs"

	"tricorn/bridge"
	"tricorn/bridge/networks"
	"tricorn/bridge/transactions"
)

// ensures that transactionsDB implements transactions.DB.
var _ transactions.DB = (*transactionsDB)(nil)

// ErrTransactions indicates that there was an error in the database.
var ErrTransactions = errs.Class("transactions repository")

// transactionsDB provide access to transactions DB.
//
// architecture: Database
type transactionsDB struct {
	conn *sql.DB
}

// Create inserts transaction to database.
func (transactionsDB *transactionsDB) Create(ctx context.Context, transaction transactions.Transaction) (transactions.ID, error) {
	var id transactions.ID

	query := "INSERT INTO transactions(network_id,tx_hash,sender,block_number,seen_at) VALUES($1,$2,$3,$4,$5) RETURNING id"
	row := transactionsDB.conn.QueryRowContext(ctx, query, transaction.NetworkID, transaction.TxHash, transaction.Sender,
		transaction.BlockNumber, transaction.SeenAt)

	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrTransactions.Wrap(bridge.ErrNoTransaction)
		}

		return 0, ErrTransactions.Wrap(err)
	}

	return id, nil
}

// Exists returns nil if there is a new txHash for specified networkID.
func (transactionsDB *transactionsDB) Exists(ctx context.Context, networkID networks.ID, txHash []byte) error {
	query := "SELECT EXISTS(SELECT tx_Hash FROM transactions WHERE network_id = $1 AND tx_hash = $2)"
	row := transactionsDB.conn.QueryRowContext(ctx, query, networkID, txHash)

	var exist bool
	if err := row.Scan(&exist); err != nil {
		return ErrTransactions.Wrap(err)
	}

	if exist {
		return ErrTransactions.Wrap(bridge.ErrTransactionAlreadyExists)
	}

	return nil
}

// Get returns transaction by id from database.
func (transactionsDB *transactionsDB) Get(ctx context.Context, id transactions.ID) (transactions.Transaction, error) {
	transaction := transactions.Transaction{
		ID: id,
	}

	query := "SELECT network_id,tx_hash,sender,block_number,seen_at FROM transactions WHERE id = $1"
	row := transactionsDB.conn.QueryRowContext(ctx, query, id)

	if err := row.Scan(&transaction.NetworkID, &transaction.TxHash, &transaction.Sender, &transaction.BlockNumber, &transaction.SeenAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return transaction, ErrTransactions.Wrap(bridge.ErrNoTransaction)
		}

		return transaction, ErrTransactions.Wrap(err)
	}

	return transaction, nil
}
