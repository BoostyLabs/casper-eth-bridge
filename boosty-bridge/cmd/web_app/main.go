package main

import (
	"context"
	"errors"
	"net"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"

	"tricorn"
	"tricorn/internal/config/envparse"
	"tricorn/internal/logger/zaplog"
	"tricorn/internal/process"
	web_app "tricorn/web_app"
)

// Error is a default error type for golden-gate gateway web-app cli.
var Error = errs.Class("gateway web-app cli")

// Config is the global configuration for golden-gate web-app.
type Config struct {
	Server     web_app.Config
	ServerName string `env:"SERVER_NAME"`
}

// commands.
var (
	rootCmd = &cobra.Command{
		Use:   "tricorn-gateway-web-app",
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
	process.OnSigInt(func() {
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
	err = env.ParseWithFuncs(config, envparse.EvmParseOpts(), envOpt)
	if err != nil {
		log.Error("could not parse ENV config", Error.Wrap(err))
		return Error.Wrap(err)
	}

	var server *web_app.Server
	{ // server setup.
		listener, err := net.Listen("tcp", config.Server.Address)
		if err != nil {
			return Error.Wrap(err)
		}

		server = web_app.NewServer(
			config.Server,
			log,
			listener,
		)
	}

	webapp := tricorn.New(log, nil, nil, server, config.ServerName)

	return ignoreContextCancellationError(errs.Combine(webapp.Run(ctx), webapp.Close()))
}

// ignoreContextCancellationError ignores cancellation and stopping errors since they are expected.
func ignoreContextCancellationError(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
