package discovery

import (
	"context"
	"fmt"

	"libvirt.org/go/libvirt"
)

// KVMConfig contient les paramètres de connexion à un hôte libvirt.
// L'URI est typiquement du type qemu+ssh://user@host/system pour une connexion à distance par clé SSH.
type KVMConfig struct {
	Name string `yaml:"name"`
	URI  string `yaml:"uri"` // ex: qemu+ssh://root@kvm-host01/system
}

// DiscoverKVM se connecte à un hôte libvirt et retourne les VMs actives avec leur IP (via l'agent QEMU ou le bail DHCP).
func DiscoverKVM(ctx context.Context, cfg KVMConfig) ([]DiscoveredVM, error) {
	conn, err := libvirt.NewConnect(cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt host %s: %w", cfg.Name, err)
	}
	defer conn.Close()

	domains, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains on %s: %w", cfg.Name, err)
	}

	var results []DiscoveredVM

	for _, domain := range domains {
		name, err := domain.GetName()
		if err != nil {
			continue
		}

		ip := extractDomainIP(&domain)
		if ip == "" {
			domain.Free()
			continue // pas d'IP disponible (agent QEMU absent ou bail DHCP non trouvé)
		}

		results = append(results, DiscoveredVM{
			Name:       name,
			IP:         ip,
			Hypervisor: "kvm",
			PowerState: "running",
		})

		domain.Free()
	}

	return results, nil
}

// extractDomainIP tente d'abord via l'agent QEMU (le plus fiable), puis via le bail DHCP libvirt.
func extractDomainIP(domain *libvirt.Domain) string {
	// 1. Tentative via l'agent QEMU (nécessite qemu-guest-agent installé dans la VM)
	ifaces, err := domain.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT)
	if err == nil {
		for _, iface := range ifaces {
			for _, addr := range iface.Addrs {
				if addr.Type == libvirt.IP_ADDR_TYPE_IPV4 {
					return addr.Addr
				}
			}
		}
	}

	// 2. Fallback : bail DHCP connu de libvirt (moins fiable si IP statique)
	ifaces, err = domain.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	if err == nil {
		for _, iface := range ifaces {
			for _, addr := range iface.Addrs {
				if addr.Type == libvirt.IP_ADDR_TYPE_IPV4 {
					return addr.Addr
				}
			}
		}
	}

	return ""
}
