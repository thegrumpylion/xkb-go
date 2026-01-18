# XKB-Go Architecture

## Overview

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
- Mod3, Mod4 (usually Super/Win), Mod5 (usually AltGr)

**Virtual Modifiers** (user-defined, mapped to real):

- Alt, Super, Hyper, Meta, NumLock, LevelThree, etc.

### Groups (Layouts)

Different keyboard layouts the user can switch between (max 4).

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

## Data Structures

### Context

Top-level container holding configuration shared across keymaps.

```go
type Context struct {
    ctx          context.Context
    mu           sync.RWMutex
    flags        ContextFlags
    logger       *slog.Logger
    includePaths []string
}
```

### Keymap

Compiled keyboard configuration. Immutable after creation.

```go
type Keymap struct {
    ctx *Context

    keycodeNames   map[Keycode]string
    keycodesByName map[string]Keycode
    minKeycode     Keycode
    maxKeycode     Keycode

    types     map[string]*KeyType
    keys      map[Keycode]*Key
    modNames  [8]string
    virtualMods map[string]ModMask
    leds      map[string]*LED
    groupNames []string
    numGroups  int
}

type Key struct {
    keycode Keycode
    name    string
    groups  []KeyGroup
    repeats bool
    vmodmap ModMask
}

type KeyGroup struct {
    keyType *KeyType
    levels  []KeyLevel
}

type KeyLevel struct {
    syms []Keysym
}
```

### State

Tracks current keyboard state for key translation.

```go
type State struct {
    keymap *Keymap

    // Modifier state
    baseMods    ModMask
    latchedMods ModMask
    lockedMods  ModMask

    // Group state
    baseGroup    Group
    latchedGroup Group
    lockedGroup  Group

    // Cached effective values
    effectiveMods  ModMask
    effectiveGroup Group
}
```

### Compose

Handles multi-key sequences (dead keys).

```go
type ComposeTable struct {
    locale string
    root   *composeNode  // Trie structure
}

type ComposeState struct {
    table   *ComposeTable
    current *composeNode
    status  ComposeStatus
}
```

---

## Key Translation Algorithm

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

## Data Flow Example

```
User presses 'Q' key with Shift held on US keyboard:

1. Hardware → Scancode: 16

2. evdev (Linux) → Keycode: 24 (scancode + 8)

3. Wayland compositor sends:
   - wl_keyboard.key(key=24, state=pressed)
   - wl_keyboard.modifiers(depressed=Shift, ...)

4. Application: state.UpdateMask(Shift, 0, 0, 0, 0, 0)

5. Application: state.KeyGetOneSym(24)
   - key = keymap.keys[24]  // Key "AD01"
   - keyType = "ALPHABETIC"
   - maskedMods = Shift & (Shift|Lock) = Shift
   - level = Level2
   → Keysym: 0x0051 (Q)

6. Application: KeysymToUTF32(0x0051)
   → 'Q' (rune 0x51)
```

---

## RMLVO Compilation

Build keymaps from rules files (RMLVO = Rules, Model, Layout, Variant, Options).

```
RuleNames → Rules Parser → KcCGST → Component Loader → Keymap String → Parser
```

1. Parse rules file (e.g., `/usr/share/X11/xkb/rules/evdev`)
2. Resolve RMLVO to component specifiers
3. Load component files from include paths
4. Resolve `include` and `augment` directives
5. Assemble complete keymap string
6. Parse with standard parser

---

## Memory Model

Unlike libxkbcommon (reference counting), xkb-go uses Go's garbage collector:

- **Context**: Lives as long as referenced
- **Keymap**: Immutable, safe to share across goroutines
- **State**: Mutable, one per keyboard device, not goroutine-safe
- **ComposeState**: Mutable, one per keyboard, not goroutine-safe
