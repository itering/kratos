package grpc

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"google.golang.org/grpc"
)

// ServerOption is gRPC server option.
type ServerOption func(o *Server)

// ServerMiddleware with server middleware.
func ServerMiddleware(m middleware.Middleware) ServerOption {
	return func(o *Server) {
		o.globalMiddleware = m
	}
}

// Server is a gRPC server wrapper.
type Server struct {
	globalMiddleware  middleware.Middleware
	serviceMiddleware map[interface{}]middleware.Middleware
}

// NewServer creates a gRPC server by options.
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		serviceMiddleware: make(map[interface{}]middleware.Middleware),
	}
	for _, o := range opts {
		o(srv)
	}
	return srv
}

// Use use a middleware to the transport.
func (s *Server) Use(srv interface{}, m middleware.Middleware) {
	s.serviceMiddleware[srv] = m
}

// UnaryInterceptor returns a unary server interceptor.
func (s *Server) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = transport.NewContext(ctx, transport.Transport{Kind: "GRPC"})
		ctx = NewContext(ctx, ServerInfo{Server: info.Server, FullMethod: info.FullMethod})
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		}
		if m, ok := s.serviceMiddleware[info.Server]; ok {
			h = m(h)
		}
		if s.globalMiddleware != nil {
			h = s.globalMiddleware(h)
		}
		reply, err := h(ctx, req)
		if err != nil {
			return nil, err
		}
		return reply, nil
	}
}
