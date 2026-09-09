package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Aertom/vm-monitor/backend/internal/api"
	"github.com/Aertom/vm-monitor/backend/internal/config"
	"github.com/Aertom/vm-monitor/backend/internal/scheduler"
	"github.com/Aertom/vm-monitor/backend/internal/store"
)

func main() {
	configPath := os.Getenv("VMMONITOR_CONFIG")
	if configPath == "" {
		configPath = "config/inventory.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	st := store.New()

	sched := scheduler.New(cfg, st)
	sched.Start(15 * time.Minute)

	a := &api.API{Store: st}
	router := a.Router()

	addr := os.Getenv("VMMONITOR_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("vmmonitor listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
