package networks

import "context"

// Nonces is exposing access to network nonces db.
//
// architecture: DB
type Nonces interface {
	// Create inserts network nonce to database.
	Create(ctx context.Context, networkNonce NetworkNonce) error
	// Get returns nonce by network id from database.
	Get(ctx context.Context, networkID ID) (int64, error)
	// List returns list of network ids from database.
	List(ctx context.Context) ([]ID, error)
	// Update updates network nonce in database.
	Update(ctx context.Context, networkNonce NetworkNonce) error
	// Increment increments nonce in the database.
	Increment(ctx context.Context, NetworkID ID) error
}

// NetworkNonce describes nonce by network id. The nonce is needed to generate a unique signature and check whether
// the signature has already been used on the smart contract.
type NetworkNonce struct {
	NetworkID ID
	Nonce     int64
}
