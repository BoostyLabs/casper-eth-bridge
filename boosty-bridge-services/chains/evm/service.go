// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package evm

import (
	"context"
	"crypto/ecdsa"
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/zeebo/errs"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/contracts/evm"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/contracts/evm/bridge"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger"
)

// ensures that Service implement chains.Connector.
var _ chains.Connector = (*Service)(nil)

// CurveCoordinatesSize defines size of public key curve coordinates which
// zero-padded to the adjusted underlying field size.
const CurveCoordinatesSize = 64

// wei defines the smallest unit of coin in the Ethereum network.
const wei uint64 = 1e18

// Error is connector default error type.
var Error = errs.Class("connector service")

// Service is handling ethereum connector related logic.
//
// architecture: Service
type Service struct {
	gctx   context.Context
	config Config
	log    logger.Logger

	mutex            sync.Mutex
	eventSubscribers []chains.EventSubscriber

	ethClient   *ethclient.Client
	wsEthClient *ethclient.Client

	instance *bridge.Bridge // contract instance.
	transfer Transfer       // bridge contract clent.

	bridge chains.Bridge

	wg sync.WaitGroup
}

// New is Service constructor.
func New(gctx context.Context, config Config, log logger.Logger, bridge chains.Bridge, instance *bridge.Bridge, transfer Transfer,
	ethClient *ethclient.Client, wsEthClient *ethclient.Client) *Service {
	return &Service{
		gctx:             gctx,
		config:           config,
		eventSubscribers: make([]chains.EventSubscriber, 0, config.NumOfSubscribers),
		log:              log,
		bridge:           bridge,
		instance:         instance,
		transfer:         transfer,
		ethClient:        ethClient,
		wsEthClient:      wsEthClient,
	}
}

// Metadata returns ethereum network metadata.
func (service *Service) Metadata(ctx context.Context) chains.NetworkMetadata {
	return chains.NetworkMetadata{
		ID:          networks.IDEth,
		Name:        service.GetChainName(),
		NodeAddress: service.config.NodeAddress,
		Type:        networks.TypeEVM,
		IsTestnet:   service.config.IsTestnet,
	}
}

// KnownTokens returns tokens known by this connector.
func (service *Service) KnownTokens(ctx context.Context) chains.Tokens {
	// TODO: read from db mb.
	return chains.Tokens{}
}

// BridgeOut initiates transfer.
func (service *Service) BridgeOut(ctx context.Context, transfer chains.TokenOutRequest) ([]byte, error) {
	publicKeyByte, err := service.bridge.PublicKey(ctx, networks.TypeEVM)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	if len(publicKeyByte) < CurveCoordinatesSize {
		return nil, Error.New("invalid public key curve coordinates")
	}

	x := big.NewInt(0).SetBytes(publicKeyByte[:32])
	y := big.NewInt(0).SetBytes(publicKeyByte[32:])
	publicKeyECDSA := ecdsa.PublicKey{
		Curve: btcec.S256(),
		X:     x,
		Y:     y,
	}
	ownerAddress := crypto.PubkeyToAddress(publicKeyECDSA)

	sign := func(data []byte) ([]byte, error) {
		singIn := chains.SignRequest{
			// TODO: fix it.
			NetworkId: networks.TypeEVM,
			Data:      data,
		}

		return service.bridge.Sign(ctx, singIn)
	}

	auth, err := evm.NewKeyedTransactorWithChainID(ctx, ownerAddress, big.NewInt(int64(service.config.ChainID)), sign)
	if err != nil {
		return nil, err
	}

	auth.Value = big.NewInt(0)
	tr, err := service.instance.BridgeOut(auth, common.BytesToAddress(transfer.Token.Bytes()), common.BytesToAddress(transfer.To.Bytes()),
		transfer.Amount, transfer.TransactionID, transfer.From.NetworkName, transfer.From.Address)
	if err != nil {
		return nil, err
	}

	return tr.Hash().Bytes(), nil
}

// EstimateTransfer estimates transfer fee and time.
func (service *Service) EstimateTransfer(ctx context.Context) (chains.Estimation, error) {
	gasPrice, err := service.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return chains.Estimation{}, Error.Wrap(err)
	}

	gasLimit := new(big.Int).SetUint64(service.config.GasLimit)
	feeWei := gasPrice.Mul(gasPrice, gasLimit)

	fee := new(big.Float).Quo(new(big.Float).SetInt(feeWei), new(big.Float).SetUint64(wei))

	estimation := chains.Estimation{
		Fee:                   fee.String(),
		FeePercentage:         service.config.FeePercentage,
		EstimatedConfirmation: service.config.ConfirmationTime,
	}

	return estimation, nil
}

