// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package communication

import (
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/transfers"
)

// Communication provides access to all services communication with bridge methods.
type Communication interface {
	// Networks provides access to the networks.Bridge rpc methods.
	Networks() networks.Bridge

	// Transfers provides access to the transfers.Bridge rpc methods.
	Transfers() transfers.Bridge

	// Bridge provides access to the chains.Bridge rpc methods.
	Bridge() chains.Bridge

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
