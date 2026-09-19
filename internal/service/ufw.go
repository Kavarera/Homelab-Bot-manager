package service

import (
	"context"
	"fmt"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/executor"
	"time"
)

// FirewallService defines operations for managing firewall rules.
type FirewallService interface {
	TogglePort53(ctx context.Context, action domain.UFWAction) error
}

type firewallService struct {
	remoteExec executor.RemoteExecutor
	timeout    time.Duration
}

// NewFirewallService creates a new FirewallService instance.
func NewFirewallService(remoteExec executor.RemoteExecutor) FirewallService {
	return &firewallService{
		remoteExec: remoteExec,
		timeout:    15 * time.Second,
	}
}

// TogglePort53 allows or denies TCP/UDP port 53 via UFW on the remote machine.
func (s *firewallService) TogglePort53(ctx context.Context, action domain.UFWAction) error {
	var rule string
	switch action {
	case domain.UFWActionAllow:
		rule = "allow"
	case domain.UFWActionDeny:
		rule = "deny"
	default:
		return fmt.Errorf("unsupported ufw action: %s", action)
	}

	remoteCommand := fmt.Sprintf("sudo ufw %s 53/tcp && sudo ufw %s 53/udp", rule, rule)

	execCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := s.remoteExec.ExecuteRemote(execCtx, remoteCommand)
	if err != nil {
		return fmt.Errorf("failed to execute remote ufw command: %w", err)
	}

	return nil
}
