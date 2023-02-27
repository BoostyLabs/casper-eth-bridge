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
)

// ensures that networkTokensDB implements networks.NetworkTokens.
var _ networks.NetworkTokens = (*networkTokensDB)(nil)

// ErrNetworkTokens indicates that there was an error in the database.
var ErrNetworkTokens = errs.Class("network tokens repository")

// networkTokensDB provide access to network tokens DB.
//
// architecture: Database
type networkTokensDB struct {
	conn *sql.DB
}

// Create inserts network token to database.
func (networkTokensDB *networkTokensDB) Create(ctx context.Context, networkToken networks.NetworkToken) error {
	query := "INSERT INTO network_tokens(network_id,token_id,contract_key,decimals) VALUES($1,$2,$3,$4)"
	_, err := networkTokensDB.conn.ExecContext(ctx, query, networkToken.NetworkID, networkToken.TokenID, networkToken.ContractAddress,
		networkToken.Decimals)
	return ErrNetworkTokens.Wrap(err)
}

// Get returns network token by network id and token id from database.
func (networkTokensDB *networkTokensDB) Get(ctx context.Context, networkID networks.ID, tokenID int64) (networks.NetworkToken, error) {
	networkToken := networks.NetworkToken{
		NetworkID: networkID,
		TokenID:   tokenID,
	}

	query := "SELECT contract_key, decimals FROM network_tokens WHERE network_id = $1 AND token_id = $2"
	row := networkTokensDB.conn.QueryRowContext(ctx, query, networkID, tokenID)

	if err := row.Scan(&networkToken.ContractAddress, &networkToken.Decimals); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return networkToken, ErrNetworkTokens.Wrap(bridge.ErrNoNetworkToken)
		}

		return networkToken, ErrNetworkTokens.Wrap(err)
	}

	return networkToken, nil
}

// List returns list of network tokens by token id from database.
func (networkTokensDB *networkTokensDB) List(ctx context.Context, tokenID int64) (_ []networks.NetworkToken, err error) {
	networkTokens := make([]networks.NetworkToken, 0)

	query := "SELECT network_id, contract_key, decimals FROM network_tokens WHERE token_id = $1"
	rows, err := networkTokensDB.conn.QueryContext(ctx, query, tokenID)
	if err != nil {
		return networkTokens, Error.Wrap(err)
	}

	defer func() {
		err = errs.Combine(err, rows.Close())
	}()

	for rows.Next() {
		networkToken := networks.NetworkToken{
			TokenID: tokenID,
		}

		err := rows.Scan(&networkToken.NetworkID, &networkToken.ContractAddress, &networkToken.Decimals)
		if err != nil {
			return networkTokens, Error.Wrap(err)
		}

		networkTokens = append(networkTokens, networkToken)
	}

	return networkTokens, nil
}

// Update updates network token in database.
func (networkTokensDB *networkTokensDB) Update(ctx context.Context, networkToken networks.NetworkToken) error {
	query := "UPDATE network_tokens SET contract_key = $1, decimals = $2 WHERE network_id = $3 AND token_id = $4"
	result, err := networkTokensDB.conn.ExecContext(ctx, query, networkToken.ContractAddress, networkToken.Decimals, networkToken.NetworkID,
		networkToken.TokenID)
	if err != nil {
		return ErrNetworkTokens.Wrap(err)
	}

	rowNum, err := result.RowsAffected()
	if rowNum == 0 && err == nil {
		return ErrNetworkTokens.Wrap(bridge.ErrNoNetworkToken)
	}

	return ErrNetworkTokens.Wrap(err)
}
