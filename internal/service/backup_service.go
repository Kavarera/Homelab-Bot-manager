package service

import (
	"context"
	"errors"
	"fmt"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/executor"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var containerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)

// BackupService defines operations for inspecting remote Docker containers and dumping databases.
type BackupService interface {
	ListPostgresContainers(ctx context.Context) ([]domain.PostgresContainer, error)
	DumpAndEncrypt(ctx context.Context, containerName string, basePath string) (*domain.BackupResult, error)
}

type backupService struct {
	remoteExec executor.RemoteExecutor
}

// NewBackupService creates a new BackupService with the provided RemoteExecutor.
func NewBackupService(remoteExec executor.RemoteExecutor) BackupService {
	return &backupService{
		remoteExec: remoteExec,
	}
}

// ListPostgresContainers queries running Docker containers on the remote host via SSH and filters for PostgreSQL images.
func (s *backupService) ListPostgresContainers(ctx context.Context) ([]domain.PostgresContainer, error) {
	cmd := `sudo docker ps --format "{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}" --no-trunc 2>/dev/null || docker ps --format "{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}" --no-trunc`
	out, err := s.remoteExec.ExecuteRemote(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to query docker ps on remote server: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	var containers []domain.PostgresContainer

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}

		id := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		image := strings.TrimSpace(parts[2])
		status := strings.TrimSpace(parts[3])

		// Clean leading slash if any
		name = strings.TrimPrefix(name, "/")

		// Filter for postgres image
		imageLower := strings.ToLower(image)
		if strings.Contains(imageLower, "postgres") || strings.Contains(imageLower, "postgresql") {
			containers = append(containers, domain.PostgresContainer{
				ID:     id,
				Name:   name,
				Image:  image,
				Status: status,
			})
		}
	}

	return containers, nil
}

// DumpAndEncrypt streams pg_dumpall from the specified container, compresses with Gzip,
// encrypts with AES-256-GCM, and saves a local copy on the homelab filesystem.
func (s *backupService) DumpAndEncrypt(ctx context.Context, containerName string, basePath string) (*domain.BackupResult, error) {
	cleanContainer := strings.TrimSpace(containerName)
	if !containerNameRegex.MatchString(cleanContainer) {
		return nil, fmt.Errorf("invalid container name: %q", containerName)
	}

	if basePath == "" {
		basePath = "data/backup_db"
	}

	// 1. Stream dump via SSH without creating temporary files on VPS
	cmd := fmt.Sprintf(`sudo docker exec %s pg_dumpall -U postgres 2>/dev/null || docker exec %s pg_dumpall -U postgres`, cleanContainer, cleanContainer)
	rawDump, err := s.remoteExec.ExecuteRemote(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute pg_dumpall on container %s: %w", cleanContainer, err)
	}

	if len(rawDump) == 0 {
		return nil, errors.New("pg_dumpall produced empty output")
	}

	rawSize := int64(len(rawDump))

	// 2. Compress with Gzip
	compressed, err := CompressGzip(rawDump)
	if err != nil {
		return nil, fmt.Errorf("failed to compress database dump: %w", err)
	}

	// 3. Generate random 256-bit encryption key
	rawKey, keyHex, err := GenerateAES256Key()
	if err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// 4. Encrypt with AES-256-GCM
	encryptedData, err := EncryptAESGCM(compressed, rawKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt compressed backup: %w", err)
	}

	encryptedSize := int64(len(encryptedData))
	now := time.Now()

	// 5. Structure storage folder: {basePath}/{container}/{YYYY}/{MM}/{DD}/
	filename := fmt.Sprintf(
		"backup-%s-%04d%02d%02d_%02d%02d%02d.sql.gz.enc",
		cleanContainer,
		now.Year(), int(now.Month()), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
	)

	targetDir := filepath.Join(
		basePath,
		cleanContainer,
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", int(now.Month())),
		fmt.Sprintf("%02d", now.Day()),
	)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local backup directory %s: %w", targetDir, err)
	}

	fullLocalPath := filepath.Join(targetDir, filename)
	if err := os.WriteFile(fullLocalPath, encryptedData, 0600); err != nil {
		return nil, fmt.Errorf("failed to save encrypted backup file to %s: %w", fullLocalPath, err)
	}

	return &domain.BackupResult{
		ContainerName:      cleanContainer,
		Filename:           filename,
		LocalPath:          fullLocalPath,
		EncryptedData:      encryptedData,
		KeyHex:             keyHex,
		RawSizeBytes:       rawSize,
		EncryptedSizeBytes: encryptedSize,
		CreatedAt:          now,
	}, nil
}
