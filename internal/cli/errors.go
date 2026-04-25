package cli

import "fmt"

// ExitError lets a subcommand request a specific process exit code from
// main.go. Validation failures use code 2 so CI can distinguish them
// from generic errors (code 1).
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }

func exitErr(code int, format string, a ...any) error {
	return &ExitError{Code: code, Err: fmt.Errorf(format, a...)}
}
