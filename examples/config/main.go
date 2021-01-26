package main

import (
	"flag"
	"log"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/source/file"
)

// Service is service config.
type Service struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Server is server config.
type Server struct {
	Network string `json:"network"`
	Address string `json:"address"`
}

var flagconf string

func init() {
	flag.StringVar(&flagconf, "conf", "config.yaml", "config path, eg: -conf ../configs")
}

func main() {
	flag.Parse()
	conf := config.New(config.WithSource(
		file.NewSource(flagconf),
	))
	if err := conf.Load(); err != nil {
		panic(err)
	}

	var (
		sc Service
		hc Server
		gc Server
	)
	if err := conf.Value("service").Scan(&sc); err != nil {
		panic(err)
	}
	if err := conf.Value("server.http").Scan(&hc); err != nil {
		panic(err)
	}
	if err := conf.Value("server.grpc").Scan(&gc); err != nil {
		panic(err)
	}

	log.Printf("service: %+v\n", sc)
	log.Printf("http: %+v\n", hc)
	log.Printf("grpc: %+v\n", gc)
}
