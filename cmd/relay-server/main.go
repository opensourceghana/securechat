package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/opensourceghana/securechat/pkg/network"
)

func main() {
	var (
		addr = flag.String("addr", "0.0.0.0", "Server address")
		port = flag.Int("port", 8080, "Server port")
	)
	flag.Parse()

	// Create server
	server := network.NewServer(network.ServerOptions{
		Addr: *addr,
		Port: *port,
	})

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		if err := server.Stop(); err != nil {
			log.Printf("Error stopping server: %v", err)
		}
		os.Exit(0)
	}()

	// Start server
	log.Printf("Starting relay server on %s:%d", *addr, *port)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
