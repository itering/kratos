package main

import (
	"context"
	"log"

	"github.com/go-kratos/kratos/v2/errors"
	pb "github.com/go-kratos/kratos/v2/examples/helloworld/helloworld"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/status"
	transgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	transhttp "github.com/go-kratos/kratos/v2/transport/http"
)

func main() {
	callHTTP()
	callGRPC()
}

func callHTTP() {
	client, err := transhttp.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Get("http://127.0.0.1:8000/helloworld/kratos")
	if err != nil {
		log.Fatal(err)
	}
	var reply pb.HelloReply
	if err := transhttp.DecodeResponse(resp, &reply); err != nil {
		log.Fatal(err)
	}
	log.Printf("hello %s\n", reply.Message)
	// returns error
	_, err = client.Get("http://127.0.0.1:8000/helloworld/error")
	if err != nil {
		log.Printf("SayHello error: %v\n", err)
	}
	if errors.IsInvalidArgument(err) {
		log.Printf("SayHello error is invalid argument: %v\n", err)
	}
}

func callGRPC() {
	conn, err := transgrpc.NewClient(
		"127.0.0.1:9000",
		transgrpc.ClientInsecure(),
		transgrpc.ClientMiddleware(
			middleware.Chain(
				status.Client(),
				recovery.Recovery(),
			),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	client := pb.NewGreeterClient(conn)
	reply, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "kratos"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SayHello %+v\n", reply)
	// returns error
	_, err = client.SayHello(context.Background(), &pb.HelloRequest{Name: "error"})
	if err != nil {
		log.Printf("SayHello error: %v\n", err)
	}
	if errors.IsInvalidArgument(err) {
		log.Printf("SayHello error is invalid argument: %v\n", err)
	}
}
