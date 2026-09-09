package store

import (
	"sync"
	"time"
)

// VM représente une machine virtuelle telle que connue par le store.
type VM struct {
	Hostname    string            `json:"hostname"`
	IP          string            `json:"ip"`
	Family      string            `json:"family"`
	Hypervisor  string            `json:"hypervisor"`
	AppVersions map[string]string `json:"app_versions,omitempty"`
	LastSeen    time.Time         `json:"last_seen"`
	LastError   string            `json:"last_error,omitempty"`
}

// Group représente un groupe de VMs liées (une par famille sm/cm/ws/oa).
type Group struct {
	ID          string         `json:"id"`
	VMs         map[string]*VM `json:"vms"` // clé = family
	InUseBy     string         `json:"in_use_by,omitempty"`
	CheckedOutAt *time.Time    `json:"checked_out_at,omitempty"`
}

// Store est le store en mémoire thread-safe des groupes/VMs.
type Store struct {
	mu     sync.RWMutex
	groups map[string]*Group
}

// New crée un store vide.
func New() *Store {
	return &Store{
		groups: make(map[string]*Group),
	}
}

// ReplaceGroups remplace entièrement les groupes issus d'une collecte complète,
// en conservant le statut checkout/checkin existant lorsque le groupe persiste.
func (s *Store) ReplaceGroups(newGroups map[string]*Group) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, ng := range newGroups {
		if old, ok := s.groups[id]; ok {
			ng.InUseBy = old.InUseBy
			ng.CheckedOutAt = old.CheckedOutAt
		}
	}
	s.groups = newGroups
}

// ListGroups retourne une copie de la liste des groupes.
func (s *Store) ListGroups() []*Group {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*Group, 0, len(s.groups))
	for _, g := range s.groups {
		out = append(out, g)
	}
	return out
}

// Checkout marque un groupe comme utilisé par `name`.
func (s *Store) Checkout(groupID, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.groups[groupID]
	if !ok {
		return false
	}
	now := time.Now()
	g.InUseBy = name
	g.CheckedOutAt = &now
	return true
}

// Checkin libère un groupe.
func (s *Store) Checkin(groupID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.groups[groupID]
	if !ok {
		return false
	}
	g.InUseBy = ""
	g.CheckedOutAt = nil
	return true
}

// UpsertDiscoveredVM fusionne les informations issues de la découverte
// automatique (hyperviseur : IP, hostname, type) avec les VMs déjà connues
// du store (issues de la collecte SSH : versions d'applis, appartenance à un
// groupe). Si la VM (identifiée par hostname) n'existe dans aucun groupe,
// elle est rattachée à un groupe "orphelin" nommé d'après son hostname, en
// attendant qu'une collecte SSH complète reconstruise les groupes réels via
// /etc/hosts.
func (s *Store) UpsertDiscoveredVM(hostname, ip, hypervisor, family string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Cherche si cette VM (par hostname) est déjà rattachée à un groupe existant.
	for _, g := range s.groups {
		if vm, ok := g.VMs[family]; ok && vm.Hostname == hostname {
			vm.IP = ip
			vm.Hypervisor = hypervisor
			vm.LastSeen = now
			return
		}
	}

	// Sinon, upsert dans un groupe orphelin dédié à cette VM, en attendant la
	// prochaine collecte SSH qui reconstruira le vrai groupe via /etc/hosts.
	orphanID := "discovered-" + hostname
	g, ok := s.groups[orphanID]
	if !ok {
		g = &Group{
			ID:  orphanID,
			VMs: make(map[string]*VM),
		}
		s.groups[orphanID] = g
	}

	vm, ok := g.VMs[family]
	if !ok {
		vm = &VM{Family: family}
		g.VMs[family] = vm
	}
	vm.Hostname = hostname
	vm.IP = ip
	vm.Hypervisor = hypervisor
	vm.LastSeen = now
}
