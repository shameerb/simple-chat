package main

import (
	"flag"
	"log"

	"github.com/shameerb/tcp-chat-redis/pkg/client"
)

var (
	user       = flag.String("user", "", "username of the client")
	redisAddr  = flag.String("redis_addr", "localhost:6379", "redis address to connect to as host:port")
	serverAddr = flag.String("server_addr", "localhost:3000", "server address => host:port")
)

func main() {
	c := client.NewClient(*redisAddr, *serverAddr, *user)
	if err := c.Run(); err != nil {
		log.Fatalf("failed to start the client: %s", err)
	}
}
