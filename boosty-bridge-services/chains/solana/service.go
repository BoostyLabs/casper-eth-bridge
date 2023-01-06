// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package solana

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/portto/solana-go-sdk/client"
	"github.com/zeebo/errs"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger"
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

	bridge       chains.Bridge
	solanaClient *client.Client

	mutex            sync.Mutex
	eventSubscribers []chains.EventSubscriber
	wg               sync.WaitGroup
}

// NewService is constructor for Service.
func NewService(gctx context.Context, config Config, log logger.Logger, bridge chains.Bridge, solanaClient *client.Client) *Service {
	return &Service{
		gctx:             gctx,
		config:           config,
		log:              log,
		bridge:           bridge,
		solanaClient:     solanaClient,
		eventSubscribers: make([]chains.EventSubscriber, 0),
	}
}

// Metadata returns metadata of the network this connector provides.
func (service *Service) Metadata(ctx context.Context) chains.NetworkMetadata {
	return chains.NetworkMetadata{
		ID:          networks.IDSolana,
		Name:        service.GetChainName(),
		NodeAddress: service.config.NodeAddress,
		Type:        networks.TypeSolana,
		IsTestnet:   service.config.IsTestnet,
	}
}

// KnownTokens returns tokens known by this connector.
func (service *Service) KnownTokens(ctx context.Context) chains.Tokens {
	// TODO: read from db mb.
	return chains.Tokens{}
}

// BridgeOut initiates outbound bridge transaction.
func (service *Service) BridgeOut(ctx context.Context, req chains.TokenOutRequest) ([]byte, error) {
	// TODO implement.
	return nil, nil
}

// ReadEvents reads real-time events from node and old events from blocks and notifies subscribers.
func (service *Service) ReadEvents(ctx context.Context, fromBlock uint64) error {
	// TODO implement.
	return nil
}

// EstimateTransfer estimates a potential transfer.
func (service *Service) EstimateTransfer(ctx context.Context) (chains.Estimation, error) {
	// TODO implement.
	return chains.Estimation{}, nil
}

// GetChainName returns chain name.
func (service *Service) GetChainName() string {
	return service.config.ChainName
}

// BridgeInSignature returns signature for user to send bridgeIn transaction.
func (service *Service) BridgeInSignature(context.Context, chains.BridgeInSignatureRequest) (chains.BridgeInSignatureResponse, error) {
	// TODO: add implementation.
	return chains.BridgeInSignatureResponse{}, nil
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

// CloseClient closes HTTP node client.
func (service *Service) CloseClient() {
	// TODO: add implementation.
}

// CloseWsClient closes WS node client.
func (service *Service) CloseWsClient() {
	// TODO: add implementation.
}
