package networks

import (
	"context"
)

// Service contains networks specific business rules.
//
// architecture: Service
type Service struct {
	bridge Bridge
}

// NewService is a constructor for networks service.
func NewService(bridge Bridge) *Service {
	return &Service{
		bridge: bridge,
	}
}

// Connected returns list of supported networks.
func (service *Service) Connected(ctx context.Context) ([]Network, error) {
	networksList, err := service.bridge.ConnectedNetworks(ctx)
	return networksList, err
}

// SupportedTokens returns list of network supported tokens.
func (service *Service) SupportedTokens(ctx context.Context, networkID uint32) ([]Token, error) {
	tokensList, err := service.bridge.SupportedTokens(ctx, networkID)
	return tokensList, err
}
