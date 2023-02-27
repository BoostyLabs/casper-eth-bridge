package bridge

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeebo/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"tricorn/bridge/networks"
	"tricorn/bridge/transactions"
	"tricorn/bridge/transfers"
	"tricorn/chains"
	"tricorn/internal/logger"
	"tricorn/pkg/signature"
	"tricorn/signer"
)

// Error is bridge default error type.
var Error = errs.Class("bridge service")

// Service is handling bridge related logic.
//
// architecture: Service
type Service struct {
	log logger.Logger

	signer         Signer
	nonces         networks.Nonces
	networkTokens  networks.NetworkTokens
	networkBlocks  networks.NetworkBlocks
	transactions   transactions.DB
	tokenTransfers transfers.TokenTransfers
	tokens         Tokens

	mutex      sync.Mutex
	connectors map[networks.Name]Connector
}

// New is Service constructor.
func New(log logger.Logger, signer Signer, nonces networks.Nonces, networkTokens networks.NetworkTokens,
	tokens Tokens, transactions transactions.DB, tokenTransfers transfers.TokenTransfers, networkBlocks networks.NetworkBlocks) *Service {
	return &Service{
		log:            log,
		signer:         signer,
		nonces:         nonces,
		tokenTransfers: tokenTransfers,
		networkBlocks:  networkBlocks,
		networkTokens:  networkTokens,
		transactions:   transactions,
		tokens:         tokens,
		connectors:     make(map[networks.Name]Connector),
	}
}

// ListConnectedNetworks returns list of connected networks.
func (service *Service) ListConnectedNetworks(ctx context.Context) ([]networks.Network, error) {
	connectedNetworks := make([]networks.Network, 0)
	connectors := service.GetConnectors()

	for _, connector := range connectors {
		network, err := connector.Network(ctx)
		if err != nil {
			return connectedNetworks, Error.Wrap(err)
		}

		connectedNetworks = append(connectedNetworks, network)
	}

	return connectedNetworks, nil
}

// ListSupportedTokens returns list of tokens, supported by network.
func (service *Service) ListSupportedTokens(ctx context.Context, id uint32) ([]networks.SupportedToken, error) {
	supportedTokens := make([]networks.SupportedToken, 0)

	_, networkID, err := service.parseNetworkDataFromIDAndValidate(id)
	if err != nil {
		return supportedTokens, Error.Wrap(err)
	}

	tokens, err := service.tokens.List(ctx, networkID)
	if err != nil {
		return supportedTokens, Error.Wrap(err)
	}

	for _, token := range tokens {
		supportedTokenNetworks, err := service.networkTokens.List(ctx, token.ID)
		if err != nil {
			return supportedTokens, Error.Wrap(err)
		}

		supportedTokens = append(supportedTokens, networks.SupportedToken{
			ID:        token.ID,
			ShortName: token.ShortName,
			LongName:  token.LongName,
			Addresses: supportedTokenNetworks,
		})
	}

	return supportedTokens, nil
}

// parseNetworkDataFromIDAndValidate returns network data and validateNetworkName for networks and connected connectors.
func (service *Service) parseNetworkDataFromIDAndValidate(id uint32) (networkName networks.Name, networkID networks.ID, err error) {
	networkID = networks.ID(id)
	networkName, ok := networks.IDToNetworkName[networkID]
	if !ok {
		return networkName, networkID, ErrNotConnectedNetwork
	}

	return networkName, networkID, service.validateNetworkName(networkName)
}

// validateNetworkName returns error if network name is incorrect or appropriate connector is not connected.
func (service *Service) validateNetworkName(networkName networks.Name) error {
	if err := networkName.Validate(); err != nil {
		return err
	}

	_, ok := service.connectors[networkName]
	if !ok {
		return ErrNotConnectedNetwork
	}

	return nil
}

