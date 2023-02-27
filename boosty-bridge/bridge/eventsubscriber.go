// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package bridge

import (
	"github.com/google/uuid"

	"tricorn/chains"
)

// EventSubscriber defines event subscriber entity.
type EventSubscriber struct {
	ID uuid.UUID

	EventsChan chan chains.EventVariant
}

// GetID return subscriber id.
func (s *EventSubscriber) GetID() uuid.UUID {
	return s.ID
}

// NotifyWithEvent notifies subscribers with event.
func (s *EventSubscriber) NotifyWithEvent(event chains.EventVariant) {
	s.EventsChan <- event
}

// ReceiveEvents returns subscriber events chan.
func (s *EventSubscriber) ReceiveEvents() chan chains.EventVariant {
	return s.EventsChan
}
