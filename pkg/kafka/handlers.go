package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"k8s.io/client-go/kubernetes"

	"github.com/kube-pilot-labs/k8s-pilot-agent/pkg/config"
	"github.com/kube-pilot-labs/k8s-pilot-agent/pkg/kube"
)

// MessageHandlerFunc is the type for handler functions for Kafka messages.
type MessageHandlerFunc func(clientset *kubernetes.Clientset, m kafka.Message) error

// TopicHandlers is a map of topics and their corresponding handler functions.
var TopicHandlers map[string]MessageHandlerFunc

func init() {
	TopicHandlers = map[string]MessageHandlerFunc{
		config.GetConfig().CreateDeployTopic: handleCreateDeploy,
	}
}

// handleCreateDeploy is the handler function for the deploy topic.
func handleCreateDeploy(clientset *kubernetes.Clientset, m kafka.Message) error {
	var cmd kube.DeployCommand
	if err := json.Unmarshal(m.Value, &cmd); err != nil {
		return fmt.Errorf("JSON parsing error: %v", err)
	}
	if err := kube.CreateDeployment(clientset, cmd); err != nil {
		return fmt.Errorf("deployment creation failed: %v", err)
	}
	return nil
}
