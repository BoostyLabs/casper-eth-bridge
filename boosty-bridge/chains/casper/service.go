// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package casper

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/casper-ecosystem/casper-golang-sdk/keypair"
	"github.com/casper-ecosystem/casper-golang-sdk/sdk"
	"github.com/casper-ecosystem/casper-golang-sdk/serialization"
	"github.com/casper-ecosystem/casper-golang-sdk/types"
	"github.com/google/uuid"
	"github.com/zeebo/errs"

	"tricorn/bridge/networks"
	"tricorn/chains"
	"tricorn/internal/eventparsing"
	"tricorn/internal/logger"
)

// ensures that Service implement chains.Connector.
var _ chains.Connector = (*Service)(nil)

// ErrConnector indicates that there was an error in the service.
var ErrConnector = errs.Class("connector service")

// Service is handling connector related logic.
//
// architecture: Service
type Service struct {
	gctx   context.Context
	config Config
	log    logger.Logger

	bridge chains.Bridge
	casper Casper
	signer Signer
	events *http.Client

	mutex            sync.Mutex
	eventSubscribers []chains.EventSubscriber
	wg               sync.WaitGroup
}

// NewService is constructor for Service.
func NewService(gctx context.Context, config Config, log logger.Logger, bridge chains.Bridge, casper Casper, signer Signer) *Service {
	eventsClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

	return &Service{
		gctx:   gctx,
		config: config,
		log:    log,
		bridge: bridge,
		casper: casper,
		signer: signer,
		events: eventsClient,
	}
}

// Network returns supported by connector network.
// TODO: place token contract to token.
func (service *Service) Network(ctx context.Context) networks.Network {
	id := networks.IDCasperTest
	if !service.config.IsTestnet {
		id = networks.IDCasper
	}

	return networks.Network{
		ID:             id,
		Name:           service.GetChainName(),
		Type:           networks.TypeCasper,
		IsTestnet:      service.config.IsTestnet,
		NodeAddress:    service.config.RPCNodeAddress,
		BridgeContract: service.config.BridgeContractAddress,
		GasLimit:       service.config.GasLimit,
	}
}

// KnownTokens returns tokens known by this connector.
func (service *Service) KnownTokens(ctx context.Context) chains.Tokens {
	// TODO: read from db mb.
	return chains.Tokens{}
}