// TransfersInfo returns list of transfers of triggering transaction on selected network.
func (service *Service) TransfersInfo(ctx context.Context, networkName string, txHash string) ([]transfers.Transfer, error) {
	transfersList := make([]transfers.Transfer, 0)

	internalNetworkName, err := service.parseNetworkNameAndValidate(networkName)
	if err != nil {
		return transfersList, Error.Wrap(err)
	}

	networkID := networks.NetworkNameToID[internalNetworkName]
	transactionHash, err := networks.StringToBytes(networkID, txHash)
	if err != nil {
		return transfersList, Error.Wrap(err)
	}

	tokenTransfer, err := service.tokenTransfers.GetByNetworkAndTx(ctx, networkID, transactionHash)
	if err != nil {
		return transfersList, Error.Wrap(err)
	}

	transfersList, err = service.parseTransfers(ctx, []transfers.TokenTransfer{tokenTransfer})
	if err != nil {
		return transfersList, Error.Wrap(err)
	}

	return transfersList, nil
}

// parseNetworkNameAndValidate parses network name and id from string and validateNetworkName for networks and connected connectors.
func (service *Service) parseNetworkNameAndValidate(stringNetworkName string) (networkName networks.Name, err error) {
	networkName = networks.Name(stringNetworkName)

	return networkName, service.validateNetworkName(networks.Name(stringNetworkName))
}

// parseTransfers parses slice of transfers.Transfer from slice of transfers.TokenTransfer.
func (service *Service) parseTransfers(ctx context.Context, tokenTransfers []transfers.TokenTransfer) ([]transfers.Transfer, error) {
	transfersList := make([]transfers.Transfer, 0, len(tokenTransfers))
	for _, tokenTransfer := range tokenTransfers {
		triggeringTransaction, err := service.transactions.Get(ctx, tokenTransfer.TriggeringTx)
		if err != nil {
			return transfersList, err
		}

		var outboundTx transfers.StringTxHash
		if tokenTransfer.Status != transfers.StatusWaiting && tokenTransfer.Status != transfers.StatusConfirming {
			outboundTransaction, err := service.transactions.Get(ctx, tokenTransfer.OutboundTx)
			if err != nil {
				return transfersList, err
			}

			outboundTx = parseStringTxHash(outboundTransaction.NetworkID, outboundTransaction.TxHash)
		}

		transfer := transfers.Transfer{
			ID:           transfers.ID(tokenTransfer.ID),
			Amount:       tokenTransfer.Amount,
			Sender:       parseNetworkAddress(tokenTransfer.SenderNetworkID, tokenTransfer.SenderAddress),
			Recipient:    parseNetworkAddress(tokenTransfer.RecipientNetworkID, tokenTransfer.RecipientAddress),
			Status:       tokenTransfer.Status,
			TriggeringTx: parseStringTxHash(triggeringTransaction.NetworkID, triggeringTransaction.TxHash),
			OutboundTx:   outboundTx,
			CreatedAt:    triggeringTransaction.SeenAt,
		}

		transfersList = append(transfersList, transfer)
	}

	return transfersList, nil
}

// parseStringTxHash returns transfers.StringTxHash by network id and hash.
func parseStringTxHash(networkID networks.ID, hash []byte) transfers.StringTxHash {
	stringTxHash := transfers.StringTxHash{
		NetworkName: networks.IDToNetworkName[networkID].String(),
	}
	stringTxHash.Hash.SetBytes(hash)

	return stringTxHash
}

// parseNetworkAddress returns networks.Address by network id and address.
func parseNetworkAddress(id int64, address []byte) networks.Address {
	networkID := networks.ID(id)

	return networks.Address{
		NetworkName: networks.IDToNetworkName[networkID].String(),
		Address:     networks.BytesToString(networkID, address),
	}
}

