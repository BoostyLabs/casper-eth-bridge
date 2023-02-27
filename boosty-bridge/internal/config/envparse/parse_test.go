package envparse_test

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tricorn/internal/config/envparse"
)

func TestEthParseOpts(t *testing.T) {
	opts := envparse.EvmParseOpts()
	require.Len(t, opts, 2)

	assert.Contains(t, opts, reflect.TypeOf(common.Address{}))
	assert.Contains(t, opts, reflect.TypeOf(common.Hash{}))
}
