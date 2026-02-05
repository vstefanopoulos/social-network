package gorpc

import (
	"context"
	"errors"
	tele "social-network/shared/go/telemetry"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//TODO convert string keys back to context interface keys

// this interface exists to enforce usage of typed context keys instead of just strings
type contextKeys interface {
	GetKeys() []string
}

var ErrBadContextValues = errors.New("bad context keys passed to interceptor creator, at least a key doesn't follow the validation requirements")

// this package holds gRPC interceptors for both client and server
// they can add specified values from context to metadata and vice versa, so that we can seamlessly propagate context values from one service to another
// the propagation only happens from client to server, cause context can only go one direction.
// the propagation works by taking the context values and putting them metadata, and then from metadata back to the context
// this package can also support adding logic that runs before and after each request
// there are two types of interceptors, for streams and for unary requests

/*


===============================================================================================
=====================================  SERVER =================================================
===============================================================================================


*/

// UnaryServerInterceptorWithContextKeys returns a server interceptor that adds specified metadata values to context.
//
// IMPORTANT: Only "a-z", "0-9", and "-_." characters allowed for keys
func UnaryServerInterceptorWithContextKeys(contextKeys contextKeys) (grpc.UnaryServerInterceptor, error) {
	if !validateContextKeys(contextKeys.GetKeys()...) {
		return nil, ErrBadContextValues
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		ctx = addMetadataToContext(ctx, md, contextKeys.GetKeys()...)
		tele.Debug(ctx, "unary server grpc interceptor intercepting @1", "method", info.FullMethod, "request", req)
		m, err := handler(ctx, req)
		return m, err
	}, nil
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func newWrappedServerStream(ctx context.Context, s grpc.ServerStream) grpc.ServerStream {
	return &wrappedServerStream{
		ServerStream: s,
		ctx:          ctx,
	}
}

func (w *wrappedServerStream) RecvMsg(m any) error {
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedServerStream) SendMsg(m any) error {
	return w.ServerStream.SendMsg(m)
}

// TODO UNTESTED
// StreamServerInterceptorWithContextKeys returns a server interceptor that adds specified metadata values to context.
//
// IMPORTANT: Only "a-z", "0-9", and "-_." characters allowed for keys
func StreamServerInterceptorWithContextKeys(contextKeys contextKeys) (grpc.StreamServerInterceptor, error) {
	if !validateContextKeys(contextKeys.GetKeys()...) {
		return nil, ErrBadContextValues
	}

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, _ := metadata.FromIncomingContext(ss.Context())
		ctx := addMetadataToContext(ss.Context(), md, contextKeys.GetKeys()...)
		tele.Debug(ctx, "stream server grpc interceptor intercepting @1", "method", info.FullMethod)
		wrapped := newWrappedServerStream(ctx, ss)
		return handler(srv, wrapped)
	}, nil
}

/*





===============================================================================================
=====================================  CLIENT =================================================
===============================================================================================

*/

// UnaryClientInterceptorWithContextKeys returns a client interceptor that adds specified context values to outgoing metadata.
//
// IMPORTANT: Only "a-z", "0-9", and "-_." characters allowed for keys
func UnaryClientInterceptorWithContextKeys(contextKeys contextKeys) (grpc.UnaryClientInterceptor, error) {
	if !validateContextKeys(contextKeys.GetKeys()...) {
		return nil, ErrBadContextValues
	}

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		pairs := createPairs(ctx, contextKeys.GetKeys()...)
		tele.Debug(ctx, "unary client grpc interceptor intercepting @1 @2", "method", method, "target", cc.Target(), "request", req, "reply", reply)
		ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
		err := invoker(ctx, method, req, reply, cc, opts...)
		return err
	}, nil
}

type wrappedClientStream struct {
	grpc.ClientStream
}

func newWrappedClientStream(s grpc.ClientStream) grpc.ClientStream {
	return &wrappedClientStream{s}
}

func (w *wrappedClientStream) RecvMsg(m any) error {
	return w.ClientStream.RecvMsg(m)
}

func (w *wrappedClientStream) SendMsg(m any) error {
	return w.ClientStream.SendMsg(m)
}

// StreamClientInterceptorWithContextKeys returns a client interceptor that adds specified context values to outgoing metadata.
//
// IMPORTANT: Only "a-z", "0-9", and "-_." characters allowed for keys
func StreamClientInterceptorWithContextKeys(contextKeys contextKeys) (grpc.StreamClientInterceptor, error) {

	if !validateContextKeys(contextKeys.GetKeys()...) {
		return nil, ErrBadContextValues
	}

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// creating pairs of key values, ex. ["key1", "val1", "key2", "val2"]
		pairs := createPairs(ctx, contextKeys.GetKeys()...)
		ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
		tele.Debug(ctx, "stream client grpc interceptor intercepting @1 @2", "method", method, "target", cc.Target())
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		return newWrappedClientStream(clientStream), err
	}, nil
}

/*




===============================================================================================
=====================================  UTILS =================================================
===============================================================================================

*/

/* FROM GRPC DOCUMENTATION, relating to metadata keys

Only the following ASCII characters are allowed in keys:
- digits: 0-9
- uppercase letters: A-Z (normalized to lower)
- lowercase letters: a-z
- special characters: -_.

Uppercase letters are automatically converted to lowercase.

Keys beginning with "grpc-" are reserved for grpc-internal use only and may
result in errors if set in metadata.

*/

// validateContextKeys validates the keys so that nothing bad happens during context value propagation due to the above limitations
func validateContextKeys(keys ...string) bool {
	for _, key := range keys {
		for _, r := range []rune(key) {
			if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && r != '_' && r != '.' {
				return false
			}
		}

		if strings.HasPrefix(key, "grpc-") {
			return false
		}
	}

	return true
}

// addMetadataToContext adds metadata to a context
//
// this is needed because while the metadata exists in the ctx that appears on the server side, the values are inside the metadata and not the ctx in the same form that they were in the client side
func addMetadataToContext(ctx context.Context, md metadata.MD, keys ...string) context.Context {
	for _, key := range keys {
		vals := md.Get(key)
		for _, val := range vals {
			ctx = context.WithValue(ctx, key, val)
		}
	}
	return ctx
}

// createPairs creates pairs values alternating between contextKey and string, meant to be used to append metadata to existing context
// ex. ["key1", "val1", "key2", "val2"]
func createPairs(ctx context.Context, keys ...string) []string {
	pairs := make([]string, 0, len(keys)*2)
	for _, key := range keys {
		val, ok := ctx.Value(key).(string)
		if !ok {
			continue
		}
		pairs = append(pairs, key)
		pairs = append(pairs, val)
	}
	return pairs
}
