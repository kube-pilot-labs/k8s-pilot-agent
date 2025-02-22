package config

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	KafkaBroker       string `mapstructure:"kafka_broker"`
	CreateDeployTopic string `mapstructure:"create_deploy_topic"`
}

var (
	c    *Config
	once sync.Once
)

func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if config.KafkaBroker == "" {
		return nil, errors.New("please provide the kafka_broker value")
	}
	if config.CreateDeployTopic == "" {
		return nil, errors.New("please provide the create_deploy_topic value")
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
