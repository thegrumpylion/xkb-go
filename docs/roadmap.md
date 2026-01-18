# XKB-Go Implementation Roadmap

## Status Summary

| Component | Status | Notes |
|-----------|--------|-------|
| Core Types | ✅ Done | `types.go` |
| Keysym Utilities | ✅ Done | `keysym.go` (partial table) |
| Context | ✅ Done | `context.go` |
| Keymap Structure | ✅ Done | `keymap.go` |
| State Machine | ✅ Done | `state.go` |
| XKB Parser (Lexer) | ✅ Done | `lexer.go` |
| XKB Parser (Full) | ✅ Done | `parser.go` |
| Compose Parser | ✅ Done | `compose_parser.go` |
| Compose State | ✅ Done | `compose.go` |
| RMLVO Compilation | ❌ Not Done | `NewKeymapFromNames()` |
| Full Keysym Tables | ❌ Not Done | Generated from headers |

**Test Coverage:** 80.5%

---

## What Works Today

### Wayland Client Use Case (Primary)

```go
// 1. Create context
ctx := xkb.NewContext(xkb.ContextNoFlags)

// 2. Parse keymap from compositor (wl_keyboard.keymap event)
keymap, err := ctx.NewKeymapFromString(keymapData, xkb.KeymapFormatTextV1)

// 3. Create state
state := keymap.NewState()

// 4. Update modifier state (wl_keyboard.modifiers event)
state.UpdateMask(depressed, latched, locked, 0, 0, group)

// 5. Translate keys (wl_keyboard.key event)
keysym := state.KeyGetOneSym(keycode)
utf8 := state.KeyGetUTF8(keycode)

// 6. Optional: Compose/dead key handling
composeTable, _ := ctx.NewComposeTableFromLocale("en_US.UTF-8", 0)
composeState := composeTable.NewState(0)
result := composeState.Feed(keysym)
if composeState.GetStatus() == xkb.ComposeComposed {
    utf8 = composeState.GetUTF8()
}
```

### Tested With Real Data

- System keymaps from `/usr/share/X11/xkb/`
- System compose files from `/usr/share/X11/locale/`
- Wayland compositor keymaps (dumped via `xkbcli`)

---

## Implementation Details

### Phase 1-3: Core (✅ Complete)

**Files:** `types.go`, `keysym.go`, `context.go`, `keymap.go`, `state.go`, `errors.go`, `testing.go`

- [x] Core types: `Keysym`, `Keycode`, `ModMask`, `Level`, `Group`
- [x] Modifier constants and indices
- [x] `KeyDirection`, `StateComponent` flags
- [x] Keysym ↔ name conversion (partial table)
- [x] Keysym ↔ UTF-8/UTF-32 conversion
- [x] Context with include paths and logging (`log/slog`)
- [x] Keymap structure with types, keys, groups, levels
- [x] State machine with three-component modifier/group state
- [x] Key translation: `KeyGetOneSym`, `KeyGetSyms`, `KeyGetUTF8`, `KeyGetUTF32`
- [x] Modifier queries: `ModNameIsActive`, `ModIndexIsActive`, `SerializeMods`
- [x] `UpdateMask()` for Wayland
- [x] `UpdateKey()` for evdev

### Phase 4-9: XKB Parser (✅ Complete)

**Files:** `lexer.go`, `parser.go`

- [x] Lexer with all token types
- [x] Line/column tracking for errors
- [x] Comment handling (`//` and `/* */`)
- [x] String escape sequences
- [x] Hex and octal number literals
- [x] `xkb_keycodes` section: keycode assignments, aliases, indicators, min/max
- [x] `xkb_types` section: type definitions, modifiers, level mappings, preserve
- [x] `xkb_compat` section: interpret statements, virtual modifiers, indicators
- [x] `xkb_symbols` section: key definitions, groups, levels, modifier_map
- [x] `xkb_geometry` section: parsed and ignored
- [x] `ctx.NewKeymapFromString()` fully functional
- [x] Tested with real system keymaps

### Phase 10: Compose Tables (✅ Complete)

