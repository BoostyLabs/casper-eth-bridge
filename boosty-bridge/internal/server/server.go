package server

import (
	"context"
)

// Server describes server behaviour.
type Server interface {
	// Run runs underlying server connection.
	Run(ctx context.Context) (err error)
	// Close closes underlying server connection.
	Close() error
}
