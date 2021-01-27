package logging

import (
	"context"
	"path"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	transgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
)

// GRPCServer is a gRPC logging middleware.
func GRPCServer(logger log.Logger) middleware.Middleware {
	infoLog := log.Info(logger)
	errLog := log.Error(logger)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				service string
				method  string
			)
			g, ok := transgrpc.FromContext(ctx)
			if ok {
				service = path.Dir(g.FullMethod)[1:]
				method = path.Base(g.FullMethod)
			}
			reply, err := handler(ctx, req)
			if err != nil {
				errLog.Print(
					"system", "grpc",
					"kind", "server",
					"grpc.service", service,
					"grpc.method", method,
					"grpc.code", errors.Code(err),
					"grpc.error", err.Error(),
				)
				return nil, err
			}
			infoLog.Print(
				"system", "grpc",
				"kind", "server",
				"grpc.service", service,
				"grpc.method", method,
				"grpc.code", 0,
			)
			return reply, nil
		}
	}
}

// HTTPServer is a gRPC logging middleware.
func HTTPServer(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		}
	}
}
