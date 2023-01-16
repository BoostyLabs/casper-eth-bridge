// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package mockcommunication

import (
	"context"
	"math/big"
	"time"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/transfers"
	"github.com/ethereum/go-ethereum/common"
)

// ensures that MockCommunication implement communication.Communication.
var _ communication.Communication = (*MockCommunication)(nil)

// MockCommunication is a implementation of Communication protocol.
type MockCommunication struct{}

// New is a constructor for mock communication.Communication.
func New() communication.Communication {
	return &MockCommunication{}
}

// Bridge provides access to the chains.Bridge rpc methods.
func (rpc *MockCommunication) Bridge() chains.Bridge {
	return &BridgeMock{
		signImpl: func(ctx context.Context, req chains.SignRequest) ([]byte, error) {
			return []byte{}, nil
		},
		publicKeyImpl: func(ctx context.Context, networkId networks.Type) ([]byte, error) {
			return []byte{}, nil
		},
	}
}

// ensures that BridgeMock implements chains.Bridge.
var _ chains.Bridge = (*BridgeMock)(nil)

// BridgeMock provides access to the chains.Bridge.
type BridgeMock struct {
	signImpl      func(ctx context.Context, req chains.SignRequest) ([]byte, error)
	publicKeyImpl func(ctx context.Context, networkId networks.Type) ([]byte, error)
}

// Sign returns signed data for specific network.
func (bridgeMock *BridgeMock) Sign(ctx context.Context, req chains.SignRequest) ([]byte, error) {
	return bridgeMock.signImpl(ctx, req)
}

// PublicKey returns public key for specific network.
func (bridgeMock *BridgeMock) PublicKey(ctx context.Context, networkId networks.Type) ([]byte, error) {
	return bridgeMock.publicKeyImpl(ctx, networkId)
}

// Networks provides access to the networks.Bridge rpc methods.
func (rpc *MockCommunication) Networks() networks.Bridge {
	return &NetworksMock{
		connectedNetworksImpl: func(ctx context.Context) ([]networks.Network, error) {
			return []networks.Network{
				{
					ID:        0,
					Name:      "kefir",
					Type:      networks.TypeEVM,
					IsTestnet: true,
				},
				{
					ID:        1,
					Name:      "Karlson",
					Type:      networks.TypeCasper,
					IsTestnet: true,
				},
			}, nil
		},
		supportedTokensImpl: func(ctx context.Context, networkID uint32) ([]networks.Token, error) {
			return []networks.Token{
				{
					ID:        1,
					ShortName: "eth",
					LongName:  "ethereum",
					Wraps:     nil,
				},
			}, nil
		},
	}
}

// ensures that networksRPC implements networks.Bridge.
var _ networks.Bridge = (*NetworksMock)(nil)

// NetworksMock provides access to the networks.Bridge.
type NetworksMock struct {
	connectedNetworksImpl func(ctx context.Context) ([]networks.Network, error)
	supportedTokensImpl   func(ctx context.Context, networkID uint32) ([]networks.Token, error)
}

// ConnectedNetworks returns all connected networks to the bridge.
func (networksMock *NetworksMock) ConnectedNetworks(ctx context.Context) ([]networks.Network, error) {
	return networksMock.connectedNetworksImpl(ctx)
}

// SetConnectedNetworks sets the mock implementation for ConnectedNetworks.
func (networksMock *NetworksMock) SetConnectedNetworks(impl func(ctx context.Context) ([]networks.Network, error)) {
	networksMock.connectedNetworksImpl = impl
}

// SupportedTokens returns all supported tokens by certain network.
func (networksMock *NetworksMock) SupportedTokens(ctx context.Context, networkID uint32) ([]networks.Token, error) {
	return networksMock.supportedTokensImpl(ctx, networkID)
}

// SetSupportedTokens sets the mock implementation for SupportedTokens.
func (networksMock *NetworksMock) SetSupportedTokens(impl func(ctx context.Context, networkID uint32) ([]networks.Token, error)) {
	networksMock.supportedTokensImpl = impl
}