// BridgeOut initiates outbound bridge transaction.
func (service *Service) BridgeOut(ctx context.Context, req chains.TokenOutRequest) ([]byte, error) {
	respPubKey, err := service.bridge.PublicKey(ctx, networks.TypeCasper)
	if err != nil {
		return nil, ErrConnector.Wrap(err)
	}

	publicKey := keypair.PublicKey{
		Tag:        keypair.KeyTagEd25519,
		PubKeyData: respPubKey,
	}

	standardPayment := new(big.Int).SetUint64(service.config.GasLimit)

	deployParams := sdk.NewDeployParams(publicKey, strings.ToLower(service.config.ChainName.String()), nil, 0)
	payment := sdk.StandardPayment(standardPayment)

	// token contract.
	tokenContractFixedBytes := types.FixedByteArray(req.Token)
	tokenContract := types.CLValue{
		Type:      types.CLTypeByteArray,
		ByteArray: &tokenContractFixedBytes,
	}
	tokenContractBytes, err := serialization.Marshal(tokenContract)
	if err != nil {
		return nil, ErrConnector.Wrap(err)
	}

	// amount.
	amount := types.CLValue{
		Type: types.CLTypeU256,
		U256: req.Amount,
	}
	amountBytes, err := serialization.Marshal(amount)
	if err != nil {
		return nil, ErrConnector.Wrap(err)
	}

	// transaction_id.
	transactionID := types.CLValue{
		Type: types.CLTypeU256,
		U256: req.TransactionID,
	}
	transactionIDBytes, err := serialization.Marshal(transactionID)
	if err != nil {
		return nil, ErrConnector.Wrap(err)
	}

	//  source chain.
	sourceChain := types.CLValue{
		Type:   types.CLTypeString,
		String: &req.From.NetworkName,
	}
	sourceChainBytes, err := serialization.Marshal(sourceChain)
	if err != nil {
		return nil, ErrConnector.Wrap(err)
	}

	// source address.
	sourceAddress := types.CLValue{
		Type:   types.CLTypeString,
		String: &req.From.Address,
	}
	sourceAddressBytes, err := serialization.Marshal(sourceAddress)
	if err != nil {
		return nil, ErrConnector.Wrap(err)
	}

	// recipient.
	var recipientHashBytes [32]byte
	copy(recipientHashBytes[:], req.To)

	recipient := types.CLValue{
		Type: types.CLTypeKey,
		Key: &types.Key{
			Type:    types.KeyTypeAccount,
			Account: recipientHashBytes,
		},
	}
	recipientBytes, err := serialization.Marshal(recipient)
	if err != nil {
		return nil, ErrConnector.Wrap(err)
	}

	args := map[string]sdk.Value{
		"token_contract": {
			IsOptional:  false,
			Tag:         types.CLTypeByteArray,
			StringBytes: hex.EncodeToString(tokenContractBytes),
		},
		"amount": {
			Tag:         types.CLTypeU256,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(amountBytes),
		},
		"transaction_id": {
			Tag:         types.CLTypeU256,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(transactionIDBytes),
		},
		"source_chain": {
			Tag:         types.CLTypeString,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(sourceChainBytes),
		},
		"source_address": {
			Tag:         types.CLTypeString,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(sourceAddressBytes),
		},
		"recipient": {
			Tag:         types.CLTypeKey,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(recipientBytes),
		},
	}

	keyOrder := []string{
		"token_contract",
		"amount",
		"transaction_id",
		"source_chain",
		"source_address",
		"recipient",
	}
	runtimeArgs := sdk.NewRunTimeArgs(args, keyOrder)

	contractHexBytes, err := hex.DecodeString(service.config.BridgeContractAddress)
	if err != nil {
		return nil, ErrConnector.Wrap(err)
	}

	var contractHashBytes [32]byte
	copy(contractHashBytes[:], contractHexBytes)
	session := sdk.NewStoredContractByHash(contractHashBytes, "bridge_out", *runtimeArgs)

	deploy := sdk.MakeDeploy(deployParams, payment, session)

	reqSign := chains.SignRequest{
		NetworkId: networks.TypeCasper,
		Data:      deploy.Hash,
	}
	signature, err := service.bridge.Sign(ctx, reqSign)
	if err != nil {
		return nil, ErrConnector.Wrap(err)
	}

	signatureKeypair := keypair.Signature{
		Tag:           keypair.KeyTagEd25519,
		SignatureData: signature,
	}

	approval := sdk.Approval{
		Signer:    publicKey,
		Signature: signatureKeypair,
	}

	deploy.Approvals = append(deploy.Approvals, approval)

	hash, err := service.casper.PutDeploy(*deploy)
	if err != nil {
		return nil, ErrConnector.Wrap(err)
	}

	txHash, err := hex.DecodeString(hash)

	return txHash, ErrConnector.Wrap(err)
}

