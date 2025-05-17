package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"coffee-loyalty-system/cmd"
)

func main() {
	server, err := cmd.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)
	case sig := <-quit:
		log.Printf("Received signal: %v", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Failed to shutdown server: %v", err)
		}
	}
}