// readEventsFromBlock reads node events in a given interval of blocks and notifies subscribers.
func (service *Service) readEventsFromBlock(ctx context.Context, fromBlock, toBlock uint64) error {
	// if we start from scratch and client does not have any events it will be sufficient to read batch
	// of events from 0 to last block. On the other hand if client already has some events.
	// Reading small chunks of events would be more appropriate, because logs reading/parsing
	// is time-consuming operation.

	var lastProcessedBlock = fromBlock
	for {
		switch {
		case lastProcessedBlock < toBlock:
			err := service.readOldEvents(ctx, lastProcessedBlock, lastProcessedBlock+listeningLimit)
			if err != nil {
				return Error.Wrap(err)
			}
		case lastProcessedBlock >= toBlock:
			return service.readOldEvents(ctx, lastProcessedBlock, toBlock)
		default:
			return nil
		}
		lastProcessedBlock += listeningLimit
	}
}

func (service *Service) readOldEvents(ctx context.Context, fromBlock, toBlock uint64) error {
	topics := make([]common.Hash, 0)
	topics = append(topics, service.config.EventsFundIn, service.config.EventsFundOut)

	from := new(big.Int).SetUint64(fromBlock)
	to := new(big.Int).SetUint64(toBlock)

	query := ethereum.FilterQuery{
		FromBlock: from,
		ToBlock:   to,
		Addresses: []common.Address{service.config.BridgeContractAddress},
		Topics:    [][]common.Hash{topics},
	}

	logs, err := service.ethClient.FilterLogs(ctx, query)
	if err != nil {
		return Error.Wrap(err)
	}

	for _, log := range logs {
		// check is func need to be closed because of app/stream context.
		select {
		case <-service.gctx.Done():
			return nil
		case <-ctx.Done():
			return nil
		default:
		}

		event, err := parseLog(service.instance, log, service.config.EventsFundIn, service.config.EventsFundOut)
		if err != nil {
			return Error.Wrap(err)
		}

		service.Notify(ctx, event)
	}

	return nil
}

// SubscribeEvents is real time events streaming from blockchain to events subscribers.
func (service *Service) subscribeEvents(ctx context.Context) error {
	topics := make([]common.Hash, 0)
	topics = append(topics, service.config.EventsFundIn, service.config.EventsFundOut)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{service.config.BridgeContractAddress},
		Topics:    [][]common.Hash{topics},
	}

	logsCh := make(chan types.Log)
	subscriptions, err := service.wsEthClient.SubscribeFilterLogs(ctx, query, logsCh)
	if err != nil {
		return Error.Wrap(err)
	}

	for {
		select {
		case <-service.gctx.Done():
			return nil
		case <-ctx.Done():
			return nil
		case err = <-subscriptions.Err():
			return Error.Wrap(err)
		case log := <-logsCh:
			event, err := parseLog(service.instance, log, service.config.EventsFundIn, service.config.EventsFundOut)
			if err != nil && !Error.Has(ErrBlockchainRework) {
				return Error.Wrap(err)
			}
			if Error.Has(ErrBlockchainRework) {
				continue
			}

			service.Notify(ctx, event)
		}
	}
}

// parseLog parses log data to internal object by contract instance.
func parseLog(instance *bridge.Bridge, log types.Log, fundInEventHash, fundOutEventHash common.Hash) (chains.EventVariant, error) {
	// handle blockchain rework.
	if log.Removed {
		return chains.EventVariant{}, Error.Wrap(ErrBlockchainRework)
	}

	switch log.Topics[0] {
	case fundInEventHash:
		fundIn, err := instance.ParseBridgeFundsIn(log)
		if err != nil {
			return chains.EventVariant{}, Error.Wrap(err)
		}

		txInfo := chains.TransactionInfo{
			Hash:        fundIn.Raw.TxHash.Bytes(),
			BlockNumber: fundIn.Raw.BlockNumber,
			Sender:      fundIn.Raw.Address.Bytes(),
		}

		event := chains.EventVariant{
			Type: chains.EventTypeIn,
			EventFundsIn: chains.EventFundsIn{
				From: fundIn.Sender.Bytes(),
				To: networks.Address{
					NetworkName: fundIn.DestinationChain,
					Address:     fundIn.DestinationAddress,
				},
				Amount: fundIn.Amount.String(),
				Token:  fundIn.Token.Bytes(),
				Tx:     txInfo,
			},
		}

		return event, nil
	case fundOutEventHash:
		fundOut, err := instance.ParseBridgeFundsOut(log)
		if err != nil {
			return chains.EventVariant{}, Error.Wrap(err)
		}

		txInfo := chains.TransactionInfo{
			Hash:        fundOut.Raw.TxHash.Bytes(),
			BlockNumber: fundOut.Raw.BlockNumber,
			Sender:      fundOut.Raw.Address.Bytes(),
		}

		event := chains.EventVariant{
			Type: chains.EventTypeOut,
			EventFundsOut: chains.EventFundsOut{
				From: networks.Address{
					NetworkName: fundOut.SourceChain,
					Address:     fundOut.SourceAddress,
				},
				To:     fundOut.Recipient.Bytes(),
				Amount: fundOut.Amount.String(),
				Token:  fundOut.Token.Bytes(),
				Tx:     txInfo,
			},
		}

		return event, nil
	default:
		return chains.EventVariant{}, Error.New("unknown log type")
	}
}

