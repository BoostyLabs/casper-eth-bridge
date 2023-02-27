// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package rpc

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	bridgeconnectorpb "github.com/BoostyLabs/golden-gate-communication/go-gen/bridge-connector"
	connectorpb "github.com/BoostyLabs/golden-gate-communication/go-gen/connector"
	networkspb "github.com/BoostyLabs/golden-gate-communication/go-gen/networks"
	pb_transfers "github.com/BoostyLabs/golden-gate-communication/go-gen/transfers"

	"tricorn/bridge"
	"tricorn/bridge/networks"
	"tricorn/bridge/transfers"
	"tricorn/chains"
)

// ensures that connectorRPC implements bridge.Connector.
var _ bridge.Connector = (*connectorRPC)(nil)

// connectorRPC provides access to the Connector.
type connectorRPC struct {
	gctx context.Context

	client bridgeconnectorpb.ConnectorClient

	mutex            sync.Mutex
	eventSubscribers []bridge.EventSubscriber
}

// Network returns supported by connector network.
func (connectorRPC *connectorRPC) Network(ctx context.Context) (networks.Network, error) {
	network, err := connectorRPC.client.Network(ctx, &emptypb.Empty{})
	if err != nil {
		return networks.Network{}, Error.Wrap(err)
	}
	var ntype networks.Type
	switch network.Type {
	case networkspb.NetworkType_NT_CASPER:
		ntype = networks.TypeCasper
	case networkspb.NetworkType_NT_EVM:
		ntype = networks.TypeEVM
	case networkspb.NetworkType_NT_SOLANA:
		ntype = networks.TypeSolana
	}
	return networks.Network{
		ID:             networks.ID(network.GetId()),
		Name:           networks.Name(network.GetName()),
		Type:           ntype,
		IsTestnet:      network.GetIsTestnet(),
		NodeAddress:    network.GetNodeAddress(),
		TokenContract:  network.GetTokenContract(),
		BridgeContract: network.GetBridgeContract(),
		GasLimit:       network.GetGasLimit(),
	}, nil
}

// KnownTokens returns tokens known by this connector.
func (connectorRPC *connectorRPC) KnownTokens(ctx context.Context) (chains.Tokens, error) {
	tokensResponse, err := connectorRPC.client.KnownTokens(ctx, &emptypb.Empty{})
	if err != nil {
		return chains.Tokens{}, Error.Wrap(err)
	}

	var tokens []chains.Token
	for _, token := range tokensResponse.GetTokens() {
		tokens = append(tokens, chains.Token{
			ID:      token.GetId(),
			Address: token.GetAddress().GetAddress(),
		})
	}

	return chains.Tokens{
		Tokens: tokens,
	}, nil
}

// EventStream initiates event stream from the network.
func (connectorRPC *connectorRPC) EventStream(ctx context.Context, fromBlock uint64) error {
	stream, err := connectorRPC.client.EventStream(ctx, &connectorpb.EventsRequest{
		BlockNumber: &fromBlock,
	})
	if err != nil {
		return Error.Wrap(err)
	}

	for {
		select {
		case <-connectorRPC.gctx.Done():
			return nil
		case <-ctx.Done():
			return nil
		default:
		}

		pbEvent, err := stream.Recv()
		if err != nil {
			return err
		}

		connectorRPC.Notify(ctx, toEventVariant(pbEvent))
	}
}

// BridgeOut initiates outbound bridge transaction.
func (connectorRPC *connectorRPC) BridgeOut(ctx context.Context, req chains.TokenOutRequest) (chains.TokenOutResponse, error) {
	bridgeOutResponse, err := connectorRPC.client.BridgeOut(ctx, &connectorpb.TokenOutRequest{
		Amount: req.Amount.String(),
		Token: &connectorpb.Address{
			Address: req.Token,
		},
		To: &connectorpb.Address{
			Address: req.To,
		},
		From: &pb_transfers.StringNetworkAddress{
			NetworkName: req.From.NetworkName,
			Address:     req.From.Address,
		},
		TransactionId: req.TransactionID.Uint64(),
	})
	if err != nil {
		return chains.TokenOutResponse{}, Error.Wrap(err)
	}

	return chains.TokenOutResponse{
		Txhash: bridgeOutResponse.GetTxhash(),
	}, nil
}

// EstimateTransfer estimates a potential transfer.
func (connectorRPC *connectorRPC) EstimateTransfer(ctx context.Context, req transfers.EstimateTransfer) (chains.Estimation, error) {
	estimation, err := connectorRPC.client.EstimateTransfer(ctx, &pb_transfers.EstimateTransferRequest{
		SenderNetwork:    req.SenderNetwork,
		RecipientNetwork: req.RecipientNetwork,
		TokenId:          req.TokenID,
		Amount:           req.Amount,
	})
	if err != nil {
		return chains.Estimation{}, Error.Wrap(err)
	}

	return chains.Estimation{
		Fee:                   estimation.GetFee(),
		FeePercentage:         estimation.GetFeePercentage(),
		EstimatedConfirmation: estimation.GetEstimatedConfirmation(),
	}, nil
}

