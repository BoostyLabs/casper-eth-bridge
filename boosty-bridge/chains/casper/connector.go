// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package casper

import (
	"context"
	"math/big"

	"github.com/casper-ecosystem/casper-golang-sdk/sdk"

	"tricorn/bridge/networks"
)

// Casper exposes access to the casper sdk methods.
type Casper interface {
	// PutDeploy deploys a contract or sends a transaction and returns deployment hash.
	PutDeploy(deploy sdk.Deploy) (string, error)
	// GetBlockNumberByHash returns block number by deploy hash.
	GetBlockNumberByHash(hash string) (int, error)
	// GetEventsByBlockNumbers returns events for range of block numbers.
	GetEventsByBlockNumbers(fromBlockNumber uint64, toBlockNumber uint64, bridgeInEventHash string) ([]Event, error)
	// GetCurrentBlockNumber returns current block number.
	GetCurrentBlockNumber() (uint64, error)
}

// Signer exposes access to the signer methods.
type Signer interface {
	// GetBridgeInSignature generates signature for inbound bridge transaction.
	GetBridgeInSignature(ctx context.Context, bridgeIn BridgeInSignature) ([]byte, error)
	// GetTransferOutSignature generates signature for outbound transfer transaction.
	GetTransferOutSignature(ctx context.Context, transferOut TransferOutSignature) ([]byte, error)
}

// BridgeInSignature describes values to generate signature for bridgeIn method.
type BridgeInSignature struct {
	Prefix             string
	BridgeHash         []byte
	TokenPackageHash   []byte
	AccountAddress     []byte
	Amount             *big.Int
	GasCommission      *big.Int
	Deadline           *big.Int
	Nonce              *big.Int
	DestinationChain   string
	DestinationAddress string
}

// TransferOutSignature describes values to generate signature for transferOut method.
type TransferOutSignature struct {
	Prefix           string
	BridgeHash       []byte
	TokenPackageHash []byte
	AccountAddress   []byte
	Recipient        []byte
	Amount           *big.Int
	GasCommission    *big.Int
	Nonce            *big.Int
}

// Config defines configurable values for casper service.
// TODO: add gasLimit for In and Out.
type Config struct {
	RPCNodeAddress   string `env:"RPC_NODE_ADDRESS"`
	EventNodeAddress string `env:"EVENT_NODE_ADDRESS"`

	BridgeEventsHash      string        `env:"BRIDGE_EVENTS_HASH"`
	ChainName             networks.Name `env:"CHAIN_NAME"`
	GasLimit              uint64        `env:"GAS_LIMIT"`
	BridgeContractAddress string        `env:"BRIDGE_CONTRACT_ADDRESS"`
	IsTestnet             bool          `env:"IS_TESTNET"`
	FeePercentage         string        `env:"FEE_PERCENTAGE"`
	EstimatedConfirmation uint32        `env:"ESTIMATED_CONFIRMATION"`
	BridgeInPrefix        string        `env:"BRIDGE_IN_PREFIX"`
	TransferOutPrefix     string        `env:"TRANSFER_OUT_PREFIX"`
	SignatureValidityTime uint32        `env:"SIGNATURE_VALIDITY_TIME"`
}

// Event describes event structure in casper network.
type (
	Event struct {
		DeployProcessed DeployProcessed `json:"DeployProcessed"`
	}

	DeployProcessed struct {
		DeployHash      string          `json:"deploy_hash"`
		Account         string          `json:"account"`
		BlockHash       string          `json:"block_hash"`
		ExecutionResult ExecutionResult `json:"execution_result"`
	}

	ExecutionResult struct {
		Success Success `json:"Success"`
	}

	Success struct {
		Effect Effect `json:"effect"`
	}

	Effect struct {
		Transforms []Transform `json:"transforms"`
	}

	Transform struct {
		Key       string      `json:"key"`
		Transform interface{} `json:"transform"`
	}
)

const (
	// WriteCLValueKey defines that transform key is WriteCLValue. This key stores the type and data of the transforming event.
	WriteCLValueKey string = "WriteCLValue"
	// BytesKey defines that WriteCLValue key is bytes. This key stores data of the transforming event.
	BytesKey string = "bytes"
)

const (
	// fundInType defines fund in event type.
	fundInType = 0
	// fundOutType defines fund out event type.
	fundOutType = 1
)
