// Copyright (C) 2023 Creditor Corp. Group.
// See LICENSE for copying information.

package currencyrates

import (
	"github.com/google/uuid"
)

// EventSubscriber defines event subscriber entity.
type EventSubscriber struct {
	ID uuid.UUID

	TokenPriceChan chan TokenPrice
}

// GetID return subscriber id.
func (s *EventSubscriber) GetID() uuid.UUID {
	return s.ID
}

// NotifyWithEvent notifies subscribers with event.
func (s *EventSubscriber) NotifyWithEvent(event TokenPrice) {
	s.TokenPriceChan <- event
}

// ReceiveEvents returns subscriber events chan.
func (s *EventSubscriber) ReceiveEvents() chan TokenPrice {
	return s.TokenPriceChan
}
