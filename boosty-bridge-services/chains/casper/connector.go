// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package casper

import (
	"github.com/casper-ecosystem/casper-golang-sdk/sdk"
)

// Casper exposes access to the casper sdk methods.
type Casper interface {
	// PutDeploy deploys a contract or sends a transaction and returns deployment hash.
	PutDeploy(deploy sdk.Deploy) (string, error)
	// GetBlockNumberByHash returns block number by deploy hash.
	GetBlockNumberByHash(hash string) (int, error)
	// GetEventsByBlockNumbers returns events for range of block numbers.
	GetEventsByBlockNumbers(fromBlockNumber uint64, toBlockNumber uint64, bridgeInEventHash string, bridgeOutEventHash string) ([]Event, error)
	// GetCurrentBlockNumber returns current block number.
	GetCurrentBlockNumber() (uint64, error)
}

// Config defines configurable values for casper service.
type Config struct {
	RPCNodeAddress   string `env:"RPC_NODE_ADDRESS"`
	EventNodeAddress string `env:"EVENT_NODE_ADDRESS"`

	BridgeInEventHash           string `env:"BRIDGE_IN_EVENT_HASH"`
	BridgeOutEventHash          string `env:"BRIDGE_OUT_EVENT_HASH"`
	ChainName                   string `env:"CHAIN_NAME"`
	StandardPaymentForBridgeOut uint64 `env:"STANDARD_PAYMENT_FOR_BRIDGE_OUT"`
	BridgeContractPackageHash   string `env:"BRIDGE_CONTRACT_PACKAGE_HASH"`
	IsTestnet                   bool   `env:"IS_TESTNET"`
	Fee                         string `env:"FEE"`
	FeePercentage               string `env:"FEE_PERCENTAGE"`
	EstimatedConfirmation       uint32 `env:"ESTIMATED_CONFIRMATION"`
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