// History returns paginated transfer history for user.
func (service *Service) History(ctx context.Context, offset, limit uint64, sig, pubKey []byte, networkID uint32) (transfers.Page, error) {
	page := transfers.Page{
		Transfers:  make([]transfers.Transfer, 0),
		Offset:     int64(offset),
		Limit:      int64(limit),
		TotalCount: 0,
	}

	_, userNetworkID, err := service.parseNetworkDataFromIDAndValidate(networkID)
	if err != nil {
		return page, Error.Wrap(err)
	}

	address, err := decodeSignature(userNetworkID, pubKey, sig)
	if err != nil {
		return page, Error.Wrap(err)
	}

	tokenTransfers, err := service.tokenTransfers.ListByUser(ctx, offset, limit, address, userNetworkID)
	if err != nil {
		return page, Error.Wrap(err)
	}

	transferList, err := service.parseTransfers(ctx, tokenTransfers)
	if err != nil {
		return page, Error.Wrap(err)
	}

	totalCount, err := service.tokenTransfers.CountByUser(ctx, userNetworkID, address)
	if err != nil {
		return page, Error.Wrap(err)
	}

	page.Transfers = transferList
	page.TotalCount = int64(totalCount)

	return page, nil
}

// decodeSignature decodes signature and returns public key/account hash of sender.
func decodeSignature(networkID networks.ID, publicKey, sig []byte) ([]byte, error) {
	switch networkID {
	case networks.IDEth, networks.IDPolygon, networks.IDGoerli, networks.IDMumbai, networks.IDBNB, networks.IDBNBTest, networks.IDAvalanche, networks.IDAvalancheTest:
		pubKey, err := signature.RecoverEVMPublicKeyFrom(sig, authenticationMsg)
		if err != nil {
			return nil, err
		}

		address, err := signature.EVMPublicKeySecp256k1ToAddress(pubKey)
		if err != nil {
			return nil, err
		}

		return networks.StringToBytes(networkID, address.String())
	case networks.IDCasper, networks.IDCasperTest:
		return publicKey, nil
	}
	return nil, nil
}

// EstimateTransfer estimates a potential transfer.
func (service *Service) EstimateTransfer(ctx context.Context, transfer transfers.EstimateTransfer) (chains.Estimation, error) {
	_, err := service.parseNetworkNameAndValidate(transfer.SenderNetwork)
	if err != nil {
		return chains.Estimation{}, Error.Wrap(err)
	}

	recipientNetworkName, err := service.parseNetworkNameAndValidate(transfer.RecipientNetwork)
	if err != nil {
		return chains.Estimation{}, Error.Wrap(err)
	}

	amount, ok := new(big.Int).SetString(transfer.Amount, 10)
	if !ok || amount.Int64() < 0 {
		return chains.Estimation{}, Error.Wrap(ErrInvalidAmount)
	}

	estimation, err := service.connectors[recipientNetworkName].EstimateTransfer(ctx, transfer)
	if err != nil {
		return chains.Estimation{}, Error.Wrap(err)
	}

	return estimation, nil
}

// GetBridgeInSignature returns signature for user to send bridgeIn transaction.
func (service *Service) GetBridgeInSignature(ctx context.Context, request transfers.BridgeInSignatureRequest) (BridgeInSignatureResponse, error) {
	senderNetworkName, err := service.parseNetworkNameAndValidate(request.Sender.NetworkName)
	if err != nil {
		return BridgeInSignatureResponse{}, Error.Wrap(err)
	}

	// TODO: uncomment after fix.
	// recipientNetworkName, err := service.parseNetworkNameAndValidate(request.Destination.NetworkName)
	// if err != nil {
	//	return BridgeInSignatureResponse{}, Error.Wrap(err)
	// }.

	amount, ok := new(big.Int).SetString(request.Amount, 10)
	if !ok || amount.Int64() < 0 {
		return BridgeInSignatureResponse{}, Error.Wrap(ErrInvalidAmount)
	}

	senderNetworkID := networks.NetworkNameToID[senderNetworkName]

	nonce, err := service.nonces.Get(ctx, senderNetworkID)
	if err != nil {
		return BridgeInSignatureResponse{}, Error.Wrap(err)
	}

	token, err := service.networkTokens.Get(ctx, senderNetworkID, int64(request.TokenID))
	if err != nil {
		return BridgeInSignatureResponse{}, Error.Wrap(err)
	}

	senderAddress, err := networks.StringToBytes(senderNetworkID, request.Sender.Address)
	if err != nil {
		return BridgeInSignatureResponse{}, Error.Wrap(err)
	}

	// TODO: uncomment after fix.
	// destinationEstimation, err := service.connectors[recipientNetworkName].EstimateTransfer(ctx, transfers.EstimateTransfer{
	//	SenderNetwork:    request.Sender.NetworkName,
	//	RecipientNetwork: request.Destination.NetworkName,
	//	TokenID:          request.TokenID,
	//	Amount:           request.Amount,
	// })
	// if err != nil {
	//	return BridgeInSignatureResponse{}, Error.Wrap(err)
	// }
	//
	// gasCommission, ok := new(big.Int).SetString(destinationEstimation.Fee, 10)
	// if !ok {
	//	return BridgeInSignatureResponse{}, Error.New("couldn't parse gas commission")
	// }.

	gasCommission := new(big.Int).SetInt64(0)

	bridgeInSignature, err := service.connectors[senderNetworkName].BridgeInSignature(ctx, BridgeInSignatureRequest{
		User:          senderAddress,
		Nonce:         big.NewInt(nonce),
		Token:         token.ContractAddress,
		Amount:        amount,
		Destination:   request.Destination,
		GasCommission: gasCommission,
	})
	if err != nil {
		return BridgeInSignatureResponse{}, Error.Wrap(err)
	}

	senderName := networks.Name(request.Sender.NetworkName)
	senderID, ok := networks.NetworkNameToID[senderName]
	if !ok {
		return BridgeInSignatureResponse{}, Error.Wrap(networks.ErrTransactionNameInvalid)
	}

	return bridgeInSignature, service.nonces.Increment(ctx, senderID)
}

