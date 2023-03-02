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
	"tricorn/bridge/transfers"
)

// ensures that tokenTransfersDB implements transfers.TokenTransfers.
var _ transfers.TokenTransfers = (*tokenTransfersDB)(nil)

// ErrTokenTransfers indicates that there was an error in the database.
var ErrTokenTransfers = errs.Class("token transfers repository")

// tokenTransfersDB provide access to token transfers DB.
//
// architecture: Database
type tokenTransfersDB struct {
	conn *sql.DB
}

// Create inserts token transfer to database.
func (tokenTransfersDB *tokenTransfersDB) Create(ctx context.Context, tokenTransfer transfers.TokenTransfer) error {
	query := `INSERT INTO token_transfers(triggering_tx,outbound_tx,token_id,amount,status,sender_network_id,sender_address,
		recipient_network_id,recipient_address) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := tokenTransfersDB.conn.ExecContext(ctx, query, tokenTransfer.TriggeringTx, tokenTransfer.OutboundTx, tokenTransfer.TokenID,
		tokenTransfer.Amount.Bytes(), tokenTransfer.Status, tokenTransfer.SenderNetworkID, tokenTransfer.SenderAddress,
		tokenTransfer.RecipientNetworkID, tokenTransfer.RecipientAddress)
	return ErrTokenTransfers.Wrap(err)
}

// Get returns token transfer by id from database.
func (tokenTransfersDB *tokenTransfersDB) Get(ctx context.Context, id int64) (transfers.TokenTransfer, error) {
	var (
		tokenTransfer transfers.TokenTransfer
		amount        []byte
		outboundTx    sql.NullInt64
		triggeringTx  sql.NullInt64
	)

	query := `SELECT id,triggering_tx,outbound_tx,token_id,amount,status,sender_network_id,sender_address,recipient_network_id,recipient_address 
	FROM token_transfers WHERE id = $1`
	row := tokenTransfersDB.conn.QueryRowContext(ctx, query, id)

	if err := row.Scan(&tokenTransfer.ID, &triggeringTx, &outboundTx, &tokenTransfer.TokenID, &amount,
		&tokenTransfer.Status, &tokenTransfer.SenderNetworkID, &tokenTransfer.SenderAddress, &tokenTransfer.RecipientNetworkID,
		&tokenTransfer.RecipientAddress); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tokenTransfer, ErrTokenTransfers.Wrap(bridge.ErrNoTokenTransfer)
		}

		return tokenTransfer, ErrTokenTransfers.Wrap(err)
	}

	tokenTransfer.Amount.SetBytes(amount)
	if triggeringTx.Valid {
		tokenTransfer.TriggeringTx = transactions.ID(triggeringTx.Int64)
	}
	if outboundTx.Valid {
		tokenTransfer.OutboundTx = transactions.ID(outboundTx.Int64)
	}

	return tokenTransfer, nil
}

// GetByAllParams returns transfer by from, to, amount, tokenAddress parameters from the database.
func (tokenTransfersDB *tokenTransfersDB) GetByAllParams(ctx context.Context, params transfers.TokenTransfer) (transfers.TokenTransfer, error) {
	var (
		tokenTransfer transfers.TokenTransfer
		outboundTx    sql.NullInt64
		triggeringTx  sql.NullInt64
		amount        []byte
	)

	query := `SELECT id,triggering_tx,outbound_tx,token_id,amount,status,sender_network_id,sender_address,recipient_network_id,recipient_address
	          FROM token_transfers
	          WHERE token_id = $1 AND amount=$2 AND sender_address = $3 AND recipient_address = $4
			  ORDER BY id DESC`

	row := tokenTransfersDB.conn.QueryRowContext(ctx, query, params.TokenID, params.Amount.Bytes(), params.SenderAddress, params.RecipientAddress)

	if err := row.Scan(&tokenTransfer.ID, &triggeringTx, &outboundTx, &tokenTransfer.TokenID, &amount,
		&tokenTransfer.Status, &tokenTransfer.SenderNetworkID, &tokenTransfer.SenderAddress, &tokenTransfer.RecipientNetworkID,
		&tokenTransfer.RecipientAddress); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tokenTransfer, ErrTokenTransfers.Wrap(bridge.ErrNoTokenTransfer)
		}

		return tokenTransfer, ErrTokenTransfers.Wrap(err)
	}

	tokenTransfer.Amount.SetBytes(amount)
	if triggeringTx.Valid {
		tokenTransfer.TriggeringTx = transactions.ID(triggeringTx.Int64)
	}
	if outboundTx.Valid {
		tokenTransfer.OutboundTx = transactions.ID(outboundTx.Int64)
	}

	return tokenTransfer, nil
}

// GetByNetworkAndTx returns token transfer by network and hash from database.
func (tokenTransfersDB *tokenTransfersDB) GetByNetworkAndTx(ctx context.Context, networkID networks.ID, txHash []byte) (transfers.TokenTransfer, error) {
	var (
		tokenTransfer transfers.TokenTransfer
		outboundTx    sql.NullInt64
		triggeringTx  sql.NullInt64
		amount        []byte
	)

	query := `SELECT tt.id,tt.triggering_tx,tt.outbound_tx,tt.token_id,tt.amount,tt.status,tt.sender_network_id,tt.sender_address,tt.recipient_network_id,tt.recipient_address 
	    FROM token_transfers as tt
	    LEFT JOIN transactions as txt ON tt.triggering_tx = txt.id
        LEFT JOIN transactions as txo ON tt.outbound_tx = txo.id
        WHERE txt.network_id = $1 AND txt.tx_hash = $2`
	row := tokenTransfersDB.conn.QueryRowContext(ctx, query, networkID, txHash)

	if err := row.Scan(&tokenTransfer.ID, &triggeringTx, &outboundTx, &tokenTransfer.TokenID, &amount,
		&tokenTransfer.Status, &tokenTransfer.SenderNetworkID, &tokenTransfer.SenderAddress, &tokenTransfer.RecipientNetworkID,
		&tokenTransfer.RecipientAddress); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tokenTransfer, ErrTokenTransfers.Wrap(bridge.ErrNoTokenTransfer)
		}

		return tokenTransfer, ErrTokenTransfers.Wrap(err)
	}

	if triggeringTx.Valid {
		tokenTransfer.TriggeringTx = transactions.ID(triggeringTx.Int64)
	}
	if outboundTx.Valid {
		tokenTransfer.OutboundTx = transactions.ID(outboundTx.Int64)
	}
	tokenTransfer.Amount.SetBytes(amount)

	return tokenTransfer, nil
}

// ListByUser returns selected list of token transfers by user address and network id from database.
func (tokenTransfersDB *tokenTransfersDB) ListByUser(ctx context.Context, offset, limit uint64, userWalletAddress []byte, networkID networks.ID) (_ []transfers.TokenTransfer, err error) {
	tokenTransfers := make([]transfers.TokenTransfer, 0)

	selectQuery := `SELECT tt.id, tt.triggering_tx, tt.outbound_tx, tt.token_id, tt.amount, tt.status, tt.sender_network_id,
   	    tt.sender_address, tt.recipient_network_id, tt.recipient_address
        FROM token_transfers as tt 
        LEFT JOIN transactions as txt ON tt.triggering_tx = txt.id
        LEFT JOIN transactions as txo ON tt.outbound_tx = txo.id
        WHERE tt.sender_network_id = $1 AND tt.sender_address = $2
        ORDER BY txt.seen_at DESC
        LIMIT $3 OFFSET $4`
	rows, err := tokenTransfersDB.conn.QueryContext(ctx, selectQuery, networkID, userWalletAddress, limit, offset)
	if err != nil {
		return tokenTransfers, Error.Wrap(err)
	}

	defer func() {
		err = errs.Combine(err, rows.Close())
	}()

	for rows.Next() {
		var (
			tokenTransfer transfers.TokenTransfer
			outboundTx    sql.NullInt64
			triggeringTx  sql.NullInt64
			amount        []byte
		)
		if err := rows.Scan(&tokenTransfer.ID, &triggeringTx, &outboundTx, &tokenTransfer.TokenID, &amount,
			&tokenTransfer.Status, &tokenTransfer.SenderNetworkID, &tokenTransfer.SenderAddress, &tokenTransfer.RecipientNetworkID,
			&tokenTransfer.RecipientAddress); err != nil {
			return tokenTransfers, Error.Wrap(err)
		}

		if triggeringTx.Valid {
			tokenTransfer.TriggeringTx = transactions.ID(triggeringTx.Int64)
		}
		if outboundTx.Valid {
			tokenTransfer.OutboundTx = transactions.ID(outboundTx.Int64)
		}
		tokenTransfer.Amount.SetBytes(amount)

		tokenTransfers = append(tokenTransfers, tokenTransfer)
	}

	return tokenTransfers, nil
}

// CountByUser counts total amount of transactions for user in one network.
func (tokenTransfersDB *tokenTransfersDB) CountByUser(ctx context.Context, networkID networks.ID, userWalletAddress []byte) (amount uint64, err error) {
	query := `SELECT COUNT(*) as total_value FROM token_transfers as tt WHERE tt.sender_network_id = $1 AND tt.sender_address = $2`
	row := tokenTransfersDB.conn.QueryRowContext(ctx, query, networkID, userWalletAddress)

	if err = row.Scan(&amount); err != nil {
		return 0, ErrTokenTransfers.Wrap(err)
	}

	return amount, nil
}

// Update updates token transfer in database.
func (tokenTransfersDB *tokenTransfersDB) Update(ctx context.Context, tokenTransfer transfers.TokenTransfer) error {
	query := `UPDATE token_transfers SET triggering_tx = $1, outbound_tx = $2, token_id = $3, amount = $4, status = $5, sender_network_id = $6,
	sender_address = $7, recipient_network_id = $8, recipient_address = $9 WHERE id = $10`
	result, err := tokenTransfersDB.conn.ExecContext(ctx, query, tokenTransfer.TriggeringTx, tokenTransfer.OutboundTx, tokenTransfer.TokenID,
		tokenTransfer.Amount.Bytes(), tokenTransfer.Status, tokenTransfer.SenderNetworkID, tokenTransfer.SenderAddress,
		tokenTransfer.RecipientNetworkID, tokenTransfer.RecipientAddress, tokenTransfer.ID)
	if err != nil {
		return ErrTokenTransfers.Wrap(err)
	}

	rowNum, err := result.RowsAffected()
	if rowNum == 0 && err == nil {
		return ErrTokenTransfers.Wrap(bridge.ErrNoTokenTransfer)
	}

	return ErrTokenTransfers.Wrap(err)
}
