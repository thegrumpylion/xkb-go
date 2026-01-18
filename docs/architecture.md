# XKB-Go Architecture

A pure Go implementation of the XKB (X Keyboard Extension) library, compatible with libxkbcommon.

## Overview

XKB-Go provides keyboard handling for Wayland and X11 applications without requiring CGO or the libxkbcommon C library at runtime.

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Application                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   ┌─────────┐     ┌─────────┐     ┌─────────────┐                   │
│   │ Context │────▶│ Keymap  │────▶│    State    │                   │
│   └─────────┘     └─────────┘     └─────────────┘                   │
│        │               ▲                  │                          │
│        │               │                  ▼                          │
│        │          ┌─────────┐      ┌─────────────┐                  │
│        │          │ Parser  │      │   Compose   │                  │
│        │          └─────────┘      │    State    │                  │
│        │               ▲           └─────────────┘                  │
│        ▼               │                  ▲                          │
│   ┌─────────┐     ┌─────────┐      ┌─────────────┐                  │
│   │ Include │     │ Keymap  │      │   Compose   │                  │
│   │  Paths  │     │  Text   │      │    Table    │                  │
│   └─────────┘     └─────────┘      └─────────────┘                  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Core Concepts

### Keycodes

Physical key identifiers. In Linux evdev, keycodes are scancode + 8.

```
Scancode 16 (Q key) → Keycode 24
Scancode 41 (` key) → Keycode 49
```

XKB uses symbolic names for keycodes:
- `<AD01>` = keycode 24 (first key of row A-D, i.e., Q on QWERTY)
- `<TLDE>` = keycode 49 (tilde key)
- `<LFSH>` = keycode 50 (left shift)

### Keysyms

Abstract symbols representing what a key produces. Defined in X11 headers.

```
XKB_KEY_a        = 0x0061
XKB_KEY_A        = 0x0041
XKB_KEY_Shift_L  = 0xffe1
XKB_KEY_Return   = 0xff0d
XKB_KEY_dead_acute = 0xfe51
```

Keysyms above 0x01000000 are Unicode codepoints + 0x01000000.

### Modifiers

Keys that modify the behavior of other keys.

**Real Modifiers** (8 total, hardware-mapped):
- Shift, Lock (Caps Lock), Control
- Mod1 (usually Alt), Mod2 (usually Num Lock)
- Mod3, Mod4 (usually Super/Win), Mod5 (usually ISO_Level3_Shift/AltGr)

**Virtual Modifiers** (user-defined, mapped to real):
- Alt, Super, Hyper, Meta, NumLock, LevelThree, etc.

### Groups (Layouts)

Different keyboard layouts the user can switch between.
- Group 1: English (US)
- Group 2: Russian
- Group 3: German
- Group 4: (max 4 groups)

### Levels

Different outputs from the same key based on modifier state.
- Level 1: `a` (no modifiers)
- Level 2: `A` (Shift)
- Level 3: `æ` (AltGr on some layouts)
- Level 4: `Æ` (Shift+AltGr)

### Key Types

Rules defining how modifiers select levels for a key.

```
type "ALPHABETIC" {
    modifiers = Shift + Lock;
    map[None]       = Level1;  // a
    map[Shift]      = Level2;  // A
    map[Lock]       = Level2;  // A (Caps Lock)
    map[Shift+Lock] = Level1;  // a (Shift cancels Caps)
}
```

---

## Layer 1: Context

The top-level container holding configuration shared across keymaps.

```go
type Context struct {
    includePaths []string
    logLevel     LogLevel
    logFn        func(level LogLevel, format string, args ...any)
    flags        ContextFlags
}
```

**Responsibilities:**
- Manage include paths for keymap file resolution
- Provide logging infrastructure
- Factory for Keymap and ComposeTable objects

**Why separate from Keymap?**
- Multiple keymaps can share the same context
- Logging and include paths are environment-level concerns
- Follows libxkbcommon's design for compatibility

---

## Layer 2: Keymap (Immutable)

The compiled keyboard configuration. Created once, never modified.

```go
type Keymap struct {
    ctx *Context

    // Keycode ↔ name mapping
    keycodeNames   map[Keycode]string
    keycodesByName map[string]Keycode
    minKeycode     Keycode
    maxKeycode     Keycode

    // Key type definitions
    types     map[string]*KeyType
    typesList []*KeyType

    // Per-key information
    keys map[Keycode]*Key

    // Modifier definitions
    modNames    [8]string           // Real modifier names
    virtualMods map[string]ModMask  // Virtual → real mapping

    // LED/indicator definitions
    leds map[string]*LED

    // Group names
    groupNames []string
    numGroups  int
}
```

### Sub-structures

```go
type KeyType struct {
    name      string
    mods      ModMask              // Modifiers this type considers
    numLevels int
    entries   []KeyTypeEntry       // Modifier → level mapping
}

