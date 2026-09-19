package executor

import (
	"context"
	"os/exec"
)

// LocalExecutor executes commands locally on the host OS.
type LocalExecutor struct{}

// NewLocalExecutor creates a new instance of LocalExecutor.
func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{}
}

// LookPath searches for an executable in the directories named by the PATH environment variable.
func (e *LocalExecutor) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// Execute runs a command with context on the local machine and returns combined output.
func (e *LocalExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}