// CancelTransfer cancels a pending transfer.
func (service *Service) CancelTransfer(ctx context.Context, transfer transfers.CancelSignatureRequest) (transfers.CancelSignatureResponse, error) {
	networkName, networkID, err := service.parseNetworkDataFromIDAndValidate(transfer.NetworkID)
	if err != nil {
		return transfers.CancelSignatureResponse{}, Error.Wrap(err)
	}

	tokenTransfer, err := service.tokenTransfers.GetByNetworkAndTx(ctx, networkID, transfer.Signature)
	if err != nil {
		return transfers.CancelSignatureResponse{}, Error.Wrap(err)
	}

	if tokenTransfer.Status != transfers.StatusWaiting {
		return transfers.CancelSignatureResponse{}, Error.Wrap(ErrInvalidTransferStatus)
	}

	nonce, err := service.nonces.Get(ctx, networkID)
	if err != nil {
		return transfers.CancelSignatureResponse{}, Error.Wrap(err)
	}

	token, err := service.networkTokens.Get(ctx, networkID, tokenTransfer.TokenID)
	if err != nil {
		return transfers.CancelSignatureResponse{}, Error.Wrap(err)
	}

	estimation, err := service.connectors[networkName].EstimateTransfer(ctx, transfers.EstimateTransfer{
		SenderNetwork:    networkName.String(),
		RecipientNetwork: networkName.String(),
		TokenID:          uint32(tokenTransfer.TokenID),
		Amount:           tokenTransfer.Amount.String(),
	})
	if err != nil {
		return transfers.CancelSignatureResponse{}, Error.Wrap(err)
	}

	commission, ok := new(big.Int).SetString(estimation.Fee, 10)
	if !ok {
		return transfers.CancelSignatureResponse{}, Error.New("couldn't parse commission")
	}

	// TODO: Unify type for addresses, mb use byte format.
	cancelSignatureRequest := chains.CancelSignatureRequest{
		Nonce:      new(big.Int).SetInt64(nonce),
		Token:      common.BytesToAddress(token.ContractAddress),
		Recipient:  common.BytesToAddress(transfer.PublicKey),
		Commission: commission,
		Amount:     &tokenTransfer.Amount,
	}

	cancelSignatureResponse, err := service.connectors[networkName].CancelSignature(ctx, cancelSignatureRequest)
	if err != nil {
		return transfers.CancelSignatureResponse{}, Error.Wrap(err)
	}

	cancelTransferResponse := transfers.CancelSignatureResponse{
		Status:     string(tokenTransfer.Status),
		Nonce:      uint64(nonce),
		Signature:  cancelSignatureResponse.Signature,
		Token:      token.ContractAddress,
		Recipient:  tokenTransfer.SenderAddress,
		Commission: commission.String(),
		Amount:     tokenTransfer.Amount.String(),
	}

	return cancelTransferResponse, service.nonces.Increment(ctx, networkID)
}

