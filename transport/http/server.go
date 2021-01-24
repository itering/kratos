package http

import (
	"context"
	"net/http"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/gorilla/mux"
)

// SupportPackageIsVersion1 These constants should not be referenced from any other code.
const SupportPackageIsVersion1 = true

// ServiceRegistrar wraps a single method that supports service registration.
type ServiceRegistrar interface {
	RegisterService(desc *ServiceDesc, impl interface{})
}

// ServiceDesc represents a HTTP service's specification.
type ServiceDesc struct {
	ServiceName string
	HandlerType interface{}
	Methods     []MethodDesc
	Metadata    interface{}
}

type serverMethodHandler func(srv interface{}, ctx context.Context, req *http.Request) (interface{}, error)

// MethodDesc represents a HTTP service's method specification.
type MethodDesc struct {
	Path    string
	Method  string
	Handler serverMethodHandler
}

// ServerOption is HTTP server option.
type ServerOption func(*serverOptions)

// serverOptions is HTTP server options.
type serverOptions struct {
	requestDecoder  ServerDecodeRequestFunc
	responseEncoder ServerEncodeResponseFunc
	errorEncoder    ServerEncodeErrorFunc
	middleware      middleware.Middleware
}

// ServerDecodeRequestFunc is decode request func.
type ServerDecodeRequestFunc func(in interface{}, req *http.Request) error

// ServerEncodeResponseFunc is encode response func.
type ServerEncodeResponseFunc func(out interface{}, res http.ResponseWriter, req *http.Request) error

// ServerEncodeErrorFunc is encode error func.
type ServerEncodeErrorFunc func(err error, res http.ResponseWriter, req *http.Request)

// ServerRequestDecoder with decode request option.
func ServerRequestDecoder(fn ServerEncodeErrorFunc) ServerOption {
	return func(o *serverOptions) {
		o.errorEncoder = fn
	}
}

// ServerResponseEncoder with response handler option.
func ServerResponseEncoder(fn ServerEncodeResponseFunc) ServerOption {
	return func(o *serverOptions) {
		o.responseEncoder = fn
	}
}

// ServerErrorEncoder with error handler option.
func ServerErrorEncoder(fn ServerEncodeErrorFunc) ServerOption {
	return func(o *serverOptions) {
		o.errorEncoder = fn
	}
}

// ServerMiddleware with server middleware option.
func ServerMiddleware(m ...middleware.Middleware) ServerOption {
	return func(o *serverOptions) {
		o.middleware = middleware.Chain(m[0], m[1:]...)
	}
}

// Server is a HTTP server wrapper.
type Server struct {
	router      *mux.Router
	opts        serverOptions
	middlewares map[interface{}]middleware.Middleware
}

// NewServer creates a HTTP server by options.
func NewServer(opts ...ServerOption) *Server {
	options := serverOptions{
		requestDecoder:  DefaultRequestDecoder,
		responseEncoder: DefaultResponseEncoder,
		errorEncoder:    DefaultErrorEncoder,
	}
	for _, o := range opts {
		o(&options)
	}
	return &Server{
		opts:        options,
		router:      mux.NewRouter(),
		middlewares: make(map[interface{}]middleware.Middleware),
	}
}

// Use use a middleware to the transport.
func (s *Server) Use(srv interface{}, m ...middleware.Middleware) {
	s.middlewares[srv] = middleware.Chain(m[0], m[1:]...)
}

// Handle registers a new route with a matcher for the URL path.
func (s *Server) Handle(path string, handler http.Handler) {
	s.router.Handle(path, handler)
}

// HandleFunc registers a new route with a matcher for the URL path.
func (s *Server) HandleFunc(path string, h func(http.ResponseWriter, *http.Request)) {
	s.router.HandleFunc(path, h)
}

// ServeHTTP should write reply headers and data to the ResponseWriter and then return.
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx := transport.NewContext(req.Context(), transport.Transport{Kind: "HTTP"})
	ctx = NewContext(ctx, ServerInfo{Request: req, Response: res})
	s.router.ServeHTTP(res, req.WithContext(ctx))
}

// RegisterService registers a service and its implementation to the HTTP server.
func (s *Server) RegisterService(sd *ServiceDesc, ss interface{}) {
	for _, method := range sd.Methods {
		s.registerHandle(ss, method)
	}
}

func (s *Server) registerHandle(srv interface{}, md MethodDesc) {
	s.router.HandleFunc(md.Path, func(res http.ResponseWriter, req *http.Request) {

		handler := func(ctx context.Context, in interface{}) (interface{}, error) {
			return md.Handler(srv, ctx, req)
		}
		if m, ok := s.middlewares[srv]; ok {
			handler = m(handler)
		}
		if s.opts.middleware != nil {
			handler = s.opts.middleware(handler)
		}

		reply, err := handler(req.Context(), req)
		if err != nil {
			s.opts.errorEncoder(err, res, req)
			return
		}

		if err := s.opts.responseEncoder(reply, res, req); err != nil {
			s.opts.errorEncoder(err, res, req)
			return
		}

	}).Methods(md.Method)
}
