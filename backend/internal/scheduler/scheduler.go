package scheduler

import (
	"log"
	"time"

	"github.com/Aertom/vm-monitor/backend/internal/collector"
	"github.com/Aertom/vm-monitor/backend/internal/config"
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

// Start lance une collecte immédiate puis une collecte périodique en arrière-plan.
func (s *Scheduler) Start(interval time.Duration) {
	s.runOnce()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) runOnce() {
	log.Println("scheduler: starting collection cycle")
	vms, groups := s.collector.CollectAll()
	s.store.ReplaceAll(vms, groups)
	log.Printf("scheduler: collection done, %d VMs, %d groups", len(vms), len(groups))
}
