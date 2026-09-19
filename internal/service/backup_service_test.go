package service_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"hs1-bot/internal/service"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockRemoteExecutor struct {
	executeRemoteFunc func(ctx context.Context, cmd string) ([]byte, error)
}

func (m *mockRemoteExecutor) ExecuteRemote(ctx context.Context, cmd string) ([]byte, error) {
	if m.executeRemoteFunc != nil {
		return m.executeRemoteFunc(ctx, cmd)
	}
	return nil, nil
}

func TestBackupService_ListPostgresContainers(t *testing.T) {
	dockerPsOutput := `fdb5ca78937d|payrollpro_web|ghcr.io/kavarera/payrollpro-web:latest|Up 2 weeks
e181b18c3814|payroll_worker_default|ghcr.io/kavarera/payrollpro-app:latest|Up 2 weeks
a71393fc8159|postgres-db|postgres:latest|Up 4 months
b92139fc1111|custom-db|bitnami/postgresql:15|Up 1 month
711d62f487ad|porto-static|image-porto-static|Up 3 months
`

	mock := &mockRemoteExecutor{
		executeRemoteFunc: func(ctx context.Context, cmd string) ([]byte, error) {
			if strings.Contains(cmd, "docker ps") {
				return []byte(dockerPsOutput), nil
			}
			return nil, fmt.Errorf("unexpected command: %s", cmd)
		},
	}

	backupSvc := service.NewBackupService(mock)
	containers, err := backupSvc.ListPostgresContainers(context.Background())
	if err != nil {
		t.Fatalf("failed to list postgres containers: %v", err)
	}

	if len(containers) != 2 {
		t.Fatalf("expected 2 postgres containers, got %d", len(containers))
	}

	if containers[0].Name != "postgres-db" || containers[0].Image != "postgres:latest" {
		t.Errorf("unexpected first container: %+v", containers[0])
	}
	if containers[1].Name != "custom-db" || containers[1].Image != "bitnami/postgresql:15" {
		t.Errorf("unexpected second container: %+v", containers[1])
	}
}

func TestBackupService_DumpAndEncrypt(t *testing.T) {
	fakeSQLDump := []byte("-- PostgreSQL database dump\nCREATE TABLE employees (id INT, name TEXT);\nINSERT INTO employees VALUES (1, 'Alice');\n")

	mock := &mockRemoteExecutor{
		executeRemoteFunc: func(ctx context.Context, cmd string) ([]byte, error) {
			if strings.Contains(cmd, "pg_dumpall") {
				return fakeSQLDump, nil
			}
			return nil, fmt.Errorf("unexpected command: %s", cmd)
		},
	}

	tempDir := t.TempDir()
	backupSvc := service.NewBackupService(mock)

	result, err := backupSvc.DumpAndEncrypt(context.Background(), "postgres-db", tempDir)
	if err != nil {
		t.Fatalf("failed to dump and encrypt: %v", err)
	}

	if result.ContainerName != "postgres-db" {
		t.Errorf("expected container name 'postgres-db', got %s", result.ContainerName)
	}
	if result.KeyHex == "" || len(result.KeyHex) != 64 {
		t.Errorf("expected 64-character hex key, got %q", result.KeyHex)
	}
	if result.RawSizeBytes != int64(len(fakeSQLDump)) {
		t.Errorf("expected raw size %d, got %d", len(fakeSQLDump), result.RawSizeBytes)
	}

	// Verify local file exists on disk
	if _, err := os.Stat(result.LocalPath); err != nil {
		t.Fatalf("local backup file not found at %s: %v", result.LocalPath, err)
	}

	// Verify directory structure: tempDir/postgres-db/YYYY/MM/DD/
	if !strings.HasPrefix(result.LocalPath, filepath.Join(tempDir, "postgres-db")) {
		t.Errorf("unexpected local file path structure: %s", result.LocalPath)
	}

	// Read file from disk, decrypt and decompress to verify integrity
	fileBytes, err := os.ReadFile(result.LocalPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	keyBytes, err := hex.DecodeString(result.KeyHex)
	if err != nil {
		t.Fatalf("failed to decode hex key: %v", err)
	}

	decryptedCompressed, err := service.DecryptAESGCM(fileBytes, keyBytes)
	if err != nil {
		t.Fatalf("failed to decrypt saved file: %v", err)
	}

	decompressedSQL, err := service.DecompressGzip(decryptedCompressed)
	if err != nil {
		t.Fatalf("failed to decompress decrypted file: %v", err)
	}

	if string(decompressedSQL) != string(fakeSQLDump) {
		t.Errorf("expected recovered SQL %q, got %q", string(fakeSQLDump), string(decompressedSQL))
	}
}
