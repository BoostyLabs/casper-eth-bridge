package transfers

import (
	"context"

	"github.com/zeebo/errs"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
)

// Error that error was from transfers service.
var Error = errs.Class("transfers service")

// Service contains transfers specific business rules.
//
// architecture: Service
type Service struct {
	bridge Bridge
}

// NewService is a constructor for transfers service.
func NewService(bridge Bridge) *Service {
	return &Service{
		bridge: bridge,
	}
}

// Info returns list of transfers of triggering transaction.
func (service *Service) Info(ctx context.Context, triggeringTransactionHash string) ([]Transfer, error) {
	transfers, err := service.bridge.Info(ctx, triggeringTransactionHash)
	return transfers, Error.Wrap(err)
}

// Estimate returns approximate information about transfer fee and time.
func (service *Service) Estimate(ctx context.Context, sender, recipient networks.Type, tokenID uint32, amount string) (Estimate, error) {
	estimate, err := service.bridge.Estimate(ctx, sender, recipient, tokenID, amount)
	return estimate, Error.Wrap(err)
}

// Cancel cancels a transfer in the CONFIRMING status, returning the funds to the sender after deducting the commission for issuing the transaction.
func (service *Service) Cancel(ctx context.Context, id ID, signature, pubKey []byte) error {
	return Error.Wrap(service.bridge.Cancel(ctx, id, signature, pubKey))
}

// History returns paginated list of transfers.
func (service *Service) History(ctx context.Context, offset, limit uint64, signature, pubKey []byte, networkID uint32) (Page, error) {
	history, err := service.bridge.History(ctx, offset, limit, signature, pubKey, networkID)
	return history, Error.Wrap(err)
}

// BridgeInSignature returns signature for user to send bridgeIn transaction.
func (service *Service) BridgeInSignature(ctx context.Context, req BridgeInSignatureRequest) (BridgeInSignatureResponse, error) {
	signature, err := service.bridge.BridgeInSignature(ctx, req)
	return signature, Error.Wrap(err)
}
