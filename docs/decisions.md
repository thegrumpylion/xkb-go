# Design Decisions

This document records key design decisions for xkb-go.

---

## D1: Wayland-First, Not X11

**Decision**: xkb-go is Wayland-first. X11-specific features are not implemented.

**Rationale**:
- Wayland is the modern Linux display protocol
- X11-specific functions require X11 connection and add complexity
- Wayland sends complete keymaps as strings, simplifying our primary use case

**What we skip**:
- `xkb_x11_*` functions (require X11 connection)
- X11-specific state synchronization

**What we keep**:
- Everything needed for `wl_keyboard` handling
- Compose/dead key support (works the same)
- RMLVO compilation (useful for testing, not X11-specific)

---

## D2: Go-Style Error Handling

**Decision**: Use Go's standard error handling with returned `error` values.

**Rationale**:
- Idiomatic Go
- Explicit error handling
- Works with `errors.Is`, `errors.As`, `fmt.Errorf` wrapping

**Example**:
```go
keymap, err := ctx.NewKeymapFromString(data, xkb.KeymapFormatTextV1)
if err != nil {
    return fmt.Errorf("parse keymap: %w", err)
}
```

**Not doing**: C-style "store last error in context" pattern.

---

## D3: Use log/slog for Logging

**Decision**: Use Go's `log/slog` package for structured logging.

**Rationale**:
- Standard library (Go 1.21+)
- Structured logging with levels
- Users can inject their own `*slog.Logger` for custom handling
- Integrates with existing logging infrastructure

**API**:
```go
// Default: uses slog.Default()
ctx := xkb.NewContext(xkb.ContextNoFlags)

// Custom logger
ctx.SetLogger(myLogger)
```

**Not doing**: Custom log callback function signature like libxkbcommon.

---

## D4: Thread Safety Model

**Decision**:
- `Context` is safe for concurrent use (protected by mutex)
- `Keymap` is immutable after creation, safe to share
- `State` is NOT safe for concurrent use (one per keyboard device)
- `ComposeState` is NOT safe for concurrent use

**Rationale**:
- Matches real-world usage patterns
- Context is typically shared across the application
- Keymap is read-only after parsing
- State is per-device and updated on each key event

---

## D5: Feature Compatibility, Not API Compatibility

**Decision**: Provide feature compatibility with libxkbcommon, not API compatibility.

**Rationale**:
- Go has different idioms than C
- Can provide better Go-native API
- Users should not need to translate C patterns

**Examples**:
| libxkbcommon (C) | xkb-go (Go) |
|------------------|-------------|
| `xkb_context_new(flags)` | `xkb.NewContext(flags)` |
| `xkb_context_unref(ctx)` | (garbage collected) |
| `xkb_keymap_new_from_string(ctx, str, fmt, flags)` | `ctx.NewKeymapFromString(data, fmt)` |
| Returns `NULL` on error | Returns `error` |

---

## D6: RMLVO Support (Deferred)

**Decision**: Implement RMLVO compilation in a later phase.

**RMLVO** = Rules + Model + Layout + Variant + Options

**Rationale**:
- Primary use case (Wayland) receives complete keymap strings
- RMLVO is useful for testing ("give me US layout")
- Rules file parsing adds complexity
- Can be added after core parser works

**Plan**:
1. First: Parse complete keymaps from strings
2. Later: Add RMLVO for convenient testing and compositor use

---

## D7: Geometry Section Ignored

**Decision**: Parse but ignore `xkb_geometry` section.

**Rationale**:
- Geometry describes physical keyboard layout for visualization
- Not needed for key translation
- libxkbcommon also ignores it
- Reduces implementation scope

---

## D8: Package Structure

**Decision**: Single package `xkb` with internal sub-packages.

```
github.com/thegrumpylion/xkb-go/
├── xkb.go              # Public API
├── keymap.go           # Keymap type
├── state.go            # State type
├── compose.go          # Compose types
├── keysym.go           # Keysym utilities
├── keysym_tables.go    # Generated tables
├── internal/
│   └── parser/         # Keymap parser (internal)
└── cmd/
    └── xkb-go/         # CLI tools (optional)
```

**Rationale**:
- Simple import: `import "github.com/thegrumpylion/xkb-go"`
- Parser is internal implementation detail
- Follows Go conventions

---

## D9: Keysym Table Generation

**Decision**: Generate keysym tables from X11 header files.

**Source files**:
- `keysymdef.h` - Core keysyms
- `XF86keysym.h` - Multimedia keys

**Generated**:
- `keysym_tables.go` - Name ↔ keysym mappings
- `keysym_unicode.go` - Keysym → Unicode mappings

**Rationale**:
- Canonical source of truth
- ~2500 keysyms, impractical to maintain by hand
- Can regenerate when X11 headers update

---

## D10: Compose File Locations

**Decision**: Follow XDG and libX11 conventions for compose file search.

**Search order**:
1. `$XCOMPOSEFILE` environment variable
2. `~/.XCompose`
3. `$XDG_CONFIG_HOME/XCompose` (if exists)
4. System locale file: `/usr/share/X11/locale/<locale>/Compose`

**Rationale**:
- Compatible with existing user configurations
- Follows established conventions
