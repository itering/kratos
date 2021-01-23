package grpc

import (
	"context"

	"google.golang.org/grpc"
)

// ClientOption is gRPC client option.
type ClientOption func(o *Client)

// DecodeErrorFunc is encode error func.
type DecodeErrorFunc func(ctx context.Context, err error) error

// ErrorDecoder with client error decoder.
func ErrorDecoder(d DecodeErrorFunc) ClientOption {
	return func(o *Client) {
		o.errorDecoder = d
	}
}

// Client is grpc transport client.
type Client struct {
	errorDecoder DecodeErrorFunc
}

// NewClient new a grpc transport client.
func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		errorDecoder: DefaultErrorDecoder,
	}
	for _, o := range opts {
		o(client)
	}
	return client
}

// Interceptor returns a unary server interceptor.
func (c *Client) Interceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
			return c.errorDecoder(ctx, err)
		}
		return nil
	}
}
