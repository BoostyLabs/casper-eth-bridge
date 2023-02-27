package transactions

import (
	"context"
	"time"

	"tricorn/bridge/networks"
)

// DB is exposing access to transactions db.
//
// architecture: DB
type DB interface {
	// Create inserts transaction to database.
	Create(ctx context.Context, transaction Transaction) (ID, error)
	// Get returns transaction by id from database.
	Get(ctx context.Context, id ID) (Transaction, error)
	// Exists returns nil if there is a new txHash for specified networkID.
	Exists(ctx context.Context, networkID networks.ID, txHash []byte) error
}

// ID defines internal transaction id.
type ID int

// Transaction describes transaction sent by the sender. A blockchain transaction is nothing but data transmission across the network
// of computers in a blockchain system. A transaction refers to a contract, agreement, transfer, or exchange of assets between two or more parties.
// In our case, the difference between a transfer and a transaction was that the transfer is a complex concept, it is type 2 transaction(In and Out).
type Transaction struct {
	ID          ID
	NetworkID   networks.ID
	TxHash      []byte
	Sender      []byte
	BlockNumber int64
	SeenAt      time.Time
}
