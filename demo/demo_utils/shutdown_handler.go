package demoutils

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// ShutdownHandler sets up a signal handler to gracefully shut down the application when an interrupt or termination signal is received.
//
// The provided callback function will be called when the signal is received.
func ShutdownHandler(callback func()) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Set up a channel to listen for OS signals.
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

		// Wait for an OS signal.
		<-signalCh
		log.Println("Received interrupt signal, shutting down...")

		// Stop receiving OS signals.
		signal.Stop(signalCh)

		// Call the provided callback function to perform the shutdown tasks.
		callback()
	}()
	return &wg
}
