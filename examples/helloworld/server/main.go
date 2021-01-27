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
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/status"
	srvgrpc "github.com/go-kratos/kratos/v2/server/grpc"
	srvhttp "github.com/go-kratos/kratos/v2/server/http"
	"github.com/go-kratos/kratos/v2/transport"
	transgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	transhttp "github.com/go-kratos/kratos/v2/transport/http"
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

func loggerInfo(logger log.Logger) middleware.Middleware {
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
	logger := stdlog.NewLogger(stdlog.Writer(os.Stdout), stdlog.Skip(4))
	defer logger.Close()

	log := log.NewHelper("main", logger)

	s := &server{}
	app := kratos.New()

	httpTrans := transhttp.NewServer(transhttp.ServerMiddleware(
		middleware.Chain(
			logging.HTTPServer(logger),
			status.Server(),
			recovery.Recovery(),
		),
	))
	grpcTrans := transgrpc.NewServer(transgrpc.ServerMiddleware(
		middleware.Chain(
			logging.GRPCServer(logger),
			status.Server(),
			recovery.Recovery(),
		),
	))

	httpTrans.Use(s, loggerInfo(logger))
	grpcTrans.Use(s, loggerInfo(logger))

	httpServer := srvhttp.NewServer(srvhttp.Address(":8000"), srvhttp.Transport(httpTrans))
	grpcServer := srvgrpc.NewServer(srvgrpc.Address(":9000"), srvgrpc.Transport(grpcTrans))

	pb.RegisterGreeterServer(grpcServer, s)
	pb.RegisterGreeterHTTPServer(httpTrans, s)

	app.Append(httpServer)
	app.Append(grpcServer)

	if err := app.Run(); err != nil {
		log.Error(err)
	}
}
