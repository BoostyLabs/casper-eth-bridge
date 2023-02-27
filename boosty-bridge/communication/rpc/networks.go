package rpc

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	bridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/gateway-bridge"
	networkspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/networks"

	"tricorn/bridge/networks"
	"tricorn/communication"
)

// ensures that networksRPC implements networks.Bridge.
var _ networks.Bridge = (*networksRPC)(nil)

// networksRPC provides access to the networks.Bridge.
type networksRPC struct {
	isConnected bool
	client      bridgepb.GatewayBridgeClient
}

// ConnectedNetworks returns all connected networks to the bridge.
func (networksRPC *networksRPC) ConnectedNetworks(ctx context.Context) ([]networks.Network, error) {
	if !networksRPC.isConnected {
		return []networks.Network{}, communication.ErrNotConnected
	}

	pbNetworksList, err := networksRPC.client.ConnectedNetworks(ctx, &emptypb.Empty{})
	if err != nil {
		return []networks.Network{}, Error.Wrap(err)
	}

	allNetworks := make([]networks.Network, 0, len(pbNetworksList.Networks))
	for _, networkpb := range pbNetworksList.GetNetworks() {
		network := networks.Network{
			ID:             networks.ID(networkpb.GetId()),
			Name:           networks.Name(networkpb.GetName()),
			Type:           networks.Type(networkpb.GetType()),
			IsTestnet:      networkpb.GetIsTestnet(),
			NodeAddress:    networkpb.GetNodeAddress(),
			TokenContract:  networkpb.GetTokenContract(),
			BridgeContract: networkpb.GetBridgeContract(),
			GasLimit:       networkpb.GetGasLimit(),
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
	if !networksRPC.isConnected {
		return []networks.Token{}, communication.ErrNotConnected
	}

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
