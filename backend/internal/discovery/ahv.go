package discovery

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AHVConfig contient les paramètres de connexion à Prism Central/Element.
type AHVConfig struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"` // ex: https://prism.local:9440
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Insecure bool   `yaml:"insecure"`
}

type ahvVMListRequest struct {
	Kind   string `json:"kind"`
	Length int    `json:"length"`
	Offset int    `json:"offset"`
}

type ahvNicIPEndpoint struct {
	IP string `json:"ip"`
}

type ahvNic struct {
	IPEndpointList []ahvNicIPEndpoint `json:"ip_endpoint_list"`
}

type ahvVMSpec struct {
	Name     string `json:"name"`
	Resources struct {
		PowerState  string   `json:"power_state"`
		NicList     []ahvNic `json:"nic_list"`
	} `json:"resources"`
}

type ahvVMEntity struct {
	Spec ahvVMSpec `json:"spec"`
}

type ahvVMListResponse struct {
	Entities []ahvVMEntity `json:"entities"`
}

// DiscoverAHV interroge l'API Prism v2 (/vms/list) et retourne les VMs sous tension avec leur IP.
func DiscoverAHV(ctx context.Context, cfg AHVConfig) ([]DiscoveredVM, error) {
	endpoint := fmt.Sprintf("%s/api/nutanix/v2.0/vms/list", cfg.URL)

	reqBody, err := json.Marshal(ahvVMListRequest{Kind: "vm", Length: 500, Offset: 0})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Insecure}, //nolint:gosec
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.SetBasicAuth(cfg.Username, cfg.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach prism %s: %w", cfg.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prism %s returned status %d", cfg.Name, resp.StatusCode)
	}

	var listResp ahvVMListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode prism response: %w", err)
	}

	var results []DiscoveredVM
	for _, entity := range listResp.Entities {
		if entity.Spec.Resources.PowerState != "ON" {
			continue
		}
		ip := ""
		for _, nic := range entity.Spec.Resources.NicList {
			if len(nic.IPEndpointList) > 0 {
				ip = nic.IPEndpointList[0].IP
				break
			}
		}
		if ip == "" {
			continue
		}
		results = append(results, DiscoveredVM{
			Name:       entity.Spec.Name,
			IP:         ip,
			Hypervisor: "ahv",
			PowerState: entity.Spec.Resources.PowerState,
		})
	}

	return results, nil
}