// ReadEvents reads real-time events from node and old events from blocks and notifies subscribers.
func (service *Service) ReadEvents(ctx context.Context, fromBlock uint64) error {
	service.wg.Add(2)

	go func(ctx context.Context) {
		defer service.wg.Done()

		currentBlockNumber, err := service.casper.GetCurrentBlockNumber()
		if err != nil {
			service.log.Error("could not get current block number", err)
		}

		// when the fromBlock is not sent, we skip reading old events.
		if fromBlock == 0 || currentBlockNumber <= fromBlock {
			return
		}

		err = service.readEventsFromBlock(ctx, fromBlock, currentBlockNumber)
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

// readEventsFromBlock reads node events from blocks and notifies subscribers.
func (service *Service) readEventsFromBlock(ctx context.Context, fromBlock uint64, toBlock uint64) error {
	events, err := service.casper.GetEventsByBlockNumbers(fromBlock, toBlock, service.config.BridgeEventsHash)
	if err != nil {
		return ErrConnector.Wrap(err)
	}

	for _, event := range events {
		eventFunds, err := service.parseEventFromTransform(event, event.DeployProcessed.ExecutionResult.Success.Effect.Transforms[0])
		if err != nil {
			return ErrConnector.Wrap(err)
		}

		service.Notify(ctx, eventFunds)
	}

	return nil
}

func (service *Service) parseEventFromTransform(event Event, transform Transform) (chains.EventVariant, error) {
	transformMap, ok := transform.Transform.(map[string]interface{})
	if !ok {
		return chains.EventVariant{}, ErrConnector.New("couldn't parse map to transform")
	}

	writeCLValue, ok := transformMap[WriteCLValueKey].(map[string]interface{})
	if !ok {
		return chains.EventVariant{}, ErrConnector.New("couldn't parse map to transform map")
	}

	bytes, ok := writeCLValue[BytesKey].(string)
	if !ok {
		return chains.EventVariant{}, ErrConnector.New("couldn't parse string to bytes key")
	}

	eventData := eventparsing.EventData{
		Bytes: bytes,
	}

	eventType, err := eventData.GetEventType()
	if err != nil {
		return chains.EventVariant{}, ErrConnector.Wrap(err)
	}

	var (
		tokenContractAddress []byte
		chainName            string
		chainAddress         string
		amountStr            string
		userWalletAddress    []byte
	)

	if eventType == fundInType {
		tokenContractAddress, err = hex.DecodeString(eventData.GetTokenContractAddress())
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}

		chainName, err = eventData.GetChainName()
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}

		chainAddress, err = eventData.GetChainAddress()
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}

		amount, err := eventData.GetAmount()
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}
		amountStr = strconv.Itoa(amount)

		gasCommission, err := eventData.GetGasCommission()
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}
		gasCommissionStr := strconv.Itoa(gasCommission)
		// TODO: add in event later.
		_ = gasCommissionStr

		stableCommissionPercent, err := eventData.GetStableCommissionPercent()
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}
		stableCommissionPercentStr := strconv.Itoa(stableCommissionPercent)
		// TODO: add in event later.
		_ = stableCommissionPercentStr

		nonce, err := eventData.GetNonce()
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}
		nonceStr := strconv.Itoa(nonce)
		// TODO: add in event later.
		_ = nonceStr

		userWalletAddress, err = hex.DecodeString(eventData.GetUserWalletAddress())
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}
	}

	if eventType == fundOutType {
		tokenContractAddress, err = hex.DecodeString(eventData.GetTokenContractAddress())
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}

		chainName, err = eventData.GetChainName()
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}

		chainAddress, err = eventData.GetChainAddress()
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}

		amount, err := eventData.GetAmount()
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}
		amountStr = strconv.Itoa(amount)

		transactionID, err := eventData.GetTransactionID()
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}
		transactionIDStr := strconv.Itoa(transactionID)
		// TODO: add in event later.
		_ = transactionIDStr

		userWalletAddress, err = hex.DecodeString(eventData.GetUserWalletAddress())
		if err != nil {
			return chains.EventVariant{}, ErrConnector.Wrap(err)
		}
	}

	hash, err := hex.DecodeString(event.DeployProcessed.DeployHash)
	if err != nil {
		return chains.EventVariant{}, ErrConnector.Wrap(err)
	}

	sender, err := hex.DecodeString(event.DeployProcessed.Account)
	if err != nil {
		return chains.EventVariant{}, ErrConnector.Wrap(err)
	}

	blockNumber, err := service.casper.GetBlockNumberByHash(event.DeployProcessed.BlockHash)
	if err != nil {
		return chains.EventVariant{}, ErrConnector.Wrap(err)
	}

	transactionInfo := chains.TransactionInfo{
		Hash:        hash,
		BlockNumber: uint64(blockNumber),
		Sender:      sender,
	}

	var eventFunds chains.EventVariant
	switch eventType {
	case chains.EventTypeIn.Int():
		eventFunds = chains.EventVariant{
			Type: chains.EventType(eventType),
			EventFundsIn: chains.EventFundsIn{
				From: userWalletAddress,
				To: networks.Address{
					NetworkName: chainName,
					Address:     chainAddress,
				},
				Amount: amountStr,
				Token:  tokenContractAddress,
				Tx:     transactionInfo,
			},
		}
	case chains.EventTypeOut.Int():
		eventFunds = chains.EventVariant{
			Type: chains.EventType(eventType),
			EventFundsOut: chains.EventFundsOut{
				From: networks.Address{
					NetworkName: chainName,
					Address:     chainAddress,
				},
				To:     userWalletAddress,
				Amount: amountStr,
				Token:  tokenContractAddress,
				Tx:     transactionInfo,
			},
		}
	default:
		return chains.EventVariant{}, ErrConnector.New("invalid event type")
	}

	tokenIn := hex.EncodeToString(eventFunds.EventFundsIn.Token)
	eventFunds.EventFundsIn.Token, err = hex.DecodeString(eventparsing.TagHash.String() + tokenIn)
	if err != nil {
		return chains.EventVariant{}, ErrConnector.Wrap(err)
	}

	from := hex.EncodeToString(eventFunds.EventFundsIn.From)
	eventFunds.EventFundsIn.From, err = hex.DecodeString(eventparsing.TagAccount.String() + from)
	if err != nil {
		return chains.EventVariant{}, ErrConnector.Wrap(err)
	}

	tokenOut := hex.EncodeToString(eventFunds.EventFundsOut.Token)
	eventFunds.EventFundsOut.Token, err = hex.DecodeString(eventparsing.TagHash.String() + tokenOut)
	if err != nil {
		return chains.EventVariant{}, ErrConnector.Wrap(err)
	}

	to := hex.EncodeToString(eventFunds.EventFundsOut.To)
	eventFunds.EventFundsOut.To, err = hex.DecodeString(eventparsing.TagAccount.String() + to)
	if err != nil {
		return chains.EventVariant{}, ErrConnector.Wrap(err)
	}

	return eventFunds, nil
}

