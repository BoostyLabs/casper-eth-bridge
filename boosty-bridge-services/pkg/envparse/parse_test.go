package envparse_test

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/envparse"
)

func TestEthParseOpts(t *testing.T) {
	opts := envparse.EthParseOpts()
	require.Len(t, opts, 2)

	assert.Contains(t, opts, reflect.TypeOf(common.Address{}))
	assert.Contains(t, opts, reflect.TypeOf(common.Hash{}))
}
