package session

import (
	"context"
	"sync"
	"time"
)

// MemoryStore is an in-memory, thread-safe implementation of Store.
type MemoryStore struct {
	mu         sync.RWMutex
	sessions   map[int64]*Session
	defaultTTL time.Duration
}

// NewMemoryStore creates a new MemoryStore with the specified default TTL (e.g. 1 hour).
func NewMemoryStore(defaultTTL time.Duration) *MemoryStore {
	if defaultTTL <= 0 {
		defaultTTL = 1 * time.Hour
	}
	return &MemoryStore{
		sessions:   make(map[int64]*Session),
		defaultTTL: defaultTTL,
	}
}

// Get retrieves the active session for a given user ID. If expired, it is purged.
func (m *MemoryStore) Get(ctx context.Context, userID int64) (*Session, error) {
	m.mu.RLock()
	sess, exists := m.sessions[userID]
	m.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	if sess.IsExpired() {
		_ = m.Delete(ctx, userID)
		return nil, nil
	}

	return sess, nil
}

// Save saves or updates a session in memory.
func (m *MemoryStore) Save(ctx context.Context, session *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.UserID] = session
	return nil
}

// Delete removes a user's session from memory.
func (m *MemoryStore) Delete(ctx context.Context, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, userID)
	return nil
}

// CleanupExpired removes all sessions that have passed their expiration timestamp.
func (m *MemoryStore) CleanupExpired(ctx context.Context) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	cleaned := 0

	for userID, sess := range m.sessions {
		if now.After(sess.ExpiresAt) {
			delete(m.sessions, userID)
			cleaned++
		}
	}

	return cleaned
}

// StartAutoCleanup starts a background worker that periodically purges expired sessions.
func (m *MemoryStore) StartAutoCleanup(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.CleanupExpired(ctx)
			}
		}
	}()
}
