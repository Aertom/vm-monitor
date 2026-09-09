package discovery

import (
	"context"
	"log"
	"sync"
)

// HypervisorsConfig regroupe la config de tous les hyperviseurs à interroger.
type HypervisorsConfig struct {
	ESXi []ESXiConfig `yaml:"esxi"`
	AHV  []AHVConfig  `yaml:"ahv"`
	KVM  []KVMConfig  `yaml:"kvm"`
}

// DiscoverAll interroge en parallèle tous les hyperviseurs configurés et agrège les VMs découvertes.
// Les erreurs individuelles sont loggées mais n'interrompent pas la découverte globale.
func DiscoverAll(ctx context.Context, cfg HypervisorsConfig) []DiscoveredVM {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results []DiscoveredVM
	)

	append_ := func(vms []DiscoveredVM, err error, source string) {
		if err != nil {
			log.Printf("discovery: %s failed: %v", source, err)
			return
		}
		mu.Lock()
		results = append(results, vms...)
		mu.Unlock()
	}

	for _, c := range cfg.ESXi {
		wg.Add(1)
		go func(c ESXiConfig) {
			defer wg.Done()
			vms, err := DiscoverESXi(ctx, c)
			append_(vms, err, "esxi/"+c.Name)
		}(c)
	}

	for _, c := range cfg.AHV {
		wg.Add(1)
		go func(c AHVConfig) {
			defer wg.Done()
			vms, err := DiscoverAHV(ctx, c)
			append_(vms, err, "ahv/"+c.Name)
		}(c)
	}

	for _, c := range cfg.KVM {
		wg.Add(1)
		go func(c KVMConfig) {
			defer wg.Done()
			vms, err := DiscoverKVM(ctx, c)
			append_(vms, err, "kvm/"+c.Name)
		}(c)
	}

	wg.Wait()
	return results
}
