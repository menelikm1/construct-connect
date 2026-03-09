package telegram

import (
	"sync"

	"github.com/google/uuid"
)

// Session holds per-chat state — mainly the last browse results so users
// can refer to listings by number (e.g. /listing 2) instead of typing UUIDs.
type Session struct {
	LastListings []uuid.UUID // index 0 = listing "1" in bot messages
}

type SessionStore struct {
	mu   sync.RWMutex
	data map[int64]*Session
}

func newSessionStore() *SessionStore {
	return &SessionStore{data: make(map[int64]*Session)}
}

func (s *SessionStore) get(chatID int64) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if sess, ok := s.data[chatID]; ok {
		return sess
	}
	return &Session{}
}

func (s *SessionStore) setListings(chatID int64, ids []uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[chatID]; !ok {
		s.data[chatID] = &Session{}
	}
	s.data[chatID].LastListings = ids
}
