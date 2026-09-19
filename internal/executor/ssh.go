package executor

import (
	"context"
	"fmt"
)

// SSHConfig holds connection options for SSH commands.
type SSHConfig struct {
	Host           string
	PrivateKeyPath string
}

// RemoteExecutor defines an interface for running commands on a remote host.
type RemoteExecutor interface {
	ExecuteRemote(ctx context.Context, remoteCommand string) ([]byte, error)
}

// SSHExecutor implements RemoteExecutor using the ssh CLI through a CommandExecutor.
type SSHExecutor struct {
	executor CommandExecutor
	cfg      SSHConfig
}

// NewSSHExecutor creates a new SSHExecutor.
func NewSSHExecutor(executor CommandExecutor, cfg SSHConfig) *SSHExecutor {
	return &SSHExecutor{
		executor: executor,
		cfg:      cfg,
	}
}

// ExecuteRemote executes a shell command on the remote host via SSH.
func (s *SSHExecutor) ExecuteRemote(ctx context.Context, remoteCommand string) ([]byte, error) {
	args := []string{
		"-i", s.cfg.PrivateKeyPath,
		"-o", "StrictHostKeyChecking=no",
		s.cfg.Host,
		remoteCommand,
	}

	out, err := s.executor.Execute(ctx, "ssh", args...)
	if err != nil {
		return out, fmt.Errorf("remote ssh execution failed on %s: %s (error: %w)", s.cfg.Host, string(out), err)
	}

	return out, nil
}
