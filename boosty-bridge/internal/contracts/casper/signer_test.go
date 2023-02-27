package casper_test

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	casper_chain "tricorn/chains/casper"
	"tricorn/internal/contracts/casper"
	"tricorn/signer"
)

var (
	expectedBridgeInSignature    = "704a998ecd23af18ea1e6b61139975982fd8cedb5625eac9451b47b4658e48df24441537f57102d9f050448de5279d0ee7e4e24cac5e5deaa9c3eeeecbb1fb82"
	expectedTransferOutSignature = "613cec1ab8d753f495f7e512659f4156655ce2a216fc2c67f003530a5b0d18362212fb1fff7ee6a35c54ff8bf3c865a12bab6e872138c424bf5134105fb39618"

	privateKeySecp256k1 = "cc903a2179a5c47acef21d732c0693848c6c33e626fd6651b3773732bde6e127"

	bridgeInTokenPackageHash    = "c62925cefa47af5eb44a2c46de1055315b9371e3b34c47cb4ec6e30d5ab18ef6"
	transferOutTokenPackageHash = "f7d8a923e6de29974a313945d5feedf9b43732ccad5e635d43a4b8b239e6a16f"
	bridgeHash                  = "23e2dafc78abbb9a5159aef578eafd1794774838ddae2cfc8ed5165ee67b471d"

	accountAddress   = "5e681784dab76326249cf0d5f413806c366bbe4ed04508349a2e4b162fdcea5a"
	recipientAddress = "e94daaff79c2ab8d9c31d9c3058d7d0a0dd31204a5638dc1451fa67b2e3fb88c"
)

func TestCasper_GetBridgeInSignature(t *testing.T) {
	ctx := context.Background()

	privateKeyECDSA, err := crypto.HexToECDSA(privateKeySecp256k1)
	require.NoError(t, err)

	signer := casper.NewSigner(func(b []byte, _ signer.Type) ([]byte, error) {
		signature, err := crypto.Sign(b, privateKeyECDSA)
		return signature, err
	})

	t.Run("get bridgeIn signature", func(t *testing.T) {
		bridgeHashBytes, err := hex.DecodeString(bridgeHash)
		require.NoError(t, err)

		tokenPackageHashBytes, err := hex.DecodeString(bridgeInTokenPackageHash)
		require.NoError(t, err)

		accountAddressBytes, err := hex.DecodeString(accountAddress)
		require.NoError(t, err)

		sig, err := signer.GetBridgeInSignature(ctx, casper_chain.BridgeInSignature{
			Prefix:             "TRICORN_BRIDGE_IN",
			BridgeHash:         bridgeHashBytes,
			TokenPackageHash:   tokenPackageHashBytes,
			AccountAddress:     accountAddressBytes,
			Amount:             big.NewInt(1000000000000),
			GasCommission:      big.NewInt(1000),
			Deadline:           big.NewInt(1672943628),
			Nonce:              big.NewInt(555),
			DestinationChain:   "DEST",
			DestinationAddress: "DESTADDR",
		})
		require.NoError(t, err)
		require.NotEmpty(t, sig)
		require.Equal(t, expectedBridgeInSignature, hex.EncodeToString(sig))
	})
}

func TestCasper_GetTransferOutSignature(t *testing.T) {
	ctx := context.Background()

	privateKeyECDSA, err := crypto.HexToECDSA(privateKeySecp256k1)
	require.NoError(t, err)

	signer := casper.NewSigner(func(b []byte, _ signer.Type) ([]byte, error) {
		signature, err := crypto.Sign(b, privateKeyECDSA)
		return signature, err
	})

	t.Run("get transferOut signature", func(t *testing.T) {
		tokenPackageHashChainBytes, err := hex.DecodeString(transferOutTokenPackageHash)
		require.NoError(t, err)

		accountAddressBytes, err := hex.DecodeString(accountAddress)
		require.NoError(t, err)

		recipientBytes, err := hex.DecodeString(recipientAddress)
		require.NoError(t, err)

		sig, err := signer.GetTransferOutSignature(ctx, casper_chain.TransferOutSignature{
			Prefix:           "BBCSP/TR_OUT",
			TokenPackageHash: tokenPackageHashChainBytes,
			AccountAddress:   accountAddressBytes,
			Recipient:        recipientBytes,
			Amount:           big.NewInt(969999999000),
			GasCommission:    big.NewInt(30000001000),
			Nonce:            big.NewInt(556),
		})
		require.NoError(t, err)
		require.NotEmpty(t, sig)
		require.Equal(t, expectedTransferOutSignature, hex.EncodeToString(sig))
	})
}
