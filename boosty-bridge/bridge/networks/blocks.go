package networks

import "context"

// NetworkBlocks is exposing access to network blocks db.
//
// architecture: DB
type NetworkBlocks interface {
	// Create inserts network block to database.
	Create(ctx context.Context, networkBlock NetworkBlock) error
	// Get returns last seen block by network id from database.
	Get(ctx context.Context, networkID ID) (int64, error)
	// Update updates network block in database.
	Update(ctx context.Context, networkBlock NetworkBlock) error
}

// NetworkBlock describes last seen block by network id. The last іуут block is required for uninterrupted event tracking.
type NetworkBlock struct {
	NetworkID     ID
	LastSeenBlock int64
}
