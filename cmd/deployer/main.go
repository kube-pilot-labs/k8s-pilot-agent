package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kube-pilot-labs/k8s-command-agent/pkg/kafka"
	"github.com/kube-pilot-labs/k8s-command-agent/pkg/kube"
)

func main() {
	clientset, err := kube.NewKubeClient()
	if err != nil {
		log.Fatalf("failed to create Kubernetes client: %v", err)
	}

	// Settings can be managed via environment variables or flags (e.g., kafkaBroker, topic, etc.)
	kafkaBroker := "localhost:9092"
	topic := "deploy-commands"

	// Execute Kafka message consumption in a separate goroutine
	go kafka.ConsumeDeployCommands(kafkaBroker, topic, clientset)

	// Handle graceful shutdown (signal capture)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Println("Shutdown signal received, exiting...")
}
