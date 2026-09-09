package discovery

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// ESXiConfig contient les paramètres de connexion à un hôte ESXi ou vCenter.
type ESXiConfig struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`      // ex: https://vcenter.local/sdk
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Insecure bool   `yaml:"insecure"`
}

// DiscoveredVM représente une VM découverte, indépendamment de l'hyperviseur.
type DiscoveredVM struct {
	Name       string
	IP         string
	Hypervisor string
	PowerState string
}

// DiscoverESXi se connecte à vCenter/ESXi et retourne la liste des VMs sous tension avec leur IP.
func DiscoverESXi(ctx context.Context, cfg ESXiConfig) ([]DiscoveredVM, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid vcenter url: %w", err)
	}
	u.User = url.UserPassword(cfg.Username, cfg.Password)

	client, err := govmomi.NewClient(ctx, u, cfg.Insecure)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", cfg.Name, err)
	}
	defer client.Logout(ctx)

	finder := find.NewFinder(client.Client, true)
	datacenters, err := finder.DatacenterList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to list datacenters: %w", err)
	}

	var results []DiscoveredVM

	for _, dc := range datacenters {
		finder.SetDatacenter(dc)

		m := view.NewManager(client.Client)
		v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
		if err != nil {
			return nil, fmt.Errorf("failed to create container view: %w", err)
		}
		defer v.Destroy(ctx)

		var vms []mo.VirtualMachine
		err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary", "guest"}, &vms)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve vms: %w", err)
		}

		for _, vm := range vms {
			if vm.Summary.Runtime.PowerState != types.VirtualMachinePowerStatePoweredOn {
				continue
			}
			ip := ""
			if vm.Guest != nil {
				ip = vm.Guest.IpAddress
			}
			if ip == "" {
				continue // pas d'IP remontée par VMware Tools, on ignore pour l'instant
			}
			results = append(results, DiscoveredVM{
				Name:       vm.Summary.Config.Name,
				IP:         ip,
				Hypervisor: "esxi",
				PowerState: string(vm.Summary.Runtime.PowerState),
			})
		}
	}

	return results, nil
}

// property est importé pour usage futur (filtrage de propriétés additionnelles).
var _ = property.Filter{}
