// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package controllers

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	bridgeconnectorpb "github.com/BoostyLabs/golden-gate-communication/go-gen/bridge-connector"
	connectorpb "github.com/BoostyLabs/golden-gate-communication/go-gen/connector"
	networkspb "github.com/BoostyLabs/golden-gate-communication/go-gen/networks"
	transferspb "github.com/BoostyLabs/golden-gate-communication/go-gen/transfers"

	"tricorn/bridge/networks"
	"tricorn/chains"
	"tricorn/internal/logger"
)

// ensures that connector implements bridgeconnectorpb.ConnectorServer.
var _ bridgeconnectorpb.ConnectorServer = (*Connector)(nil)

// Error is an internal error type for connector controller.
var Error = errs.Class("connector controller")

// Connector is controller that handles all connector related methods.
type Connector struct {
	gctx context.Context
	log  logger.Logger

	connector chains.Connector
}

// NewConnector is a constructor for connector controller.
func NewConnector(globalCtx context.Context, log logger.Logger, connector chains.Connector) *Connector {
	connectorController := &Connector{
		gctx:      globalCtx,
		log:       log,
		connector: connector,
	}

	return connectorController
}

// Network returns supported by connector network.
func (s *Connector) Network(ctx context.Context, req *emptypb.Empty) (*networkspb.Network, error) {
	network := s.connector.Network(ctx)

	var ntype networkspb.NetworkType
	switch network.Type {
	case networks.TypeCasper:
		ntype = networkspb.NetworkType_NT_CASPER
	case networks.TypeEVM:
		ntype = networkspb.NetworkType_NT_EVM
	case networks.TypeSolana:
		ntype = networkspb.NetworkType_NT_SOLANA
	}
	return &networkspb.Network{
		Id:             uint32(network.ID),
		Name:           network.Name.String(),
		Type:           ntype,
		IsTestnet:      network.IsTestnet,
		NodeAddress:    network.NodeAddress,
		TokenContract:  network.TokenContract,
		BridgeContract: network.BridgeContract,
		GasLimit:       network.GasLimit,
	}, nil
}

// KnownTokens returns tokens known by this connector.
func (s *Connector) KnownTokens(ctx context.Context, req *emptypb.Empty) (*connectorpb.ConnectorTokens, error) {
	var resp connectorpb.ConnectorTokens

	tokens := s.connector.KnownTokens(ctx)

	for _, token := range tokens.Tokens {
		t := connectorpb.ConnectorTokens_ConnectorToken{
			Id: token.ID,
			Address: &connectorpb.Address{
				Address: token.Address,
			},
		}
		resp.Tokens = append(resp.Tokens, &t)
	}

	return &resp, nil
}

// EventStream initiates event stream from the network.
func (s *Connector) EventStream(req *connectorpb.EventsRequest, stream bridgeconnectorpb.Connector_EventStreamServer) error {
	s.log.Debug(fmt.Sprintf("time: %s, connected to EventStream with block number %d", time.Now().Format(time.RFC1123), req.GetBlockNumber()))
	s.log.Debug("")

	group, ctx := errgroup.WithContext(stream.Context())

	group.Go(func() error {
		err := s.connector.ReadEvents(ctx, req.GetBlockNumber())
		if err != nil {
			s.log.Error("couldn't read events", err)
			return status.Error(codes.Internal, err.Error())
		}

		return nil
	})

	group.Go(func() error {
		subscriber := s.connector.AddEventSubscriber()

		for {
			select {
			case eventFund, ok := <-subscriber.ReceiveEvents():
				if !ok {
					err := Error.New("events chan unexpectedly closed")
					s.log.Error("", err)
					return status.Error(codes.Internal, err.Error())
				}

				var resp connectorpb.Event
				switch eventFund.Type {
				case chains.EventTypeIn:
					resp = connectorpb.Event{
						Variant: &connectorpb.Event_FundsIn{
							FundsIn: &connectorpb.EventFundsIn{
								From: &connectorpb.Address{
									Address: eventFund.EventFundsIn.From,
								},
								To: &transferspb.StringNetworkAddress{
									NetworkName: eventFund.EventFundsIn.To.NetworkName,
									Address:     eventFund.EventFundsIn.To.Address,
								},
								Amount: eventFund.EventFundsIn.Amount,
								Token: &connectorpb.Address{
									Address: eventFund.EventFundsIn.Token,
								},
								Tx: &connectorpb.TransactionInfo{
									Hash:        eventFund.EventFundsIn.Tx.Hash,
									Blocknumber: eventFund.EventFundsIn.Tx.BlockNumber,
									Sender:      eventFund.EventFundsIn.Tx.Sender,
								},
							},
						},
					}
				case chains.EventTypeOut:
					resp = connectorpb.Event{
						Variant: &connectorpb.Event_FundsOut{
							FundsOut: &connectorpb.EventFundsOut{
								From: &transferspb.StringNetworkAddress{
									NetworkName: eventFund.EventFundsOut.From.NetworkName,
									Address:     eventFund.EventFundsOut.From.Address,
								},
								To: &connectorpb.Address{
									Address: eventFund.EventFundsOut.To,
								},
								Amount: eventFund.EventFundsOut.Amount,
								Token: &connectorpb.Address{
									Address: eventFund.EventFundsOut.Token,
								},
								Tx: &connectorpb.TransactionInfo{
									Hash:        eventFund.EventFundsOut.Tx.Hash,
									Blocknumber: eventFund.EventFundsOut.Tx.BlockNumber,
									Sender:      eventFund.EventFundsOut.Tx.Sender,
								},
							},
						},
					}
				default:
					err := Error.New("invalid event type")
					s.log.Error("", err)
					return status.Error(codes.Internal, err.Error())
				}

				s.logEvent(eventFund.Type, &resp)

				if err := stream.Send(&resp); err != nil {
					s.log.Error("couldn't send event fund", Error.Wrap(err))
					return status.Error(codes.Internal, Error.Wrap(err).Error())
				}
			case <-s.gctx.Done():
				s.connector.RemoveEventSubscriber(subscriber.GetID())

				return nil
			case <-ctx.Done():
				s.connector.RemoveEventSubscriber(subscriber.GetID())

				return nil
			}
		}
	})

	return group.Wait()
}

