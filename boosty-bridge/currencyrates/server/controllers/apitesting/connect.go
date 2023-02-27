package apitesting

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	bridgeoraclepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-oracle"
)

// ConnectToOracle initiates connection with oracle server.
func ConnectToOracle(addr string) (bridgeoraclepb.BridgeOracleClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return bridgeoraclepb.NewBridgeOracleClient(conn), nil
}
