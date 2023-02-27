// Copyright (C) 2023 Creditor Corp. Group.
// See LICENSE for copying information.

package chainlink

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/zeebo/errs"

	"tricorn/currencyrates"
)

// ensures that Service implement currencyrates.CurrencyRates.
var _ currencyrates.CurrencyRates = (*Service)(nil)

// Service is a implementation of currencyrates.CurrencyRates.
type Service struct {
	httpClient http.Client

	baseURL string
}

// New is constructor for Service.
func New(baseURL string) currencyrates.CurrencyRates {
	httpClient := http.Client{
		Timeout: time.Second * 10,
	}

	return &Service{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// GetPrice returns the price for `from` token relative to the `to`.
func (s *Service) GetPrice(ctx context.Context, from string, to string) (_ currencyrates.Currency, err error) {
	fullURL := fmt.Sprintf("%s?fsym=%s&tsyms=%s", s.baseURL, from, to)

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return currencyrates.Currency{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return currencyrates.Currency{}, err
	}

	defer func() {
		err = errs.Combine(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		rr, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return currencyrates.Currency{}, err
		}

		return currencyrates.Currency{}, errs.New(string(rr))
	}

	var currency map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&currency); err != nil {
		return currencyrates.Currency{}, err
	}

	if errorMessage, ok := currency["Message"]; ok && currency["Response"] == "Error" {
		return currencyrates.Currency{}, errors.New(errorMessage.(string))
	}

	price, ok := currency[to]
	if !ok {
		return currencyrates.Currency{}, fmt.Errorf("token %s does not exist", to)
	}

	parsedPrice, ok := price.(float64)
	if !ok {
		return currencyrates.Currency{}, fmt.Errorf("token price %s is not parsed as float", to)
	}

	return currencyrates.Currency{
		Symbol: to,
		Price:  parsedPrice,
	}, nil
}

// Convert converts token amount.
func (s *Service) Convert(ctx context.Context, from string, to string, amount *big.Float) (*big.Float, error) {
	if amount.Cmp(big.NewFloat(0)) <= 0 {
		return nil, fmt.Errorf("amount %s is less than or equal to 0", amount.String())
	}

	priceCurrency, err := s.GetPrice(ctx, from, to)
	if err != nil {
		return nil, err
	}

	price := big.NewFloat(0).SetFloat64(priceCurrency.Price)
	amount.Mul(amount, price)

	return amount, nil
}
