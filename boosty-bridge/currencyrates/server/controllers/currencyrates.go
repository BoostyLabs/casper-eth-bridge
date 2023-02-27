// Copyright (C) 2023 Creditor Corp. Group.
// See LICENSE for copying information.

package controllers

import (
	"context"

	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	bridgeoraclepb "github.com/BoostyLabs/golden-gate-communication/go-gen/bridge-oracle"

	"tricorn/currencyrates"
	"tricorn/internal/logger"
)

// ensures that CurrencyRates implements bridgeoraclepb.BridgeOracleServer.
var _ bridgeoraclepb.BridgeOracleServer = (*CurrencyRates)(nil)

// Error is an internal error type for currencyrates controller.
var Error = errs.Class("currencyrates controller")

// CurrencyRates is controller that handles all currency rates related routes.
type CurrencyRates struct {
	gctx context.Context
	log  logger.Logger

	currencyrates *currencyrates.Service
}

// NewCurrencyRates is a constructor for currencyrates controller.
func NewCurrencyRates(gctx context.Context, log logger.Logger, currencyrates *currencyrates.Service) *CurrencyRates {
	currencyRatesController := &CurrencyRates{
		gctx:          gctx,
		log:           log,
		currencyrates: currencyrates,
	}

	return currencyRatesController
}

// PriceStream returns price for specific token.
func (c *CurrencyRates) PriceStream(_ *emptypb.Empty, stream bridgeoraclepb.BridgeOracle_PriceStreamServer) error {
	group, ctx := errgroup.WithContext(stream.Context())

	group.Go(func() error {
		err := c.currencyrates.SubscribeToTokenPrice(ctx)
		if err != nil {
			c.log.Error("couldn't read events", err)
			return status.Error(codes.Internal, err.Error())
		}

		return nil
	})

	group.Go(func() error {
		subscriber := c.currencyrates.AddEventSubscriber()

		for {
			select {
			case tokenPriceEvent, ok := <-subscriber.ReceiveEvents():
				if !ok {
					err := Error.New("events chan unexpectedly closed")
					c.log.Error("", err)
					return status.Error(codes.Internal, err.Error())
				}

				priceUpdate := bridgeoraclepb.PriceUpdate{
					TokenName: tokenPriceEvent.TokenName,
					Amount:    tokenPriceEvent.Amount,
					Decimals:  tokenPriceEvent.Decimals,
					LastUpdate: &timestamppb.Timestamp{
						Seconds: tokenPriceEvent.LastUpdate.Unix(),
					},
				}

				if err := stream.Send(&priceUpdate); err != nil {
					c.log.Error("couldn't send price update", Error.Wrap(err))
					return status.Error(codes.Internal, Error.Wrap(err).Error())
				}
			case <-c.gctx.Done():
				c.currencyrates.RemoveEventSubscriber(subscriber.GetID())

				return nil
			case <-ctx.Done():
				c.currencyrates.RemoveEventSubscriber(subscriber.GetID())

				return nil
			}
		}
	})

	return group.Wait()
}
