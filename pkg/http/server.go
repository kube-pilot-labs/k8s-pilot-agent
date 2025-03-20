package http

import (
	"log"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func StartServer() {
	http.HandleFunc("/healthz", HealthHandler)

	port := ":8080"
	log.Printf("HTTP server started on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("HTTP server failed to start: %v", err)
	}
}
