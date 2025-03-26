package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/kube-pilot-labs/k8s-pilot-agent/pkg/config"
	"github.com/kube-pilot-labs/k8s-pilot-agent/pkg/kafka"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	if !kafka.IsConnected(cfg.KafkaBroker, 5*time.Second) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Kafka connection failed"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func StartServer(ctx context.Context, shutdownComplete chan<- struct{}) {
	http.HandleFunc("/ping", PingHandler)
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
