// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package evm

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// ErrBlockchainRework represents error for events that have been removed due to blockchain rework.
var ErrBlockchainRework = errors.New("blockchain event removed due to rework")

// listeningLimit defines the limit for listing event.
const listeningLimit = 2500

// Config contains eth configurable values.
type Config struct {
	NodeAddress              string         `env:"NODE_ADDRESS"`
	WsNodeAddress            string         `env:"WS_NODE_ADDRESS" help:"node which supports websocket connection"`
	ChainID                  uint32         `env:"CHAIN_ID"`
	ChainName                string         `env:"CHAIN_NAME"`
	IsTestnet                bool           `env:"IS_TESTNET"`
	BridgeContractAddress    common.Address `env:"BRIDGE_CONTRACT_ADDRESS"`
	BridgeOutMethodName      string         `env:"BRIDGE_OUT_METHOD_NAME"`
	EventsFundIn             common.Hash    `env:"FUND_IN_EVENT_HASH"`
	EventsFundOut            common.Hash    `env:"FUND_OUT_EVENT_HASH"`
	GasIncreasingCoefficient float64        `env:"GAS_INCREASING_COEFFICIENT"`
	ConfirmationTime         uint32         `env:"CONFIRMATION_TIME"`
	FeePercentage            string         `env:"FEE_PERCENTAGE"`
	GasLimit                 uint64         `env:"GAS_LIMIT"` // TODO: count by tx.
	NumOfSubscribers         int            `env:"NUM_OF_SUBSCRIBERS"`
	SignatureValidityTime    uint32         `env:"SIGNATURE_VALIDITY_TIME"`
}

// Transfer exposes access to the evm transfer methods.
type Transfer interface {
	// TransferOut initiates outbound bridge transaction only for contract owner.
	TransferOut(ctx context.Context, transferOut TransferOutRequest) error
	// GetBridgeInSignature generates signature for inbound bridge transaction.
	GetBridgeInSignature(ctx context.Context, bridgeIn GetBridgeInSignatureRequest) ([]byte, error)
	// BridgeIn initiates inbound bridge transaction.
	BridgeIn(ctx context.Context, bridgeIn BridgeInRequest) (string, error)
	// Close closes underlying client connection.
	Close()
}

// TransferOutRequest describes values for TransferOut method.
type TransferOutRequest struct {
	Token      common.Address
	Recipient  common.Address
	Amount     *big.Int
	Commsision *big.Int
	Nonce      *big.Int
}

// GetBridgeInSignatureRequest describes values for GetBridgeInSignature method.
type GetBridgeInSignatureRequest struct {
	User               common.Address
	Token              common.Address
	Amount             *big.Int
	GasCommission      *big.Int
	DestinationChain   string
	DestinationAddress string
	Deadline           *big.Int
	Nonce              *big.Int
}

// BridgeInRequest describes values for BridgeIn method.
type BridgeInRequest struct {
	Token              common.Address
	Amount             *big.Int
	GasCommission      *big.Int
	DestinationChain   string
	DestinationAddress string
	Deadline           *big.Int
	Nonce              *big.Int
	Signature          []byte
}
