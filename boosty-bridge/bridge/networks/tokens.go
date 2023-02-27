package networks

import (
	"context"
)

// NetworkTokens is exposing access to network tokens db.
//
// architecture: DB
type NetworkTokens interface {
	// Create inserts network token to database.
	Create(ctx context.Context, networkToken NetworkToken) error
	// Get returns network token by network id and token id from database.
	Get(ctx context.Context, networkID ID, tokenID int64) (NetworkToken, error)
	// List returns list of network tokens by token id from database.
	List(ctx context.Context, tokenID int64) ([]NetworkToken, error)
	// Update updates network token in database.
	Update(ctx context.Context, networkToken NetworkToken) error
}

// NetworkToken describes decimals by token in specific network. The decimal number is required to convert currencies.
type NetworkToken struct {
	NetworkID       ID
	TokenID         int64
	ContractAddress []byte
	Decimals        int64
}

// SupportedToken describes supported tokens for network.
type SupportedToken struct {
	ID        int64
	ShortName string
	LongName  string
	Addresses []NetworkToken
}
