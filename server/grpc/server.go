package grpc

import (
	"context"
	"net"
	"time"

	pb "github.com/go-kratos/kratos/v2/api/kratos/config/grpc"
	"github.com/go-kratos/kratos/v2/server"
	"google.golang.org/grpc"
)

var _ server.Server = (*Server)(nil)

// Option is grpc option.
type Option func(*options)

type options struct {
	network  string
	address  string
	timeout  time.Duration
	unaryInt grpc.UnaryServerInterceptor
}

// Network with server network.
func Network(network string) Option {
	return func(o *options) {
		o.network = network
	}
}

// Address with server address.
func Address(addr string) Option {
	return func(o *options) {
		o.address = addr
	}
}

// Timeout with server timeout.
func Timeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

// UnaryInterceptor with grpc unary interceptor.
func UnaryInterceptor(in grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.unaryInt = in
	}
}

// Apply apply config.
func Apply(c *pb.ServerConfig) Option {
	return func(o *options) {
		o.network = c.Network
		o.address = c.Address
		if c.Timeout != nil {
			o.timeout = c.Timeout.AsDuration()
		}
	}
}

// Server is a gRPC server wrapper.
type Server struct {
	*grpc.Server
	opts options
}

// NewServer creates a gRPC server by options.
func NewServer(opts ...Option) *Server {
	options := options{
		network: "tcp",
		address: ":9000",
		timeout: time.Second,
	}
	for _, o := range opts {
		o(&options)
	}
	return &Server{
		opts:   options,
		Server: grpc.NewServer(),
	}
}

// Start start the gRPC server.
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen(s.opts.network, s.opts.address)
	if err != nil {
		return err
	}
	return s.Serve(lis)
}

// Stop stop the gRPC server.
func (s *Server) Stop(ctx context.Context) error {
	s.GracefulStop()
	return nil
}
