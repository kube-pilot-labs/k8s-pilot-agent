package kafka

import (
	"context"
	"log"
	"time"

	"github.com/kube-pilot-labs/k8s-pilot-agent/pkg/config"
	"github.com/segmentio/kafka-go"
	"k8s.io/client-go/kubernetes"
)

var (
	TopicReaders = map[string]*kafka.Reader{}
	TopicWriters = map[string]*kafka.Writer{}
)

func InitializeTopicReaders(ctx context.Context, shutdownComplete chan<- struct{}) {
	cfg := config.GetConfig()
	createDeploymentReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.KafkaBroker},
		Topic:    cfg.CreateDeployTopic,
		GroupID:  cfg.CreateDeployTopic + ".reader",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	TopicReaders[cfg.CreateDeployTopic] = createDeploymentReader

	writer := &kafka.Writer{
		Addr:      kafka.TCP(cfg.KafkaBroker),
		Topic:     cfg.HealthCheckTopic,
		Balancer:  &kafka.LeastBytes{},
		BatchSize: 1,
	}
	TopicWriters[cfg.HealthCheckTopic] = writer

	go func() {
		<-ctx.Done()
		for topic, reader := range TopicReaders {
			log.Printf("Closing Kafka reader for topic: %s", topic)
			if err := reader.Close(); err != nil {
				log.Printf("Error closing Kafka reader for topic %s: %v", topic, err)
			}
		}
		for topic, writer := range TopicWriters {
			log.Printf("Closing Kafka writer for topic: %s", topic)
			if err := writer.Close(); err != nil {
				log.Printf("Error closing Kafka writer for topic %s: %v", topic, err)
			}
		}
		close(shutdownComplete)
	}()
}

func IsConnected(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cfg := config.GetConfig()
	writer, exists := TopicWriters[cfg.HealthCheckTopic]
	if !exists {
		return false
	}

	err := writer.WriteMessages(ctx, kafka.Message{
		Value: []byte(""),
	})
	if err != nil {
		if ctx.Err() != nil {
			log.Printf("Kafka connection check timed out: %v", err)
		} else {
			log.Printf("Kafka connection failed: %v", err)
		}
		return false
	}

	return true
}

// ConsumeRequests continuously consumes Kafka messages and calls the corresponding handler function.
func ConsumeRequests(ctx context.Context, topic string, clientset *kubernetes.Clientset) {
	reader, exists := TopicReaders[topic]
	if !exists {
		log.Printf("Kafka reader for topic %s not initialized", topic)
		return
	}

	log.Printf("[%s] Kafka topic consumption started", topic)

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