// Transfers provides access to the transfers.Bridge rpc methods.
func (rpc *MockCommunication) Transfers() transfers.Bridge {
	return &transfersMock{
		estimateImpl: func(ctx context.Context, sender, recipient networks.Name, tokenID uint32, amount string) (transfers.Estimate, error) {
			return transfers.Estimate{
				Fee:                       "0.333",
				FeePercentage:             "4",
				EstimatedConfirmationTime: "47",
			}, nil
		},
		infoImpl: func(ctx context.Context, txHash string) ([]transfers.Transfer, error) {
			amount := new(big.Int)
			amount, ok := amount.SetString("1000000000000000000", 10)
			if !ok {
				return nil, nil
			}

			return []transfers.Transfer{
				{
					ID:     17,
					Amount: *amount,
					Sender: networks.Address{
						NetworkName: "GOERLI",
						Address:     "0xB7F14E1C560Fc97b08F1327329D59F6db5FD2009",
					},
					Recipient: networks.Address{
						NetworkName: "CASPER-TESTNET",
						Address:     "account-hash-3c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23",
					},
					Status: transfers.StatusFinished,
					TriggeringTx: transfers.StringTxHash{
						NetworkName: "GOERLI",
						Hash:        common.HexToHash("0x879e513fcdf956e8f65ba9267b1e278cb8afca733bd0b0f8456bc8fe6d2c3a62"),
					},
					OutboundTx: transfers.StringTxHash{
						NetworkName: "CASPER-TESTNET",
						Hash:        common.HexToHash("7e4b5e5419c26c224c4654fddf127e597fa9c966f9e41a4ae0b5702b3bd24abc"),
					},
					CreatedAt: time.Now().UTC().AddDate(0, 0, -1),
				},
			}, nil
		},
		cancelImpl: func(ctx context.Context, id transfers.ID, signature, pubKey []byte) error {
			return nil
		},
		historyImpl: func(ctx context.Context, offset, limit uint64, signature, pubKey []byte, networkID uint32) (transfers.Page, error) {
			amount := new(big.Int)
			amount, ok := amount.SetString("1000000000000000000", 10)
			if !ok {
				return transfers.Page{}, nil
			}

			return transfers.Page{
				Transfers: []transfers.Transfer{
					{
						ID:     17,
						Amount: *amount,
						Sender: networks.Address{
							NetworkName: "GOERLI",
							Address:     "0xB7F14E1C560Fc97b08F1327329D59F6db5FD2009",
						},
						Recipient: networks.Address{
							NetworkName: "CASPER-TESTNET",
							Address:     "account-hash-3c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23",
						},
						Status: transfers.StatusFinished,
						TriggeringTx: transfers.StringTxHash{
							NetworkName: "GOERLI",
							Hash:        common.HexToHash("0x879e513fcdf956e8f65ba9267b1e278cb8afca733bd0b0f8456bc8fe6d2c3a62"),
						},
						OutboundTx: transfers.StringTxHash{
							NetworkName: "CASPER-TESTNET",
							Hash:        common.HexToHash("7e4b5e5419c26c224c4654fddf127e597fa9c966f9e41a4ae0b5702b3bd24abc"),
						},
						CreatedAt: time.Now().UTC().AddDate(0, 0, -1),
					},
				},
				Offset:     0,
				Limit:      10,
				TotalCount: 1,
			}, nil
		},
		bridgeInSignatureImpl: func(ctx context.Context, req transfers.BridgeInSignatureRequest) (transfers.BridgeInSignatureResponse, error) {
			return transfers.BridgeInSignatureResponse{}, nil
		},
		cancelSignatureImpl: func(ctx context.Context, req transfers.CancelSignatureRequest) (transfers.CancelSignatureResponse, error) {
			return transfers.CancelSignatureResponse{}, nil
		},
	}
}

// ensures that transfersMock implements transfers.Bridge.
var _ transfers.Bridge = (*transfersMock)(nil)

