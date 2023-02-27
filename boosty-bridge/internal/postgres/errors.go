// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package postgres

import (
	"errors"
	"strings"
)

const (
	// pgErrorClassIntegrityConstraintViolation is the class of PostgreSQL errors indicating
	// integrity constraint violations.
	pgErrorClassIntegrityConstraintViolation = "23"
)

// FromError returns the 5-character PostgreSQL error code string associated
// with the given error, if any.
func FromError(err error) string {
	var sqlStateErr errWithSQLState
	if errors.As(err, &sqlStateErr) {
		return sqlStateErr.SQLState()
	}
	return ""
}

// IsConstraintError checks if given error is about constraint violation.
func IsConstraintError(err error) bool {
	errCode := FromError(err)
	return strings.HasPrefix(errCode, pgErrorClassIntegrityConstraintViolation)
}

// errWithSQLState is an interface supported by error classes corresponding
// to PostgreSQL errors from certain drivers. An effort is
// apparently underway to get lib/pq to add this interface.
type errWithSQLState interface {
	SQLState() string
}
