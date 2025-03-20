package config

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	KafkaBroker       string `mapstructure:"KAFKA_BROKER"`
	CreateDeployTopic string `mapstructure:"KAFKA_CREATE_DEPLOY_TOPIC"`
}

var (
	c    *Config
	once sync.Once
)

func loadConfig() (*Config, error) {
	viper.AutomaticEnv()
	viper.BindEnv("KAFKA_BROKER")
	viper.BindEnv("KAFKA_CREATE_DEPLOY_TOPIC")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if config.KafkaBroker == "" {
		return nil, errors.New("KAFKA_BROKER is not set")
	}
	if config.CreateDeployTopic == "" {
		return nil, errors.New("KAFKA_CREATE_DEPLOY_TOPIC is not set")
	}

	return &config, nil
}

func InitConfig() error {
	var err error
	once.Do(func() {
		c, err = loadConfig()
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
	})
	return err
}

func GetConfig() *Config {
	return c
}
