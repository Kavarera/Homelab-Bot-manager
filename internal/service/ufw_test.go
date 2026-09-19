package service

import (
	"context"
	"errors"
	"hs1-bot/internal/domain"
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

func TestFirewallService_TogglePort53_Allow(t *testing.T) {
	var executedCmd string
	mockExec := &mockRemoteExecutor{
		executeRemoteFunc: func(ctx context.Context, cmd string) ([]byte, error) {
			executedCmd = cmd
			return []byte("Rules updated"), nil
		},
	}

	svc := NewFirewallService(mockExec)
	err := svc.TogglePort53(context.Background(), domain.UFWActionAllow)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(executedCmd, "allow 53/tcp") || !strings.Contains(executedCmd, "allow 53/udp") {
		t.Errorf("command does not contain allow rules: %s", executedCmd)
	}
}

func TestFirewallService_TogglePort53_Deny(t *testing.T) {
	var executedCmd string
	mockExec := &mockRemoteExecutor{
		executeRemoteFunc: func(ctx context.Context, cmd string) ([]byte, error) {
			executedCmd = cmd
			return []byte("Rules updated"), nil
		},
	}

	svc := NewFirewallService(mockExec)
	err := svc.TogglePort53(context.Background(), domain.UFWActionDeny)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(executedCmd, "deny 53/tcp") || !strings.Contains(executedCmd, "deny 53/udp") {
		t.Errorf("command does not contain deny rules: %s", executedCmd)
	}
}

func TestFirewallService_TogglePort53_Error(t *testing.T) {
	mockExec := &mockRemoteExecutor{
		executeRemoteFunc: func(ctx context.Context, cmd string) ([]byte, error) {
			return nil, errors.New("connection timed out")
		},
	}

	svc := NewFirewallService(mockExec)
	err := svc.TogglePort53(context.Background(), domain.UFWActionAllow)
	if err == nil {
		t.Error("expected error from remote executor, got nil")
	}
}
