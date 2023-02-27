// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package communication

import (
	"context"
	"errors"

	"tricorn/bridge"
	"tricorn/bridge/networks"
	"tricorn/bridge/transfers"
	"tricorn/chains"
)

// ErrNotConnected indicated that underlying Communication connection is not established.
var ErrNotConnected = errors.New("connection is not established")

// Communication provides access to all services communication with bridge methods.
type Communication interface {
	// Networks provides access to the networks.Bridge rpc methods.
	Networks() networks.Bridge

	// Transfers provides access to the transfers.Bridge rpc methods.
	Transfers() transfers.Bridge

	// Bridge provides access to the chains.Bridge rpc methods.
	Bridge() chains.Bridge

	// Connector provides access to the bridge.Connector rpc methods.
	Connector(ctx context.Context) bridge.Connector

	// CurrencyRates provides access to the bridge.CurrencyRates rpc methods.
	CurrencyRates() bridge.CurrencyRates

	// ConnectWithPing will try to establish connection.
	ConnectWithPing(ctx context.Context) error

	// Signer provides access to the bridge.Signer rpc methods.
	Signer() bridge.Signer

	// Close closes underlying communication connection.
	Close() error
}

// Mode holds possible communication modes.
type Mode string

const (
	// ModeGRPC indicates that we use GRPC for communication.
	ModeGRPC Mode = "GRPC"
	// ModeDEV indicates that instead of real communication we use mock proxy.
	ModeDEV Mode = "DEV"
)
