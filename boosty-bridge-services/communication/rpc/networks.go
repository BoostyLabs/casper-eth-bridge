package rpc

import (
	"context"

	bridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/gateway-bridge"
	networkspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/networks"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
)

// ensures that networksRPC implements networks.Bridge.
var _ networks.Bridge = (*networksRPC)(nil)

// networksRPC provides access to the networks.Bridge.
type networksRPC struct {
	client bridgepb.GatewayBridgeClient
}

// ConnectedNetworks returns all connected networks to the bridge.
func (networksRPC *networksRPC) ConnectedNetworks(ctx context.Context) ([]networks.Network, error) {
	pbNetworksList, err := networksRPC.client.ConnectedNetworks(ctx, &emptypb.Empty{})
	if err != nil {
		return []networks.Network{}, Error.Wrap(err)
	}

	allNetworks := make([]networks.Network, 0, len(pbNetworksList.Networks))
	for _, networkpb := range pbNetworksList.GetNetworks() {
		network := networks.Network{
			ID:        networkpb.GetId(),
			Name:      networkpb.GetName(),
			IsTestnet: networkpb.GetIsTestnet(),
		}

		switch networkpb.GetType() {
		case networkspb.NetworkType_NT_CASPER:
			network.Type = networks.TypeCasper
		case networkspb.NetworkType_NT_EVM:
			network.Type = networks.TypeEVM
		}

		allNetworks = append(allNetworks, network)
	}

	return allNetworks, nil
}

// SupportedTokens returns all supported tokens by certain network.
func (networksRPC *networksRPC) SupportedTokens(ctx context.Context, networkID uint32) ([]networks.Token, error) {
	pbTokensList, err := networksRPC.client.SupportedTokens(ctx, &networkspb.SupportedTokensRequest{
		NetworkId: networkID,
	})
	if err != nil {
		return []networks.Token{}, Error.Wrap(err)
	}

	tokens := make([]networks.Token, 0, len(pbTokensList.GetTokens()))
	for _, tokenpb := range pbTokensList.GetTokens() {
		token := networks.Token{
			ID:        tokenpb.GetId(),
			ShortName: tokenpb.GetShortName(),
			LongName:  tokenpb.GetLongName(),
		}

		wraps := make([]networks.WrappedIn, 0, len(tokenpb.GetAddresses()))
		for _, addresspb := range tokenpb.GetAddresses() {
			wrapper := networks.WrappedIn{
				NetworkID:            addresspb.GetNetworkId(),
				SmartContractAddress: addresspb.GetAddress(),
			}

			wraps = append(wraps, wrapper)
		}
		token.Wraps = wraps

		tokens = append(tokens, token)
	}

	return tokens, nil
}
