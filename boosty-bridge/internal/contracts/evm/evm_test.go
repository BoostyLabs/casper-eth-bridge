package evm_test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/BoostyLabs/evmsignature"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"tricorn/internal/contracts/evm"
)

func TestClient(t *testing.T) {
	t.Run("toEthSignedMessageHash", func(t *testing.T) {
		ethSignedMessageHash := evm.ToEthSignedMessageHash([]byte("1212"))

		expectedHash, err := hex.DecodeString("c4e8036a01f1b83ad304f8738e5cf0cf99dbe114cb05083ec89ce18c5860844e")
		require.NoError(t, err)

		require.Equal(t, expectedHash, ethSignedMessageHash)
	})

	t.Run("ToEVMSignature", func(t *testing.T) {
		signature, err := hex.DecodeString("aac4e8036a01f1b83ad304f8738e5cf0cf99dbe114cb05083ec89ce18c5860844ec4e8036a01f1b83ad304f8738e5cf0cf99dbe114cb05083ec89ce18c58608400")
		require.NoError(t, err)

		ethSignedMessageHash, err := evm.ToEVMSignature(signature)
		require.NoError(t, err)

		expectedHash, err := hex.DecodeString("aac4e8036a01f1b83ad304f8738e5cf0cf99dbe114cb05083ec89ce18c5860844ec4e8036a01f1b83ad304f8738e5cf0cf99dbe114cb05083ec89ce18c5860841b")
		require.NoError(t, err)

		require.Equal(t, expectedHash, ethSignedMessageHash)
	})

	t.Run("compare addresses", func(t *testing.T) {
		address := "e7f1725E7734CE288F8367e1Bb143E90bb3F0512"

		address1, err := hex.DecodeString(strings.ToLower(address))
		require.NoError(t, err)

		address2 := common.HexToAddress(address).Bytes()

		require.Equal(t, address1, address2)
	})

	t.Run("compare numbers", func(t *testing.T) {
		number := 999

		require.Equal(t, "3e7", fmt.Sprintf("%x", number))
		number1 := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", number))
		require.EqualValues(t, "00000000000000000000000000000000000000000000000000000000000003e7", number1)

		numberByte1, err := hex.DecodeString(string(number1))
		require.NoError(t, err)

		numberByte2 := big.NewInt(int64(number)).Bytes()

		require.NotEqual(t, numberByte1, numberByte2)
	})

	t.Run("compare strings", func(t *testing.T) {
		destinationChain := "Solana"

		require.Equal(t, "536f6c616e61", fmt.Sprintf("%x", destinationChain))

		destinationChain1 := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", destinationChain))

		destinationChainByte1, err := hex.DecodeString(string(destinationChain1))
		require.NoError(t, err)

		destinationChainByte2 := []byte(destinationChain)

		require.NotEqual(t, destinationChainByte1, destinationChainByte2)
	})
}
