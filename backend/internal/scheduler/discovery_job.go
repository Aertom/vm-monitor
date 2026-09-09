package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/yourorg/vmmonitor/internal/discovery"
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

// StartWithDiscovery démarre la boucle périodique : découverte hyperviseurs puis collecte SSH.
func (s *Scheduler) StartWithDiscovery(interval time.Duration, hvCfg discovery.HypervisorsConfig) {
	ticker := time.NewTicker(interval)

	runOnce := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		s.RunDiscovery(ctx, hvCfg)
		s.collectAllOnce(ctx)
	}

	runOnce()
	go func() {
		for range ticker.C {
			runOnce()
		}
	}()
}
