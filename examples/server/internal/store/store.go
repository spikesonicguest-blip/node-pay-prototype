package store

import (
	"errors"
	"sync"
	"nodepay-example-merchant/internal/models"
)

var (
	ErrNotFound = errors.New("not found")
)

type Store struct {
	mu      sync.RWMutex
	charges map[string]models.Charge
}

func New() *Store {
	return &Store{
		charges: make(map[string]models.Charge),
	}
}

func (s *Store) CreateCharge(c models.Charge) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.charges[c.ID] = c
	return nil
}

func (s *Store) GetCharge(id string) (models.Charge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.charges[id]
	if !ok {
		return models.Charge{}, ErrNotFound
	}
	return c, nil
}
