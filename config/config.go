package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	filePath = "./data/config.yaml"
)

type Config struct {
	RealIP struct {
		SrcIP string `yaml:"src_ip"`
		DstIP string `yaml:"dst_ip"`
	} `yaml:"real_ip"`
	ErrorIP struct {
		SrcIP string `yaml:"src_ip"`
		DstIP string `yaml:"dst_ip"`
	} `yaml:"error_ip"`
}

func Get() (*Config, error) {
	var cfg Config

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML file: %w", err)
	}

	return &cfg, nil
}
