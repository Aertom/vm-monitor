package store

import (
	"errors"
	"sync"
	"time"

	"github.com/Aertom/vm-monitor/backend/internal/models"
)

var ErrAlreadyCheckedOut = errors.New("group already checked out by another user")
var ErrGroupNotFound = errors.New("group not found")

// Store est un stockage en mémoire, thread-safe, des VMs/groupes et du
// statut de checkout. À remplacer plus tard par une persistance réelle
// si besoin (SQLite/Postgres) sans changer l'interface publique.
type Store struct {
	mu     sync.RWMutex
	vms    []models.VM
	groups []models.Group
}

func New() *Store {
	return &Store{}
}

func (s *Store) ReplaceAll(vms []models.VM, groups []models.Group) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// on conserve le statut de checkout existant lors du rafraîchissement
	prevCheckout := make(map[string]models.Group)
	for _, g := range s.groups {
		if g.InUseBy != nil {
			prevCheckout[g.ID] = g
		}
	}

	for i := range groups {
		if prev, ok := prevCheckout[groups[i].ID]; ok {
			groups[i].InUseBy = prev.InUseBy
			groups[i].CheckedAt = prev.CheckedAt
		}
	}

	s.vms = vms
	s.groups = groups
}

func (s *Store) GetGroups() []models.Group {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Group, len(s.groups))
	copy(out, s.groups)
	return out
}

func (s *Store) Checkout(groupID, user string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.groups {
		if s.groups[i].ID == groupID {
			if s.groups[i].InUseBy != nil && *s.groups[i].InUseBy != user {
				return ErrAlreadyCheckedOut
			}
			now := time.Now()
			s.groups[i].InUseBy = &user
			s.groups[i].CheckedAt = &now
			return nil
		}
	}
	return ErrGroupNotFound
}

func (s *Store) Checkin(groupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.groups {
		if s.groups[i].ID == groupID {
			s.groups[i].InUseBy = nil
			s.groups[i].CheckedAt = nil
			return nil
		}
	}
	return ErrGroupNotFound
}
