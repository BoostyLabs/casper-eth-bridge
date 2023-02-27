package hexutils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tricorn/pkg/hexutils"
)

func TestIsHex(t *testing.T) {
	tests := []struct {
		hash string
		res  bool
	}{
		{
			hash: "018a88e3dd7409f195fd52db2d3cba5d72ca6709bf1d94121bf3748801b40f6f5c",
			res:  true,
		},
		{
			hash: "0x3095F955Da700b96215CFfC9Bc64AB2e69eB7DAB",
			res:  true,
		},
		{
			hash: "018a88e3dd7409f195fd52db2d3cba5d72ca6709bf1d94121bf3748801b40f6f5cb",
			res:  false,
		},
		{
			hash: "",
			res:  false,
		},
	}

	for _, test := range tests {
		res := hexutils.IsHex(test.hash)
		assert.Equal(t, test.res, res)
	}
}

func TestHas0xPrefix(t *testing.T) {
	tests := []struct {
		hash string
		res  bool
	}{
		{
			hash: "3095F955Da700b96215CFfC9Bc64AB2e69eB7DAB",
			res:  false,
		},
		{
			hash: "0x3095F955Da700b96215CFfC9Bc64AB2e69eB7DAB",
			res:  true,
		},
	}

	for _, test := range tests {
		res := hexutils.Has0xPrefix(test.hash)
		assert.Equal(t, test.res, res)
	}
}

func TestHasTagPrefix(t *testing.T) {
	tests := []struct {
		hash string
		res  bool
	}{
		{
			hash: "31783c4d47a3030add05472a685b20c2f8ec2fe64b309b601347be0968167c0d67",
			res:  false,
		},
		{
			hash: "01783c4d47a3030add05472a685b20c2f8ec2fe64b309b601347be0968167c0d67",
			res:  true,
		},
	}

	for _, test := range tests {
		res := hexutils.HasTagPrefix(test.hash)
		assert.Equal(t, test.res, res)
	}
}

func TestToHexString(t *testing.T) {
	tests := []struct {
		str string
		res string
	}{
		{
			str: "31783c4d47a3030add05472a685b20c2f8ec2fe64b309b601347be0968167c0d6",
			res: "031783c4d47a3030add05472a685b20c2f8ec2fe64b309b601347be0968167c0d6",
		},
		{
			str: "01783c4d47a3030add05472a685b20c2f8ec2fe64b309b601347be0968167c0d67",
			res: "01783c4d47a3030add05472a685b20c2f8ec2fe64b309b601347be0968167c0d67",
		},
	}

	for _, test := range tests {
		res := hexutils.ToHexString(test.str)
		assert.Equal(t, test.res, res)
	}
}
