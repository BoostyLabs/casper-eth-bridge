// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package mock

import (
	"context"

	"tricorn/chains/evm"
)

// Client is mock implementation of evm.Transfer protocol.
type Client struct{}

// New is a constructor for mock evm.Transfer.
func New() evm.Transfer {
	return &Client{}
}

// TransferOutSignature generates signature for transfer out transaction.
func (c *Client) TransferOutSignature(ctx context.Context, transferOut evm.TransferOutRequest) ([]byte, error) {
	return []byte{}, nil
}

// TransferOut initiates outbound bridge transaction only for contract owner.
func (c *Client) TransferOut(ctx context.Context, transferOut evm.TransferOutRequest) error {
	return nil
}

// GetBridgeInSignature generates signature for inbound bridge transaction.
func (c *Client) GetBridgeInSignature(ctx context.Context, bridgeIn evm.GetBridgeInSignatureRequest) ([]byte, error) {
	return []byte{}, nil
}

// BridgeIn initiates inbound bridge transaction.
func (c *Client) BridgeIn(ctx context.Context, bridgeIn evm.BridgeInRequest) (string, error) {
	return "", nil
}

// Close closes underlying client connection.
func (c *Client) Close() {}
