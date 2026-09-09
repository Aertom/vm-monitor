package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/Aertom/vm-monitor/backend/internal/collector"
	"github.com/Aertom/vm-monitor/backend/internal/config"
	"github.com/Aertom/vm-monitor/backend/internal/discovery"
	"github.com/Aertom/vm-monitor/backend/internal/models"
	"github.com/Aertom/vm-monitor/backend/internal/store"
)

type Scheduler struct {
	collector *collector.Collector
	store     *store.Store
	stopCh    chan struct{}
}

func New(cfg *config.Config, st *store.Store) *Scheduler {
	return &Scheduler{
		collector: collector.New(cfg),
		store:     st,
		stopCh:    make(chan struct{}),
	}
}

// Start lance une collecte SSH immédiate puis périodique en arrière-plan,
// sans découverte automatique des hyperviseurs.
func (s *Scheduler) Start(interval time.Duration) {
	s.runOnce(context.Background(), nil)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.runOnce(context.Background(), nil)
			case <-s.stopCh:
				return
			}
		}
	}()
}

// StartWithDiscovery démarre une boucle périodique : découverte hyperviseurs (libvirt/vmware/ahv)
// puis collecte SSH enrichissement (versions d'applis, /etc/hosts), puis reconstruction des groupes.
// L'ordre d'arguments est (hvCfg, interval) pour matcher l'appel dans main.go.
func (s *Scheduler) StartWithDiscovery(hvCfg discovery.HypervisorsConfig, interval time.Duration) {
	ticker := time.NewTicker(interval)

	runOnce := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		s.runOnce(ctx, &hvCfg)
	}

	runOnce()
	go func() {
		for range ticker.C {
			runOnce()
		}
	}()
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}

// runOnce exécute un cycle complet :
// 1. Si hvCfg != nil : découverte hyperviseurs (ajoute les VMs découvertes à l'inventaire)
// 2. Collecte SSH (hostname, versions d'applis, /etc/hosts, reconstruction des groupes)
// 3. Conversion des types models → store et remplacement atomique du store
func (s *Scheduler) runOnce(ctx context.Context, hvCfg *discovery.HypervisorsConfig) {
	log.Println("scheduler: starting collection cycle")

	// Étape 1 : Découverte hyperviseurs (optionnelle)
	if hvCfg != nil {
		start := time.Now()
		discoveredVMs := discovery.DiscoverAll(ctx, *hvCfg)
		log.Printf("discovery: found %d VMs across all hypervisors in %s", len(discoveredVMs), time.Since(start))

		for _, vm := range discoveredVMs {
			s.store.UpsertDiscoveredVM(vm.Name, vm.IP, vm.Hypervisor)
		}
	}

	// Étape 2 : Collecte SSH (hostname, versions, /etc/hosts, reconstruction groupes)
	modelVMs, modelGroups := s.collector.CollectAll()

	// Étape 3 : Conversion et mise à jour atomique du store
	storeGroups := toStoreGroups(modelGroups)
	s.store.ReplaceGroups(storeGroups)

	log.Printf("scheduler: collection done, %d VMs, %d groups", len(modelVMs), len(storeGroups))
}

// toStoreGroups convertit []models.Group (VMs en slice) vers map[string]*store.Group (VMs indexées par famille).
// Chaque VM du groupe models est insérée dans store avec sa famille comme clé.
func toStoreGroups(modelGroups []models.Group) map[string]*store.Group {
	result := make(map[string]*store.Group)

	for _, mg := range modelGroups {
		sg := &store.Group{
			ID:  mg.ID,
			VMs: make(map[string]*store.VM),
		}

		// Conversion du statut checkout/checkin depuis models vers store
		if mg.InUseBy != nil {
			sg.InUseBy = *mg.InUseBy
		}
		sg.CheckedOutAt = mg.CheckedAt

		// Chaque VM modèles devient une store.VM, indexée par sa famille
		for _, mv := range mg.VMs {
			sv := &store.VM{
				Hostname:    mv.Hostname,
				IP:          mv.IP,
				Family:      string(mv.Family),
				Hypervisor:  mv.Hypervisor,
				AppVersions: mv.AppVersions,
				LastSeen:    mv.LastChecked,
			}
			if !mv.Reachable {
				sv.LastError = mv.Error
			}
			sg.VMs[string(mv.Family)] = sv
		}

		result[mg.ID] = sg
	}

	return result
}
