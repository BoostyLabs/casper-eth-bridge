package server

import (
	"context"
)

// Server provides access to all server methods.
type Server interface {
	// Run runs underlying server connection.
	Run(ctx context.Context) (err error)
	// Close closes underlying server connection.
	Close() error
}
