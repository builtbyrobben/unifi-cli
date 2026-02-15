package cmd

import (
	"context"
	"fmt"
	"runtime"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

// VersionCmd prints version information.
type VersionCmd struct{}

func (cmd *VersionCmd) Run(_ context.Context) error {
	fmt.Printf("unifi-cli %s\n", VersionString())
	fmt.Printf("  Commit: %s\n", commit)
	fmt.Printf("  Built:  %s\n", date)
	fmt.Printf("  OS:     %s/%s\n", runtime.GOOS, runtime.GOARCH)

	return nil
}

// VersionString returns the version string.
func VersionString() string {
	if version == "dev" {
		return "dev (no version)"
	}

	return version
}

// ExitError wraps an error with an exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e == nil {
		return ""
	}

	if e.Err != nil {
		return e.Err.Error()
	}

	return fmt.Sprintf("exit code %d", e.Code)
}

// Unwrap returns the underlying error.
func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}
