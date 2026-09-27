// Package cmd holds CLI orchestration: runInit/runList/runDoctor/runTag.
// main.go only parses flags and calls these helpers.
package cmd

import "fmt"

// Exit codes: 0 ok, 1 generic/push failed, 2 already-exists, 3 bad args.
const (
	CodeOK      = 0
	CodeGeneric = 1
	CodeExists  = 2
	CodeBadArgs = 3
)

// ExitError carries a process exit code with a message.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }

// BadArgs marks flag/config usage errors (exit 3).
func BadArgs(format string, args ...any) *ExitError {
	return &ExitError{Code: CodeBadArgs, Err: fmt.Errorf(format, args...)}
}

// Exists marks already-exists / no-op errors (exit 2).
func Exists(format string, args ...any) *ExitError {
	return &ExitError{Code: CodeExists, Err: fmt.Errorf(format, args...)}
}

// Generic marks push failures and other runtime errors (exit 1).
func Generic(format string, args ...any) *ExitError {
	return &ExitError{Code: CodeGeneric, Err: fmt.Errorf(format, args...)}
}

// GenericErr wraps an existing error as exit 1.
func GenericErr(err error) *ExitError {
	return &ExitError{Code: CodeGeneric, Err: err}
}

// CodeOf extracts the exit code from an error (default 1).
func CodeOf(err error) int {
	if err == nil {
		return CodeOK
	}
	if ee, ok := err.(*ExitError); ok {
		return ee.Code
	}
	return CodeGeneric
}
