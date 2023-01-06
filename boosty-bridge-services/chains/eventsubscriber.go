// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package chains

import (
	"github.com/google/uuid"
)

// EventSubscriber defines event subscriber entity.
type EventSubscriber struct {
	ID uuid.UUID

	EventsChan chan EventVariant
}

// GetID return subscriber id.
func (s *EventSubscriber) GetID() uuid.UUID {
	return s.ID
}

// NotifyWithEvent notifies subscribers with event.
func (s *EventSubscriber) NotifyWithEvent(event EventVariant) {
	s.EventsChan <- event
}

// ReceiveEvents returns subscriber events chan.
func (s *EventSubscriber) ReceiveEvents() chan EventVariant {
	return s.EventsChan
}
