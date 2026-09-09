package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ESXiConfig contient les paramètres de connexion à un serveur ESXi/vCenter.
type ESXiConfig struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Insecure bool   `yaml:"insecure"`
}

// AHVConfig contient les paramètres de connexion à un cluster Nutanix AHV.
type AHVConfig struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Insecure bool   `yaml:"insecure"`
}

// KVMConfig contient les paramètres de connexion à un hyperviseur KVM.
type KVMConfig struct {
	Name string `yaml:"name"`
	URI  string `yaml:"uri"`
}

// HypervisorsConfig regroupe la config de tous les hyperviseurs à interroger.
type HypervisorsConfig struct {
	ESXi []ESXiConfig `yaml:"esxi"`
	AHV  []AHVConfig  `yaml:"ahv"`
	KVM  []KVMConfig  `yaml:"kvm"`
}

// LoadHypervisors charge la configuration des hyperviseurs depuis un fichier YAML.
func LoadHypervisors(path string) (*HypervisorsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read hypervisors config file %s: %w", path, err)
	}

	var cfg HypervisorsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse hypervisors config file %s: %w", path, err)
	}

	return &cfg, nil
}
