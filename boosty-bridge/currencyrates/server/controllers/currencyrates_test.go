// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package controllers_test

import (
	"context"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"

	"tricorn/currencyrates/server/controllers/apitesting"
	"tricorn/internal/config/envparse"
)

func TestRates(t *testing.T) {
	err := godotenv.Overload("./apitesting/configs/.test.currencyrates.env")
	if err != nil {
		t.Fatalf("could not load currencyrates testing file: %v", err)
	}

	config := new(apitesting.Config)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EvmParseOpts(), envOpt)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	apitesting.RatesRun(t, func(ctx context.Context, t *testing.T) {
		currencyRatesClient, err := apitesting.ConnectToOracle(config.GrpcServerAddress)
		require.NoError(t, err, "can't create currency rates service client")

		t.Run("Seed", func(t *testing.T) {
			_, err := currencyRatesClient.PriceStream(ctx, &emptypb.Empty{})
			require.NoError(t, err)
			// TODO: add stream values checks.
		})
	})
}
