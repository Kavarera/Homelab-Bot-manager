package session

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStore_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore(1 * time.Hour)

	sess := NewSession(12345, "test_flow", "step_1", 1*time.Hour)
	sess.Set("key1", "value1")

	err := store.Save(ctx, sess)
	if err != nil {
		t.Fatalf("expected no error saving session, got: %v", err)
	}

	retrieved, err := store.Get(ctx, 12345)
	if err != nil {
		t.Fatalf("expected no error getting session, got: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected session to exist, got nil")
	}

	if retrieved.FlowID != "test_flow" || retrieved.CurrentStep != "step_1" {
		t.Errorf("unexpected session data: %+v", retrieved)
	}

	val, ok := retrieved.GetString("key1")
	if !ok || val != "value1" {
		t.Errorf("expected key1='value1', got '%s'", val)
	}
}

func TestMemoryStore_Expiration(t *testing.T) {
	ctx := context.Background()
	// TTL very short: 10 milliseconds
	store := NewMemoryStore(10 * time.Millisecond)

	sess := NewSession(999, "quick_flow", "init", 10*time.Millisecond)
	_ = store.Save(ctx, sess)

	time.Sleep(25 * time.Millisecond)

	retrieved, err := store.Get(ctx, 999)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if retrieved != nil {
		t.Error("expected expired session to return nil on Get")
	}
}

func TestMemoryStore_CleanupExpired(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore(10 * time.Millisecond)

	sess1 := NewSession(1, "f1", "s1", 10*time.Millisecond)
	sess2 := NewSession(2, "f2", "s2", 1*time.Hour)

	_ = store.Save(ctx, sess1)
	_ = store.Save(ctx, sess2)

	time.Sleep(20 * time.Millisecond)

	cleaned := store.CleanupExpired(ctx)
	if cleaned != 1 {
		t.Errorf("expected 1 session cleaned, got %d", cleaned)
	}

	s2, _ := store.Get(ctx, 2)
	if s2 == nil {
		t.Error("expected active session 2 to still exist")
	}
}
