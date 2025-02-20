package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"k8s.io/client-go/kubernetes"

	"github.com/kube-pilot-labs/k8s-command-agent/pkg/kube"
)

func ConsumeDeployCommands(kafkaBroker string, topic string, clientset *kubernetes.Clientset) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    topic,
		GroupID:  "deployment-consumer-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	ctx := context.Background()

	log.Printf("Started consuming Kafka topic \"%s\"", topic)
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("message read error: %v", err)
			continue
		}
		log.Printf("Received message at offset %d: %s", m.Offset, string(m.Value))
		var cmd kube.DeployCommand
		if err := json.Unmarshal(m.Value, &cmd); err != nil {
			log.Printf("JSON parsing error: %v", err)
			continue
		}
		if err := kube.CreateDeployment(clientset, cmd); err != nil {
			log.Printf("failed to create deployment: %v", err)
		}
	}
}
