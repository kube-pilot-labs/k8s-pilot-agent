package http

import (
	"context"
	"log"
	"net/http"
	"time"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func StartServer(ctx context.Context, shutdownComplete chan<- struct{}) {
	http.HandleFunc("/healthz", HealthHandler)

	port := ":8080"
	server := &http.Server{
		Addr: port,
	}

	defer func() {
		log.Println("Context canceled, shutting down HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
		close(shutdownComplete)
	}()

	go func() {
		log.Printf("HTTP server started on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed to start: %v", err)
		}
	}()
	<-ctx.Done()
}
