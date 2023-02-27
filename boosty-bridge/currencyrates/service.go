// Copyright (C) 2023 Creditor Corp. Group.
// See LICENSE for copying information.

package currencyrates

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"

	"tricorn/internal/logger"
)

// ErrCurrencyRates indicates that there was an error in the service.
var ErrCurrencyRates = errs.Class("currencyrates service")

// Service is handling currencyrates related logic.
//
// architecture: Service
type Service struct {
	gctx   context.Context
	config Config
	log    logger.Logger

	mutex            sync.Mutex
	eventSubscribers []EventSubscriber

	currencyRates CurrencyRates
}

// NewService is constructor for Service.
func NewService(gctx context.Context, config Config, log logger.Logger, currencyRates CurrencyRates) *Service {
	return &Service{
		gctx:             gctx,
		config:           config,
		log:              log,
		eventSubscribers: make([]EventSubscriber, 0),
		currencyRates:    currencyRates,
	}
}

// SubscribeToTokenPrice is real time events streaming for tokens price.
func (service *Service) SubscribeToTokenPrice(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(service.config.EventsReadingIntervalInSeconds) * time.Second)

	for range ticker.C {
		select {
		case <-service.gctx.Done():
			return nil
		case <-ctx.Done():
			return nil
		default:
		}

		// TODO: get token symbol dynamically.
		currency, err := service.currencyRates.GetPrice(ctx, "USDT", "ETH")
		if err != nil {
			service.log.Error("could not get currency price", ErrCurrencyRates.Wrap(err))
		}

		event := TokenPrice{
			TokenName:  currency.Symbol,
			Amount:     fmt.Sprintf("%f", currency.Price),
			Decimals:   18, // TODO: get it from db table.
			LastUpdate: time.Now().UTC(),
		}
		service.Notify(ctx, event)
	}

	return nil
}

// AddEventSubscriber adds subscriber to event publisher.
func (service *Service) AddEventSubscriber() EventSubscriber {
	subscriber := EventSubscriber{
		ID:             uuid.New(),
		TokenPriceChan: make(chan TokenPrice),
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
func (service *Service) Notify(ctx context.Context, event TokenPrice) {
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
