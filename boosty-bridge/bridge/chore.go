package bridge

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"tricorn/bridge/networks"
	"tricorn/internal/logger"
)

// chore responsible for reading events from connectors.
//
// architecture: Chore
type chore struct {
	log logger.Logger

	service       *Service
	networkBlocks networks.NetworkBlocks
}

// NewChore instantiates chore.
func NewChore(log logger.Logger, service *Service, networkBlocks networks.NetworkBlocks) *chore {
	return &chore{
		log:           log,
		service:       service,
		networkBlocks: networkBlocks,
	}
}

// Run runs events reading from connectors.
func (chore *chore) Run(ctx context.Context, networkName networks.Name, connector Connector) {
	go chore.eventsReading(ctx, networkName, connector)
}

// eventsStreaming runs events streaming and receiving from connector to db accordingly.
func (chore *chore) eventsReading(ctx context.Context, networkName networks.Name, connector Connector) {
	group, ctx := errgroup.WithContext(ctx)
	subscriber := connector.AddEventSubscriber()

	group.Go(func() error {
		fromBlock, err := chore.networkBlocks.Get(ctx, networks.NetworkNameToID[networkName])
		if err != nil && !errors.Is(err, ErrNoNetworkBlock) {
			chore.log.Error("", Error.Wrap(err))
			return nil
		}
		if errors.Is(err, ErrNoNetworkBlock) {
			networkID, ok := networks.NetworkNameToID[networkName]
			if !ok {
				err = fmt.Errorf("no network with such name %v", networkName)
				chore.log.Error("", Error.Wrap(err))
				return nil
			}

			err = chore.service.networkBlocks.Create(ctx, networks.NetworkBlock{
				NetworkID:     networkID,
				LastSeenBlock: 0,
			})
			if err != nil {
				chore.log.Error("", Error.Wrap(err))
				return nil
			}
		}

		err = connector.EventStream(ctx, uint64(fromBlock))
		if err != nil {
			chore.log.Error("", Error.Wrap(err))
			return nil
		}

		return nil
	})

	group.Go(func() error {
		err := chore.receiveEvents(ctx, subscriber, networkName, connector)
		if err != nil {
			chore.log.Error("couldn't receive events", Error.Wrap(err))
			return nil
		}

		return nil
	})
}

// receiveEvents reads events from connector subscriber.
func (chore *chore) receiveEvents(ctx context.Context, subscriber EventSubscriber, networkName networks.Name, connector Connector) error {
	for {
		select {
		case eventFund, ok := <-subscriber.ReceiveEvents():
			if !ok {
				err := Error.New("events chan unexpectedly closed")
				chore.log.Error("", Error.Wrap(err))
				return status.Error(codes.Internal, Error.Wrap(err).Error())
			}

			if err := chore.service.separateEvent(ctx, eventFund, networkName); err != nil {
				chore.log.Error("couldn't separate event", Error.Wrap(err))
				return status.Error(codes.Internal, Error.Wrap(err).Error())
			}

			networkID, ok := networks.NetworkNameToID[networkName]
			if !ok {
				err := Error.New("network %v is not connected", networkName)
				chore.log.Error("", err)
				return status.Error(codes.Internal, Error.Wrap(err).Error())
			}
			err := chore.networkBlocks.Update(ctx, networks.NetworkBlock{
				NetworkID:     networkID,
				LastSeenBlock: int64(eventFund.Block()),
			})
			if err != nil {
				chore.log.Error("couldn't update network block", Error.Wrap(err))
				return status.Error(codes.Internal, Error.Wrap(err).Error())
			}
		case <-ctx.Done():
			connector.RemoveEventSubscriber(subscriber.GetID())
			return nil
		}
	}
}
