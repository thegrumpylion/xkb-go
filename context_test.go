package xkb

import (
	"log/slog"
	"os"
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext(ContextNoFlags)
	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}

	// Should have default include paths (if system directories exist)
	paths := ctx.IncludePaths()
	t.Logf("Default include paths: %v", paths)
}

func TestNewContextNoDefaultIncludes(t *testing.T) {
	ctx := NewContext(ContextNoDefaultIncludes)
	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}

	paths := ctx.IncludePaths()
	if len(paths) != 0 {
		t.Errorf("Expected no include paths with ContextNoDefaultIncludes, got %v", paths)
	}
}

func TestContextIncludePaths(t *testing.T) {
	ctx := NewContext(ContextNoDefaultIncludes)

	// Append paths
	ctx.AppendIncludePath("/path/a")
	ctx.AppendIncludePath("/path/b")

	paths := ctx.IncludePaths()
	if len(paths) != 2 {
		t.Fatalf("Expected 2 paths, got %d", len(paths))
	}
	if paths[0] != "/path/a" || paths[1] != "/path/b" {
		t.Errorf("Unexpected paths: %v", paths)
	}

	// Prepend path
	ctx.PrependIncludePath("/path/first")
	paths = ctx.IncludePaths()
	if len(paths) != 3 {
		t.Fatalf("Expected 3 paths, got %d", len(paths))
	}
	if paths[0] != "/path/first" {
		t.Errorf("Expected /path/first first, got %s", paths[0])
	}

	// Clear paths
	ctx.ClearIncludePaths()
	if ctx.NumIncludePaths() != 0 {
		t.Errorf("Expected 0 paths after clear, got %d", ctx.NumIncludePaths())
	}
}

func TestContextLogger(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	// Default logger should be set
	if ctx.Logger() == nil {
		t.Error("Default logger should not be nil")
	}

	// Set custom logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	ctx.SetLogger(logger)

	if ctx.Logger() != logger {
		t.Error("Logger not set correctly")
	}

	// Set nil logger (should use no-op)
	ctx.SetLogger(nil)
	if ctx.Logger() == nil {
		t.Error("Setting nil logger should create no-op logger")
	}
}

func TestContextFlags(t *testing.T) {
	ctx := NewContext(ContextNoDefaultIncludes | ContextNoEnvironmentNames)

	flags := ctx.Flags()
	if flags&ContextNoDefaultIncludes == 0 {
		t.Error("Expected ContextNoDefaultIncludes flag")
	}
	if flags&ContextNoEnvironmentNames == 0 {
		t.Error("Expected ContextNoEnvironmentNames flag")
	}
}

func TestContextNewKeymapFromStringNotImplemented(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	_, err := ctx.NewKeymapFromString([]byte("test"), KeymapFormatTextV1)
	if err == nil {
		t.Error("Expected error for unimplemented function")
	}
}

func TestContextNewKeymapFromStringUnsupportedFormat(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	_, err := ctx.NewKeymapFromString([]byte("test"), KeymapFormat(99))
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestContextNewKeymapFromNames(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	_, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})
	if err == nil {
		t.Error("Expected error for unimplemented function")
	}
}

func TestContextNewComposeTableFromLocale(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	_, err := ctx.NewComposeTableFromLocale("en_US.UTF-8", ComposeCompileNoFlags)
	if err == nil {
		t.Error("Expected error for unimplemented function")
	}
}

func TestContextNewComposeTableFromFile(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	_, err := ctx.NewComposeTableFromFile("/nonexistent", "en_US.UTF-8", ComposeCompileNoFlags)
	if err == nil {
		t.Error("Expected error for unimplemented function")
	}
}

func TestContextConcurrentAccess(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	// Test concurrent access to include paths
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				ctx.AppendIncludePath("/test/path")
				_ = ctx.IncludePaths()
				_ = ctx.NumIncludePaths()
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
