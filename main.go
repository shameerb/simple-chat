package main

import (
	"log"
	"net"
)

func main() {
	// create an instance of server and // listen to commands on a go routine in a channel
	s := newServer()
	go s.run()

	// create a tcp server
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("TCP server failed to start %s\n", err)
	}
	defer listener.Close()

	// accept new connections on the tcp port and create a new client
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connections: %s\n", err)
		}

		c := s.newClient(conn)
		// keep a continous loop on a goroutine listening on this client and send the commands to the main chan
		// todo: stop reading from the conn after quit. Use a context to exit the goroutine on a cancel on the context. Next Steps
		go c.readInput()
	}
}
