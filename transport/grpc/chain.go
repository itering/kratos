package grpc

import (
	"context"

	"google.golang.org/grpc"
)

// ChainUnaryClient .
func ChainUnaryClient(ints ...grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		chain := func(in grpc.UnaryClientInterceptor, invoker grpc.UnaryInvoker) grpc.UnaryInvoker {
			return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				return in(ctx, method, req, reply, cc, invoker, opts...)
			}
		}
		next := invoker
		for i := len(ints) - 1; i >= 0; i-- {
			next = chain(ints[i], next)
		}
		return next(ctx, method, req, reply, cc, opts...)
	}
}

// ChainUnaryServer .
func ChainUnaryServer(ints ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		chain := func(in grpc.UnaryServerInterceptor, handler grpc.UnaryHandler) grpc.UnaryHandler {
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				return in(ctx, req, info, handler)
			}
		}
		next := handler
		for i := len(ints) - 1; i >= 0; i-- {
			next = chain(ints[i], next)
		}
		return next(ctx, req)
	}
}
