package gorpc

import (
	"fmt"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// CreateGRPCServer creates a gRPC server and registers the given service handler.
//
// Usage:
//
//	startServer, endServer, err := gorpc.CreateGRpcServer[users.UserServiceServer](users.RegisterUserServiceServer, &service, []ct.CtxKey{ct.Key1, ct.Key2})
//
// Type Parameter T:
//   - T: Pass it the service interface that the handler implements.
//
// Parameters:
//   - register: gRPC-generated registration function for the service.
//   - handler: Implementation of the service interface.
//   - contextKeys: The keys of context values that will propagate from the client to this server, they will be selected from metadata and put into the incoming context.
func CreateGRpcServer[T any](register func(grpc.ServiceRegistrar, T), handler T, port string, contextKeys contextKeys) (func() error, func(), error) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to listen on %s: %v", port, err)
	}

	customUnaryInterceptor, err := UnaryServerInterceptorWithContextKeys(contextKeys)
	if err != nil {
		return nil, nil, err
	}
	customStreamInterceptor, err := StreamServerInterceptorWithContextKeys(contextKeys)
	if err != nil {
		return nil, nil, err
	}
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()), //TODO check that this works? and opts
		grpc.UnaryInterceptor(customUnaryInterceptor),
		grpc.StreamInterceptor(customStreamInterceptor),
	)

	register(grpcServer, handler)

	startServer := func() error { return grpcServer.Serve(listener) }
	stopServer := func() { grpcServer.GracefulStop() }
	return startServer, stopServer, nil
}
