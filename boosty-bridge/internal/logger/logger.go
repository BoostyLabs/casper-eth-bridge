// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package logger

// Logger exposes functionality to write messages in stdout.
type Logger interface {
	// Error is used to send formatted as error message.
	Error(msg string, err error)
	// Debug is used to send formatted debug message.
	Debug(msg string)
	// Warn is used to send formatted warning message.
	Warn(msg string)
}
