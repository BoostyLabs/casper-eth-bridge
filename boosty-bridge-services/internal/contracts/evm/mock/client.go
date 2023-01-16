// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package mock

import (
	"context"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/evm"
)

// MockClient is a implementation of evm.Transfer protocol.
type MockClient struct{}

// New is a constructor for mock evm.Transfer.
func New() evm.Transfer {
	return &MockClient{}
}

// TransferOutSignature generates signature for transfer out transaction.
func (c *MockClient) TransferOutSignature(ctx context.Context, transferOut evm.TransferOutRequest) ([]byte, error) {
	return []byte{}, nil
}

// TransferOut initiates outbound bridge transaction only for contract owner.
func (m *MockClient) TransferOut(ctx context.Context, transferOut evm.TransferOutRequest) error {
	return nil
}

// GetBridgeInSignature generates signature for inbound bridge transaction.
func (m *MockClient) GetBridgeInSignature(ctx context.Context, bridgeIn evm.GetBridgeInSignatureRequest) ([]byte, error) {
	return []byte{}, nil
}

// BridgeIn initiates inbound bridge transaction.
func (m *MockClient) BridgeIn(ctx context.Context, bridgeIn evm.BridgeInRequest) (string, error) {
	return "", nil
}

// Close closes underlying client connection.
func (m *MockClient) Close() {}
