// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package zaplog

import (
	"go.uber.org/zap"

	"tricorn/internal/logger"
)

// ensures that zaplog implements logger.Logger.
var _ logger.Logger = (*zaplog)(nil)

// zaplog is an implementation of logger.Logger using zap.
type zaplog struct {
	client *zap.Logger
}

// NewLog is a constructor for a logger.Logger.
func NewLog() logger.Logger {
	return &zaplog{
		client: zap.NewExample(),
	}
}

// Debug is used to send formatted debug message.
func (log zaplog) Debug(msg string) {
	log.client.Debug(msg)
}

// Warn is used to send formatted warn message.
func (log zaplog) Warn(msg string) {
	log.client.Warn(msg)
}

// Error is used to send formatted as error message.
func (log zaplog) Error(msg string, err error) {
	log.client.Error(msg, zap.Error(err))
}