// SubscribeEvents is real time events streaming from blockchain to events subscribers.
func (service *Service) subscribeEvents(ctx context.Context) error {
	var body io.Reader
	req, err := http.NewRequest(http.MethodGet, service.config.EventNodeAddress, body)
	if err != nil {
		return ErrConnector.Wrap(err)
	}

	resp, err := service.events.Do(req)
	if err != nil {
		return ErrConnector.Wrap(err)
	}

	for {
		select {
		case <-service.gctx.Done():
			return nil
		case <-ctx.Done():
			return nil
		default:
		}

		reader := bufio.NewReader(resp.Body)
		rawBody, err := reader.ReadBytes('\n')
		if err != nil {
			return ErrConnector.Wrap(err)
		}

		switch {
		case string(rawBody) == ":" || len(string(rawBody)) < 5:
			continue
		case len(string(rawBody)) >= 5 && string(rawBody)[:5] == "data:":
			rawBody = []byte(string(rawBody)[5:])
		}

		var event Event
		err = json.Unmarshal(rawBody, &event)
		if err != nil {
			// continue execution because event has unsupported structure.
			continue
		}

		transforms := event.DeployProcessed.ExecutionResult.Success.Effect.Transforms
		if len(transforms) == 0 {
			continue
		}

		for _, transform := range transforms {
			select {
			case <-service.gctx.Done():
				return nil
			case <-ctx.Done():
				return nil
			default:
			}

			if transform.Key == service.config.BridgeEventsHash {
				eventFunds, err := service.parseEventFromTransform(event, transform)
				if err != nil {
					return ErrConnector.Wrap(err)
				}

				service.Notify(ctx, eventFunds)
			}
		}
	}
}

