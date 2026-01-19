package expectations

import (
	"fmt"
	"sync"

	"andboson/mock-server/internal/models"
)

// Store holds the expectations and request history in memory.
// It is safe for concurrent use.
type Store struct {
	expectations []models.Expectation
	history      []models.HistoryItem
	mu           sync.RWMutex
}

// NewStore creates a new empty Store.
func NewStore() *Store {
	return &Store{
		expectations: make([]models.Expectation, 0),
		history:      make([]models.HistoryItem, 0),
	}
}

// AddExpectation adds a new expectation to the store.
// It compiles the expectation's regexes before adding.
func (s *Store) AddExpectation(e models.Expectation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := e.Compile(); err != nil {
		return fmt.Errorf("failed to compile regexp: %w", err)
	}

	s.expectations = append(s.expectations, e)

	return nil
}

// AddHistory adds a recorded request to the history.
func (s *Store) AddHistory(item models.HistoryItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = append(s.history, item)
}

// FindMatch searches for an expectation that matches the given method, path, and body.
// It returns the matcher expectation and true if found, otherwise an empty expectation and false.
func (s *Store) FindMatch(method, path, body string) (models.Expectation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, e := range s.expectations {
		if e.Match(method, path, body) {
			return e, true
		}
	}

	return models.Expectation{}, false
}

// GetHistory returns requests history (in reverse order)
func (s *Store) GetHistory(reverse bool) []models.HistoryItem {
	if !reverse {
		return s.history
	}

	reversed := append([]models.HistoryItem{}, s.history...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}

	return reversed
}

// DumpAvailableExpectations return available expectations
func (s *Store) DumpAvailableExpectations() []models.Expectation {
	return append([]models.Expectation{}, s.expectations...)
}

// AddExpectations adds multiple expectations to the store.
func (s *Store) AddExpectations(expectations []models.Expectation) error {
	for i, expectation := range expectations {
		if err := s.AddExpectation(expectation); err != nil {
			return fmt.Errorf("failed to add expectation at index %d: %w", i, err)
		}
	}

	return nil
}
