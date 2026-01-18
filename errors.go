package xkb

import (
	"errors"
	"fmt"
)

// Sentinel errors for xkb operations.
var (
	// ErrNotImplemented indicates a feature is not yet implemented.
	ErrNotImplemented = errors.New("not implemented")

	// ErrUnsupportedFormat indicates an unsupported keymap or compose format.
	ErrUnsupportedFormat = errors.New("unsupported format")

	// ErrInvalidKeymap indicates the keymap data is invalid.
	ErrInvalidKeymap = errors.New("invalid keymap")

	// ErrInvalidSyntax indicates a syntax error in the input.
	ErrInvalidSyntax = errors.New("invalid syntax")

	// ErrFileNotFound indicates a required file was not found.
	ErrFileNotFound = errors.New("file not found")

	// ErrInvalidKeycode indicates an invalid keycode.
	ErrInvalidKeycode = errors.New("invalid keycode")

	// ErrInvalidKeysym indicates an invalid keysym.
	ErrInvalidKeysym = errors.New("invalid keysym")
)

// Error wraps an underlying error with operation context.
type Error struct {
	Op   string // Operation that failed (e.g., "NewKeymapFromString")
	Path string // File path, if applicable
	Line int    // Line number, if applicable
	Col  int    // Column number, if applicable
	Err  error  // Underlying error
}

func (e *Error) Error() string {
	if e.Path != "" {
		if e.Line > 0 {
			if e.Col > 0 {
				return fmt.Sprintf("%s: %s:%d:%d: %v", e.Op, e.Path, e.Line, e.Col, e.Err)
			}
			return fmt.Sprintf("%s: %s:%d: %v", e.Op, e.Path, e.Line, e.Err)
		}
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Is implements errors.Is for Error.
func (e *Error) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// SyntaxError provides detailed information about a parsing error.
type SyntaxError struct {
	File    string // Source file name
	Line    int    // Line number (1-based)
	Col     int    // Column number (1-based)
	Message string // Error message
}

func (e *SyntaxError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Col, e.Message)
	}
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Col, e.Message)
}

func (e *SyntaxError) Is(target error) bool {
	return target == ErrInvalidSyntax
}
