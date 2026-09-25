package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Clients []ClientConfig `yaml:"clients"`
	GRPC    GRPCConfig     `yaml:"grpc"`
}

type ClientConfig struct {
	ID        string       `yaml:"id"`
	Enabled   bool         `yaml:"enabled"`
	MQTT      MQTTConfig   `yaml:"mqtt"`
	Sparkplug SparkplugCfg `yaml:"sparkplug"`
}

type MQTTConfig struct {
	BrokerURL    string `yaml:"broker_url"`
	ClientID     string `yaml:"client_id"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	KeepAliveSec int    `yaml:"keep_alive_sec"`
	QoS          byte   `yaml:"qos"`
}

type SparkplugCfg struct {
	GroupID   string `yaml:"group_id"`
	NodeID    string `yaml:"node_id"`
	Namespace string `yaml:"namespace"`
}

type GRPCConfig struct {
	SocketPath        string `yaml:"socket_path"`
	SocketPermissions int    `yaml:"socket_permissions"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if len(cfg.Clients) == 0 {
		return nil, errors.New("config requires at least one client")
	}
	return &cfg, validate(&cfg)
}

func validate(cfg *Config) error {
	for _, client := range cfg.Clients {
		if client.ID == "" {
			return errors.New("client id is required")
		}
		if client.MQTT.BrokerURL == "" {
			return fmt.Errorf("client %s: mqtt broker_url is required", client.ID)
		}
		if client.Sparkplug.GroupID == "" || client.Sparkplug.NodeID == "" || client.Sparkplug.Namespace == "" {
			return fmt.Errorf("client %s: sparkplug group_id, node_id, namespace are required", client.ID)
		}
	}
	if cfg.GRPC.SocketPath == "" {
		return errors.New("grpc socket_path is required")
	}
	if cfg.GRPC.SocketPermissions == 0 {
		cfg.GRPC.SocketPermissions = 0666
	}
	return nil
}
