package status

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/golang/protobuf/proto"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandlerFunc is middleware error handler.
type HandlerFunc func(error) error

// Option is recovery option.
type Option func(*options)

type options struct {
	handler HandlerFunc
}

// Handler with status handler.
func Handler(h HandlerFunc) Option {
	return func(o *options) {
		o.handler = h
	}
}

// Server is an error middleware.
func Server(opts ...Option) middleware.Middleware {
	options := options{
		handler: func(err error) error {
			se, ok := err.(*errors.StatusError)
			if !ok {
				se = &errors.StatusError{
					Code:    2,
					Reason:  "Unknown",
					Message: "Unknown: " + err.Error(),
				}
			}
			gs := status.Newf(codes.Code(se.Code), "%s: %s", se.Reason, se.Message)
			details := []proto.Message{
				&errdetails.ErrorInfo{
					Reason:   se.Reason,
					Metadata: map[string]string{"message": se.Message},
				},
			}
			gs, err = gs.WithDetails(details...)
			if err != nil {
				return err
			}
			return gs.Err()
		},
	}
	for _, o := range opts {
		o(&options)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			reply, err := handler(ctx, req)
			if err != nil {
				return nil, options.handler(err)
			}
			return reply, nil
		}
	}
}

// Client is an error middleware.
func Client(opts ...Option) middleware.Middleware {
	options := options{
		handler: func(err error) error {
			gs := status.Convert(err)
			for _, detail := range gs.Details() {
				switch d := detail.(type) {
				case *errdetails.ErrorInfo:
					return &errors.StatusError{
						Code:    int32(gs.Code()),
						Reason:  d.Reason,
						Message: d.Metadata["message"],
					}
				}
			}
			return &errors.StatusError{Code: int32(gs.Code())}
		},
	}
	for _, o := range opts {
		o(&options)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			reply, err := handler(ctx, req)
			if err != nil {
				return nil, options.handler(err)
			}
			return reply, nil
		}
	}
}
