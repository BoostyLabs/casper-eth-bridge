package casper

import (
	"context"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/casper-ecosystem/casper-golang-sdk/keypair"
	"github.com/casper-ecosystem/casper-golang-sdk/sdk"
	"github.com/casper-ecosystem/casper-golang-sdk/serialization"
	"github.com/casper-ecosystem/casper-golang-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"tricorn/bridge/networks"
	casper_chain "tricorn/chains/casper"
)

// Transfer describes sign func to sign transaction and casper client to send transaction.
type Transfer struct {
	casper casper_chain.Casper

	sign func([]byte) ([]byte, error)
}

// NewTransfer is constructor for Transfer.
func NewTransfer(casper casper_chain.Casper, sign func([]byte) ([]byte, error)) *Transfer {
	return &Transfer{
		casper: casper,
		sign:   sign,
	}
}

const (
	// Ed25519PublicKeySize defines that Ed25519 public key bytes size is 32 bytes.
	Ed25519PublicKeySize = 32
)

// BridgeInRequest describes values to calls bridgeIn method.
type BridgeInRequest struct {
	PublicKey                   keypair.PublicKey
	ChainName                   string
	StandardPaymentForBridgeOut int64
	BridgeContractPackageHash   string

	TokenContractAddress []byte
	Amount               *big.Int
	GasCommission        *big.Int
	Deadline             *big.Int
	Nonce                *big.Int
	DestinationChain     string
	DestinationAddress   string
	Signature            []byte
}

