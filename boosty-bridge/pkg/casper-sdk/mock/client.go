// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package mock

import (
	"github.com/casper-ecosystem/casper-golang-sdk/sdk"

	"tricorn/chains/casper"
)

// MockRpcClient is a implementation of casper.Casper protocol.
type MockRpcClient struct{}

// New is a constructor for mock casper.Casper.
func New() casper.Casper {
	return &MockRpcClient{}
}

// PutDeploy deploys a contract or sends a transaction and returns deployment hash.
func (c *MockRpcClient) PutDeploy(deploy sdk.Deploy) (string, error) {
	return "a48d854d52746d159ecdde76bddea780159d223dc558379931dc4dbd07ea5261", nil
}

// GetBlockNumberByHash returns block number by deploy hash.
func (c *MockRpcClient) GetBlockNumberByHash(hash string) (int, error) {
	return 1, nil
}

// GetEventsByBlockNumbers returns events for range of block numbers.
func (c *MockRpcClient) GetEventsByBlockNumbers(fromBlockNumber uint64, toBlockNumber uint64, bridgeEventsHash string) ([]casper.Event, error) {
	return nil, nil
}

// GetCurrentBlockNumber returns events for range of block numbers.
func (c *MockRpcClient) GetCurrentBlockNumber() (uint64, error) {
	return 0, nil
}
