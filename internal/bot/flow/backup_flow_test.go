package flow_test

import (
	"context"
	"fmt"
	"hs1-bot/internal/bot/flow"
	"hs1-bot/internal/bot/session"
	"hs1-bot/internal/bot/ui"
	"hs1-bot/internal/domain"
	"testing"
	"time"
)

type mockBackupService struct {
	listCalled bool
	dumpCalled bool
	targetName string
}

func (m *mockBackupService) ListPostgresContainers(ctx context.Context) ([]domain.PostgresContainer, error) {
	m.listCalled = true
	return []domain.PostgresContainer{
		{ID: "a71393fc8159", Name: "postgres-db", Image: "postgres:latest", Status: "Up 4 months"},
		{ID: "b92139fc1111", Name: "app-postgres", Image: "bitnami/postgresql:15", Status: "Up 2 weeks"},
	}, nil
}

func (m *mockBackupService) DumpAndEncrypt(ctx context.Context, containerName string, basePath string) (*domain.BackupResult, error) {
	m.dumpCalled = true
	m.targetName = containerName
	return &domain.BackupResult{
		ContainerName:      containerName,
		Filename:           fmt.Sprintf("backup-%s-test.sql.gz.enc", containerName),
		LocalPath:          fmt.Sprintf("/home/kava/backup_db/%s/2026/09/20/backup-%s-test.sql.gz.enc", containerName, containerName),
		EncryptedData:      []byte("MOCK_ENCRYPTED_BYTES"),
		KeyHex:             "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		RawSizeBytes:       102400,
		EncryptedSizeBytes: 51200,
		CreatedAt:          time.Now(),
	}, nil
}

func TestBackupFlow_FullCycle(t *testing.T) {
	mockSvc := &mockBackupService{}
	store := session.NewMemoryStore(1 * time.Hour)
	engine := flow.NewEngine(store, 1*time.Hour)
	engine.Register(flow.NewBackupFlow(mockSvc, "/home/kava/backup_db"))

	const userID int64 = 77777

	// 1. Start Flow
	startCtx := createTestBotContext(userID, ui.ButtonBackupDB)
	err := engine.StartFlow(startCtx, flow.BackupFlowID)
	if err != nil {
		t.Fatalf("failed to start backup flow: %v", err)
	}
	if !mockSvc.listCalled {
		t.Error("expected ListPostgresContainers to be called")
	}

	// 2. Select Container "postgres-db"
	selectCtx := createTestBotContext(userID, "postgres-db")
	handled, err := engine.HandleActiveFlow(selectCtx)
	if err != nil || !handled {
		t.Fatalf("expected container selection to be handled: %v", err)
	}

	// 3. Confirm Backup
	confirmCtx := createTestBotContext(userID, ui.ButtonConfirm)
	handled, err = engine.HandleActiveFlow(confirmCtx)
	if err != nil || !handled {
		t.Fatalf("expected confirmation to be handled: %v", err)
	}

	if !mockSvc.dumpCalled {
		t.Error("expected DumpAndEncrypt to be called")
	}
	if mockSvc.targetName != "postgres-db" {
		t.Errorf("expected target name 'postgres-db', got %s", mockSvc.targetName)
	}
}
