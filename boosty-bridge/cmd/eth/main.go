// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/btcsuite/btcd/btcec"
	"github.com/caarlos0/env/v6"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
	"google.golang.org/grpc"

	bridge_connectorpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-connector"

	"tricorn"
	"tricorn/bridge/networks"
	"tricorn/chains"
	"tricorn/chains/communication/controllers"
	"tricorn/chains/evm"
	"tricorn/communication"
	"tricorn/communication/mockcommunication"
	"tricorn/communication/rpc"
	"tricorn/internal/config/envparse"
	"tricorn/internal/contracts/evm/bridge"
	"tricorn/internal/contracts/evm/client"
	"tricorn/internal/logger/zaplog"
	"tricorn/internal/server"
	grpc_server "tricorn/internal/server/grpc"
	"tricorn/signer"
)

// Error is a default error type for eth connector cli.
var Error = errs.Class("eth connector cli")

// Config is the global configuration to run connector server.
type Config struct {
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	Service           evm.Config
	Communication     rpc.Config
	CommunicationMode communication.Mode `env:"COMMUNICATION_MODE"`
	ServerName        string             `env:"SERVER_NAME"`
	Bridge            client.Config
}

// commands.
var (
	rootCmd = &cobra.Command{
		Use:   "connector",
		Short: "cli for interacting with eth connector project",
	}
	runCmd = &cobra.Command{
		Use:         "run",
		Short:       "runs the program",
		RunE:        cmdRun,
		Annotations: map[string]string{"type": "run"},
	}
)

func init() {
	rootCmd.AddCommand(runCmd)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func cmdRun(cmd *cobra.Command, args []string) (err error) {
	var (
		comm          communication.Communication
		service       *evm.Service
		server        server.Server
		signerAddress common.Address
	)

	ctx, cancel := context.WithCancel(context.Background())
	onSigInt(func() {
		// starting graceful exit on context cancellation.
		cancel()
	})

	log := zaplog.NewLog()

	err = godotenv.Overload("./configs/.eth.env")
	if err != nil {
		log.Error("could not load eth config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	err = godotenv.Overload("./configs/.env")
	if err != nil {
		log.Error("could not load config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	config := new(Config)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EvmParseOpts(), envOpt)
	if err != nil {
		log.Error("could not parse config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	{ // Communication setup.
		switch config.CommunicationMode {
		case communication.ModeGRPC:
			comm, err = rpc.New(config.Communication, log, true)
			if err != nil {
				return Error.Wrap(err)
			}

			// TODO: fix it.
			publicKey, err := comm.Bridge().PublicKey(ctx, networks.TypeEVM)
			if err != nil {
				return Error.Wrap(err)
			}

			if len(publicKey) < evm.CurveCoordinatesSize {
				return Error.New("invalid public key curve coordinates")
			}

			x := big.NewInt(0).SetBytes(publicKey[:32])
			y := big.NewInt(0).SetBytes(publicKey[32:])
			publicKeyECDSA := ecdsa.PublicKey{
				Curve: btcec.S256(),
				X:     x,
				Y:     y,
			}

			signerAddress = crypto.PubkeyToAddress(publicKeyECDSA)
		default:
			comm = mockcommunication.New()
			signerAddress = common.Address{}
		}
	}

	{ // Eth server setup.
		// connect client to default http connection node.
		ethClient, err := ethclient.Dial(config.Service.NodeAddress)
		if err != nil {
			return Error.Wrap(err)
		}

		instance, err := bridge.NewBridge(config.Service.BridgeContractAddress, ethClient)
		if err != nil {
			return Error.Wrap(err)
		}

		sign := func(data []byte, dataType signer.Type) ([]byte, error) {
			singIn := chains.SignRequest{
				// TODO: fix it.
				NetworkId: networks.TypeEVM,
				Data:      data,
				DataType:  dataType,
			}

			return comm.Bridge().Sign(ctx, singIn)
		}

		transfer, err := client.NewClient(ctx, config.Bridge, signerAddress, sign)
		if err != nil {
			return Error.Wrap(err)
		}

		service = evm.New(
			ctx,
			config.Service,
			log,
			comm.Bridge(),
			instance,
			transfer,
			ethClient,
		)
	}

	{ // Server setup.
		controller := controllers.NewConnector(ctx, log, service)

		registerServer := func(grpcServer *grpc.Server) {
			bridge_connectorpb.RegisterConnectorServer(grpcServer, controller)
		}

		server = grpc_server.NewServer(log, registerServer, config.ServerName, config.GrpcServerAddress)
	}

	connector := tricorn.New(log, comm, service, server, config.ServerName)

	return ignoreContextCancellationError(errs.Combine(connector.Run(ctx), connector.Close()))
}

// onSigInt fires in SIGINT or SIGTERM event (usually CTRL+C).
func onSigInt(onSigInt func()) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		onSigInt()
	}()
}

// ignoreContextCancellationError ignores cancellation and stopping errors since they are expected.
func ignoreContextCancellationError(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
