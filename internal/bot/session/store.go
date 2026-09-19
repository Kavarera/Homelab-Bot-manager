package session

import "context"

// Store defines the interface for storing and retrieving user sessions.
type Store interface {
	// Get retrieves the active session for a given user ID.
	Get(ctx context.Context, userID int64) (*Session, error)
	// Save saves or updates the session.
	Save(ctx context.Context, session *Session) error
	// Delete removes the session for a given user ID.
	Delete(ctx context.Context, userID int64) error
	// CleanupExpired removes all expired sessions and returns the count of deleted sessions.
	CleanupExpired(ctx context.Context) int
}
