// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package apitesting

import (
	"context"
	"github.com/caarlos0/env/v6"
	"testing"
	"time"

	bridge_signerpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-signer"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	peer "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services"
	signer "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/cmd/signer/db"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/database/dbtesting"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/envparse"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger/zaplog"
	signer_service "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/signer"
	chains_server "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/signer/server"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/signer/server/controllers"
)

// SignerConfig contains configurable values for signer project.
type SignerConfig struct {
	Database          string `env:"DATABASE"`
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	Signer            signer_service.Config
	ServerName        string `env:"SERVER_NAME"`
}

func SignerRun(t *testing.T, test func(ctx context.Context, t *testing.T, db signer.DB)) {
	ctx, cancel := context.WithCancel(context.Background())
	log := zaplog.NewLog()

	err := godotenv.Overload("./apitesting/configs/.test.signer.env")
	if err != nil {
		t.Fatalf("could not load signer testing file: %v", err)
	}

	config := new(SignerConfig)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EthParseOpts(), envOpt)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	masterDB := dbtesting.Database{
		Name: "Postgres",
		URL:  "postgresql://postgres:1212@127.0.0.1/private_keys?sslmode=disable",
	}

	db, err := dbtesting.CreateMasterDB(ctx, t.Name(), "Test", 0, masterDB)
	require.NoError(t, err)
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

	err = db.CreateSchema(ctx)
	require.NoError(t, err)

	service := signer_service.NewService(config.Signer, db.PrivateKeys())
	controller := controllers.NewSigner(log, service)

	registerServer := func(grpcServer *grpc.Server) {
		bridge_signerpb.RegisterBridgeSignerServer(grpcServer, controller)
	}
	server := chains_server.NewServer(ctx, log, config.GrpcServerAddress, registerServer)

	signer := peer.New(log, nil, nil, server, config.ServerName)

	var group errgroup.Group
	group.Go(func() error {
		return signer.Run(ctx)
	})

	time.Sleep(time.Second)

	group.Go(func() error {
		test(ctx, t, db)
		cancel()
		return nil
	})

	err = group.Wait()
	if err != nil {
		log.Error("could not run test/server", err)
		return
	}
}
