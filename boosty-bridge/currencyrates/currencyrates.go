// Copyright (C) 2023 Creditor Corp. Group.
// See LICENSE for copying information.

package currencyrates

import (
	"context"
	"math/big"
	"time"
)

// Config contains currencyrates configurable values.
type Config struct {
	CurrencyRateBaseURL            string `env:"CURRENCY_RATE_BASE_URL"`
	EventsReadingIntervalInSeconds uint32 `env:"EVENTS_READING_INTERVAL_IN_SECONDS"`
}

// TokenPrice describes the token price and decimals at a given time.
type TokenPrice struct {
	TokenName  string
	Amount     string
	Decimals   uint32
	LastUpdate time.Time
}

// CurrencyRates provides access to currency rate functions.
type CurrencyRates interface {
	// GetPrice returns the price for `from` token relative to the `to`.
	GetPrice(ctx context.Context, from string, to string) (Currency, error)
	// Convert converts token amount.
	Convert(ctx context.Context, from string, to string, amount *big.Float) (*big.Float, error)
}

// Currency describes symbol and price of a token.
type Currency struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}
