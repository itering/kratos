package grpc

import (
	"context"
	"net"

	"github.com/go-kratos/kratos/v2/server"
	"google.golang.org/grpc"
)

var _ server.Server = (*Server)(nil)

// Server is a gRPC server wrapper.
type Server struct {
	*grpc.Server

	network string
	address string
}

// NewServer creates a gRPC server by options.
func NewServer(network, address string, opts ...grpc.ServerOption) *Server {
	return &Server{
		network: network,
		address: address,
		Server:  grpc.NewServer(opts...),
	}
}

// Start start the gRPC server.
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen(s.network, s.address)
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
