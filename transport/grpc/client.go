package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// ClientOption is gRPC client option.
type ClientOption func(o *Client)

// DecodeErrorFunc is encode error func.
type DecodeErrorFunc func(ctx context.Context, err error) error

// ClientErrorDecoder with client error decoder.
func ClientErrorDecoder(d DecodeErrorFunc) ClientOption {
	return func(c *Client) {
		c.errorDecoder = d
	}
}

// ClientUnaryInterceptor with client unary interceptor.
func ClientUnaryInterceptor(ints ...grpc.UnaryClientInterceptor) ClientOption {
	return func(c *Client) {
		c.ints = ints
	}
}

// ClientContext with client context.
func ClientContext(ctx context.Context) ClientOption {
	return func(c *Client) {
		c.ctx = ctx
	}
}

// ClientTimeout with client timeout.
func ClientTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// ClientInsecure with client insecure.
func ClientInsecure() ClientOption {
	return func(c *Client) {
		c.insecure = true
	}
}

// Client is grpc transport client.
type Client struct {
	ctx          context.Context
	insecure     bool
	timeout      time.Duration
	ints         []grpc.UnaryClientInterceptor
	errorDecoder DecodeErrorFunc
}

// NewClient new a grpc transport client.
func NewClient(target string, opts ...ClientOption) (*grpc.ClientConn, error) {
	client := &Client{
		ctx:          context.Background(),
		insecure:     false,
		timeout:      500 * time.Millisecond,
		errorDecoder: DefaultErrorDecoder,
	}
	for _, o := range opts {
		o(client)
	}
	var grpcOpts = []grpc.DialOption{
		grpc.WithTimeout(client.timeout),
		grpc.WithUnaryInterceptor(
			client.chainUnary(client.unaryInterceptor()),
		),
	}
	if client.insecure {
		grpcOpts = []grpc.DialOption{grpc.WithInsecure()}
	}
	return grpc.DialContext(client.ctx, target, grpcOpts...)
}

func (c *Client) unaryInterceptor() grpc.UnaryClientInterceptor {
	in := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
			return c.errorDecoder(ctx, err)
		}
		return nil
	}
	return c.chainUnary(append(c.ints, in)...)
}

func (c *Client) chainUnary(ints ...grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {
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