// BridgeOut initiates outbound bridge transaction.
func (s *Connector) BridgeOut(ctx context.Context, req *connectorpb.TokenOutRequest) (*connectorpb.TokenOutResponse, error) {
	s.logBridgeOut(req)

	amount, ok := big.NewInt(0).SetString(req.Amount, 10)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, Error.New(fmt.Sprintf("couldn't set int as %s to big.Int", req.Amount)).Error())
	}

	transactionID := big.NewInt(0).SetUint64(req.TransactionId)

	tokenOutRequest := chains.TokenOutRequest{
		Amount: amount,
		Token:  req.Token.Address,
		To:     req.To.Address,
		From: networks.Address{
			NetworkName: req.From.NetworkName,
			Address:     req.From.Address,
		},
		TransactionID: transactionID,
	}

	txhash, err := s.connector.BridgeOut(ctx, tokenOutRequest)
	if err != nil {
		s.log.Error("couldn't sent transaction to bridge out", Error.Wrap(err))
		return nil, status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	s.log.Debug(fmt.Sprintf("tx hash - %s", hex.EncodeToString(txhash)))
	s.log.Debug("")

	resp := connectorpb.TokenOutResponse{
		Txhash: txhash,
	}

	return &resp, nil
}

// EstimateTransfer estimates a potential transfer.
func (s *Connector) EstimateTransfer(ctx context.Context, request *transferspb.EstimateTransferRequest) (*transferspb.EstimateTransferResponse, error) {
	if s.connector.GetChainName().String() != request.RecipientNetwork {
		return nil, status.Error(codes.InvalidArgument, "recipient network is invalid")
	}

	estimatedTransfer, err := s.connector.EstimateTransfer(ctx)
	if err != nil {
		s.log.Error("could not estimate transfer", Error.Wrap(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := transferspb.EstimateTransferResponse{
		Fee:                   estimatedTransfer.Fee,
		FeePercentage:         estimatedTransfer.FeePercentage,
		EstimatedConfirmation: estimatedTransfer.EstimatedConfirmation,
	}

	return &response, nil
}

// BridgeInSignature returns signature for user to send bridgeIn transaction.
func (s *Connector) BridgeInSignature(ctx context.Context, req *transferspb.BridgeInSignatureWithNonceRequest) (*transferspb.BridgeInSignatureResponse, error) {
	amount, ok := big.NewInt(0).SetString(req.Amount, 10)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, Error.New(fmt.Sprintf("couldn't set amount %s to big.Int", req.Amount)).Error())
	}

	gasCommission, ok := big.NewInt(0).SetString(req.GasCommission, 10)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, Error.New(fmt.Sprintf("couldn't set gas commission %s to big.Int", req.GasCommission)).Error())
	}

	// TODO: uncomment after casper fixing.
	// amount, ok := new(big.Float).SetString(req.Amount)
	// if !ok {
	// 	return nil, Error.New("amount is invalid")
	// }
	// // TODO: give wei from bridge db.
	// amountInWei, _ := new(big.Float).Mul(amount, new(big.Float).SetUint64(wei)).Int64()
	// amountBigInt := big.NewInt(amountInWei).

	nonce := big.NewInt(0).SetUint64(req.Nonce)

	getSignature := chains.BridgeInSignatureRequest{
		User:   req.GetSender(),
		Nonce:  nonce,
		Token:  hex.EncodeToString(req.GetToken()),
		Amount: amount,
		Destination: networks.Address{
			NetworkName: req.GetDestination().GetNetworkName(),
			Address:     req.GetDestination().GetAddress(),
		},
		GasCommission: gasCommission,
	}

	signatureResponse, err := s.connector.BridgeInSignature(ctx, getSignature)
	if err != nil {
		s.log.Error("could not get signature", Error.Wrap(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	token, err := hex.DecodeString(signatureResponse.Token)
	if err != nil {
		s.log.Error("could not decode token", Error.Wrap(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := transferspb.BridgeInSignatureResponse{
		Token:         token,
		Amount:        signatureResponse.Amount.String(),
		GasCommission: signatureResponse.GasCommission,
		Destination: &transferspb.StringNetworkAddress{
			NetworkName: signatureResponse.Destination.NetworkName,
			Address:     signatureResponse.Destination.Address,
		},
		Deadline:  signatureResponse.Deadline,
		Nonce:     signatureResponse.Nonce.Uint64(),
		Signature: signatureResponse.Signature,
	}

	return &response, nil
}

// CancelSignature returns signature for user to return funds.
func (s *Connector) CancelSignature(ctx context.Context, req *transferspb.CancelSignatureRequest) (*transferspb.CancelSignatureResponse, error) {
	commission, ok := new(big.Int).SetString(req.Commission, 10)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, Error.New(fmt.Sprintf("couldn't set commission %s to big.Int", req.Commission)).Error())
	}

	amount, ok := big.NewInt(0).SetString(req.Amount, 10)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, Error.New(fmt.Sprintf("couldn't set amount %s to big.Int", req.Amount)).Error())
	}

	cancelSignature := chains.CancelSignatureRequest{
		Nonce:      new(big.Int).SetUint64(req.Nonce),
		Token:      common.BytesToAddress(req.Token),
		Recipient:  common.BytesToAddress(req.Recipient),
		Commission: commission,
		Amount:     amount,
	}
	signatureResponse, err := s.connector.CancelSignature(ctx, cancelSignature)
	if err != nil {
		s.log.Error("could not get signature", Error.Wrap(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := transferspb.CancelSignatureResponse{
		Signature: signatureResponse.Signature,
	}
	return &response, nil
}

func (s *Connector) logEvent(eventType chains.EventType, event *connectorpb.Event) {
	s.log.Debug(fmt.Sprintf("time: %s, send event to bridge with params: ", time.Now().Format(time.RFC1123)))
	s.log.Debug(fmt.Sprintf("event type: %d", eventType))

	switch eventType {
	case chains.EventTypeIn:
		s.log.Debug(fmt.Sprintf("from: %s", hex.EncodeToString(event.GetFundsIn().GetFrom().GetAddress())))
		s.log.Debug(fmt.Sprintf("to network name: %s", event.GetFundsIn().GetTo().GetNetworkName()))
		s.log.Debug(fmt.Sprintf("to address: %s", event.GetFundsIn().GetTo().GetAddress()))
		s.log.Debug(fmt.Sprintf("amount: %s", event.GetFundsIn().GetAmount()))
		s.log.Debug(fmt.Sprintf("token: %s", hex.EncodeToString(event.GetFundsIn().GetToken().GetAddress())))
		s.log.Debug(fmt.Sprintf("tx hash: %s", hex.EncodeToString(event.GetFundsIn().GetTx().GetHash())))
		s.log.Debug(fmt.Sprintf("block number: %d", event.GetFundsIn().GetTx().GetBlocknumber()))
		s.log.Debug(fmt.Sprintf("sender: %s", hex.EncodeToString(event.GetFundsIn().GetTx().GetSender())))
		s.log.Debug("")
	case chains.EventTypeOut:
		s.log.Debug(fmt.Sprintf("from network name: %s", event.GetFundsOut().GetFrom().GetNetworkName()))
		s.log.Debug(fmt.Sprintf("from address: %s", event.GetFundsOut().GetFrom().GetAddress()))
		s.log.Debug(fmt.Sprintf("to: %s", hex.EncodeToString(event.GetFundsOut().GetTo().GetAddress())))
		s.log.Debug(fmt.Sprintf("amount: %s", event.GetFundsOut().GetAmount()))
		s.log.Debug(fmt.Sprintf("token: %s", hex.EncodeToString(event.GetFundsOut().GetToken().GetAddress())))
		s.log.Debug(fmt.Sprintf("tx hash: %s", hex.EncodeToString(event.GetFundsOut().GetTx().GetHash())))
		s.log.Debug(fmt.Sprintf("block number: %d", event.GetFundsOut().GetTx().GetBlocknumber()))
		s.log.Debug(fmt.Sprintf("sender: %s", hex.EncodeToString(event.GetFundsOut().GetTx().GetSender())))
		s.log.Debug("")
	}
}

func (s *Connector) logBridgeOut(req *connectorpb.TokenOutRequest) {
	s.log.Debug(fmt.Sprintf("time: %s, called BridgeOut with request:", time.Now().Format(time.RFC1123)))
	s.log.Debug(fmt.Sprintf("amount: %s", req.GetAmount()))
	s.log.Debug(fmt.Sprintf("token: %s", hex.EncodeToString(req.GetToken().GetAddress())))
	s.log.Debug(fmt.Sprintf("to: %s", hex.EncodeToString(req.GetTo().GetAddress())))
	s.log.Debug(fmt.Sprintf("from network name: %s", req.GetFrom().GetNetworkName()))
	s.log.Debug(fmt.Sprintf("from address: %s", req.GetFrom().GetAddress()))
	s.log.Debug(fmt.Sprintf("transaction id: %d", req.GetTransactionId()))
	s.log.Debug("")
}
