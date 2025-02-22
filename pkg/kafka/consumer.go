package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
	"k8s.io/client-go/kubernetes"
)

// ConsumeRequests continuously consumes Kafka messages and calls the corresponding handler function.
func ConsumeRequests(kafkaBroker string, topic string, clientset *kubernetes.Clientset) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    topic,
		GroupID:  topic + ".reader",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	ctx := context.Background()

	log.Printf("[%s] Kafka topic consumption started", topic)
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("[%s] error reading message: %v", topic, err)
			continue
		}
		log.Printf("[%s] received message at offset %d: %s", topic, m.Offset, string(m.Value))

		handler, exists := TopicHandlers[topic]
		if !exists {
			log.Printf("[%s] no handler defined for this topic", topic)
			continue
		}

		if err := handler(clientset, m); err != nil {
			log.Printf("[%s] message processing failed: %v", topic, err)
		}
	}
}
