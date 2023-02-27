package transfers

import (
	"context"
	"math/big"

	"tricorn/bridge/networks"
	"tricorn/bridge/transactions"
)

// TokenTransfers is exposing access to token transfers db.
//
// architecture: DB
type TokenTransfers interface {
	// Create inserts token transfer to database.
	Create(ctx context.Context, tokenTransfer TokenTransfer) error
	// Get returns token transfer by id from database.
	Get(ctx context.Context, id int64) (TokenTransfer, error)
	// GetByNetworkAndTx returns token transfer by network and hash from database.
	GetByNetworkAndTx(ctx context.Context, networkID networks.ID, txHash []byte) (TokenTransfer, error)
	// GetByAllParams returns transfer by from, to, amount, tokenAddress parameters from the database.
	GetByAllParams(ctx context.Context, tokenTransfer TokenTransfer) (TokenTransfer, error)
	// ListByUser returns selected list of token transfers by user address and network id from database.
	ListByUser(ctx context.Context, offset, limit uint64, userWalletAddress []byte, networkID networks.ID) ([]TokenTransfer, error)
	// CountByUser counts total amount of transactions for user in one network.
	CountByUser(ctx context.Context, networkID networks.ID, userWalletAddress []byte) (amount uint64, err error)
	// Update updates token transfer in database.
	Update(ctx context.Context, tokenTransfer TokenTransfer) error
}

// TokenTransfer describes a transfer between networks.
type TokenTransfer struct {
	ID                 int64
	TriggeringTx       transactions.ID
	OutboundTx         transactions.ID
	TokenID            int64
	Amount             big.Int
	Status             Status
	SenderNetworkID    int64
	SenderAddress      []byte
	RecipientNetworkID int64
	RecipientAddress   []byte
}
