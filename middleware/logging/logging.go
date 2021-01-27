package logging

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	transgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	transhttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Server is a logging middleware.
func Server(logger log.Logger) middleware.Middleware {
	infoLog := log.Info(logger)
	errLog := log.Error(logger)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				path   string
				method string
			)
			tr, _ := transport.FromContext(ctx)
			switch tr.Kind {
			case "HTTP":
				h, ok := transhttp.FromContext(ctx)
				if ok {
					path = h.Request.URL.Path
					method = h.Request.Method
				}
			case "GRPC":
				g, ok := transgrpc.FromContext(ctx)
				if ok {
					path = g.FullMethod
				}
			}
			reply, err := handler(ctx, req)
			if err != nil {
				errLog.Print(
					"system", tr.Kind,
					"kind", "server",
					"http.path", path,
					"http.method", method,
					"code", errors.Code(err),
					"error", err.Error(),
				)
				return nil, err
			}
			infoLog.Print(
				"system", tr.Kind,
				"kind", "server",
				"grpc.service", path,
				"grpc.method", "POST",
				"code", 0,
			)
			return reply, nil
		}
	}
}
