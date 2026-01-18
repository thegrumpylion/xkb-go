# XKB-Go Roadmap

## Core

- [x] Types (`Keysym`, `Keycode`, `ModMask`, `Level`, `Group`)
- [x] Context with include paths and logging
- [x] Keymap structure and query methods
- [x] State machine (`UpdateMask`, `UpdateKey`, `KeyGetOneSym`, `KeyGetUTF8`)
- [x] Error types

## Keysym

- [x] Keysym <-> UTF-8/UTF-32 conversion
- [x] Keysym <-> name conversion
- [x] Full keysym tables (generated from X11 headers)

## XKB Parser

- [x] Lexer (all token types, comments, escapes, hex/octal)
- [x] `xkb_keycodes` section
- [x] `xkb_types` section
- [x] `xkb_compat` section
- [x] `xkb_symbols` section
- [x] `xkb_geometry` section (parsed and ignored)
- [x] Strict validation (matching libxkbcommon)
- [x] `NewKeymapFromString()`

## Compose

- [x] Compose file parser with include support
- [x] Trie-based compose table
- [x] `NewComposeTableFromLocale()`
- [x] `NewComposeTableFromFile()`
- [x] Compose state machine (`Feed`, `GetStatus`, `GetOneSym`, `GetUTF8`)

## Testing

- [x] Unit tests (80%+ coverage)
- [x] Fuzz tests (lexer, parser, compose)
- [x] Edge case tests
- [x] Benchmarks
- [x] Tested with real system keymaps and compose files

## Not Implemented

- [ ] RMLVO compilation (`NewKeymapFromNames`)
- [ ] `NewKeymapFromFile`
- [ ] `Keymap.GetAsString()`
