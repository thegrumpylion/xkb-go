package xkb

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// ContextFlags controls context behavior.
type ContextFlags uint32

const (
	// ContextNoFlags is the default context flags.
	ContextNoFlags ContextFlags = 0

	// ContextNoDefaultIncludes prevents adding default include paths.
	// By default, the context includes system XKB data paths.
	ContextNoDefaultIncludes ContextFlags = 1 << 0

	// ContextNoEnvironmentNames prevents reading RMLVO names from environment.
	// Affects XKB_DEFAULT_RULES, XKB_DEFAULT_MODEL, etc.
	ContextNoEnvironmentNames ContextFlags = 1 << 1
)

// Context is the top-level xkb object.
// It holds configuration shared across keymaps such as include paths and logging.
// Context is safe for concurrent use.
type Context struct {
	mu           sync.RWMutex
	flags        ContextFlags
	logger       *slog.Logger
	includePaths []string
}

// NewContext creates a new xkb context with the given flags.
func NewContext(flags ContextFlags) *Context {
	c := &Context{
		flags:  flags,
		logger: slog.Default(),
	}

	if flags&ContextNoDefaultIncludes == 0 {
		c.addDefaultIncludePaths()
	}

	return c
}

// addDefaultIncludePaths adds the standard XKB data directories.
func (c *Context) addDefaultIncludePaths() {
	// User XKB directory
	if home, err := os.UserHomeDir(); err == nil {
		userXKB := filepath.Join(home, ".xkb")
		if info, err := os.Stat(userXKB); err == nil && info.IsDir() {
			c.includePaths = append(c.includePaths, userXKB)
		}
	}

	// XDG config directory
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		xkbDir := filepath.Join(xdgConfig, "xkb")
		if info, err := os.Stat(xkbDir); err == nil && info.IsDir() {
			c.includePaths = append(c.includePaths, xkbDir)
		}
	}

	// System XKB data directories
	systemPaths := []string{
		"/usr/share/X11/xkb",
		"/usr/local/share/X11/xkb",
	}
	for _, p := range systemPaths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			c.includePaths = append(c.includePaths, p)
		}
	}
}

// SetLogger sets the logger for this context.
// If logger is nil, a no-op logger is used.
func (c *Context) SetLogger(logger *slog.Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if logger == nil {
		// Create a no-op logger
		logger = slog.New(slog.NewTextHandler(nil, nil))
	}
	c.logger = logger
}

// Logger returns the current logger.
func (c *Context) Logger() *slog.Logger {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.logger
}

// log is a convenience method for logging.
func (c *Context) log(level slog.Level, msg string, args ...any) {
	c.mu.RLock()
	logger := c.logger
	c.mu.RUnlock()

	if logger != nil {
		logger.Log(nil, level, msg, args...)
	}
}

// IncludePaths returns a copy of the current include paths.
func (c *Context) IncludePaths() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	paths := make([]string, len(c.includePaths))
	copy(paths, c.includePaths)
	return paths
}

// AppendIncludePath adds a path to the end of the include path list.
// Later paths have lower priority when resolving includes.
func (c *Context) AppendIncludePath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.includePaths = append(c.includePaths, path)
}

// PrependIncludePath adds a path to the front of the include path list.
// Earlier paths have higher priority when resolving includes.
func (c *Context) PrependIncludePath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.includePaths = append([]string{path}, c.includePaths...)
}

// ClearIncludePaths removes all include paths.
func (c *Context) ClearIncludePaths() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.includePaths = nil
}

// NumIncludePaths returns the number of include paths.
func (c *Context) NumIncludePaths() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.includePaths)
}

// Flags returns the context flags.
func (c *Context) Flags() ContextFlags {
	return c.flags
}

// resolvePath looks for a file in the include paths.
// Returns the full path if found, empty string otherwise.
func (c *Context) resolvePath(filename string) string {
	c.mu.RLock()
	paths := c.includePaths
	c.mu.RUnlock()

	for _, base := range paths {
		full := filepath.Join(base, filename)
		if _, err := os.Stat(full); err == nil {
			return full
		}
	}
	return ""
}

// NewKeymapFromString parses a keymap from XKB text format.
// This is the primary method for Wayland clients, which receive
// the complete keymap as a string from the compositor.
//
// The format parameter must be KeymapFormatTextV1.
func (c *Context) NewKeymapFromString(text []byte, format KeymapFormat) (*Keymap, error) {
	if format != KeymapFormatTextV1 {
		return nil, &Error{Op: "NewKeymapFromString", Err: ErrUnsupportedFormat}
	}

	parser := NewParser(text)
	keymap, err := parser.Parse()
	if err != nil {
		return nil, &Error{Op: "NewKeymapFromString", Err: err}
	}

	keymap.ctx = c
	return keymap, nil
}

// NewKeymapFromNames builds a keymap from RMLVO names.
// This looks up the rules file and assembles keymap components.
//
// If names is nil, default values are used.
// Empty fields in names fall back to defaults.
func (c *Context) NewKeymapFromNames(names *RuleNames) (*Keymap, error) {
	// TODO: Implement RMLVO compilation
	return nil, &Error{Op: "NewKeymapFromNames", Err: ErrNotImplemented}
}

// NewComposeTableFromLocale loads a compose table for the given locale.
// The locale string should be in the form "language_TERRITORY.encoding"
// (e.g., "en_US.UTF-8").
//
// The compose table is searched for in:
// 1. $XCOMPOSEFILE environment variable
// 2. ~/.XCompose
// 3. System locale compose file
func (c *Context) NewComposeTableFromLocale(locale string, flags ComposeCompileFlags) (*ComposeTable, error) {
	// TODO: Implement compose table loading
	return nil, &Error{Op: "NewComposeTableFromLocale", Err: ErrNotImplemented}
}

// NewComposeTableFromFile loads a compose table from a specific file.
func (c *Context) NewComposeTableFromFile(path string, locale string, flags ComposeCompileFlags) (*ComposeTable, error) {
	// TODO: Implement compose table loading
	return nil, &Error{Op: "NewComposeTableFromFile", Err: ErrNotImplemented}
}
