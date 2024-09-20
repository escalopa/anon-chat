package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	log "github.com/catalystgo/logger/cli"
)

func gracefulShutdown(stop func(ctx context.Context) error) {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	sig := <-stopChan
	fmt.Println() // New line after ^C
	log.Warnf("Received signal: %v", sig)
	log.Info("Shutting down server...")

	if err := stop(context.Background()); err != nil {
		log.Errorf("Error stopping server: %v", err)
	}
}

const losthost = "localhost"

func getAddress() string {
	name, err := os.Hostname()
	if err != nil {
		log.Debugf("Error getting hostname: %v", err)
		return losthost
	}

	addrs, err := net.LookupHost(name)
	if err != nil {
		log.Debugf("Error looking up host %s: %v", name, err)
		return losthost
	}

	for _, a := range addrs {
		return a // Return the first address
	}

	return losthost
}