// ReadEvents initiates events reading. Reading logic divided into two parts.
// First part is reading from the last block which was processed to the latest block in blockchain
// Second part is real-time reading of new events that just occurred.
func (service *Service) ReadEvents(ctx context.Context, fromBlock uint64) error {
	blockNum, err := service.ethClient.BlockNumber(ctx)
	if err != nil {
		return Error.Wrap(err)
	}

	service.wg.Add(2)
	go func(ctx context.Context) {
		defer service.wg.Done()

		if fromBlock == 0 || fromBlock > blockNum {
			return
		}

		err = service.readEventsFromBlock(ctx, fromBlock, blockNum)
		if err != nil {
			service.log.Error("could not read past events", err)
		}
	}(ctx)
	go func(ctx context.Context) {
		defer service.wg.Done()

		err := service.subscribeEvents(ctx)
		if err != nil {
			service.log.Error("could not read real time events", err)
		}
	}(ctx)
	service.wg.Wait()

	return nil
}

// GetChainName returns chain name.
func (service *Service) GetChainName() string {
	return service.config.ChainName
}

// BridgeInSignature returns signature for user to send bridgeIn transaction.
func (service *Service) BridgeInSignature(ctx context.Context, req chains.BridgeInSignatureRequest) (chains.BridgeInSignatureResponse, error) {
	commission := new(big.Float)
	commissionInt, _ := commission.Int64()
	commissionBigInt := big.NewInt(commissionInt)

	if req.Amount.Cmp(commissionBigInt) <= 0 {
		return chains.BridgeInSignatureResponse{}, Error.New("the amount must be greater than the gas commission")
	}

	deadlineTime := time.Now().UTC().Add(time.Second * time.Duration(service.config.SignatureValidityTime)).Unix()
	deadline := big.NewInt(0).SetInt64(deadlineTime)

	bridgeIn := GetBridgeInSignatureRequest{
		User:               req.User,
		Token:              common.HexToAddress(req.Token),
		Amount:             req.Amount,
		GasCommission:      commissionBigInt,
		DestinationChain:   req.Destination.NetworkName,
		DestinationAddress: req.Destination.Address,
		Deadline:           deadline,
		Nonce:              req.Nonce,
	}

	signature, err := service.transfer.GetBridgeInSignature(ctx, bridgeIn)

	response := chains.BridgeInSignatureResponse{
		Token:        req.Token,
		Amount:       req.Amount,
		GasComission: commissionBigInt.String(),
		Destination:  req.Destination,
		Deadline:     deadline.String(),
		Nonce:        req.Nonce,
		Signature:    signature,
	}

	return response, Error.Wrap(err)
}

// AddEventSubscriber adds subscriber to event publisher.
func (service *Service) AddEventSubscriber() chains.EventSubscriber {
	subscriber := chains.EventSubscriber{
		ID:         uuid.New(),
		EventsChan: make(chan chains.EventVariant),
	}

	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.eventSubscribers = append(service.eventSubscribers, subscriber)

	return subscriber
}

// RemoveEventSubscriber removes subscriber.
func (service *Service) RemoveEventSubscriber(id uuid.UUID) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	subIndex := 0
	for index, subscriber := range service.eventSubscribers {
		if subscriber.GetID() == id {
			subIndex = index
			break
		}
	}

	copy(service.eventSubscribers[subIndex:], service.eventSubscribers[subIndex+1:])
	service.eventSubscribers = service.eventSubscribers[:len(service.eventSubscribers)-1]
}

// Notify notifies all subscribers with events.
func (service *Service) Notify(ctx context.Context, event chains.EventVariant) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	for _, subscriber := range service.eventSubscribers {
		select {
		case <-service.gctx.Done():
			return
		case <-ctx.Done():
			return
		default:
			subscriber.NotifyWithEvent(event)
		}
	}
}

// CloseClient closes HTTP ethereum client.
func (service *Service) CloseClient() {
	service.ethClient.Close()
}

// CloseWsClient closes WS ethereum client.
func (service *Service) CloseWsClient() {
	service.wsEthClient.Close()
}
