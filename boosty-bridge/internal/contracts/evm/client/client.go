package client

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/BoostyLabs/evmsignature"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	evm_chain "tricorn/chains/evm"
	"tricorn/internal/contracts/evm"
	"tricorn/internal/contracts/evm/bridge"
	"tricorn/signer"
)

// ensures that client implement evm_chain.Transfer.
var _ evm_chain.Transfer = (*client)(nil)

// Config defines configurable values for evm client.
type Config struct {
	NodeAddress           string         `env:"NODE_ADDRESS"`
	BridgeContractAddress common.Address `env:"BRIDGE_CONTRACT_ADDRESS"`
}

// client describes evm client and values to generate and sent transaction.
type client struct {
	config Config

	ethclient *ethclient.Client
	instance  *bridge.Bridge
	auth      *bind.TransactOpts

	signerAddress common.Address
	sign          func([]byte, signer.Type) ([]byte, error)
}

// NewClient is constructor for client.
func NewClient(ctx context.Context, config Config, signerAddress common.Address, sign func([]byte, signer.Type) ([]byte, error)) (evm_chain.Transfer, error) {
	ethclient, err := ethclient.Dial(config.NodeAddress)
	if err != nil {
		return nil, err
	}

	chainID, err := ethclient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	instance, err := bridge.NewBridge(config.BridgeContractAddress, ethclient)
	if err != nil {
		return nil, err
	}

	auth, err := evm.NewKeyedTransactorWithChainID(ctx, signerAddress, chainID, sign)
	if err != nil {
		return nil, err
	}

	return &client{
		config:        config,
		ethclient:     ethclient,
		instance:      instance,
		auth:          auth,
		signerAddress: signerAddress,
		sign:          sign,
	}, nil
}

// TransferOutSignature generates signature for transfer out transaction.
func (c *client) TransferOutSignature(ctx context.Context, transferOut evm_chain.TransferOutRequest) ([]byte, error) {
	var data []byte

	amountStringWithZeros := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", transferOut.Amount))
	amountByte, err := hex.DecodeString(string(amountStringWithZeros))
	if err != nil {
		return nil, err
	}

	commissionStringWithZeros := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", transferOut.Commission))
	commissionByte, err := hex.DecodeString(string(commissionStringWithZeros))
	if err != nil {
		return nil, err
	}

	nonceStringWithZeros := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", transferOut.Nonce))
	nonceByte, err := hex.DecodeString(string(nonceStringWithZeros))
	if err != nil {
		return nil, err
	}

	data = append(data, transferOut.Token.Bytes()...)
	data = append(data, transferOut.Recipient.Bytes()...)
	data = append(data, amountByte...)
	data = append(data, commissionByte...)
	data = append(data, nonceByte...)

	dataHash := crypto.Keccak256Hash(data)

	ethSignedMessageHash := evm.ToEthSignedMessageHash(dataHash.Bytes())

	signature, err := c.sign(ethSignedMessageHash, signer.TypeDTSignature)
	if err != nil {
		return nil, err
	}

	reformedSignature, err := evm.ToEVMSignature(signature)

	return reformedSignature, err
}

// TransferOut initiates outbound bridge transaction only for contract owner.
func (c *client) TransferOut(ctx context.Context, transferOut evm_chain.TransferOutRequest) error {
	var data []byte

	amountStringWithZeros := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", transferOut.Amount))
	amountByte, err := hex.DecodeString(string(amountStringWithZeros))
	if err != nil {
		return err
	}

	commissionStringWithZeros := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", transferOut.Commission))
	commissionByte, err := hex.DecodeString(string(commissionStringWithZeros))
	if err != nil {
		return err
	}

	nonceStringWithZeros := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", transferOut.Nonce))
	nonceByte, err := hex.DecodeString(string(nonceStringWithZeros))
	if err != nil {
		return err
	}

	data = append(data, transferOut.Token.Bytes()...)
	data = append(data, transferOut.Recipient.Bytes()...)
	data = append(data, amountByte...)
	data = append(data, commissionByte...)
	data = append(data, nonceByte...)

	dataHash := crypto.Keccak256Hash(data)

	ethSignedMessageHash := evm.ToEthSignedMessageHash(dataHash.Bytes())

	signature, err := c.sign(ethSignedMessageHash, signer.TypeDTTransaction)
	if err != nil {
		return err
	}

	reformedSignature, err := evm.ToEVMSignature(signature)
	if err != nil {
		return err
	}

	_, err = c.instance.TransferOut(c.auth, transferOut.Token, transferOut.Recipient, transferOut.Amount, transferOut.Commission,
		transferOut.Nonce, reformedSignature)

	return err
}

// GetBridgeInSignature generates signature for inbound bridge transaction.
func (c *client) GetBridgeInSignature(ctx context.Context, bridgeIn evm_chain.GetBridgeInSignatureRequest) ([]byte, error) {
	var data []byte

	amountStringWithZeros := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", bridgeIn.Amount))
	amountByte, err := hex.DecodeString(string(amountStringWithZeros))
	if err != nil {
		return nil, err
	}

	gasCommissionStringWithZeros := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", bridgeIn.GasCommission))
	gasCommissionByte, err := hex.DecodeString(string(gasCommissionStringWithZeros))
	if err != nil {
		return nil, err
	}

	destinationChain := fmt.Sprintf("%x", bridgeIn.DestinationChain)
	destinationChainByte, err := hex.DecodeString(destinationChain)
	if err != nil {
		return nil, err
	}

	destinationAddress := fmt.Sprintf("%x", bridgeIn.DestinationAddress)
	destinationAddressByte, err := hex.DecodeString(destinationAddress)
	if err != nil {
		return nil, err
	}

	deadlineStringWithZeros := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", bridgeIn.Deadline))
	deadlineByte, err := hex.DecodeString(string(deadlineStringWithZeros))
	if err != nil {
		return nil, err
	}

	nonceStringWithZeros := evmsignature.CreateHexStringFixedLength(fmt.Sprintf("%x", bridgeIn.Nonce))
	nonceByte, err := hex.DecodeString(string(nonceStringWithZeros))
	if err != nil {
		return nil, err
	}

	data = append(data, bridgeIn.User.Bytes()...)
	data = append(data, bridgeIn.Token.Bytes()...)
	data = append(data, amountByte...)
	data = append(data, gasCommissionByte...)
	data = append(data, destinationChainByte...)
	data = append(data, destinationAddressByte...)
	data = append(data, deadlineByte...)
	data = append(data, nonceByte...)

	dataHash := crypto.Keccak256Hash(data)

	ethSignedMessageHash := evm.ToEthSignedMessageHash(dataHash.Bytes())

	signature, err := c.sign(ethSignedMessageHash, signer.TypeDTSignature)
	if err != nil {
		return nil, err
	}

	reformedSignature, err := evm.ToEVMSignature(signature)

	return reformedSignature, err
}

// BridgeIn initiates inbound bridge transaction.
func (c *client) BridgeIn(ctx context.Context, bridgeIn evm_chain.BridgeInRequest) (string, error) {
	tx, err := c.instance.BridgeIn(c.auth, bridgeIn.Token, bridgeIn.Amount, bridgeIn.GasCommission, bridgeIn.DestinationChain,
		bridgeIn.DestinationAddress, bridgeIn.Deadline, bridgeIn.Nonce, bridgeIn.Signature)
	if err != nil {
		return "", err
	}

	txHash := hex.EncodeToString(tx.Hash().Bytes())

	return txHash, nil
}

// Close closes underlying client connection.
func (c *client) Close() {
	c.ethclient.Close()
}
