// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package server

import (
	"context"
	"fmt"
	"net"

	"github.com/oklog/run"
	"google.golang.org/grpc"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/server"
)

// ensures that Server implement server.Server.
var _ server.Server = (*Server)(nil)

// defaultGrpcMessageSize defines the max message size in bytes the server can receive.
const defaultGrpcMessageSize = 100 * 1024 * 1024 // 1 Mib.

// Server is the representation of connector server.
type Server struct {
	log logger.Logger

	grpcServer *grpc.Server

	serverAddress string
}

// NewServer is a constructor for Server.
func NewServer(ctx context.Context, logger logger.Logger, serverAddress string, registerServer func(*grpc.Server)) *Server {
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(defaultGrpcMessageSize),
		grpc.MaxSendMsgSize(defaultGrpcMessageSize),
	)

	registerServer(grpcServer)

	return &Server{
		log:           logger,
		serverAddress: serverAddress,
		grpcServer:    grpcServer,
	}
}

// Run runs connector server until it's either closed or it errors.
func (s *Server) Run(ctx context.Context) error {
	s.log.Debug("golden-gate signer server running")

	var g run.Group

	// GRPC endpoints.
	{
		g.Add(func() error {
			s.log.Debug("Start GRPC endpoints")

			lis, err := net.Listen("tcp", s.serverAddress)
			if err != nil {
				return fmt.Errorf("failed to listen: %v", err)
			}

			return s.grpcServer.Serve(lis)
		}, func(err error) {
			s.log.Debug("Stop GRPC endpoints")
			s.grpcServer.GracefulStop()
		})
	}

	{
		g.Add(func() error {
			<-ctx.Done()
			return nil
		}, func(err error) {})
	}

	s.log.Debug(fmt.Sprintf("The signer server was terminated with: %v", g.Run()))

	return nil
}

// Close closes underlying server connection.
func (s *Server) Close() error {
	s.log.Debug("golden-gate signer server closing")
	s.grpcServer.GracefulStop()

	return nil
}
