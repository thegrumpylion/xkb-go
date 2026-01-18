# XKB-Go Implementation Roadmap

## Philosophy

- **Top-down approach**: Start with public API, implement layers as needed
- **Test-driven**: Each phase includes tests before moving on
- **Iterative**: Get basic functionality working, then expand
- **Real-world validation**: Test against actual Wayland keymaps

---

## Phase 1: Foundation

**Goal**: Basic types, keysym utilities, and public API skeleton.

### 1.1 Core Types
- [ ] `Keysym` type (uint32)
- [ ] `Keycode` type (uint32)
- [ ] `ModMask` type (uint32)
- [ ] `Level` type (uint8)
- [ ] `Group` type (uint8)
- [ ] Constants for real modifiers (Shift, Lock, Control, Mod1-5)
- [ ] `KeyDirection` enum (Up, Down)
- [ ] `StateComponent` flags

### 1.2 Keysym Tables
- [ ] Generate `keysym_names.go` from X11 keysymdefs.h
- [ ] Generate `keysym_utf.go` for keysym ↔ Unicode mapping
- [ ] `KeysymGetName(Keysym) string`
- [ ] `KeysymFromName(string) Keysym`
- [ ] `KeysymToUTF32(Keysym) rune`
- [ ] `KeysymToUTF8(Keysym) string`

### 1.3 Context
- [ ] `Context` struct
- [ ] `NewContext(flags ContextFlags) *Context`
- [ ] `ctx.SetLogLevel(LogLevel)`
- [ ] `ctx.SetLogFn(func(LogLevel, string, ...any))`
- [ ] `ctx.IncludePath() []string`
- [ ] `ctx.AppendIncludePath(string)`

### 1.4 Tests
- [ ] Keysym name round-trip tests
- [ ] Keysym to UTF conversion tests
- [ ] Context creation tests

**Deliverable**: Can look up keysym names and convert to Unicode.

---

## Phase 2: Keymap Structure

**Goal**: Define keymap data structures (no parsing yet).

### 2.1 Keymap Types
- [ ] `Keymap` struct
- [ ] `KeyType` struct with level mapping
- [ ] `Key` struct with groups
- [ ] `KeyGroup` struct with levels
- [ ] `KeyLevel` struct with keysyms
- [ ] `LED` struct for indicators

### 2.2 Keymap Methods
- [ ] `keymap.MinKeycode() Keycode`
- [ ] `keymap.MaxKeycode() Keycode`
- [ ] `keymap.KeyGetName(Keycode) string`
- [ ] `keymap.KeyByName(string) Keycode`
- [ ] `keymap.NumGroups() int`
- [ ] `keymap.GroupName(Group) string`
- [ ] `keymap.NumTypes() int`
- [ ] `keymap.ModGetIndex(string) int`

### 2.3 Manual Keymap Construction
- [ ] Builder pattern for testing
- [ ] Hardcoded US QWERTY for initial testing

**Deliverable**: Can create keymap structures programmatically.

---

## Phase 3: State

**Goal**: Keyboard state tracking and key translation.

### 3.1 State Structure
- [ ] `State` struct
- [ ] `keymap.NewState() *State`
- [ ] Three-component modifier state (base, latched, locked)
- [ ] Three-component group state

### 3.2 State Update
- [ ] `state.UpdateMask(baseMods, latchedMods, lockedMods, baseGroup, latchedGroup, lockedGroup)`
- [ ] `state.UpdateKey(keycode, direction)` (optional, for evdev)
- [ ] Effective modifier/group calculation

### 3.3 Key Translation
- [ ] `state.KeyGetSyms(keycode) []Keysym`
- [ ] `state.KeyGetOneSym(keycode) Keysym`
- [ ] `state.KeyGetUTF32(keycode) rune`
- [ ] `state.KeyGetUTF8(keycode) string`
- [ ] Level resolution through key types

### 3.4 Modifier Queries
- [ ] `state.ModNameIsActive(name, type) bool`
- [ ] `state.ModIndexIsActive(index, type) bool`
- [ ] `state.SerializeMods(components) ModMask`
- [ ] `state.SerializeGroup(components) Group`

