package rpc

import (
	"context"
	"math/big"
	"strconv"

	bridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/gateway-bridge"
	transferspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/transfers"
	"github.com/ethereum/go-ethereum/common"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/transfers"
)

// ensures that transfersRPC implements transfers.Bridge.
var _ transfers.Bridge = (*transfersRPC)(nil)

// transfersRPC provides access to the transfers.Bridge.
type transfersRPC struct {
	client bridgepb.GatewayBridgeClient
}

// Estimate returns approximate information about transfer fee and time.
func (transfersRPC *transfersRPC) Estimate(ctx context.Context, sender, recipient networks.Name, tokenID uint32, amount string) (transfers.Estimate, error) {
	estimatepb, err := transfersRPC.client.EstimateTransfer(ctx, &transferspb.EstimateTransferRequest{
		SenderNetwork:    string(sender),
		RecipientNetwork: string(recipient),
		TokenId:          tokenID,
		Amount:           amount,
	})
	if err != nil {
		return transfers.Estimate{}, Error.Wrap(err)
	}

	estimatedConfirmationTime := strconv.FormatUint(uint64(estimatepb.GetEstimatedConfirmation()), 10)
	estimate := transfers.Estimate{
		Fee:                       estimatepb.GetFee(),
		FeePercentage:             estimatepb.GetFeePercentage(),
		EstimatedConfirmationTime: estimatedConfirmationTime,
	}

	return estimate, nil
}

// Info returns list of transfers of triggering transaction.
func (transfersRPC *transfersRPC) Info(ctx context.Context, txHash string) ([]transfers.Transfer, error) {
	pbTxHash := &transferspb.StringTxHash{
		Hash: txHash,
	}
	pbTransfers, err := transfersRPC.client.Transfer(ctx, &transferspb.TransferRequest{
		TxHash: pbTxHash,
	})
	if err != nil {
		return []transfers.Transfer{}, Error.Wrap(err)
	}

	txTransfers := make([]transfers.Transfer, 0, len(pbTransfers.GetStatuses()))
	for _, pbTransfer := range pbTransfers.GetStatuses() {
		amount := new(big.Int)
		amount, ok := amount.SetString(pbTransfer.GetAmount(), 10)
		if !ok {
			return nil, Error.New("could not convert amount to big.Int")
		}
		transfer := transfers.Transfer{
			ID:     transfers.ID(pbTransfer.GetId()),
			Amount: *amount,
			Sender: networks.Address{
				NetworkName: pbTransfer.GetSender().GetNetworkName(),
				Address:     pbTransfer.GetSender().GetAddress(),
			},
			Recipient: networks.Address{
				NetworkName: pbTransfer.GetRecipient().GetNetworkName(),
				Address:     pbTransfer.GetRecipient().GetAddress(),
			},
			TriggeringTx: transfers.StringTxHash{
				NetworkName: pbTransfer.GetTriggeringTx().GetNetworkName(),
				Hash:        common.HexToHash(pbTransfer.GetTriggeringTx().GetHash()),
			},
			OutboundTx: transfers.StringTxHash{
				NetworkName: pbTransfer.GetOutboundTx().GetNetworkName(),
				Hash:        common.HexToHash(pbTransfer.GetOutboundTx().GetHash()),
			},
			CreatedAt: pbTransfer.GetCreatedAt().AsTime(),
		}

		switch pbTransfer.GetStatus() {
		case transferspb.TransferResponse_STATUS_CONFIRMING:
			transfer.Status = transfers.StatusConfirming
		case transferspb.TransferResponse_STATUS_CANCELLED:
			transfer.Status = transfers.StatusCancelled
		case transferspb.TransferResponse_STATUS_FINISHED:
			transfer.Status = transfers.StatusFinished
		case transferspb.TransferResponse_STATUS_UNSPECIFIED:
			transfer.Status = ""
		}

		txTransfers = append(txTransfers, transfer)
	}

	return txTransfers, nil
}

