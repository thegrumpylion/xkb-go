package xkb

import (
	"errors"
	"testing"
)

func TestErrorError(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "basic error",
			err:  &Error{Op: "Parse", Err: ErrInvalidSyntax},
			want: "Parse: invalid syntax",
		},
		{
			name: "with path",
			err:  &Error{Op: "Load", Path: "/foo/bar.xkb", Err: ErrFileNotFound},
			want: "Load: /foo/bar.xkb: file not found",
		},
		{
			name: "with line",
			err:  &Error{Op: "Parse", Path: "test.xkb", Line: 42, Err: ErrInvalidSyntax},
			want: "Parse: test.xkb:42: invalid syntax",
		},
		{
			name: "with line and column",
			err:  &Error{Op: "Parse", Path: "test.xkb", Line: 42, Col: 10, Err: ErrInvalidSyntax},
			want: "Parse: test.xkb:42:10: invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	inner := ErrInvalidKeymap
	err := &Error{Op: "Test", Err: inner}

	if err.Unwrap() != inner {
		t.Error("Unwrap() should return inner error")
	}
}

func TestErrorIs(t *testing.T) {
	err := &Error{Op: "Test", Err: ErrInvalidSyntax}

	if !errors.Is(err, ErrInvalidSyntax) {
		t.Error("errors.Is should match inner error")
	}
	if errors.Is(err, ErrFileNotFound) {
		t.Error("errors.Is should not match different error")
	}
}

func TestSyntaxErrorError(t *testing.T) {
	tests := []struct {
		name string
		err  *SyntaxError
		want string
	}{
		{
			name: "with file",
			err:  &SyntaxError{File: "test.xkb", Line: 10, Col: 5, Message: "unexpected token"},
			want: "test.xkb:10:5: unexpected token",
		},
		{
			name: "without file",
			err:  &SyntaxError{Line: 10, Col: 5, Message: "unexpected token"},
			want: "10:5: unexpected token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSyntaxErrorIs(t *testing.T) {
	err := &SyntaxError{Line: 1, Col: 1, Message: "test"}

	if !errors.Is(err, ErrInvalidSyntax) {
		t.Error("SyntaxError should match ErrInvalidSyntax")
	}
	if errors.Is(err, ErrFileNotFound) {
		t.Error("SyntaxError should not match ErrFileNotFound")
	}
}

func TestSentinelErrors(t *testing.T) {
	// Just verify they exist and have messages
	sentinels := []error{
		ErrNotImplemented,
		ErrUnsupportedFormat,
		ErrInvalidKeymap,
		ErrInvalidSyntax,
		ErrFileNotFound,
		ErrInvalidKeycode,
		ErrInvalidKeysym,
	}

	for _, err := range sentinels {
		if err.Error() == "" {
			t.Errorf("Sentinel error has empty message: %v", err)
		}
	}
}
