package envparse

import (
	"reflect"

	"github.com/caarlos0/env/v6"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeebo/errs"
)

// InvalidEthAddress indicates that given eth address is invalid.
var InvalidEthAddress = errs.New("invalid evm address")

// evmAddress returns env parsing options for evm address.
func evmAddress() map[reflect.Type]env.ParserFunc {
	parseOpt := make(map[reflect.Type]env.ParserFunc)

	parseAddrFunc := func(v string) (interface{}, error) {
		ok := common.IsHexAddress(v)
		if !ok {
			return nil, InvalidEthAddress
		}

		address := common.HexToAddress(v)

		return address, nil
	}

	parseOpt[reflect.TypeOf(common.Address{})] = parseAddrFunc

	return parseOpt
}

// evmHash returns env parsing options for hash.
func evmHash() map[reflect.Type]env.ParserFunc {
	parseOpt := make(map[reflect.Type]env.ParserFunc)

	parseHashFunc := func(v string) (interface{}, error) {
		hash := common.HexToHash(v)
		return hash, nil
	}

	parseOpt[reflect.TypeOf(common.Hash{})] = parseHashFunc

	return parseOpt
}

// EvmParseOpts returns options for parsing EVM address and hash.
func EvmParseOpts() map[reflect.Type]env.ParserFunc {
	parseOpt := make(map[reflect.Type]env.ParserFunc)

	for key, val := range evmAddress() {
		parseOpt[key] = val
	}

	for key, val := range evmHash() {
		parseOpt[key] = val
	}

	return parseOpt
}