// BridgeIn calls inbound bridge transaction.
func (t *Transfer) BridgeIn(ctx context.Context, bridgeIn BridgeInRequest) (string, error) {
	deployParams := sdk.NewDeployParams(bridgeIn.PublicKey, strings.ToLower(bridgeIn.ChainName), nil, 0)
	payment := sdk.StandardPayment(big.NewInt(bridgeIn.StandardPaymentForBridgeOut))

	// token contract.
	tokenContractFixedBytes := types.FixedByteArray(bridgeIn.TokenContractAddress)
	tokenContract := types.CLValue{
		Type:      types.CLTypeByteArray,
		ByteArray: &tokenContractFixedBytes,
	}
	tokenContractBytes, err := serialization.Marshal(tokenContract)
	if err != nil {
		return "", err
	}

	// amount.
	amount := types.CLValue{
		Type: types.CLTypeU256,
		U256: bridgeIn.Amount,
	}
	amountBytes, err := serialization.Marshal(amount)
	if err != nil {
		return "", err
	}

	// gas commission.
	gasCommission := types.CLValue{
		Type: types.CLTypeU256,
		U256: bridgeIn.GasCommission,
	}
	gasCommissionBytes, err := serialization.Marshal(gasCommission)
	if err != nil {
		return "", err
	}

	// deadline.
	deadline := types.CLValue{
		Type: types.CLTypeU256,
		U256: bridgeIn.Deadline,
	}
	deadlineBytes, err := serialization.Marshal(deadline)
	if err != nil {
		return "", err
	}

	// nonce.
	nonce := types.CLValue{
		Type: types.CLTypeU128,
		U128: bridgeIn.Nonce,
	}
	nonceBytes, err := serialization.Marshal(nonce)
	if err != nil {
		return "", err
	}

	// destination chain.
	destinationChain := types.CLValue{
		Type:   types.CLTypeString,
		String: &bridgeIn.DestinationChain,
	}
	destinationChainBytes, err := serialization.Marshal(destinationChain)
	if err != nil {
		return "", err
	}

	// destination address.
	destinationAddress := types.CLValue{
		Type:   types.CLTypeString,
		String: &bridgeIn.DestinationAddress,
	}
	destinationAddressBytes, err := serialization.Marshal(destinationAddress)
	if err != nil {
		return "", err
	}

	// signature.
	signatureFixedBytes := types.FixedByteArray(bridgeIn.Signature)
	signature := types.CLValue{
		Type:      types.CLTypeByteArray,
		ByteArray: &signatureFixedBytes,
	}
	signatureBytes, err := serialization.Marshal(signature)
	if err != nil {
		return "", err
	}

	args := map[string]sdk.Value{
		"token_contract": {
			IsOptional:  false,
			Tag:         types.CLTypeByteArray,
			StringBytes: hex.EncodeToString(tokenContractBytes),
		},
		"amount": {
			Tag:         types.CLTypeU256,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(amountBytes),
		},
		"gas_commission": {
			Tag:         types.CLTypeU256,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(gasCommissionBytes),
		},
		"deadline": {
			Tag:         types.CLTypeU256,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(deadlineBytes),
		},
		"nonce": {
			Tag:         types.CLTypeU128,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(nonceBytes),
		},
		"destination_chain": {
			Tag:         types.CLTypeString,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(destinationChainBytes),
		},
		"destination_address": {
			Tag:         types.CLTypeString,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(destinationAddressBytes),
		},
		"signature": {
			Tag:         types.CLTypeByteArray,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(signatureBytes),
		},
	}

	keyOrder := []string{
		"token_contract",
		"amount",
		"gas_commission",
		"deadline",
		"nonce",
		"destination_chain",
		"destination_address",
		"signature",
	}
	runtimeArgs := sdk.NewRunTimeArgs(args, keyOrder)

	contractHexBytes, err := hex.DecodeString(bridgeIn.BridgeContractPackageHash)
	if err != nil {
		return "", err
	}

	var contractHashBytes [32]byte
	copy(contractHashBytes[:], contractHexBytes)
	session := sdk.NewStoredContractByHash(contractHashBytes, "bridge_in", *runtimeArgs)

	deploy := sdk.MakeDeploy(deployParams, payment, session)

	signedTx, err := t.sign(deploy.Hash)
	if err != nil {
		return "", err
	}

	signatureKeypair := keypair.Signature{
		Tag:           keypair.KeyTagEd25519,
		SignatureData: signedTx,
	}
	approval := sdk.Approval{
		Signer:    bridgeIn.PublicKey,
		Signature: signatureKeypair,
	}
	deploy.Approvals = append(deploy.Approvals, approval)

	hash, err := t.casper.PutDeploy(*deploy)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// SetSignerRequest describes values to calls setSigner method.
type SetSignerRequest struct {
	PublicKey                   keypair.PublicKey
	ChainName                   string
	StandardPaymentForBridgeOut int64
	BridgeContractPackageHash   string
	Value                       string
}

// SetSigner sets public key in contract to verify signature.
func (t *Transfer) SetSigner(ctx context.Context, req SetSignerRequest) (string, error) {
	deployParams := sdk.NewDeployParams(req.PublicKey, strings.ToLower(req.ChainName), nil, 0)
	payment := sdk.StandardPayment(big.NewInt(req.StandardPaymentForBridgeOut))

	value := types.CLValue{
		Type:   types.CLTypeString,
		String: &req.Value,
	}
	valueBytes, err := serialization.Marshal(value)
	if err != nil {
		return "", err
	}

	args := map[string]sdk.Value{
		"signer": {
			Tag:         types.CLTypeString,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(valueBytes),
		},
	}
	keyOrder := []string{"signer"}
	runtimeArgs := sdk.NewRunTimeArgs(args, keyOrder)

	contractHexBytes, err := hex.DecodeString(req.BridgeContractPackageHash)
	if err != nil {
		return "", err
	}

	var contractHashBytes [32]byte
	copy(contractHashBytes[:], contractHexBytes)
	session := sdk.NewStoredContractByHash(contractHashBytes, "set_signer", *runtimeArgs)

	deploy := sdk.MakeDeploy(deployParams, payment, session)

	signedTx, err := t.sign(deploy.Hash)
	if err != nil {
		return "", err
	}

	signatureKeypair := keypair.Signature{
		Tag:           keypair.KeyTagEd25519,
		SignatureData: signedTx,
	}
	approval := sdk.Approval{
		Signer:    req.PublicKey,
		Signature: signatureKeypair,
	}
	deploy.Approvals = append(deploy.Approvals, approval)

	hash, err := t.casper.PutDeploy(*deploy)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// SetStableCommissionPercentRequest describes values to calls setStableCommissionPercent method.
type SetStableCommissionPercentRequest struct {
	PublicKey                   keypair.PublicKey
	ChainName                   string
	StandardPaymentForBridgeOut int64
	BridgeContractPackageHash   string
	CommissionPercent           *big.Int
}

// SetStableCommissionPercent sets commission percent in contract.
func (t *Transfer) SetStableCommissionPercent(ctx context.Context, req SetStableCommissionPercentRequest) (string, error) {
	deployParams := sdk.NewDeployParams(req.PublicKey, strings.ToLower(req.ChainName), nil, 0)
	payment := sdk.StandardPayment(big.NewInt(req.StandardPaymentForBridgeOut))

	commissionPercent := types.CLValue{
		Type: types.CLTypeU256,
		U256: req.CommissionPercent,
	}
	commissionPercentBytes, err := serialization.Marshal(commissionPercent)
	if err != nil {
		return "", err
	}

	args := map[string]sdk.Value{
		"stable_commission_percent": {
			Tag:         types.CLTypeU256,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(commissionPercentBytes),
		},
	}
	keyOrder := []string{"stable_commission_percent"}
	runtimeArgs := sdk.NewRunTimeArgs(args, keyOrder)

	contractHexBytes, err := hex.DecodeString(req.BridgeContractPackageHash)
	if err != nil {
		return "", err
	}

	var contractHashBytes [32]byte
	copy(contractHashBytes[:], contractHexBytes)
	session := sdk.NewStoredContractByHash(contractHashBytes, "set_stable_commission_percent", *runtimeArgs)

	deploy := sdk.MakeDeploy(deployParams, payment, session)

	signedTx, err := t.sign(deploy.Hash)
	if err != nil {
		return "", err
	}

	signatureKeypair := keypair.Signature{
		Tag:           keypair.KeyTagEd25519,
		SignatureData: signedTx,
	}
	approval := sdk.Approval{
		Signer:    req.PublicKey,
		Signature: signatureKeypair,
	}
	deploy.Approvals = append(deploy.Approvals, approval)

	hash, err := t.casper.PutDeploy(*deploy)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// BridgeOutRequest describes values to initiate outbound bridge transaction.
type BridgeOutRequest struct {
	PublicKey                   keypair.PublicKey
	ChainName                   string
	StandardPaymentForBridgeOut int64
	BridgeContractPackageHash   string

	Amount        *big.Int
	Token         common.Hash
	To            common.Hash
	From          networks.Address
	TransactionID *big.Int
}

// BridgeOut initiates outbound bridge transaction.
func (t *Transfer) BridgeOut(ctx context.Context, req BridgeOutRequest) (string, error) {
	deployParams := sdk.NewDeployParams(req.PublicKey, strings.ToLower(req.ChainName), nil, 0)
	payment := sdk.StandardPayment(big.NewInt(req.StandardPaymentForBridgeOut))

	// token contract.
	tokenContractFixedBytes := types.FixedByteArray(req.Token.Bytes())
	tokenContract := types.CLValue{
		Type:      types.CLTypeByteArray,
		ByteArray: &tokenContractFixedBytes,
	}
	tokenContractBytes, err := serialization.Marshal(tokenContract)
	if err != nil {
		return "", err
	}

	// amount.
	amount := types.CLValue{
		Type: types.CLTypeU256,
		U256: req.Amount,
	}
	amountBytes, err := serialization.Marshal(amount)
	if err != nil {
		return "", err
	}

	// transaction id.
	transactionID := types.CLValue{
		Type: types.CLTypeU256,
		U256: req.TransactionID,
	}
	transactionIDBytes, err := serialization.Marshal(transactionID)
	if err != nil {
		return "", err
	}

	//  source chain.
	sourceChain := types.CLValue{
		Type:   types.CLTypeString,
		String: &req.From.NetworkName,
	}
	sourceChainBytes, err := serialization.Marshal(sourceChain)
	if err != nil {
		return "", err
	}

	// source address.
	sourceAddress := types.CLValue{
		Type:   types.CLTypeString,
		String: &req.From.Address,
	}
	sourceAddressBytes, err := serialization.Marshal(sourceAddress)
	if err != nil {
		return "", err
	}

	// recipient.
	var recipientHashBytes [32]byte
	copy(recipientHashBytes[:], req.To.Bytes())

	recipient := types.CLValue{
		Type: types.CLTypeKey,
		Key: &types.Key{
			Type:    types.KeyTypeAccount,
			Account: recipientHashBytes,
		},
	}
	recipientBytes, err := serialization.Marshal(recipient)
	if err != nil {
		return "", err
	}

	args := map[string]sdk.Value{
		"token_contract": {
			IsOptional:  false,
			Tag:         types.CLTypeByteArray,
			StringBytes: hex.EncodeToString(tokenContractBytes),
		},
		"amount": {
			Tag:         types.CLTypeU256,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(amountBytes),
		},
		"transaction_id": {
			Tag:         types.CLTypeU256,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(transactionIDBytes),
		},
		"source_chain": {
			Tag:         types.CLTypeString,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(sourceChainBytes),
		},
		"source_address": {
			Tag:         types.CLTypeString,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(sourceAddressBytes),
		},
		"recipient": {
			Tag:         types.CLTypeKey,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(recipientBytes),
		},
	}

	keyOrder := []string{
		"token_contract",
		"amount",
		"transaction_id",
		"source_chain",
		"source_address",
		"recipient",
	}
	runtimeArgs := sdk.NewRunTimeArgs(args, keyOrder)

	contractHexBytes, err := hex.DecodeString(req.BridgeContractPackageHash)
	if err != nil {
		return "", err
	}

	var contractHashBytes [32]byte
	copy(contractHashBytes[:], contractHexBytes)
	session := sdk.NewStoredContractByHash(contractHashBytes, "bridge_out", *runtimeArgs)

	deploy := sdk.MakeDeploy(deployParams, payment, session)

	signature, err := t.sign(deploy.Hash)
	if err != nil {
		return "", err
	}

	signatureKeypair := keypair.Signature{
		Tag:           keypair.KeyTagEd25519,
		SignatureData: signature,
	}
	approval := sdk.Approval{
		Signer:    req.PublicKey,
		Signature: signatureKeypair,
	}
	deploy.Approvals = append(deploy.Approvals, approval)

	hash, err := t.casper.PutDeploy(*deploy)
	if err != nil {
		return "", err
	}

	return hash, err
}

// TransferOutRequest describes values to calls transferOut method.
type TransferOutRequest struct {
	PublicKey                   keypair.PublicKey
	ChainName                   string
	StandardPaymentForBridgeOut int64
	BridgeContractPackageHash   string

	TokenContractAddress []byte
	Amount               *big.Int
	GasCommission        *big.Int
	Nonce                *big.Int
	Recipient            []byte
	Signature            []byte
}

// TransferOut calls transferOut method.
func (t *Transfer) TransferOut(ctx context.Context, transferOut TransferOutRequest) (string, error) {
	deployParams := sdk.NewDeployParams(transferOut.PublicKey, strings.ToLower(transferOut.ChainName), nil, 0)
	payment := sdk.StandardPayment(big.NewInt(transferOut.StandardPaymentForBridgeOut))

	// token contract.
	tokenContractFixedBytes := types.FixedByteArray(transferOut.TokenContractAddress)
	tokenContract := types.CLValue{
		Type:      types.CLTypeByteArray,
		ByteArray: &tokenContractFixedBytes,
	}
	tokenContractBytes, err := serialization.Marshal(tokenContract)
	if err != nil {
		return "", err
	}

	// amount.
	amount := types.CLValue{
		Type: types.CLTypeU256,
		U256: transferOut.Amount,
	}
	amountBytes, err := serialization.Marshal(amount)
	if err != nil {
		return "", err
	}

	// gas commission.
	gasCommission := types.CLValue{
		Type: types.CLTypeU256,
		U256: transferOut.GasCommission,
	}
	gasCommissionBytes, err := serialization.Marshal(gasCommission)
	if err != nil {
		return "", err
	}

	// nonce.
	nonce := types.CLValue{
		Type: types.CLTypeU128,
		U128: transferOut.Nonce,
	}
	nonceBytes, err := serialization.Marshal(nonce)
	if err != nil {
		return "", err
	}

	// recipient.
	var recipientHashBytes [32]byte
	copy(recipientHashBytes[:], transferOut.Recipient)

	recipient := types.CLValue{
		Type: types.CLTypeKey,
		Key: &types.Key{
			Type:    types.KeyTypeAccount,
			Account: recipientHashBytes,
		},
	}
	recipientBytes, err := serialization.Marshal(recipient)
	if err != nil {
		return "", err
	}

	// signature.
	signatureFixedBytes := types.FixedByteArray(transferOut.Signature)
	signature := types.CLValue{
		Type:      types.CLTypeByteArray,
		ByteArray: &signatureFixedBytes,
	}
	signatureBytes, err := serialization.Marshal(signature)
	if err != nil {
		return "", err
	}

	args := map[string]sdk.Value{
		"token_contract": {
			IsOptional:  false,
			Tag:         types.CLTypeByteArray,
			StringBytes: hex.EncodeToString(tokenContractBytes),
		},
		"amount": {
			Tag:         types.CLTypeU256,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(amountBytes),
		},
		"commission": {
			Tag:         types.CLTypeU256,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(gasCommissionBytes),
		},
		"nonce": {
			Tag:         types.CLTypeU128,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(nonceBytes),
		},
		"recipient": {
			Tag:         types.CLTypeKey,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(recipientBytes),
		},
		"signature": {
			Tag:         types.CLTypeByteArray,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(signatureBytes),
		},
	}

	keyOrder := []string{
		"token_contract",
		"amount",
		"commission",
		"nonce",
		"recipient",
		"signature",
	}
	runtimeArgs := sdk.NewRunTimeArgs(args, keyOrder)

	contractHexBytes, err := hex.DecodeString(transferOut.BridgeContractPackageHash)
	if err != nil {
		return "", err
	}

	var contractHashBytes [32]byte
	copy(contractHashBytes[:], contractHexBytes)
	session := sdk.NewStoredContractByHash(contractHashBytes, "transfer_out", *runtimeArgs)

	deploy := sdk.MakeDeploy(deployParams, payment, session)

	signedTx, err := t.sign(deploy.Hash)
	if err != nil {
		return "", err
	}

	signatureKeypair := keypair.Signature{
		Tag:           keypair.KeyTagEd25519,
		SignatureData: signedTx,
	}
	approval := sdk.Approval{
		Signer:    transferOut.PublicKey,
		Signature: signatureKeypair,
	}
	deploy.Approvals = append(deploy.Approvals, approval)

	hash, err := t.casper.PutDeploy(*deploy)
	if err != nil {
		return "", err
	}

	return hash, nil
}
