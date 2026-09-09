package collector

import (
	"sort"
	"strings"

	"github.com/Aertom/vm-monitor/backend/internal/models"
)

// BuildGroups déduit les groupes de VM liées à partir des entrées /etc/hosts
// collectées sur chaque VM. Le principe : si le hostname A cite le hostname B
// (et vice versa, ou transitivement), ils appartiennent au même groupe.
// L'ID de groupe est déterministe : les hostnames triés, joints par "-".
func BuildGroups(vms []models.VM, hostsEntries map[string][]EtcHostsEntry) []models.Group {
	// union-find simple sur les hostnames
	parent := make(map[string]string)
	var find func(string) string
	find = func(x string) string {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(a, b string) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}

	for _, vm := range vms {
		if _, ok := parent[vm.Hostname]; !ok {
			parent[vm.Hostname] = vm.Hostname
		}
	}

	for hostname, entries := range hostsEntries {
		if _, ok := parent[hostname]; !ok {
			parent[hostname] = hostname
		}
		for _, e := range entries {
			if _, ok := parent[e.Hostname]; ok {
				union(hostname, e.Hostname)
			}
		}
	}

	clusters := make(map[string][]models.VM)
	for i := range vms {
		root := find(vms[i].Hostname)
		clusters[root] = append(clusters[root], vms[i])
	}

	var groups []models.Group
	for _, members := range clusters {
		names := make([]string, 0, len(members))
		for _, m := range members {
			names = append(names, m.Hostname)
		}
		sort.Strings(names)
		groupID := strings.Join(names, "-")

		for i := range members {
			members[i].GroupID = groupID
		}

		groups = append(groups, models.Group{ID: groupID, VMs: members})
	}

	sort.Slice(groups, func(i, j int) bool { return groups[i].ID < groups[j].ID })

	return groups
}
