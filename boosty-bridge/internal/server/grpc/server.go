// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/oklog/run"
	"github.com/zeebo/errs"
	"google.golang.org/grpc"

	"tricorn/internal/logger"
	"tricorn/internal/server"
)

// ensures that Server implement server.Server.
var _ server.Server = (*grpcserver)(nil)

// defaultGrpcMessageSize defines the max message size in bytes the grpcserver can receive.
const defaultGrpcMessageSize = 100 * 1024 * 1024 // 1 Mib.

// Error indicates that an error occurred in the GRPC server.
var Error = errs.Class("signer service")

// grpcserver is an implementation GRPC server.
type grpcserver struct {
	log logger.Logger

	server *grpc.Server

	address string
	name    string
}

// NewServer is a constructor for GRPC server.
func NewServer(logger logger.Logger, registerServer func(*grpc.Server), name, address string) *grpcserver {
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(defaultGrpcMessageSize),
		grpc.MaxSendMsgSize(defaultGrpcMessageSize),
	)

	registerServer(grpcServer)

	return &grpcserver{
		log:     logger,
		address: address,
		name:    name,
		server:  grpcServer,
	}
}

// Run runs connector server until it's either closed or it errors.
func (s *grpcserver) Run(ctx context.Context) error {
	s.log.Debug("GRPC server running")

	var g run.Group

	// GRPC endpoints.
	{
		g.Add(func() error {
			s.log.Debug(fmt.Sprintf("starting %s GRPC endpoints", s.name))

			listener, err := net.Listen("tcp", s.address)
			if err != nil {
				return Error.Wrap(fmt.Errorf("failed to listen: %v", err))
			}

			return s.server.Serve(listener)
		}, func(err error) {
			s.log.Debug("stopping GRPC endpoints")
			s.server.GracefulStop()
		})
	}

	{
		g.Add(func() error {
			<-ctx.Done()
			return nil
		}, func(err error) {})
	}

	s.log.Debug(fmt.Sprintf("%s GRPC server was terminated with: %v", s.name, g.Run()))

	return nil
}

// Close closes underlying server connection.
func (s *grpcserver) Close() error {
	s.log.Debug(fmt.Sprintf("%s GRPC server closing", s.name))
	s.server.GracefulStop()

	return nil
}
