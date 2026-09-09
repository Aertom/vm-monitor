package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Aertom/vm-monitor/backend/internal/config"
	"github.com/Aertom/vm-monitor/backend/internal/scheduler"
	"github.com/Aertom/vm-monitor/backend/internal/store"
)

func main() {
	// Load config from environment or default path
	configPath := os.Getenv("VMMONITOR_CONFIG")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	// Load hypervisors config from environment or default path
	hypervisorsPath := os.Getenv("VMMONITOR_HYPERVISORS_CONFIG")
	if hypervisorsPath == "" {
		hypervisorsPath = "config/hypervisors.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	hvCfg, err := config.LoadHypervisors(hypervisorsPath)
	if err != nil {
		log.Fatalf("failed to load hypervisors config: %v", err)
	}

	st := store.New()

	sched := scheduler.New(cfg, st)

	// Start scheduler with discovery
	sched.StartWithDiscovery(*hvCfg, 15*time.Minute)

	// Keep the server running
	select {}
}