### 3.5 Tests
- [ ] State update tests
- [ ] Key translation with modifiers
- [ ] ALPHABETIC type (Shift + Caps Lock interaction)
- [ ] TWO_LEVEL type
- [ ] KEYPAD type (Num Lock)

**Deliverable**: Can translate keycodes to keysyms using hardcoded keymap.

---

## Phase 4: Parser - Lexer

**Goal**: Tokenize XKB text format.

### 4.1 Token Types
- [ ] Keywords (xkb_keymap, xkb_keycodes, type, key, include, etc.)
- [ ] Identifiers
- [ ] Strings (double-quoted)
- [ ] Numbers (decimal, hex, octal)
- [ ] Operators and punctuation

### 4.2 Lexer Implementation
- [ ] `Lexer` struct with source and position
- [ ] `lexer.Next() Token`
- [ ] `lexer.Peek() Token`
- [ ] Line/column tracking for errors
- [ ] Comment handling (// and /* */)

### 4.3 Tests
- [ ] Token recognition tests
- [ ] Number parsing (decimal, hex)
- [ ] String escapes
- [ ] Comment skipping

**Deliverable**: Can tokenize XKB keymap files.

---

## Phase 5: Parser - Keycodes Section

**Goal**: Parse `xkb_keycodes` section.

### 5.1 Keycodes Parser
- [ ] Parse keycode assignments: `<TLDE> = 49;`
- [ ] Parse key name aliases: `alias <ALGR> = <RALT>;`
- [ ] Parse indicator definitions: `indicator 1 = "Caps Lock";`
- [ ] Parse min/max keycode
- [ ] Handle include directives (basic)

### 5.2 Integration
- [ ] Wire parser to Keymap structure
- [ ] Populate keycode ↔ name maps

### 5.3 Tests
- [ ] Parse evdev keycodes
- [ ] Parse aliases
- [ ] Verify keycode ranges

**Deliverable**: Can parse keycodes section from real keymap.

---

## Phase 6: Parser - Types Section

**Goal**: Parse `xkb_types` section.

### 6.1 Types Parser
- [ ] Parse type definitions
- [ ] Parse modifier specifications: `modifiers = Shift + Lock;`
- [ ] Parse level mappings: `map[Shift] = Level2;`
- [ ] Parse preserve statements
- [ ] Parse level names

### 6.2 Virtual Modifiers
- [ ] Parse virtual_modifiers declaration
- [ ] Track virtual → real modifier mapping

### 6.3 Tests
- [ ] Parse ONE_LEVEL, TWO_LEVEL, ALPHABETIC, KEYPAD
- [ ] Parse types with preserve
- [ ] Parse virtual modifier references

**Deliverable**: Can parse types section from real keymap.

---

## Phase 7: Parser - Symbols Section

**Goal**: Parse `xkb_symbols` section.

### 7.1 Symbols Parser
- [ ] Parse key definitions: `key <AD01> { [q, Q] };`
- [ ] Parse group definitions: `key <AD01> { [q, Q], [й, Й] };`
- [ ] Parse type override: `key <AD01> { type = "FOUR_LEVEL", ... };`
- [ ] Parse keysym names and Unicode notation
- [ ] Parse actions (basic)
- [ ] Parse modifier_map statements
- [ ] Parse group names

### 7.2 Include Resolution
- [ ] Parse include statements
- [ ] Merge mode handling (override, augment, replace)
- [ ] Path resolution (for file-based includes)

### 7.3 Tests
- [ ] Parse basic US layout
- [ ] Parse multi-group layout
- [ ] Parse with includes

**Deliverable**: Can parse symbols section from real keymap.

---

## Phase 8: Parser - Compat Section

**Goal**: Parse `xkb_compat` section.

### 8.1 Compat Parser
- [ ] Parse interpret statements
- [ ] Parse action specifications
- [ ] Parse indicator mappings
- [ ] Parse group compatibility

### 8.2 Compat Processing
- [ ] Apply interprets to keys
- [ ] Resolve virtual modifiers
- [ ] Set up modifier actions

### 8.3 Tests
- [ ] Parse standard compat section
- [ ] Verify modifier key setup

**Deliverable**: Can parse complete keymap file.

---

## Phase 9: Full Parser Integration

**Goal**: Parse complete keymaps end-to-end.

### 9.1 Top-Level Parser
- [ ] Parse `xkb_keymap { ... }` wrapper
- [ ] Section ordering flexibility
- [ ] Error recovery and reporting

### 9.2 API
- [ ] `ctx.NewKeymapFromString(text, format) (*Keymap, error)`
- [ ] `ctx.NewKeymapFromFile(path, format) (*Keymap, error)`
- [ ] Format validation (only XKB_KEYMAP_FORMAT_TEXT_V1)

### 9.3 Integration Tests
- [ ] Parse keymap from Wayland compositor
- [ ] Parse xkeyboard-config layouts
- [ ] Round-trip: parse → serialize → parse (optional)

**Deliverable**: Can load real keymaps from Wayland.

---

## Phase 10: Compose Tables

**Goal**: Dead key and compose sequence support.

### 10.1 Compose File Parser
- [ ] Parse Compose file format
- [ ] Handle include directives
- [ ] Build trie from sequences

### 10.2 Compose Table
- [ ] `ComposeTable` struct
- [ ] `ctx.NewComposeTableFromLocale(locale) (*ComposeTable, error)`
- [ ] `ctx.NewComposeTableFromFile(path) (*ComposeTable, error)`
- [ ] Locale file search paths

### 10.3 Compose State
- [ ] `ComposeState` struct
- [ ] `table.NewState() *ComposeState`
- [ ] `state.Feed(keysym) ComposeStatus`
- [ ] `state.GetStatus() ComposeStatus`
- [ ] `state.GetOneSym() Keysym`
- [ ] `state.GetUTF8() string`
- [ ] `state.Reset()`

### 10.4 Tests
- [ ] Dead acute + a = á
- [ ] Multi_key sequences
- [ ] Cancelled sequences
- [ ] Unknown sequences

**Deliverable**: Full dead key support.

---

## Phase 11: Polish and Optimization

**Goal**: Production readiness.

### 11.1 Performance
- [ ] Profile and optimize hot paths
- [ ] Consider string interning for names
- [ ] Benchmark against libxkbcommon

### 11.2 Completeness
- [ ] All xkb_state_* functions
- [ ] LED state tracking
- [ ] Key repeat information

### 11.3 Documentation
- [ ] GoDoc comments
- [ ] Usage examples
- [ ] Migration guide from libxkbcommon

### 11.4 Testing
- [ ] Fuzz testing for parser
- [ ] Cross-reference with libxkbcommon output
- [ ] Edge case coverage

**Deliverable**: Production-ready library.

---

## Milestones Summary

| Phase | Milestone | Enables |
|-------|-----------|---------|
| 1-3 | Basic key translation | Manual keymap construction |
| 4-9 | Parser complete | Load real keymaps |
| 10 | Compose support | International input |
| 11 | Production ready | Replace libxkbcommon |

---

## Testing Strategy

### Unit Tests
Each component has isolated tests with known inputs/outputs.

### Integration Tests
End-to-end tests with real keymap data:
```go
func TestRealKeymap(t *testing.T) {
    // Load keymap dumped from Wayland session
    data, _ := os.ReadFile("testdata/wayland_keymap.xkb")
    km, err := ctx.NewKeymapFromString(data, xkb.KeymapFormatTextV1)
    // ...
}
```

### Compatibility Tests
Compare output with libxkbcommon:
```bash
# Dump reference output
xkbcli compile-keymap --layout us | ./compare-tool
```

### Fuzz Tests
Parser fuzzing to find edge cases:
```go
func FuzzParser(f *testing.F) {
    f.Fuzz(func(t *testing.T, data []byte) {
        ctx.NewKeymapFromString(data, xkb.KeymapFormatTextV1)
    })
}
```