**Files:** `compose.go`, `compose_parser.go`

- [x] Compose file parser with include support
- [x] Trie-based compose table
- [x] `ctx.NewComposeTableFromLocale()` - locale-based lookup
- [x] `ctx.NewComposeTableFromFile()` - direct file loading
- [x] Compose state machine: `Feed`, `GetStatus`, `GetOneSym`, `GetUTF8`, `Reset`
- [x] Dead key sequences (dead_acute + a = á)
- [x] Multi_key sequences (Multi_key + o + c = ©)
- [x] Tested with real system compose files

---

## What's NOT Implemented

### RMLVO Compilation

`ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})` returns `ErrNotImplemented`.

**What it does:** Converts Rules+Model+Layout+Variant+Options into a complete keymap by:
1. Reading rules files from `/usr/share/X11/xkb/rules/`
2. Looking up component files based on RMLVO
3. Merging includes from symbols/, types/, keycodes/, compat/
4. Building the final keymap

**Who needs it:**
- Compositors (to build keymaps from user preferences)
- Testing tools (to get "us" layout without dumping)

**Wayland clients don't need it** - they receive complete keymaps from the compositor.

### Complete Keysym Tables

Current `keysym.go` has ~200 common keysyms. Full table would be ~2500.

**What's missing:**
- Many Latin-2 through Latin-10 characters
- Full multimedia keys (XF86keysym.h)
- Complete dead key set
- All currency symbols

**Impact:** Uncommon keysyms return empty names from `KeysymGetName()`.

### NewKeymapFromFile

Not implemented (not needed for Wayland clients).

### Keymap Serialization

`keymap.GetAsString()` not implemented (rarely needed).

---

## What's Next

### Option A: RMLVO Compilation

Implement `NewKeymapFromNames()` for:
- Compositors building keymaps
- Convenient testing ("give me US layout")

**Complexity:** Medium-High (rules file parsing, include resolution)

### Option B: Complete Keysym Tables

Generate full tables from X11 headers:
- `keysymdef.h` → ~2000 keysyms
- `XF86keysym.h` → ~500 multimedia keys

**Complexity:** Low (code generation)

### Option C: Polish and Testing

- Fuzz testing for parser
- Benchmark against libxkbcommon
- Edge case coverage
- Documentation improvements

**Complexity:** Low-Medium

### Option D: Use It!

The library is functional for Wayland clients. Ship it, find bugs in real usage.

---

## Architecture Notes

### Wayland vs X11 Relevance

| Component | Source | Wayland Client Needs |
|-----------|--------|---------------------|
| XKB Keymap | Compositor sends via `wl_keyboard.keymap` | ✅ Must parse |
| XKB State | Client maintains | ✅ Must implement |
| Compose Tables | Client reads from `/usr/share/X11/locale/` | Optional |
| RMLVO Rules | Compositor uses to build keymap | ❌ Not needed |

### Thread Safety

- `Context` - Safe (mutex protected)
- `Keymap` - Safe (immutable after creation)
- `State` - NOT safe (one per keyboard device)
- `ComposeState` - NOT safe (one per keyboard)

---

## File Summary

```
github.com/thegrumpylion/xkb-go/
├── types.go           # Core types (Keysym, Keycode, ModMask, etc.)
├── keysym.go          # Keysym utilities and partial tables
├── context.go         # Context, include paths, factory methods
├── keymap.go          # Keymap struct and query methods
├── state.go           # State machine for key translation
├── lexer.go           # XKB text format tokenizer
├── parser.go          # XKB keymap parser
├── compose.go         # ComposeTable and ComposeState
├── compose_parser.go  # Compose file parser
├── errors.go          # Error types
├── testing.go         # Test helpers (TestKeymap, TestComposeTable)
├── testdata/
│   └── system_keymap.xkb  # Real keymap for testing
└── docs/
    ├── architecture.md
    ├── parser.md
    ├── decisions.md
    ├── references.md
    └── roadmap.md (this file)
```

---

## Testing

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Verbose
go test -v ./...
```

Current coverage: **80.5%**
