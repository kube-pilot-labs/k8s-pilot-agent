package main

import (
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
	kafka.RegisterHandlers()

	cfg := config.GetConfig()

	go http.StartServer()
	go kafka.ConsumeRequests(cfg.KafkaBroker, cfg.CreateDeployTopic, clientset)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Println("Shutdown signal received, exiting...")
}
