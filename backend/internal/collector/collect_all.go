package collector

import (
	"fmt"
	"strings"
	"time"

	"github.com/Aertom/vm-monitor/backend/internal/config"
	"github.com/Aertom/vm-monitor/backend/internal/models"
	"github.com/Aertom/vm-monitor/backend/internal/ssh"
)

type Collector struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Collector {
	return &Collector{cfg: cfg}
}

// CollectAll se connecte à chaque VM de l'inventaire, récupère hostname,
// versions d'applis et contenu de /etc/hosts, puis reconstruit les groupes.
func (c *Collector) CollectAll() ([]models.VM, []models.Group) {
	var vms []models.VM
	hostsEntries := make(map[string][]EtcHostsEntry)

	timeout := time.Duration(c.cfg.SSH.TimeoutSeconds) * time.Second

	for _, entry := range c.cfg.VMs {
		vm := models.VM{
			ID:         entry.IP,
			IP:         entry.IP,
			Hypervisor: entry.Hypervisor,
			LastChecked: time.Now(),
		}

		client, err := ssh.NewClient(entry.IP, c.cfg.SSH.User, c.cfg.SSH.PrivateKeyPath, c.cfg.SSH.Port, timeout)
		if err != nil {
			vm.Reachable = false
			vm.Error = err.Error()
			vms = append(vms, vm)
			continue
		}

		hostnameOut, err := client.Run("hostname")
		if err != nil {
			vm.Reachable = false
			vm.Error = fmt.Sprintf("hostname: %v", err)
			vms = append(vms, vm)
			continue
		}
		vm.Hostname = strings.TrimSpace(hostnameOut)
		vm.Family = models.DetectFamily(vm.Hostname)
		vm.Reachable = true

		if versions, err := c.collectAppVersions(client, vm.Family); err == nil {
			vm.AppVersions = versions
		}

		if hostsOut, err := client.Run("cat /etc/hosts"); err == nil {
			hostsEntries[vm.Hostname] = parseEtcHosts(hostsOut)
		}

		vms = append(vms, vm)
	}

	groups := BuildGroups(vms, hostsEntries)
	return vms, groups
}

func (c *Collector) collectAppVersions(client *ssh.Client, family models.VMFamily) (map[string]string, error) {
	path := versionPathByFamily(family)
	output, err := client.Run(fmt.Sprintf("ls -1 %s 2>/dev/null", path))
	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", path, err)
	}
	return parseVersionDirs(output), nil
}