// BridgeInSignature returns signature for user to send bridgeIn transaction.
func (connectorRPC *connectorRPC) BridgeInSignature(ctx context.Context, req bridge.BridgeInSignatureRequest) (bridge.BridgeInSignatureResponse, error) {
	signature, err := connectorRPC.client.BridgeInSignature(ctx, &pb_transfers.BridgeInSignatureWithNonceRequest{
		Sender: req.User,
		Token:  req.Token,
		Nonce:  req.Nonce.Uint64(),
		Amount: req.Amount.String(),
		Destination: &pb_transfers.StringNetworkAddress{
			NetworkName: req.Destination.NetworkName,
			Address:     req.Destination.Address,
		},
		GasCommission: req.GasCommission.String(),
	})
	if err != nil {
		return bridge.BridgeInSignatureResponse{}, Error.Wrap(err)
	}

	return bridge.BridgeInSignatureResponse{
		Token:         signature.GetToken(),
		Amount:        signature.GetAmount(),
		GasCommission: signature.GetGasCommission(),
		Destination: networks.Address{
			NetworkName: signature.GetDestination().GetNetworkName(),
			Address:     signature.GetDestination().GetAddress(),
		},
		Deadline:  signature.GetDeadline(),
		Nonce:     req.Nonce,
		Signature: signature.GetSignature(),
	}, nil
}

// CancelSignature returns signature for user to return funds.
func (connectorRPC *connectorRPC) CancelSignature(ctx context.Context, req chains.CancelSignatureRequest) (chains.CancelSignatureResponse, error) {
	signature, err := connectorRPC.client.CancelSignature(ctx, &pb_transfers.CancelSignatureRequest{
		Nonce:      req.Nonce.Uint64(),
		Token:      req.Token.Bytes(),
		Recipient:  req.Recipient.Bytes(),
		Commission: req.Commission.String(),
		Amount:     req.Amount.String(),
	})
	if err != nil {
		return chains.CancelSignatureResponse{}, Error.Wrap(err)
	}

	return chains.CancelSignatureResponse{
		Signature: signature.GetSignature(),
	}, nil
}

// AddEventSubscriber adds subscriber to event publisher.
func (connectorRPC *connectorRPC) AddEventSubscriber() bridge.EventSubscriber {
	subscriber := bridge.EventSubscriber{
		ID:         uuid.New(),
		EventsChan: make(chan chains.EventVariant),
	}

	connectorRPC.mutex.Lock()
	defer connectorRPC.mutex.Unlock()
	connectorRPC.eventSubscribers = append(connectorRPC.eventSubscribers, subscriber)

	return subscriber
}

// RemoveEventSubscriber removes subscriber.
func (connectorRPC *connectorRPC) RemoveEventSubscriber(id uuid.UUID) {
	connectorRPC.mutex.Lock()
	defer connectorRPC.mutex.Unlock()

	subIndex := 0
	for index, subscriber := range connectorRPC.eventSubscribers {
		if subscriber.GetID() == id {
			subIndex = index
			break
		}
	}

	copy(connectorRPC.eventSubscribers[subIndex:], connectorRPC.eventSubscribers[subIndex+1:])
	connectorRPC.eventSubscribers = connectorRPC.eventSubscribers[:len(connectorRPC.eventSubscribers)-1]
}

// Notify notifies all subscribers with events.
func (connectorRPC *connectorRPC) Notify(ctx context.Context, event chains.EventVariant) {
	connectorRPC.mutex.Lock()
	defer connectorRPC.mutex.Unlock()

	for _, subscriber := range connectorRPC.eventSubscribers {
		select {
		case <-connectorRPC.gctx.Done():
			return
		case <-ctx.Done():
			return
		default:
			subscriber.NotifyWithEvent(event)
		}
	}
}

// toEventVariant converts *connectorpb.Event to chains.EventVariant.
func toEventVariant(pbEvent *connectorpb.Event) chains.EventVariant {
	var eventVariant chains.EventVariant

	if pbEvent.GetFundsIn() != nil {
		eventVariant = chains.EventVariant{
			Type: chains.EventTypeIn,
			EventFundsIn: chains.EventFundsIn{
				From: pbEvent.GetFundsIn().GetFrom().GetAddress(),
				To: networks.Address{
					NetworkName: pbEvent.GetFundsIn().GetTo().NetworkName,
					Address:     pbEvent.GetFundsIn().GetTo().Address,
				},
				Amount: pbEvent.GetFundsIn().GetAmount(),
				Token:  pbEvent.GetFundsIn().GetToken().GetAddress(),
				Tx: chains.TransactionInfo{
					Hash:        pbEvent.GetFundsIn().GetTx().GetHash(),
					BlockNumber: pbEvent.GetFundsIn().GetTx().GetBlocknumber(),
					Sender:      pbEvent.GetFundsIn().GetTx().GetSender(),
				},
			},
		}
	}

	if pbEvent.GetFundsOut() != nil {
		eventVariant = chains.EventVariant{
			Type: chains.EventTypeOut,
			EventFundsOut: chains.EventFundsOut{
				From: networks.Address{
					NetworkName: pbEvent.GetFundsOut().GetFrom().NetworkName,
					Address:     pbEvent.GetFundsOut().GetFrom().Address,
				},
				To:     pbEvent.GetFundsOut().GetTo().GetAddress(),
				Amount: pbEvent.GetFundsOut().GetAmount(),
				Token:  pbEvent.GetFundsOut().GetToken().GetAddress(),
				Tx: chains.TransactionInfo{
					Hash:        pbEvent.GetFundsOut().GetTx().GetHash(),
					BlockNumber: pbEvent.GetFundsOut().GetTx().GetBlocknumber(),
					Sender:      pbEvent.GetFundsOut().GetTx().GetSender(),
				},
			},
		}
	}

	return eventVariant
}
