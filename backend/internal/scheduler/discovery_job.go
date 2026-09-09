package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/Aertom/vm-monitor/backend/internal/discovery"
)

// RunDiscovery interroge tous les hyperviseurs configurés et met à jour l'inventaire
// dynamique dans le store (IP + hostname + hyperviseur d'origine), avant que le
// collecteur SSH ne vienne enrichir chaque VM (versions d'applis, /etc/hosts).
func (s *Scheduler) RunDiscovery(ctx context.Context, cfg discovery.HypervisorsConfig) {
	start := time.Now()
	vms := discovery.DiscoverAll(ctx, cfg)
	log.Printf("discovery: found %d VMs across all hypervisors in %s", len(vms), time.Since(start))

	for _, vm := range vms {
		s.store.UpsertDiscoveredVM(vm.Name, vm.IP, vm.Hypervisor)
	}
}

