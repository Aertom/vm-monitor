package store

import (
	"errors"
	"sync"
	"time"
)

// Erreurs sentinelles utilisées par la couche API.
var (
	ErrGroupNotFound     = errors.New("group not found")
	ErrAlreadyCheckedOut = errors.New("group already checked out")
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
	ID           string         `json:"id"`
	VMs          map[string]*VM `json:"vms"` // clé = family
	InUseBy      string         `json:"in_use_by,omitempty"`
	CheckedOutAt *time.Time     `json:"checked_out_at,omitempty"`
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

// GetGroups retourne l'ensemble des groupes connus (copie de la liste, pointeurs partagés).
func (s *Store) GetGroups() []*Group {
	s.mu.RLock()
	defer s.mu.RUnlock()

	groups := make([]*Group, 0, len(s.groups))
	for _, g := range s.groups {
		groups = append(groups, g)
	}
	return groups
}

// Checkout marque un groupe comme utilisé par un utilisateur donné.
// Retourne ErrGroupNotFound si le groupe n'existe pas, ErrAlreadyCheckedOut
// s'il est déjà réservé par quelqu'un d'autre.
func (s *Store) Checkout(groupID, user string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.groups[groupID]
	if !ok {
		return ErrGroupNotFound
	}
	if g.InUseBy != "" && g.InUseBy != user {
		return ErrAlreadyCheckedOut
	}

	now := time.Now()
	g.InUseBy = user
	g.CheckedOutAt = &now
	return nil
}

// Checkin libère un groupe précédemment réservé.
// Retourne ErrGroupNotFound si le groupe n'existe pas.
func (s *Store) Checkin(groupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.groups[groupID]
	if !ok {
		return ErrGroupNotFound
	}

	g.InUseBy = ""
	g.CheckedOutAt = nil
	return nil
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

// UpsertDiscoveredVM insère ou met à jour une VM découverte via l'API d'un hyperviseur,
// en l'attachant à un groupe existant si son hostname y est déjà présent, ou à un
// groupe orphelin dédié dans l'attente de la prochaine collecte SSH complète.
func (s *Store) UpsertDiscoveredVM(hostname, ip, hypervisor string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	family := "unknown"
	now := time.Now()

	// Cherche si cette VM (par hostname) est déjà rattachée à un groupe existant.
	for _, g := range s.groups {
		for fam, vm := range g.VMs {
			if vm.Hostname == hostname {
				vm.IP = ip
				vm.Hypervisor = hypervisor
				vm.LastSeen = now
				_ = fam
				return
			}
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
