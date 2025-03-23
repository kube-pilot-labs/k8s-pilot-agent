package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
	"k8s.io/client-go/kubernetes"
)

// ConsumeRequests continuously consumes Kafka messages and calls the corresponding handler function.
func ConsumeRequests(ctx context.Context, kafkaBroker string, topic string, clientset *kubernetes.Clientset, shutdownComplete chan<- struct{}) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    topic,
		GroupID:  topic + ".reader",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	log.Printf("[%s] Kafka topic consumption started", topic)

	defer func() {
		log.Printf("[%s] Closing Kafka reader", topic)
		if err := reader.Close(); err != nil {
			log.Printf("[%s] Error closing Kafka reader: %v", topic, err)
		}

		close(shutdownComplete)
		log.Printf("[%s] Kafka consumer shutdown completed", topic)
	}()

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("[%s] Context canceled, stopping consumer", topic)
				return
			}
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
