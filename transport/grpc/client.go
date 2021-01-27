package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// ClientOption is gRPC client option.
type ClientOption func(o *Client)

// ClientDecodeErrorFunc is encode error func.
type ClientDecodeErrorFunc func(err error) error

// ClientErrorDecoder with client error decoder.
func ClientErrorDecoder(d ClientDecodeErrorFunc) ClientOption {
	return func(c *Client) {
		c.errorDecoder = d
	}
}

// ClientRecoveryHandler with recovery handler.
func ClientRecoveryHandler(h RecoveryHandlerFunc) ClientOption {
	return func(c *Client) {
		c.recoveryHandler = h
	}
}

// ClientUnaryInterceptor with client unary interceptor.
func ClientUnaryInterceptor(in grpc.UnaryClientInterceptor) ClientOption {
	return func(c *Client) {
		c.in = in
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
	ctx             context.Context
	insecure        bool
	block           bool
	timeout         time.Duration
	in              grpc.UnaryClientInterceptor
	errorDecoder    ClientDecodeErrorFunc
	recoveryHandler RecoveryHandlerFunc
}

// NewClient new a grpc transport client.
func NewClient(target string, opts ...ClientOption) (*grpc.ClientConn, error) {
	client := &Client{
		ctx:             context.Background(),
		timeout:         500 * time.Millisecond,
		insecure:        false,
		errorDecoder:    DefaultErrorDecoder,
		recoveryHandler: DefaultRecoveryHandler,
	}
	for _, o := range opts {
		o(client)
	}
	var grpcOpts = []grpc.DialOption{
		grpc.WithTimeout(client.timeout),
		grpc.WithUnaryInterceptor(
			ChainUnaryClient(client.unaryInterceptor(), client.in),
		),
	}
	if client.insecure {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}
	return grpc.DialContext(client.ctx, target, grpcOpts...)
}

func (c *Client) unaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				err = c.recoveryHandler(ctx, req, rerr)
			}
		}()
		if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
			return c.errorDecoder(err)
		}
		return nil
	}
}