// History returns paginated list of transfers.
func (transfersRPC *transfersRPC) History(ctx context.Context, offset, limit uint64, signature, pubKey []byte, networkID uint32) (transfers.Page, error) {
	transferHistoryResponse, err := transfersRPC.client.TransferHistory(ctx, &transferspb.TransferHistoryRequest{
		Offset:        offset,
		Limit:         limit,
		UserSignature: signature,
		NetworkId:     networkID,
		PublicKey:     pubKey,
	})
	if err != nil {
		return transfers.Page{}, Error.Wrap(err)
	}

	var history []transfers.Transfer
	for _, transferPb := range transferHistoryResponse.GetStatuses() {
		amount := new(big.Int)
		amount, ok := amount.SetString(transferPb.GetAmount(), 10)
		if !ok {
			return transfers.Page{}, Error.New("could not convert amount to big.Int")
		}
		history = append(history, transfers.Transfer{
			ID:     transfers.ID(transferPb.GetId()),
			Amount: *amount,
			Sender: networks.Address{
				NetworkName: transferPb.GetSender().GetNetworkName(),
				Address:     transferPb.GetSender().GetAddress(),
			},
			Recipient: networks.Address{
				NetworkName: transferPb.GetRecipient().GetNetworkName(),
				Address:     transferPb.GetRecipient().GetAddress(),
			},
			Status: transfers.Status(transferPb.Status),
			TriggeringTx: transfers.StringTxHash{
				NetworkName: transferPb.GetTriggeringTx().GetNetworkName(),
				Hash:        common.HexToHash(transferPb.GetTriggeringTx().GetHash()),
			},
			OutboundTx: transfers.StringTxHash{
				NetworkName: transferPb.GetOutboundTx().GetNetworkName(),
				Hash:        common.HexToHash(transferPb.GetOutboundTx().GetHash()),
			},
			CreatedAt: transferPb.GetCreatedAt().AsTime(),
		})
	}

	page := transfers.Page{
		Transfers:  history,
		Limit:      int64(limit),
		Offset:     int64(offset),
		TotalCount: int64(transferHistoryResponse.GetTotalSize()),
	}

	return page, nil
}

// BridgeInSignature returns signature for user to send bridgeIn transaction.
func (transfersRPC *transfersRPC) BridgeInSignature(ctx context.Context, req transfers.BridgeInSignatureRequest) (transfers.BridgeInSignatureResponse, error) {
	signatureResponse, err := transfersRPC.client.BridgeInSignature(ctx, &transferspb.BridgeInSignatureRequest{
		Sender: &transferspb.StringNetworkAddress{
			NetworkName: req.Sender.NetworkName,
			Address:     req.Sender.Address,
		},
		TokenId: req.TokenID,
		Amount:  req.Amount,
		Destination: &transferspb.StringNetworkAddress{
			NetworkName: req.Destination.NetworkName,
			Address:     req.Destination.Address,
		},
	})

	response := transfers.BridgeInSignatureResponse{
		Token:         signatureResponse.GetToken(),
		Amount:        signatureResponse.GetAmount(),
		GasCommission: signatureResponse.GetGasComission(),
		Destination: networks.Address{
			NetworkName: signatureResponse.GetDestination().GetNetworkName(),
			Address:     signatureResponse.GetDestination().GetAddress(),
		},
		Deadline:  signatureResponse.GetDeadline(),
		Nonce:     signatureResponse.GetNonce(),
		Signature: signatureResponse.GetSignature(),
	}

	return response, Error.Wrap(err)
}

// CancelSignature returns signature for user to return funds.
func (transfersRPC *transfersRPC) CancelSignature(ctx context.Context, req transfers.CancelSignatureRequest) (transfers.CancelSignatureResponse, error) {
	cancelTransferResponse, err := transfersRPC.client.CancelTransfer(ctx, &transferspb.CancelTransferRequest{
		TransferId: req.TransferID,
		Signature:  req.Signature,
		NetworkId:  req.NetworkID,
		PublicKey:  req.PublicKey,
	})
	if err != nil {
		return transfers.CancelSignatureResponse{}, Error.Wrap(err)
	}

	response := transfers.CancelSignatureResponse{
		Status:     cancelTransferResponse.Status,
		Nonce:      cancelTransferResponse.Nonce,
		Signature:  cancelTransferResponse.Signature,
		Token:      cancelTransferResponse.Token,
		Recipient:  cancelTransferResponse.Recipient,
		Commission: cancelTransferResponse.Commission,
		Amount:     cancelTransferResponse.Amount,
	}

	return response, nil
}