type KeyTypeEntry struct {
    mods     ModMask
    level    Level
    preserve ModMask  // Modifiers to preserve (not consume)
}

type Key struct {
    keycode   Keycode
    name      string
    groups    []KeyGroup
    repeats   bool
    vmodmap   ModMask  // Virtual modifiers this key sets
}

type KeyGroup struct {
    keyType *KeyType
    levels  []KeyLevel
}

type KeyLevel struct {
    syms []Keysym  // Usually 1, rarely more
}
```

---

## Layer 3: Keymap Parser

Parses XKB text format into a Keymap structure.

### XKB Text Format

A complete keymap has 4-5 sections:

```
xkb_keymap {
    xkb_keycodes "name" { ... };
    xkb_types "name" { ... };
    xkb_compat "name" { ... };
    xkb_symbols "name" { ... };
    xkb_geometry "name" { ... };  // Ignored
};
```

### Parser Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│    Lexer     │────▶│    Parser    │────▶│   Keymap     │
│  (Tokenize)  │     │   (Grammar)  │     │  (Compiled)  │
└──────────────┘     └──────────────┘     └──────────────┘
       │                    │
       ▼                    ▼
  Token Stream         AST Nodes
```

**Lexer tokens:**
- Keywords: `xkb_keymap`, `xkb_keycodes`, `type`, `key`, `include`, etc.
- Identifiers: `TLDE`, `TWO_LEVEL`, `Shift`, etc.
- Strings: `"us"`, `"basic"`
- Numbers: `49`, `0xff0d`, `0x01000041`
- Operators: `=`, `+`, `[`, `]`, `{`, `}`, `;`, etc.

**Parser complexity:**
- Include resolution with merge modes
- Virtual modifier declaration and resolution
- Action parsing for xkb_compat
- Key type and symbol binding

---

## Layer 4: State (Mutable)

Tracks the current keyboard state for key translation.

```go
type State struct {
    keymap *Keymap

    // Modifier state (three components)
    baseMods    ModMask  // Currently pressed
    latchedMods ModMask  // One-shot (sticky keys)
    lockedMods  ModMask  // Toggled (Caps Lock)

    // Group state (three components)
    baseGroup    Group
    latchedGroup Group
    lockedGroup  Group

    // Cached effective values
    effectiveMods ModMask
    effectiveGroup Group
}
```

### State Update Methods

**From Wayland/X11 (server sends modifier state):**
```go
func (s *State) UpdateMask(
    baseMods, latchedMods, lockedMods ModMask,
    baseGroup, latchedGroup, lockedGroup Group,
) StateComponent
```

**From evdev (manual key tracking):**
```go
func (s *State) UpdateKey(keycode Keycode, direction KeyDirection) StateComponent
```

### Key Translation Algorithm

```
func (s *State) KeyGetSyms(keycode Keycode) []Keysym:
    1. key := s.keymap.keys[keycode]
    2. group := s.effectiveGroup % len(key.groups)
    3. keyGroup := key.groups[group]
    4. keyType := keyGroup.keyType
    5. maskedMods := s.effectiveMods & keyType.mods
    6. level := keyType.LookupLevel(maskedMods)
    7. return keyGroup.levels[level].syms
```

---

## Layer 5: Compose (Dead Keys)

Handles multi-key sequences that produce single characters.

### Compose Table

```go
type ComposeTable struct {
    locale string
    root   *composeNode
}

type composeNode struct {
    children map[Keysym]*composeNode
    result   *composeResult  // nil if not terminal
}

type composeResult struct {
    keysym Keysym
    utf8   string
}
```

### Compose State Machine

```go
type ComposeState struct {
    table   *ComposeTable
    current *composeNode
    status  ComposeStatus
}

type ComposeStatus int
const (
    ComposeNothing   ComposeStatus = iota  // No sequence in progress
    ComposeComposing                        // Sequence in progress
    ComposeComposed                         // Sequence complete
    ComposeCancelled                        // Sequence aborted
)
```

### Compose File Format

From `/usr/share/X11/locale/en_US.UTF-8/Compose`:

```
<dead_acute> <a>         : "á"   aacute
<dead_acute> <A>         : "Á"   Aacute
<Multi_key> <a> <e>      : "æ"   ae
<Multi_key> <c> <o> <p> <y> : "©" copyright
```

---

## Layer 6: Keysym Utilities

Static functions and tables for keysym handling.

