package main

import (
	"context"
	"log"

	pb "github.com/go-kratos/kratos/v2/examples/helloworld/helloworld"
	transgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc"
)

func main() {
	trans := transgrpc.NewClient()
	conn, err := grpc.DialContext(context.Background(), "127.0.0.1:9000",
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(trans.Interceptor()),
	)
	if err != nil {
		log.Fatal(err)
	}
	client := pb.NewGreeterClient(conn)
	reply, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "Kratos"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SayHello %+v\n", reply)
	// error
	_, err = client.SayHello(context.Background(), &pb.HelloRequest{Name: "error"})
	if err != nil {
		log.Printf("SayHello error: %+v", err)
	}
}
