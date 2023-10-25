package grpc

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/types"
)

// StartGRPCWeb starts a gRPC-Web server on the given address.
func StartGRPCWeb(grpcSrv *grpc.Server, config config.Config) (*http.Server, error) {
	var options []grpcweb.Option
	if config.GRPCWeb.EnableUnsafeCORS {
		options = append(options,
			grpcweb.WithOriginFunc(func(origin string) bool {
				return true
			}),
		)
	}

	var proto, addr string
	parts := strings.SplitN(config.GRPCWeb.Address, "://", 2)
	// Default to using 'tcp' to maintain backwards compatibility with configurations that don't specify
	// the network to use.
	if len(parts) != 2 {
		proto = "tcp"
		addr = config.GRPCWeb.Address
	} else {
		proto, addr = parts[0], parts[1]
	}
	listener, err := net.Listen(proto, addr)
	if err != nil {
		return nil, err
	}

	wrappedServer := grpcweb.WrapServer(grpcSrv, options...)
	grpcWebSrv := &http.Server{
		Handler:           wrappedServer,
		ReadHeaderTimeout: 500 * time.Millisecond,
	}

	errCh := make(chan error)
	go func() {
		if err := grpcWebSrv.Serve(listener); err != nil {
			errCh <- fmt.Errorf("[grpc] failed to serve: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return nil, err
	case <-time.After(time.Duration(types.ServerStartTime.Load())): // assume server started successfully
		return grpcWebSrv, nil
	}
}
