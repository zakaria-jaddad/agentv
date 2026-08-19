package config

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/assert/yaml"
)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var conf Config

	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	if conf.Token == "" {
		conf.Token = os.Getenv("AGENT_TOKEN")
	}

	return &conf, nil
}