// transfersMock provides access to the transfers.Bridge.
type transfersMock struct {
	estimateImpl          func(ctx context.Context, sender, recipient networks.Name, tokenID uint32, amount string) (transfers.Estimate, error)
	infoImpl              func(ctx context.Context, txHash string) ([]transfers.Transfer, error)
	cancelImpl            func(ctx context.Context, id transfers.ID, signature, pubKey []byte) error
	historyImpl           func(ctx context.Context, offset, limit uint64, signature, pubKey []byte, networkID uint32) (transfers.Page, error)
	bridgeInSignatureImpl func(ctx context.Context, req transfers.BridgeInSignatureRequest) (transfers.BridgeInSignatureResponse, error)
	cancelSignatureImpl   func(ctx context.Context, req transfers.CancelSignatureRequest) (transfers.CancelSignatureResponse, error)
}

// Estimate returns approximate information about transfer fee and time.
func (transfersMock *transfersMock) Estimate(ctx context.Context, sender, recipient networks.Name, tokenID uint32, amount string) (transfers.Estimate, error) {
	return transfersMock.estimateImpl(ctx, sender, recipient, tokenID, amount)
}

// SetEstimate sets Estimate mock implementation.
func (transfersMock *transfersMock) SetEstimate(impl func(ctx context.Context, sender, recipient networks.Name, tokenID uint32, amount string) (transfers.Estimate, error)) {
	transfersMock.estimateImpl = impl
}

// Info returns list of transfers of triggering transaction.
func (transfersMock *transfersMock) Info(ctx context.Context, txHash string) ([]transfers.Transfer, error) {
	return transfersMock.infoImpl(ctx, txHash)
}

// SetInfo sets Info mock implementation.
func (transfersMock *transfersMock) SetInfo(impl func(ctx context.Context, txHash string) ([]transfers.Transfer, error)) {
	transfersMock.infoImpl = impl
}

// Cancel cancels a transfer in the CONFIRMING status, returning the funds to the sender
// after deducting the commission for issuing the transaction.
func (transfersMock *transfersMock) Cancel(ctx context.Context, id transfers.ID, signature, pubKey []byte) error {
	return transfersMock.cancelImpl(ctx, id, signature, pubKey)
}

func (transfersMock *transfersMock) SetCancel(impl func(ctx context.Context, id transfers.ID, signature, pubKey []byte) error) {
	transfersMock.cancelImpl = impl
}

// History returns paginated list of transfers.
func (transfersMock *transfersMock) History(ctx context.Context, offset, limit uint64, signature, pubKey []byte, networkID uint32) (transfers.Page, error) {
	return transfersMock.historyImpl(ctx, offset, limit, signature, pubKey, networkID)
}

func (transfersMock *transfersMock) SeHistory(impl func(ctx context.Context, offset, limit uint64, signature, pubKey []byte, networkID uint32) (transfers.Page, error)) {
	transfersMock.historyImpl = impl
}

// BridgeInSignature returns signature for user to send bridgeIn transaction.
func (transfersMock *transfersMock) BridgeInSignature(ctx context.Context, req transfers.BridgeInSignatureRequest) (transfers.BridgeInSignatureResponse, error) {
	return transfersMock.bridgeInSignatureImpl(ctx, req)
}

func (transfersMock *transfersMock) BebridgeInSignature(impl func(ctx context.Context, req transfers.BridgeInSignatureRequest) (transfers.BridgeInSignatureResponse, error)) {
	transfersMock.bridgeInSignatureImpl = impl
}

// CancelSignature returns signature for user to send Cancel transaction.
func (transfersMock *transfersMock) CancelSignature(ctx context.Context, req transfers.CancelSignatureRequest) (transfers.CancelSignatureResponse, error) {
	return transfersMock.cancelSignatureImpl(ctx, req)
}

func (transfersMock *transfersMock) BeCancelSignature(impl func(ctx context.Context, req transfers.CancelSignatureRequest) (transfers.CancelSignatureResponse, error)) {
	transfersMock.cancelSignatureImpl = impl
}

// Close closes underlying rpc connection.
func (rpc *MockCommunication) Close() error {
	return nil
}
