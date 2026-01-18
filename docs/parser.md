# XKB Parser Design

## Overview

The parser converts XKB text format (v1) into a `Keymap` structure. This is the format used by Wayland compositors when sending keymap data via `wl_keyboard.keymap`.

## XKB Text Format Structure

A complete keymap looks like:

```
xkb_keymap {
    xkb_keycodes "evdev" { ... };
    xkb_types "complete" { ... };
    xkb_compat "complete" { ... };
    xkb_symbols "pc+us" { ... };
    xkb_geometry "pc(pc105)" { ... };  // Ignored
};
```

## Parser Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Source     │────▶│    Lexer     │────▶│    Parser    │
│   ([]byte)   │     │  (Tokenize)  │     │   (Parse)    │
└──────────────┘     └──────────────┘     └──────────────┘
                            │                     │
                            ▼                     ▼
                      Token Stream            *Keymap
```

### Lexer

The lexer converts source text into a stream of tokens.

**Token Types:**
- `TokenEOF` - End of file
- `TokenIdent` - Identifier (e.g., `xkb_keymap`, `Shift`, `AD01`)
- `TokenString` - Quoted string (e.g., `"evdev"`, `"us"`)
- `TokenNumber` - Numeric literal (decimal, hex, octal)
- `TokenKeycode` - Keycode name in angle brackets (e.g., `<AD01>`)
- Punctuation: `{`, `}`, `[`, `]`, `(`, `)`, `;`, `,`, `=`, `+`, `-`, `!`, `~`

**Lexer Features:**
- Line/column tracking for error messages
- Comment handling (`//` and `/* */`)
- String escape sequences
- Hex (`0x1F`) and octal (`017`) number literals

### Parser

The parser consumes tokens and builds the `Keymap` structure.

**Parsing Strategy:**
- Recursive descent parser
- Each section has its own parsing function
- Error recovery: skip to next `;` or `}` on error

---

## Section: xkb_keycodes

Maps hardware scancodes to symbolic key names.

```
xkb_keycodes "evdev" {
    minimum = 8;
    maximum = 255;

    <ESC>  = 9;
    <AE01> = 10;
    <AD01> = 24;

    alias <ALGR> = <RALT>;

    indicator 1 = "Caps Lock";
    indicator 2 = "Num Lock";
};
```

**Parsed into:**
- `keymap.minKeycode`, `keymap.maxKeycode`
- `keymap.keycodeNames` map
- `keymap.keycodesByName` map
- LED indicator names

---

## Section: xkb_types

Defines key types that map modifier combinations to shift levels.

```
xkb_types "complete" {
    virtual_modifiers NumLock, Alt, LevelThree;

    type "ONE_LEVEL" {
        modifiers = none;
        level_name[Level1] = "Any";
    };

    type "TWO_LEVEL" {
        modifiers = Shift;
        map[Shift] = Level2;
        level_name[Level1] = "Base";
        level_name[Level2] = "Shift";
    };

    type "ALPHABETIC" {
        modifiers = Shift + Lock;
        map[Shift] = Level2;
        map[Lock] = Level2;
        map[Shift+Lock] = Level1;
        level_name[Level1] = "Base";
        level_name[Level2] = "Caps";
    };
};
```

**Parsed into:**
- `keymap.virtualMods` map
- `keymap.types` map with `KeyType` structs

---

## Section: xkb_symbols

Assigns keysyms to keys, organized by groups and levels.

```
xkb_symbols "us" {
    name[Group1] = "English (US)";

    key <AD01> { [ q, Q ] };
    key <AD02> { [ w, W ] };

    // Multi-group key
    key <AD01> {
        type = "ALPHABETIC",
        symbols[Group1] = [ q, Q ],
        symbols[Group2] = [ Cyrillic_shorti, Cyrillic_SHORTI ]
    };

    // With actions
    key <LFSH> {
        type = "ONE_LEVEL",
        symbols[Group1] = [ Shift_L ],
        actions[Group1] = [ SetMods(modifiers=Shift) ]
    };

    modifier_map Shift { <LFSH>, <RTSH> };
    modifier_map Lock { <CAPS> };
};
```

**Parsed into:**
- `keymap.groupNames`
- `keymap.keys` map with groups, levels, keysyms
- Modifier mappings

---

## Section: xkb_compat

Defines compatibility mappings and interpret statements.

```
xkb_compat "complete" {
    virtual_modifiers NumLock, Alt;

    interpret Shift_L {
        action = SetMods(modifiers=Shift);
    };

    interpret Num_Lock {
        action = LockMods(modifiers=NumLock);
    };

    indicator "Caps Lock" {
        modifiers = Lock;
    };
};
```

**Parsed into:**
- Interpret rules applied to keys
- LED indicator definitions

---

## Section: xkb_geometry (Ignored)

Physical keyboard layout for visualization. We parse but ignore.

---

## Token Definitions

```go
type TokenType int

const (
    TokenEOF TokenType = iota
    TokenError

    // Literals
    TokenIdent      // identifier
    TokenString     // "quoted string"
    TokenNumber     // 123, 0x1F, 017
    TokenKeycode    // <AD01>

    // Keywords (context-sensitive, parsed as TokenIdent)
    // xkb_keymap, xkb_keycodes, xkb_types, xkb_compat, xkb_symbols
    // type, key, include, virtual_modifiers, interpret, etc.

    // Punctuation
    TokenLBrace     // {
    TokenRBrace     // }
    TokenLBracket   // [
    TokenRBracket   // ]
    TokenLParen     // (
    TokenRParen     // )
    TokenSemicolon  // ;
    TokenComma      // ,
    TokenEquals     // =
    TokenPlus       // +
    TokenMinus      // -
    TokenBang       // !
    TokenTilde      // ~
    TokenDot        // .
)
```

---

## Error Handling

Errors include source location for debugging:

```go
type ParseError struct {
    File    string
    Line    int
    Col     int
    Message string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Col, e.Message)
}
```

**Error Recovery:**
- On error, skip tokens until `;` or `}`
- Continue parsing to report multiple errors
- Return first error but try to find more

---

## Include Handling

For Wayland, keymaps are self-contained (no includes needed). But the format supports:

```
include "us(basic)"
```

**Strategy:**
- Parse include statements
- For Wayland use: warn but continue (keymap should be complete)
- For RMLVO use (future): resolve from include paths

---

## Implementation Order

1. **Lexer** - Tokenize source
2. **xkb_keycodes** - Simple, good starting point
3. **xkb_types** - Needed for key translation
4. **xkb_symbols** - Core key definitions
5. **xkb_compat** - Modifier actions
6. **Integration** - Wire to `Context.NewKeymapFromString`

---

## Test Strategy

### Unit Tests
- Lexer: tokenize known inputs
- Each section parser: parse snippets

### Integration Tests
- Parse complete keymaps dumped from Wayland
- Verify key translation matches libxkbcommon

### Test Data
```bash
# Dump current keymap
xkbcli compile-keymap --layout us > testdata/us.xkb
xkbcli compile-keymap --layout de > testdata/de.xkb
xkbcli compile-keymap --layout us --variant intl > testdata/us_intl.xkb
```

---

## Example: Minimal US Keymap

```
xkb_keymap {
    xkb_keycodes "test" {
        minimum = 8;
        maximum = 255;
        <AD01> = 24;
    };

    xkb_types "test" {
        type "ONE_LEVEL" {
            modifiers = none;
        };
    };

    xkb_compat "test" {
    };

    xkb_symbols "test" {
        key <AD01> { [ q ] };
    };
};
```
