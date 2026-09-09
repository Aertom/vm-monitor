package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type SSHConfig struct {
	User            string `yaml:"user"`
	PrivateKeyPath  string `yaml:"private_key_path"`
	Port            int    `yaml:"port"`
	TimeoutSeconds  int    `yaml:"timeout_seconds"`
}

type VMEntry struct {
	IP         string `yaml:"ip"`
	Hypervisor string `yaml:"hypervisor"`
}

type Config struct {
	SSH SSHConfig `yaml:"ssh"`
	VMs []VMEntry `yaml:"vms"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse config file %s: %w", path, err)
	}

	if cfg.SSH.Port == 0 {
		cfg.SSH.Port = 22
	}
	if cfg.SSH.TimeoutSeconds == 0 {
		cfg.SSH.TimeoutSeconds = 5
	}

	return &cfg, nil
}
