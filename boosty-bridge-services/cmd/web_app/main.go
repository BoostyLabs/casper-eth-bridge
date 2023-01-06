package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"

	peer "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/envparse"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger/zaplog"
	web_app_server "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/web_app"
)

// Error is a default error type for golden-gate gateway web-app cli.
var Error = errs.Class("gateway web-app cli")

// Config is the global configuration for golden-gate web-app.
type Config struct {
	Server     web_app_server.Config
	ServerName string `env:"SERVER_NAME"`
}

// commands.
var (
	rootCmd = &cobra.Command{
		Use:   "golden-gate-gateway-web-app",
		Short: "cli for interacting with golden-gate gateway web-app",
	}
	runCmd = &cobra.Command{
		Use:         "run",
		Short:       "runs the gateway",
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
	ctx, cancel := context.WithCancel(context.Background())
	onSigInt(func() {
		// starting graceful exit on context cancellation.
		cancel()
	})

	log := zaplog.NewLog()

	err = godotenv.Overload("./configs/.web.env")
	if err != nil {
		log.Error("could not load config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	config := new(Config)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EthParseOpts(), envOpt)
	if err != nil {
		log.Error("could not parse ENV config", Error.Wrap(err))
		return Error.Wrap(err)
	}

	var server *web_app_server.Server
	{ // server setup.
		listener, err := net.Listen("tcp", config.Server.Address)
		if err != nil {
			return Error.Wrap(err)
		}

		server = web_app_server.NewServer(
			config.Server,
			log,
			listener,
		)
	}

	webapp := peer.New(log, nil, nil, server, config.ServerName)

	return ignoreContextCancellationError(errs.Combine(webapp.Run(ctx), webapp.Close()))
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
