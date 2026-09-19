package service

import (
	"context"
	"fmt"
	"hs1-bot/internal/executor"
	"time"
)

// SystemService provides methods to inspect system status.
type SystemService interface {
	GetSystemStatus(ctx context.Context) (string, error)
}

type systemService struct {
	exec    executor.CommandExecutor
	timeout time.Duration
}

// NewSystemService creates a new SystemService instance.
func NewSystemService(exec executor.CommandExecutor) SystemService {
	return &systemService{
		exec:    exec,
		timeout: 5 * time.Second,
	}
}

// GetSystemStatus retrieves the uptime and basic status of the host server.
func (s *systemService) GetSystemStatus(ctx context.Context) (string, error) {
	execCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	out, err := s.exec.Execute(execCtx, "uptime")
	if err != nil {
		return "", fmt.Errorf("failed to get system uptime: %s (error: %w)", string(out), err)
	}

	return string(out), nil
}
