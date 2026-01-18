# XKB-Go References

## Primary Sources

### libxkbcommon Documentation

- **Main documentation**: https://xkbcommon.org/doc/current/
- **Quick Guide**: https://xkbcommon.org/doc/current/md_doc_quick_guide.html
- **Introduction to XKB**: https://xkbcommon.org/doc/current/xkb-intro.html
- **Keymap Text Format V1/V2**: https://xkbcommon.org/doc/current/keymap-text-format-v1-v2.html
- **Compose and Dead Keys**: https://xkbcommon.org/doc/current/group__compose.html

### libxkbcommon Source Code

- **GitHub Repository**: https://github.com/xkbcommon/libxkbcommon
- **Parser (yacc grammar)**: https://github.com/xkbcommon/libxkbcommon/blob/master/src/xkbcomp/parser.y
- **Lexer**: https://github.com/xkbcommon/libxkbcommon/blob/master/src/xkbcomp/scanner.c
- **State implementation**: https://github.com/xkbcommon/libxkbcommon/blob/master/src/state.c
- **Compose implementation**: https://github.com/xkbcommon/libxkbcommon/blob/master/src/compose/

### XKB Protocol Specification

- **X11 XKB Protocol**: https://www.x.org/releases/X11R7.6/doc/kbproto/xkbproto.html
- **X.Org XKB Config Guide**: https://www.x.org/releases/X11R7.5/doc/input/XKB-Config.html

### Wayland

- **Wayland Book - XKB**: https://wayland-book.com/seat/xkb.html
- **wl_keyboard protocol**: https://wayland.freedesktop.org/docs/html/apa.html#protocol-spec-wl_keyboard

---

## Keysym Definitions

### X11 Header Files

The canonical source for keysym definitions:

- **keysymdef.h**: Core keysyms (Latin, function keys, modifiers)
  - https://gitlab.freedesktop.org/xorg/proto/xorgproto/-/blob/master/include/X11/keysymdef.h

- **XF86keysym.h**: Multimedia and special keys
  - https://gitlab.freedesktop.org/xorg/proto/xorgproto/-/blob/master/include/X11/XF86keysym.h

### Keysym Ranges

```
0x0000 - 0x00ff  Latin-1 (ISO 8859-1)
0x0100 - 0x01ff  Latin-2 (ISO 8859-2)
0x0200 - 0x02ff  Latin-3 (ISO 8859-3)
0x0300 - 0x03ff  Latin-4 (ISO 8859-4)
0x0400 - 0x04ff  Kana (Japanese)
0x0500 - 0x05ff  Arabic
0x0600 - 0x06ff  Cyrillic
0x0700 - 0x07ff  Greek
0x0800 - 0x08ff  Technical
0x0900 - 0x09ff  Special
0x0a00 - 0x0aff  Publishing
0x0b00 - 0x0bff  APL
0x0c00 - 0x0cff  Hebrew
0x0d00 - 0x0dff  Thai
0x0e00 - 0x0eff  Korean
0x1000 - 0x10ff  Latin-9 (ISO 8859-15)
0x1200 - 0x12ff  Latin-8 (ISO 8859-14)
0x1300 - 0x13ff  Latin-10 (ISO 8859-16)
0x14xx - 0x1eff  Other character sets
0x20ac          Euro sign
0xfd00 - 0xfdff  3270 terminal
0xfe00 - 0xfeff  Dead keys, special
0xff00 - 0xffff  Function keys, modifiers, misc
0x01000000+     Unicode (codepoint + 0x01000000)
```

---

## Compose File Format

### Compose(5) Man Page

```bash
man Compose
```

Or online: https://linux.die.net/man/5/compose

### Compose File Locations

System compose files (from libX11):
```
/usr/share/X11/locale/<locale>/Compose
/usr/share/X11/locale/en_US.UTF-8/Compose
```

User compose file:
```
~/.XCompose
```

### Compose File Syntax

```
# Comment
<dead_acute> <a> : "á" aacute    # Sequence with result
<Multi_key> <c> <,> : "ç"        # Multi-key sequence
include "%L"                      # Include locale default
```

Directives:
- `include "path"` - Include another file
- `%H` - User's home directory
- `%L` - System locale compose file
- `%S` - System compose directory

---

## xkeyboard-config

The standard keyboard configuration database.

- **Repository**: https://gitlab.freedesktop.org/xkeyboard-config/xkeyboard-config
- **Rules**: Define how RMLVO maps to keymap components
- **Symbols**: Layout definitions
- **Types**: Key type definitions
- **Keycodes**: Keycode assignments

### File Locations

```
/usr/share/X11/xkb/
├── rules/          # RMLVO → component mapping
├── symbols/        # Layout definitions (us, de, ru, ...)
├── types/          # Key type definitions
├── keycodes/       # Keycode assignments
├── compat/         # Compatibility definitions
└── geometry/       # Physical keyboard layouts (ignored)
```

---

## Related Go Projects

### Wayland

- **go-wayland**: https://github.com/rajveermalviya/go-wayland
  - Pure Go Wayland client

- **MatthiasKunnen/go-wayland**: https://github.com/MatthiasKunnen/go-wayland
  - Wayland bindings used by wl-polkit-agent

- **neurlang/wayland**: https://github.com/neurlang/wayland
  - Pure Go Wayland (still uses libxkbcommon for keyboard)

### XKB Wrappers

- **gioui.org/app/internal/xkb**: https://pkg.go.dev/gioui.org/app/internal/xkb
  - CGO bindings for Gio UI

- **swaywm/go-wlroots/xkb**: https://pkg.go.dev/github.com/swaywm/go-wlroots/xkb
  - CGO bindings for wlroots

---

## Useful Tools

### xkbcli

Command-line tools from libxkbcommon:

```bash
# Compile and dump a keymap
xkbcli compile-keymap --layout us

# Interactive keymap tester
xkbcli interactive-wayland
xkbcli interactive-evdev

# List available layouts
xkbcli list

# Process compose file
xkbcli compile-compose
```

### xev

X11 key event viewer:
```bash
xev -event keyboard
```

### wev

Wayland event viewer:
```bash
wev
```

### xkbcomp

Legacy X11 keymap compiler:
```bash
xkbcomp $DISPLAY output.xkb
```

---

## Test Data Sources

### Keymap Dumps

Get a real keymap from your Wayland session:
```bash
# From sway/wlroots compositor
xkbcli compile-keymap --layout us --variant intl > testdata/us_intl.xkb

# Dump current X11 keymap
xkbcomp $DISPLAY testdata/current.xkb
```

### Compose Files

```bash
# Copy system compose file
cp /usr/share/X11/locale/en_US.UTF-8/Compose testdata/compose_en_US
```

---

## Technical Articles

- **A Simple Guide to XKB**: https://medium.com/@damko/a-simple-humble-but-comprehensive-guide-to-xkb-for-linux-6f1ad5e13450
- **On the Road to Pure Go X11 GUIs**: https://p.janouch.name/article-xgb.html
- **XKB Configuration Files**: https://www.charvolant.org/doug/xkb/html/node5.html
- **Arch Wiki - X Keyboard Extension**: https://wiki.archlinux.org/title/X_keyboard_extension

---

## Standards

- **ISO/IEC 9995**: Keyboard layouts for text and office systems
- **Unicode**: https://unicode.org/
- **ISO 8859**: 8-bit character encodings (Latin-1 through Latin-16)
