package main

import (
	"flag"
	"log"

	"github.com/shameerb/tcp-chat-redis/pkg/server"
)

var (
	redisAddr = flag.String("redis_addr", "localhost:6379", "redis address to connect to as host:port")
	grpcPort  = flag.String("grpc_port", "3000", "gRPC port")
)

func main() {
	c := server.NewServer(*redisAddr, *grpcPort)
	if err := c.Run(); err != nil {
		log.Fatalf("failed to start the server: %s", err)
	}
}
