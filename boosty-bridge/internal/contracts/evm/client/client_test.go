package client_test

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"tricorn/chains/evm"
	evm_client "tricorn/internal/contracts/evm/client"
	"tricorn/signer"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	expectedSignature := "648a0aeead8f01567483f201eab348a739f767f75e90fd4ff5079d2f527067975187d5adef2dd030f21558c065b5a11a2b8b8372015196843746c375ad4f8f261c"

	privateKey := "ecff8b9c717a56b30f35a75db85342a1b42fcfe8540a733c73cc9ef38a165a56" // 0x7e0f5A592322Bc973DDE62dF3f91604D21d37446.
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	require.NoError(t, err)

	var publicKey []byte
	publicKey = append(publicKey, privateKeyECDSA.PublicKey.X.Bytes()...)
	publicKey = append(publicKey, privateKeyECDSA.PublicKey.Y.Bytes()...)

	x := big.NewInt(0).SetBytes(publicKey[:32])
	y := big.NewInt(0).SetBytes(publicKey[32:])
	publicKeyECDSA := ecdsa.PublicKey{
		Curve: btcec.S256(),
		X:     x,
		Y:     y,
	}

	signerAddress := crypto.PubkeyToAddress(publicKeyECDSA)

	transfer, err := evm_client.NewClient(
		ctx,
		evm_client.Config{
			NodeAddress:           "https://goerli.infura.io/v3/b6b83346b6274c39ab08e2b2a0260235",
			BridgeContractAddress: common.HexToAddress("0xA0E532456654bcC83F584e0EDf6ba065f87f528F"),
		},
		signerAddress,
		func(data []byte, _ signer.Type) ([]byte, error) {

			signature, err := crypto.Sign(data, privateKeyECDSA)
			if err != nil {
				return nil, err
			}
			return signature, nil
		},
	)
	require.NoError(t, err)

	var signature []byte
	t.Run("get signature for bridgeIn", func(t *testing.T) {
		amount, ok := big.NewInt(0).SetString("1000000000000000000000", 10)
		require.True(t, ok)

		gasCommission, ok := big.NewInt(0).SetString("100000000000000000000", 10)
		require.True(t, ok)

		sig, err := transfer.GetBridgeInSignature(ctx, evm.GetBridgeInSignatureRequest{
			User:               common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8"),
			Token:              common.HexToAddress("0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512"),
			Amount:             amount,
			GasCommission:      gasCommission,
			DestinationChain:   "Solana",
			DestinationAddress: "4zXwdbUDWo1S5AP2CEfv4zAPRds5PQUG1dyqLLvib2xu",
			Deadline:           big.NewInt(1677754455),
			Nonce:              big.NewInt(1), // increase after a successful transaction.
		})
		require.NoError(t, err)
		require.NotEmpty(t, sig)
		require.Equal(t, expectedSignature, hex.EncodeToString(sig))

		signature = sig

	})

	// for manual testing.
	t.Skip("you need enough allowance")
	t.Run("bridge in", func(t *testing.T) {
		amount, ok := big.NewInt(0).SetString("1", 10)
		require.True(t, ok)

		gasCommission, ok := big.NewInt(0).SetString("1", 10)
		require.True(t, ok)

		_, err := transfer.BridgeIn(ctx, evm.BridgeInRequest{
			Token:              common.HexToAddress("0x9fF6D0788066982c95D26F4A74d6C700F3Dc29ec"),
			Amount:             amount,
			GasCommission:      gasCommission,
			DestinationChain:   "Solana",
			DestinationAddress: "4zXwdbUDWo1S5AP2CEfv4zAPRds5PQUG1dyqLLvib2xu",
			Deadline:           big.NewInt(time.Now().AddDate(0, 0, 1).Unix()),
			Nonce:              big.NewInt(0),
			Signature:          signature,
		})
		require.NoError(t, err)
	})

	t.Run("transfer out", func(t *testing.T) {
		amount, ok := big.NewInt(0).SetString("1", 10)
		require.True(t, ok)

		err := transfer.TransferOut(ctx, evm.TransferOutRequest{
			Token:     common.HexToAddress("0x9fF6D0788066982c95D26F4A74d6C700F3Dc29ec"),
			Recipient: signerAddress,
			Amount:    amount,
			Nonce:     big.NewInt(1),
		})
		require.NoError(t, err)
	})
}
