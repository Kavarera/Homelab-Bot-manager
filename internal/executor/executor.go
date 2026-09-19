package executor

import "context"

// CommandExecutor is an abstraction for executing commands (locally or remotely).
type CommandExecutor interface {
	// LookPath checks if the given executable is available in PATH.
	LookPath(file string) (string, error)
	// Execute runs a command with context and returns combined stdout/stderr output.
	Execute(ctx context.Context, name string, args ...string) ([]byte, error)
}
