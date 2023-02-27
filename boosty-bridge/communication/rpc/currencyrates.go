// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package rpc

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	bridgeoraclepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-oracle"

	"tricorn/currencyrates"
)

// currencyratesRPC provides access to the currencyrates.
type currencyratesRPC struct {
	client bridgeoraclepb.BridgeOracleClient
}

// PriceStream is real time events streaming for tokens price.
func (currencyratesRPC *currencyratesRPC) PriceStream(ctx context.Context, tokenPriceChan chan currencyrates.TokenPrice) error {
	stream, err := currencyratesRPC.client.PriceStream(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			close(tokenPriceChan)
			return nil
		default:
		}

		priceUpdate, err := stream.Recv()
		if err != nil {
			return err
		}

		tokenPrice := currencyrates.TokenPrice{
			TokenName:  priceUpdate.TokenName,
			Amount:     priceUpdate.Amount,
			Decimals:   priceUpdate.Decimals,
			LastUpdate: priceUpdate.LastUpdate.AsTime(),
		}
		tokenPriceChan <- tokenPrice
	}
}
