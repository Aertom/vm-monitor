package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/Aertom/vm-monitor/backend/internal/discovery"
)

// LoadHypervisors charge la configuration des hyperviseurs depuis un fichier YAML.
func LoadHypervisors(path string) (*discovery.HypervisorsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read hypervisors config file %s: %w", path, err)
	}

	var cfg discovery.HypervisorsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse hypervisors config file %s: %w", path, err)
	}

	return &cfg, nil
}
