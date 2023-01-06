// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package peer

import (
	"context"
	"fmt"

	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger"
	server "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/server"
)

// peer is the representation of a server.
type peer struct {
	log logger.Logger

	communication communication.Communication
	service       chains.Connector
	server        server.Server

	serverName string
}

// New is a constructor for peer.
func New(log logger.Logger, communication communication.Communication, service chains.Connector, server server.Server,
	serverName string) *peer {
	return &peer{
		log:           log,
		communication: communication,
		service:       service,
		server:        server,
		serverName:    serverName,
	}
}

// Run runs server until it's either closed or it errors.
func (peer *peer) Run(ctx context.Context) error {
	peer.log.Debug(fmt.Sprintf("golden-gate %s running", peer.serverName))

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return peer.server.Run(ctx)
	})

	return group.Wait()
}

// Close closes all the resources.
func (peer *peer) Close() error {
	peer.log.Debug(fmt.Sprintf("golden-gate %s closing", peer.serverName))
	var errlist errs.Group

	// closes connection with nodes.
	if peer.service != nil {
		peer.service.CloseClient()
		peer.service.CloseWsClient()
	}

	if peer.server != nil {
		errlist.Add(peer.server.Close())
	}

	if peer.communication != nil {
		errlist.Add(peer.communication.Close())
	}

	err := errlist.Err()
	if err != nil {
		peer.log.Error(fmt.Sprintf("could not close golden-gate %s", peer.serverName), err)
		return err
	}

	return nil
}