```go
// Conversion
func KeysymToUTF32(keysym Keysym) rune
func UTF32ToKeysym(r rune) Keysym
func KeysymToUTF8(keysym Keysym) string

// Naming
func KeysymGetName(keysym Keysym) string
func KeysymFromName(name string, flags KeysymFlags) Keysym

// Classification
func KeysymIsModifier(keysym Keysym) bool
func KeysymIsKeypad(keysym Keysym) bool
```

### Keysym Tables

Generated from X11 keysym headers (~2500 entries):

```go
var keysymNames = map[Keysym]string{
    0x0020: "space",
    0x0041: "A",
    0x0061: "a",
    0xff08: "BackSpace",
    0xff09: "Tab",
    0xff0d: "Return",
    0xff1b: "Escape",
    0xffe1: "Shift_L",
    0xffe2: "Shift_R",
    // ... ~2500 more
}

var keysymsByName = map[string]Keysym{
    // Inverse of above
}

// Unicode keysyms: 0x01000000 + codepoint
func KeysymToUTF32(ks Keysym) rune {
    if ks >= 0x01000000 && ks <= 0x0110ffff {
        return rune(ks - 0x01000000)
    }
    // Lookup in table for legacy keysyms
    return keysymToUnicode[ks]
}
```

---

## Package Structure

```
github.com/thegrumpylion/xkb-go/
├── xkb.go           # Public API, Context
├── keymap.go        # Keymap struct and methods
├── state.go         # State struct and methods
├── compose.go       # ComposeTable and ComposeState
├── keysym.go        # Keysym utilities
├── keysym_names.go  # Generated keysym tables
├── keysym_utf.go    # Keysym ↔ Unicode tables
├── parser/
│   ├── lexer.go     # Tokenizer
│   ├── parser.go    # Grammar parser
│   ├── ast.go       # AST nodes
│   ├── keycodes.go  # xkb_keycodes section
│   ├── types.go     # xkb_types section
│   ├── compat.go    # xkb_compat section
│   └── symbols.go   # xkb_symbols section
├── compose/
│   ├── parser.go    # Compose file parser
│   └── table.go     # ComposeTable implementation
└── internal/
    └── atom/        # String interning (optional)
```

---

## Data Flow: Complete Example

```
User presses 'Q' key with Shift held on US keyboard:

1. Hardware
   └─▶ Scancode: 16

2. evdev (Linux input)
   └─▶ Keycode: 24 (scancode + 8)

3. Wayland compositor
   └─▶ wl_keyboard.key(serial, time, key=24, state=pressed)
   └─▶ wl_keyboard.modifiers(depressed=Shift, latched=0, locked=0, group=0)

4. Application: state.UpdateMask(Shift, 0, 0, 0, 0, 0)
   └─▶ effectiveMods = Shift
   └─▶ effectiveGroup = 0

5. Application: state.KeyGetOneSym(24)
   a. key = keymap.keys[24]  // Key "AD01"
   b. group = 0 % 1 = 0
   c. keyType = "ALPHABETIC"
   d. maskedMods = Shift & (Shift|Lock) = Shift
   e. level = typeEntries[Shift] = Level2
   f. return key.groups[0].levels[1].syms[0]
   └─▶ Keysym: 0x0051 (Q)

6. Application: composeState.Feed(0x0051)
   └─▶ Status: ComposeNothing (Q doesn't start sequence)

7. Application: KeysymToUTF32(0x0051)
   └─▶ 'Q' (rune 0x51)

8. Application: insert 'Q' into text field
```

---

## Compatibility with libxkbcommon

This implementation aims for API compatibility where practical:

| libxkbcommon | xkb-go |
|--------------|--------|
| `xkb_context_new()` | `xkb.NewContext()` |
| `xkb_keymap_new_from_string()` | `ctx.NewKeymapFromString()` |
| `xkb_state_new()` | `keymap.NewState()` |
| `xkb_state_key_get_one_sym()` | `state.KeyGetOneSym()` |
| `xkb_state_key_get_utf32()` | `state.KeyGetUTF32()` |
| `xkb_state_update_mask()` | `state.UpdateMask()` |
| `xkb_compose_table_new_from_locale()` | `ctx.NewComposeTableFromLocale()` |
| `xkb_compose_state_feed()` | `composeState.Feed()` |

---

## Memory Model

Unlike libxkbcommon (reference counting), xkb-go uses Go's garbage collector:

- **Context**: Lives as long as referenced
- **Keymap**: Immutable, safe to share across goroutines
- **State**: Mutable, one per keyboard device, not goroutine-safe
- **ComposeState**: Mutable, one per keyboard, not goroutine-safe
