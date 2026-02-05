package gorpc

import (
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

// GetGRPCClient creates a gRPC client of type T with interceptors for the specified context keys.
//
// Usage:
//
//	usersServiceClient, err = gorpc.GetGRpcClient(users.NewUserServiceClient, "users:50051", []ct.CtxKey{ct.Key1, ct.Key2})
//
// Parameters:
//   - constructor: gRPC-generated constructor function for the client service.
//   - fullAddress: Full address of the service (host:port).
//   - contextKeys: Keys to propagate from outgoing context into the gRPC metadata.
func GetGRpcClient[T any](constructor func(grpc.ClientConnInterface) T, fullAddress string, contextKeys contextKeys) (T, error) {
	customUnaryInterceptor, err := UnaryClientInterceptorWithContextKeys(contextKeys)
	if err != nil {
		return *new(T), err
	}
	customStreamInterceptor, err := StreamClientInterceptorWithContextKeys(contextKeys)
	if err != nil {
		return *new(T), err
	}
	dialOpts := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]	}`),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 2 * time.Second,
			Backoff: backoff.Config{
				BaseDelay:  1 * time.Second,
				Multiplier: 1.2,
				Jitter:     0.5,
				MaxDelay:   5 * time.Second,
			},
		}),
		grpc.WithUnaryInterceptor(customUnaryInterceptor),
		grpc.WithStreamInterceptor(customStreamInterceptor),
	}

	conn, err := grpc.NewClient(fullAddress, dialOpts...)
	if err != nil {
		return *new(T), fmt.Errorf("failed to dial user service: %v", err)
	}

	return constructor(conn), nil
}
