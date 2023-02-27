// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package controllers

import (
	"context"
	"errors"

	"github.com/zeebo/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	gatewaybridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/gateway-bridge"
	networkspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/networks"
	transferspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/transfers"

	"tricorn/bridge"
	"tricorn/bridge/networks"
	"tricorn/bridge/transfers"
	"tricorn/internal/logger"
)

// ensures that Gateway implements gatewaybridgepb.GatewayBridgeServer.
var _ gatewaybridgepb.GatewayBridgeServer = (*Gateway)(nil)

// Error is an internal error type for gateway controller.
var Error = errs.Class("gateway controller")

// statusPrefixString is the first part of PbTransferStatus.
const statusPrefixString = "STATUS_"

// Gateway is controller that handles all gateway related routes.
type Gateway struct {
	log    logger.Logger
	bridge *bridge.Service
}

// NewGateway is a constructor for gateway controller.
func NewGateway(log logger.Logger, bridge *bridge.Service) *Gateway {
	gatewayController := &Gateway{
		log:    log,
		bridge: bridge,
	}

	return gatewayController
}

// ConnectedNetworks returns a list of all networks this bridge is connected to.
func (gateway *Gateway) ConnectedNetworks(ctx context.Context, empty *emptypb.Empty) (*networkspb.ConnectedNetworksResponse, error) {
	var resp networkspb.ConnectedNetworksResponse

	connectedNetworks, err := gateway.bridge.ListConnectedNetworks(ctx)
	if err != nil {
		gateway.log.Error("couldn't get connected networks", err)
		return &resp, status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	for _, network := range connectedNetworks {
		var ntype networkspb.NetworkType
		switch network.Type {
		case networks.TypeCasper:
			ntype = networkspb.NetworkType_NT_CASPER
		case networks.TypeEVM:
			ntype = networkspb.NetworkType_NT_EVM
		case networks.TypeSolana:
			ntype = networkspb.NetworkType_NT_SOLANA
		}

		respNetwork := networkspb.Network{
			Id:             uint32(network.ID),
			Name:           network.Name.String(),
			Type:           ntype,
			IsTestnet:      network.IsTestnet,
			NodeAddress:    network.NodeAddress,
			TokenContract:  network.TokenContract,
			BridgeContract: network.BridgeContract,
			GasLimit:       network.GasLimit,
		}

		resp.Networks = append(resp.Networks, &respNetwork)
	}

	return &resp, nil
}

// SupportedTokens returns a list of all tokens supported by particular network.
func (gateway *Gateway) SupportedTokens(ctx context.Context, request *networkspb.SupportedTokensRequest) (*networkspb.TokensResponse, error) {
	var resp networkspb.TokensResponse

	supportedTokens, err := gateway.bridge.ListSupportedTokens(ctx, request.GetNetworkId())
	if err != nil {
		if errors.Is(err, bridge.ErrNotConnectedNetwork) {
			gateway.log.Error("invalid network", err)
			return &resp, status.Error(codes.NotFound, Error.Wrap(err).Error())
		}

		if errors.Is(err, networks.ErrTransactionNameInvalid) {
			gateway.log.Error("invalid request", err)
			return &resp, status.Error(codes.InvalidArgument, Error.Wrap(err).Error())
		}

		gateway.log.Error("couldn't get supported tokens", err)
		return &resp, status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	for _, supportedToken := range supportedTokens {
		var addresses []*networkspb.TokensResponse_TokenAddress
		for _, supportedAddress := range supportedToken.Addresses {
			address := networkspb.TokensResponse_TokenAddress{
				NetworkId: uint32(supportedAddress.NetworkID),
				Address:   networks.BytesToString(supportedAddress.NetworkID, supportedAddress.ContractAddress),
				Decimals:  uint32(supportedAddress.Decimals),
			}

			addresses = append(addresses, &address)
		}

		token := networkspb.TokensResponse_Token{
			Id:        uint32(supportedToken.ID),
			ShortName: supportedToken.ShortName,
			LongName:  supportedToken.LongName,
			Addresses: addresses,
		}

		resp.Tokens = append(resp.Tokens, &token)
	}

	return &resp, nil
}

// EstimateTransfer estimates a potential transfer.
func (gateway *Gateway) EstimateTransfer(ctx context.Context, request *transferspb.EstimateTransferRequest) (*transferspb.EstimateTransferResponse, error) {
	var resp transferspb.EstimateTransferResponse

	estimateTransfer, err := gateway.bridge.EstimateTransfer(ctx, transfers.EstimateTransfer{
		SenderNetwork:    request.GetSenderNetwork(),
		RecipientNetwork: request.GetRecipientNetwork(),
		TokenID:          request.GetTokenId(),
		Amount:           request.GetAmount(),
	})
	if err != nil {
		if errors.Is(err, networks.ErrTransactionNameInvalid) || errors.Is(err, bridge.ErrInvalidAmount) {
			gateway.log.Error("invalid request", err)
			return &resp, status.Error(codes.InvalidArgument, Error.Wrap(err).Error())
		}

		gateway.log.Error("couldn't estimate transfer", err)
		return &resp, status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	resp.Fee = estimateTransfer.Fee
	resp.FeePercentage = estimateTransfer.FeePercentage
	resp.EstimatedConfirmation = estimateTransfer.EstimatedConfirmation

	return &resp, nil
}

// Transfer returns status of transfer.
func (gateway *Gateway) Transfer(ctx context.Context, request *transferspb.TransferRequest) (*transferspb.TransferResponse, error) {
	var resp transferspb.TransferResponse

	if request.GetTxHash() == nil {
		return &resp, status.Error(codes.InvalidArgument, Error.New("received empty request").Error())
	}

	transfersInfo, err := gateway.bridge.TransfersInfo(ctx, request.GetTxHash().GetNetworkName(), request.GetTxHash().GetHash())
	if err != nil {
		if errors.Is(err, networks.ErrTransactionNameInvalid) {
			gateway.log.Error("invalid network", err)
			return &resp, status.Error(codes.InvalidArgument, Error.Wrap(err).Error())
		}

		gateway.log.Error("couldn't get transfers info", err)
		return &resp, status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	for _, transfer := range transfersInfo {
		respTransfer := convertToPbTransfer(transfer)
		resp.Statuses = append(resp.Statuses, &respTransfer)
	}

	return &resp, nil
}

// convertToPbTransfer converts transfers.Transfer to transferspb.TransferResponse_Transfer.
func convertToPbTransfer(transfer transfers.Transfer) transferspb.TransferResponse_Transfer {
	triggeringTxHash := networks.BytesToString(networks.NetworkNameToID[networks.Name(transfer.TriggeringTx.NetworkName)], transfer.TriggeringTx.Hash.Bytes())
	outboundTxHash := networks.BytesToString(networks.NetworkNameToID[networks.Name(transfer.OutboundTx.NetworkName)], transfer.OutboundTx.Hash.Bytes())

	return transferspb.TransferResponse_Transfer{
		Id:     uint64(transfer.ID),
		Amount: transfer.Amount.String(),
		Sender: &transferspb.StringNetworkAddress{
			NetworkName: transfer.Sender.NetworkName,
			Address:     transfer.Sender.Address,
		},
		Recipient: &transferspb.StringNetworkAddress{
			NetworkName: transfer.Recipient.NetworkName,
			Address:     transfer.Recipient.Address,
		},
		Status: convertToPbTransferStatus(transfer.Status),
		TriggeringTx: &transferspb.StringTxHash{
			NetworkName: transfer.TriggeringTx.NetworkName,
			Hash:        triggeringTxHash,
		},
		OutboundTx: &transferspb.StringTxHash{
			NetworkName: transfer.OutboundTx.NetworkName,
			Hash:        outboundTxHash,
		},
		CreatedAt: &timestamppb.Timestamp{
			Seconds: transfer.CreatedAt.UTC().Unix(),
			Nanos:   int32(transfer.CreatedAt.UTC().Nanosecond()),
		},
	}
}

// convertToPbTransferStatus converts transfers.Status to transferspb.TransferResponse_Status.
func convertToPbTransferStatus(status transfers.Status) transferspb.TransferResponse_Status {
	statusString := statusPrefixString + string(status)
	statusValue := transferspb.TransferResponse_Status_value[statusString]

	return transferspb.TransferResponse_Status(statusValue)
}

// CancelTransfer generates and returns a signature to cancel a pending transfer.
func (gateway *Gateway) CancelTransfer(ctx context.Context, request *transferspb.CancelTransferRequest) (*transferspb.CancelTransferResponse, error) {
	var resp transferspb.CancelTransferResponse

	cancelTransfer, err := gateway.bridge.CancelTransfer(ctx, transfers.CancelSignatureRequest{
		TransferID: request.GetTransferId(),
		Signature:  request.GetSignature(),
		NetworkID:  request.GetNetworkId(),
		PublicKey:  request.GetPublicKey(),
	})
	if err != nil {
		if errors.Is(err, bridge.ErrNotConnectedNetwork) {
			gateway.log.Error("invalid network", err)
			return &resp, status.Error(codes.NotFound, Error.Wrap(err).Error())
		}

		if errors.Is(err, networks.ErrTransactionNameInvalid) {
			gateway.log.Error("invalid request", err)
			return &resp, status.Error(codes.InvalidArgument, Error.Wrap(err).Error())
		}

		gateway.log.Error("couldn't get cancel signature", err)
		return &resp, status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	resp.Status = cancelTransfer.Status
	resp.Nonce = cancelTransfer.Nonce
	resp.Signature = cancelTransfer.Signature
	resp.Token = cancelTransfer.Token
	resp.Recipient = cancelTransfer.Recipient
	resp.Commission = cancelTransfer.Commission
	resp.Amount = cancelTransfer.Amount

	return &resp, nil
}

// TransferHistory returns paginated transfer history for user.
func (gateway *Gateway) TransferHistory(ctx context.Context, request *transferspb.TransferHistoryRequest) (*transferspb.TransferHistoryResponse, error) {
	var resp transferspb.TransferHistoryResponse

	page, err := gateway.bridge.History(ctx, request.GetOffset(), request.GetLimit(), request.GetUserSignature(), request.GetPublicKey(), request.GetNetworkId())
	if err != nil {
		if errors.Is(err, bridge.ErrNotConnectedNetwork) {
			gateway.log.Error("invalid network", err)
			return &resp, status.Error(codes.NotFound, Error.Wrap(err).Error())
		}

		if errors.Is(err, networks.ErrTransactionNameInvalid) {
			gateway.log.Error("invalid request", err)
			return &resp, status.Error(codes.InvalidArgument, Error.Wrap(err).Error())
		}

		gateway.log.Error("couldn't get transfer history", err)
		return &resp, status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	resp.TotalSize = uint64(page.TotalCount)
	for _, transfer := range page.Transfers {
		respTransfer := convertToPbTransfer(transfer)
		resp.Statuses = append(resp.Statuses, &respTransfer)
	}

	return &resp, nil
}

// BridgeInSignature returns signature for user to send bridgeIn transaction.
func (gateway *Gateway) BridgeInSignature(ctx context.Context, request *transferspb.BridgeInSignatureRequest) (*transferspb.BridgeInSignatureResponse, error) {
	var resp transferspb.BridgeInSignatureResponse

	signature, err := gateway.bridge.GetBridgeInSignature(ctx, transfers.BridgeInSignatureRequest{
		Sender: networks.Address{
			NetworkName: request.Sender.GetNetworkName(),
			Address:     request.Sender.GetAddress(),
		},
		TokenID: request.GetTokenId(),
		Amount:  request.GetAmount(),
		Destination: networks.Address{
			NetworkName: request.Destination.GetNetworkName(),
			Address:     request.Destination.GetAddress(),
		},
	})
	if err != nil {
		if errors.Is(err, networks.ErrTransactionNameInvalid) || errors.Is(err, bridge.ErrInvalidAmount) {
			gateway.log.Error("invalid request", err)
			return &resp, status.Error(codes.InvalidArgument, Error.Wrap(err).Error())
		}

		gateway.log.Error("couldn't get bridge-in signature", err)
		return &resp, status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	resp.Token = signature.Token
	resp.Amount = signature.Amount
	resp.GasCommission = signature.GasCommission
	resp.Destination = &transferspb.StringNetworkAddress{
		NetworkName: signature.Destination.NetworkName,
		Address:     signature.Destination.Address,
	}
	resp.Deadline = signature.Deadline
	resp.Nonce = signature.Nonce.Uint64()
	resp.Signature = signature.Signature

	return &resp, nil
}
