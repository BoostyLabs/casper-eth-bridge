package apitesting

import (
	bridgesignerpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-signer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConnectToSigner initiates connection with signer server.
func ConnectToSigner(addr string) (bridgesignerpb.BridgeSignerClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return bridgesignerpb.NewBridgeSignerClient(conn), nil
}