// Sign signs data for specific network.
func (service *Service) Sign(ctx context.Context, networkType networks.Type, data []byte, dataType signer.Type) ([]byte, error) {
	signedData, err := service.signer.Sign(ctx, networkType, data, dataType)
	return signedData, Error.Wrap(err)
}

// PublicKey returns public key for specific network.
func (service *Service) PublicKey(ctx context.Context, networkType networks.Type) ([]byte, error) {
	publicKey, err := service.signer.PublicKey(ctx, networkType)
	return publicKey, Error.Wrap(err)
}

// separateEvent separates events for different processing and recording in the database.
func (service *Service) separateEvent(ctx context.Context, eventFund chains.EventVariant, networkName networks.Name) error {
	switch eventFund.Type {
	case chains.EventTypeIn:
		err := service.eventInReaction(ctx, eventFund, networkName)
		if err != nil {
			service.log.Error("eventIn reaction err: ", Error.Wrap(err))
			return status.Error(codes.Internal, Error.Wrap(err).Error())
		}
	case chains.EventTypeOut:
		err := service.eventOutReaction(ctx, eventFund, networkName)
		if err != nil {
			service.log.Error("eventOut reaction err: ", Error.Wrap(err))
			return status.Error(codes.Internal, Error.Wrap(err).Error())
		}
	default:
		err := Error.New("invalid event type")
		service.log.Error("", Error.Wrap(err))
		return status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	return nil
}

// eventInReaction performs actions after fundIn event.
func (service *Service) eventInReaction(ctx context.Context, eventFund chains.EventVariant, networkName networks.Name) error {
	recipientNetworkID := networks.NetworkNameToID[networks.Name(eventFund.EventFundsIn.To.NetworkName)]
	recipientAddress, err := networks.StringToBytes(recipientNetworkID, eventFund.EventFundsIn.To.Address)
	if err != nil {
		service.log.Error("", Error.Wrap(err))
		return status.Error(codes.Internal, err.Error())
	}

	amount, ok := new(big.Int).SetString(eventFund.EventFundsIn.Amount, 10)
	if !ok {
		service.log.Error("", Error.New("could not set amount %v", amount))
		return status.Error(codes.Internal, "could not set amount")
	}

	networkID := networks.NetworkNameToID[networkName]

	err = service.transactions.Exists(ctx, networkID, eventFund.EventFundsOut.Tx.Hash)
	if errors.Is(err, ErrTransactionAlreadyExists) {
		return nil
	}

	transactionID, err := service.transactions.Create(ctx, transactions.Transaction{
		NetworkID:   networkID,
		TxHash:      eventFund.EventFundsIn.Tx.Hash,
		Sender:      eventFund.EventFundsIn.From,
		BlockNumber: int64(eventFund.EventFundsIn.Tx.BlockNumber),
		SeenAt:      time.Now().UTC(),
	})
	if err != nil {
		service.log.Error("", Error.Wrap(err))
		return status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	tokenTransfer := transfers.TokenTransfer{
		TokenID:            1, // todo: dynamically change.
		Amount:             *amount,
		Status:             transfers.StatusConfirming,
		SenderNetworkID:    int64(networks.NetworkNameToID[networkName]),
		SenderAddress:      eventFund.EventFundsIn.From,
		RecipientNetworkID: int64(recipientNetworkID),
		RecipientAddress:   recipientAddress,
		TriggeringTx:       transactionID,
	}

	err = service.tokenTransfers.Create(ctx, tokenTransfer)
	if err != nil {
		service.log.Error("", Error.Wrap(err))
		return status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	{ // call BridgeOut.
		toAddress, err := networks.StringToBytes(recipientNetworkID, eventFund.EventFundsIn.To.Address)
		if err != nil {
			service.log.Error("", Error.Wrap(err))
			return status.Error(codes.Internal, Error.Wrap(err).Error())
		}

		token, err := service.networkTokens.Get(ctx, recipientNetworkID, 1) // TODO: add dynamic token id.
		if err != nil {
			service.log.Error("", Error.Wrap(err))
			return status.Error(codes.Internal, Error.Wrap(err).Error())
		}

		_, err = service.connectors[networks.Name(eventFund.EventFundsIn.To.NetworkName)].BridgeOut(ctx, chains.TokenOutRequest{
			Amount: amount,
			Token:  token.ContractAddress,
			To:     toAddress,
			From: networks.Address{
				NetworkName: networkName.String(),
				Address:     hex.EncodeToString(eventFund.EventFundsIn.From),
			},
			TransactionID: big.NewInt(int64(transactionID)),
		})
		if err != nil {
			service.log.Error("", Error.Wrap(err))
			return status.Error(codes.Internal, Error.Wrap(err).Error())
		}
	}

	return nil
}

// eventOutReaction performs actions after fundOut event.
func (service *Service) eventOutReaction(ctx context.Context, eventFund chains.EventVariant, networkName networks.Name) error {
	senderNetworkID := networks.NetworkNameToID[networks.Name(eventFund.EventFundsOut.From.NetworkName)]
	senderAddress, err := networks.StringToBytes(senderNetworkID, eventFund.EventFundsOut.From.Address)
	if err != nil {
		service.log.Error("", Error.Wrap(err))
		return status.Error(codes.Internal, err.Error())
	}

	networkID := networks.NetworkNameToID[networkName]
	address, err := networks.StringToBytes(networkID, eventFund.EventFundsOut.From.Address)
	if err != nil {
		service.log.Error("", Error.Wrap(err))
		return status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	err = service.transactions.Exists(ctx, networkID, eventFund.EventFundsOut.Tx.Hash)
	if errors.Is(err, ErrTransactionAlreadyExists) {
		return nil
	}

	transactionID, err := service.transactions.Create(ctx, transactions.Transaction{
		NetworkID:   networkID,
		TxHash:      eventFund.EventFundsOut.Tx.Hash,
		Sender:      address,
		BlockNumber: int64(eventFund.EventFundsOut.Tx.BlockNumber),
		SeenAt:      time.Now().UTC(),
	})
	if err != nil {
		service.log.Error("", Error.Wrap(err))
		return status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	amount, ok := new(big.Int).SetString(eventFund.EventFundsOut.Amount, 10)
	if !ok {
		service.log.Error("", Error.New("could not set amount %v", amount))
		return status.Error(codes.Internal, "could not set amount")
	}

	tokenTransfer := transfers.TokenTransfer{
		TokenID:          1, // todo: dynamically change.
		Amount:           *amount,
		SenderAddress:    senderAddress,
		RecipientAddress: eventFund.EventFundsOut.To,
	}

	tokenTransfer, err = service.tokenTransfers.GetByAllParams(ctx, tokenTransfer)
	if err != nil {
		service.log.Error("", Error.Wrap(err))
		return status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	tokenTransfer.Status = transfers.StatusFinished
	tokenTransfer.OutboundTx = transactionID
	err = service.tokenTransfers.Update(ctx, tokenTransfer)
	if err != nil {
		service.log.Error("", Error.Wrap(err))
		return status.Error(codes.Internal, Error.Wrap(err).Error())
	}

	return nil
}

// GetConnectors returns active connectors.
func (service *Service) GetConnectors() map[networks.Name]Connector {
	return service.connectors
}

// AddConnector adds connector to the active list of connectors.
func (service *Service) AddConnector(ctx context.Context, name networks.Name, connector Connector) {
	service.connectors[name] = connector
	// start event reading from connector.
	chore := NewChore(service.log, service, service.networkBlocks)
	chore.Run(ctx, name, connector)
}

// RemoveConnector removes connector to the active list of connectors.
func (service *Service) RemoveConnector(name networks.Name) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	delete(service.connectors, name)
}

// IsConnectorConnected returns bool values which defines is connector connected to bridge.
func (service *Service) IsConnectorConnected(name networks.Name) bool {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	_, exists := service.connectors[name]
	return exists
}
