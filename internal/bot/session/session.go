package session

import (
	"sync"
	"time"
)

// Session represents a stateful user conversation session.
type Session struct {
	UserID      int64          `json:"user_id"`
	FlowID      string         `json:"flow_id"`
	CurrentStep string         `json:"current_step"`
	Data        map[string]any `json:"data"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	mu          sync.RWMutex
}

// NewSession creates a new Session with the specified TTL.
func NewSession(userID int64, flowID string, initialStep string, ttl time.Duration) *Session {
	now := time.Now()
	return &Session{
		UserID:      userID,
		FlowID:      flowID,
		CurrentStep: initialStep,
		Data:        make(map[string]any),
		UpdatedAt:   now,
		ExpiresAt:   now.Add(ttl),
	}
}

// Set stores a key-value pair in the session data.
func (s *Session) Set(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Data == nil {
		s.Data = make(map[string]any)
	}
	s.Data[key] = val
	s.UpdatedAt = time.Now()
}

// Get retrieves a value from the session data.
func (s *Session) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.Data == nil {
		return nil, false
	}
	val, ok := s.Data[key]
	return val, ok
}

// GetString retrieves a string value from session data.
func (s *Session) GetString(key string) (string, bool) {
	val, ok := s.Get(key)
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetInt retrieves an int value from session data.
func (s *Session) GetInt(key string) (int, bool) {
	val, ok := s.Get(key)
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// Touch refreshes the expiration time by the given TTL from now.
func (s *Session) Touch(ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.UpdatedAt = now
	s.ExpiresAt = now.Add(ttl)
}

// IsExpired checks whether the session has expired.
func (s *Session) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Now().After(s.ExpiresAt)
}
