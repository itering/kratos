package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/errors"
	pb "github.com/go-kratos/kratos/v2/examples/helloworld/helloworld"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/log/stdlog"
	"github.com/go-kratos/kratos/v2/middleware"
	servergrpc "github.com/go-kratos/kratos/v2/server/grpc"
	serverhttp "github.com/go-kratos/kratos/v2/server/http"
	"github.com/go-kratos/kratos/v2/transport"
	transgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	transhttp "github.com/go-kratos/kratos/v2/transport/http"

	"google.golang.org/grpc"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	if in.Name == "error" {
		return nil, errors.InvalidArgument("BadRequest", "invalid argument %s", in.Name)
	}
	if in.Name == "panic" {
		panic("grpc panic")
	}
	return &pb.HelloReply{Message: fmt.Sprintf("Hello %+v", in)}, nil
}

func logger1(logger log.Logger) middleware.Middleware {
	log := log.NewHelper("logger1", logger)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {

			log.Info("before")

			return handler(ctx, req)
		}
	}
}

func logger2(logger log.Logger) middleware.Middleware {
	log := log.NewHelper("logger2", logger)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {

			resp, err := handler(ctx, req)

			log.Info("after")

			return resp, err
		}
	}
}

func logger3(logger log.Logger) middleware.Middleware {
	log := log.NewHelper("logger2", logger)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {

			tr, ok := transport.FromContext(ctx)
			if ok {
				log.Infof("transport: %+v", tr)
			}
			h, ok := transhttp.FromContext(ctx)
			if ok {
				log.Infof("http: [%s] %s", h.Request.Method, h.Request.URL.Path)
			}
			g, ok := transgrpc.FromContext(ctx)
			if ok {
				log.Infof("grpc: %s", g.FullMethod)
			}

			return handler(ctx, req)
		}
	}
}

func main() {
	logger, err := stdlog.NewLogger(stdlog.Writer(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	log := log.NewHelper("main", logger)

	s := &server{}
	app := kratos.New()

	httpTrans := transhttp.NewServer(transhttp.ServerMiddleware(logger1(logger), logger2(logger)))
	httpTrans.Use(s, logger3(logger))

	grpcTrans := transgrpc.NewServer(transgrpc.ServerMiddleware(logger1(logger), logger2(logger)))
	grpcTrans.Use(s, logger3(logger))

	httpServer := serverhttp.NewServer("tcp", ":8000", serverhttp.Handler(httpTrans))
	grpcServer := servergrpc.NewServer("tcp", ":9000", grpc.UnaryInterceptor(grpcTrans.UnaryInterceptor()))

	pb.RegisterGreeterServer(grpcServer, s)
	pb.RegisterGreeterHTTPServer(httpTrans, s)

	app.Append(httpServer)
	app.Append(grpcServer)

	if err := app.Run(); err != nil {
		log.Error(err)
	}
}
