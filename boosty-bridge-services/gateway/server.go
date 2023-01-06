package gateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/gateway/controllers"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/server"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/transfers"
)

// ensures that Server implement server.Server.
var _ server.Server = (*Server)(nil)

var (
	// Error is an error class that indicates internal http server error.
	Error = errs.Class("server")
)

// Config contains configuration for console web server.
type Config struct {
	Address       string `env:"GATEWAY_ADDRESS"`
	WebAppAddress string `env:"WEP_APP_ADDRESS"`
}

// Server represents gateway server.
//
// architecture: Endpoint
type Server struct {
	log    logger.Logger
	config Config

	listener net.Listener
	server   http.Server

	networks  *networks.Service
	transfers *transfers.Service
}

// NewServer is a constructor for gateway server.
func NewServer(config Config, log logger.Logger, listener net.Listener, networks *networks.Service, transfers *transfers.Service) *Server {
	server := &Server{
		log:       log,
		config:    config,
		listener:  listener,
		networks:  networks,
		transfers: transfers,
	}

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v0").Subrouter()

	networksController := controllers.NewNetworks(server.log, server.networks)
	networksRouter := apiRouter.PathPrefix("/networks").Subrouter()
	networksRouter.HandleFunc("", networksController.Connected).Methods(http.MethodGet)
	networksRouter.HandleFunc("/{network-id}/supported-tokens", networksController.SupportedTokens).Methods(http.MethodGet)

	transfersController := controllers.NewTransfers(server.log, server.transfers)
	transfersRouter := apiRouter.PathPrefix("/transfers").Subrouter()
	transfersRouter.HandleFunc("/history/{signature-hex}/{pub-key-hex}", transfersController.History).Methods(http.MethodGet)
	transfersRouter.HandleFunc("/{tx}", transfersController.Info).Methods(http.MethodGet)
	transfersRouter.HandleFunc("/estimate/{sender-network}/{recipient-network}/{token-id}/{amount}", transfersController.Estimate).Methods(http.MethodGet)
	transfersRouter.HandleFunc("/{transfer-id}/{signature-hex}/{pub-key-hex}", transfersController.Cancel).Methods(http.MethodDelete)
	transfersRouter.HandleFunc("/bridge-in-signature", transfersController.BridgeInSignature).Methods(http.MethodPost)

	apiRouter.PathPrefix("/docs/").Handler(http.StripPrefix("/api/v0/docs", http.FileServer(http.Dir("./gateway/docs/console"))))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{config.WebAppAddress},
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet, http.MethodDelete, http.MethodPost},
	})

	server.server = http.Server{
		Handler: c.Handler(router),
	}

	return server
}

// Run starts the server that host webapp and api endpoint.
func (server *Server) Run(ctx context.Context) (err error) {
	server.log.Debug(fmt.Sprintf("running golden-gate gateway api server on %s", server.config.Address))

	var group errgroup.Group
	group.Go(func() error {
		<-ctx.Done()
		server.log.Debug("golden-gate gateway http server gracefully exited")
		return server.server.Shutdown(ctx)
	})
	group.Go(func() error {
		err := server.server.Serve(server.listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return Error.Wrap(err)
	})

	return Error.Wrap(group.Wait())
}

// Close closes server and underlying listener.
func (server *Server) Close() error {
	server.log.Debug("golden-gate gateway http server closed")
	return Error.Wrap(server.server.Close())
}
