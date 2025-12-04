package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/PrismManager/gemstone/internal/daemon"
)

func main() {
	d, err := daemon.New()
	if err != nil {
		log.Fatalf("Failed to initialize daemon: %v", err)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gemstone daemon...")
		d.Shutdown()
		os.Exit(0)
	}()

	log.Println("Starting gemstone daemon...")
	if err := d.Run(); err != nil {
		log.Fatalf("Daemon error: %v", err)
	}
}
