package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/cmd/config/factories"
)

func main() {
	server := factories.MakeServer()

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: server.Engine,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("failed to run server: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	log.Println("Shutdown signal received. Gracefully shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("failed to shutdown server gracefully: %v\n", err)
		os.Exit(1)
	}

	log.Println("Server shutdown complete")
}
