package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kube-pilot-labs/k8s-pilot-agent/pkg/config"
	"github.com/kube-pilot-labs/k8s-pilot-agent/pkg/http"
	"github.com/kube-pilot-labs/k8s-pilot-agent/pkg/kafka"
	"github.com/kube-pilot-labs/k8s-pilot-agent/pkg/kube"
)

func main() {
	clientset, err := kube.NewKubeClient()
	if err != nil {
		log.Fatalf("failed to create k8s client: %v", err)
	}

	if err := config.InitConfig(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	kafka.InitializeTopicReaders()
	kafka.RegisterHandlers()

	cfg := config.GetConfig()

	ctx, cancel := context.WithCancel(context.Background())
	httpShutdownComplete := make(chan struct{})
	kafkaShutdownComplete := make(chan struct{})

	go http.StartServer(ctx, httpShutdownComplete)
	go kafka.ConsumeRequests(ctx, cfg.CreateDeployTopic, clientset, kafkaShutdownComplete)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Println("Shutdown signal received, exiting...")

	cancel()

	<-kafkaShutdownComplete
	<-httpShutdownComplete
	log.Println("All resources have been properly cleaned up")
}
