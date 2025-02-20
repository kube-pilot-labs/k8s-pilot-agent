package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/segmentio/kafka-go"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type DeployCommand struct {
	DeployName    string `json:"deployName"`
	Namespace     string `json:"namespace"`
	ContainerSpec struct {
		Image   string            `json:"image"`
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
		Port    int32             `json:"port"`
	} `json:"containerSpec"`
	Resources struct {
		Requests struct {
			CPU    int `json:"cpu"`
			Memory int `json:"memory"`
		} `json:"requests"`
		Limits struct {
			CPU    int `json:"cpu"`
			Memory int `json:"memory"`
		} `json:"limits"`
	} `json:"resources"`
}

func createDeployment(clientset *kubernetes.Clientset, cmd DeployCommand) error {
	var envVars []corev1.EnvVar
	for k, v := range cmd.ContainerSpec.Env {
		envVars = append(envVars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	requestsCPU := fmt.Sprintf("%dm", cmd.Resources.Requests.CPU)
	limitsCPU := fmt.Sprintf("%dm", cmd.Resources.Limits.CPU)
	requestsMem := fmt.Sprintf("%dGi", cmd.Resources.Requests.Memory)
	limitsMem := fmt.Sprintf("%dGi", cmd.Resources.Limits.Memory)

	command := strings.Fields(cmd.ContainerSpec.Command)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmd.DeployName,
			Namespace: cmd.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": cmd.DeployName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": cmd.DeployName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    cmd.DeployName,
							Image:   cmd.ContainerSpec.Image,
							Command: command,
							Args:    cmd.ContainerSpec.Args,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: cmd.ContainerSpec.Port,
								},
							},
							Env: envVars,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(requestsCPU),
									corev1.ResourceMemory: resource.MustParse(requestsMem),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(limitsCPU),
									corev1.ResourceMemory: resource.MustParse(limitsMem),
								},
							},
						},
					},
				},
			},
		},
	}

	deploymentsClient := clientset.AppsV1().Deployments(cmd.Namespace)
	_, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment: %v", err)
	}
	log.Printf("Deployment %s created in namespace \"%s\"", cmd.DeployName, cmd.Namespace)
	return nil
}

func int32Ptr(i int32) *int32 {
	return &i
}

// consumeDeployCommands receives messages from a specified Kafka topic and creates deployments.
func consumeDeployCommands(kafkaBroker string, topic string, clientset *kubernetes.Clientset) {
	// Configure Kafka Reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    topic,
		GroupID:  "deployment-consumer-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer r.Close()

	ctx := context.Background()

	log.Printf("Started consuming Kafka topic \"%s\"", topic)
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("message read error: %v", err)
			continue
		}
		log.Printf("Received message at offset %d: %s", m.Offset, string(m.Value))
		var cmd DeployCommand
		if err := json.Unmarshal(m.Value, &cmd); err != nil {
			log.Printf("JSON parsing error: %v", err)
			continue
		}

		// Call deployment creation function
		if err := createDeployment(clientset, cmd); err != nil {
			log.Printf("failed to create deployment: %v", err)
		}
	}
}

func main() {
	// Create Kubernetes API client
	// Use InClusterConfig when running inside the cluster,
	// otherwise use the kubeconfig file.
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("failed to create Kubernetes config: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create Kubernetes client: %v", err)
	}

	// Set Kafka broker and topic (can be modified using environment variables or flags as needed)
	kafkaBroker := "localhost:9092"
	topic := "deploy-commands"

	// Start consuming Kafka messages
	consumeDeployCommands(kafkaBroker, topic, clientset)
}
