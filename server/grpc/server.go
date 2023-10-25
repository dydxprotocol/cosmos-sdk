package grpc

import (
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/grpc/gogoreflection"
	reflection "github.com/cosmos/cosmos-sdk/server/grpc/reflection/v2alpha1"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection
)

// StartGRPCServer starts a gRPC server on the given address.
func StartGRPCServer(clientCtx client.Context, app types.Application, cfg config.GRPCConfig) (*grpc.Server, net.Addr, error) {
	maxSendMsgSize := cfg.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = config.DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := cfg.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = config.DefaultGRPCMaxRecvMsgSize
	}

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec()),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
	)

	app.RegisterGRPCServer(grpcSrv)

	// Reflection allows consumers to build dynamic clients that can write to any
	// Cosmos SDK application without relying on application packages at compile
	// time.
	err := reflection.Register(grpcSrv, reflection.Config{
		SigningModes: func() map[string]int32 {
			modes := make(map[string]int32, len(clientCtx.TxConfig.SignModeHandler().Modes()))
			for _, m := range clientCtx.TxConfig.SignModeHandler().Modes() {
				modes[m.String()] = (int32)(m)
			}
			return modes
		}(),
		ChainID:           clientCtx.ChainID,
		SdkConfig:         sdk.GetConfig(),
		InterfaceRegistry: clientCtx.InterfaceRegistry,
	})
	if err != nil {
		return nil, nil, err
	}

	// Reflection allows external clients to see what services and methods
	// the gRPC server exposes.
	gogoreflection.Register(grpcSrv)

	var proto, addr string
	parts := strings.SplitN(cfg.Address, "://", 2)
	// Default to using 'tcp' to maintain backwards compatibility with configurations that don't specify
	// the network to use.
	if len(parts) != 2 {
		proto = "tcp"
		addr = cfg.Address
	} else {
		proto, addr = parts[0], parts[1]
	}
	listener, err := net.Listen(proto, addr)
	if err != nil {
		return nil, nil, err
	}

	errCh := make(chan error)
	go func() {
		err = grpcSrv.Serve(listener)
		if err != nil {
			errCh <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return nil, nil, err

	case <-time.After(time.Duration(types.ServerStartTime.Load())):
		// assume server started successfully
		return grpcSrv, listener.Addr(), nil
	}
}