// EstimateTransfer estimates a potential transfer.
func (service *Service) EstimateTransfer(ctx context.Context) (chains.Estimation, error) {
	fee := new(big.Int).SetUint64(service.config.GasLimit)

	return chains.Estimation{
		Fee:                   fee.String(),
		FeePercentage:         service.config.FeePercentage,
		EstimatedConfirmation: service.config.EstimatedConfirmation,
	}, nil
}

// GetChainName returns chain name.
func (service *Service) GetChainName() networks.Name {
	return service.config.ChainName
}

// BridgeInSignature returns signature for user to send bridgeIn transaction.
func (service *Service) BridgeInSignature(ctx context.Context, req chains.BridgeInSignatureRequest) (chains.BridgeInSignatureResponse, error) {
	bridgeHash, err := hex.DecodeString(service.config.BridgeContractAddress)
	if err != nil {
		return chains.BridgeInSignatureResponse{}, ErrConnector.Wrap(err)
	}

	tokenPackageHashChainBytes, err := hex.DecodeString(req.Token)
	if err != nil {
		return chains.BridgeInSignatureResponse{}, ErrConnector.Wrap(err)
	}

	deadlineTime := time.Now().UTC().Add(time.Second * time.Duration(service.config.SignatureValidityTime)).UnixMilli()
	deadline := new(big.Int).SetInt64(deadlineTime)

	signature, err := service.signer.GetBridgeInSignature(ctx, BridgeInSignature{
		Prefix:             service.config.BridgeInPrefix,
		BridgeHash:         bridgeHash,
		TokenPackageHash:   tokenPackageHashChainBytes,
		AccountAddress:     req.User,
		Amount:             req.Amount,
		GasCommission:      req.GasCommission,
		Deadline:           deadline,
		Nonce:              req.Nonce,
		DestinationChain:   req.Destination.NetworkName,
		DestinationAddress: req.Destination.Address,
	})
	if err != nil {
		return chains.BridgeInSignatureResponse{}, ErrConnector.Wrap(err)
	}

	return chains.BridgeInSignatureResponse{
		Token:         req.Token,
		Amount:        req.Amount,
		GasCommission: req.GasCommission.String(),
		Destination:   req.Destination,
		Deadline:      deadline.String(),
		Nonce:         req.Nonce,
		Signature:     signature,
	}, nil
}

// CancelSignature returns signature for user to return funds.
func (service *Service) CancelSignature(ctx context.Context, req chains.CancelSignatureRequest) (chains.CancelSignatureResponse, error) {
	bridgeHashBytes, err := hex.DecodeString(service.config.BridgeContractAddress)
	if err != nil {
		return chains.CancelSignatureResponse{}, ErrConnector.Wrap(err)
	}

	signature, err := service.signer.GetTransferOutSignature(ctx, TransferOutSignature{
		Prefix:           service.config.TransferOutPrefix,
		BridgeHash:       bridgeHashBytes,
		TokenPackageHash: req.Token,
		AccountAddress:   req.Recipient,
		Amount:           req.Amount,
		GasCommission:    req.Commission,
		Nonce:            req.Nonce,
	})
	if err != nil {
		return chains.CancelSignatureResponse{}, ErrConnector.Wrap(err)
	}

	return chains.CancelSignatureResponse{
		Signature: signature,
	}, nil
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

// RemoveEventSubscriber removes publisher subscriber.
func (service *Service) RemoveEventSubscriber(id uuid.UUID) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	var subIndex int
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
		case <-ctx.Done():
			return
		default:
			subscriber.NotifyWithEvent(event)
		}
	}
}

// CloseClient closes HTTP node client.
func (service *Service) CloseClient() {
	// TODO: add implementation.
}
