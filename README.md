# xkb-go

A pure Go implementation of the XKB (X Keyboard Extension) library, compatible with libxkbcommon.

## Why?

libxkbcommon is the standard library for keyboard handling in Wayland and modern X11 applications. However, using it from Go requires either:

- **CGO bindings**: Breaks cross-compilation, requires C toolchain
- **purego**: Still requires libxkbcommon.so at runtime

xkb-go provides the same functionality in pure Go:

- No CGO, no C dependencies
- Cross-compile to any platform
- Single static binary
- Native Go error handling and types

## Status

**Work in Progress** - Not yet ready for production use.

See [docs/roadmap.md](docs/roadmap.md) for implementation progress.

## Usage

```go
package main

import (
    "fmt"
    "github.com/thegrumpylion/xkb-go"
)

func main() {
    // Create context
    ctx := xkb.NewContext(xkb.ContextNoFlags)

    // Load keymap from Wayland compositor
    keymap, err := ctx.NewKeymapFromString(keymapData, xkb.KeymapFormatTextV1)
    if err != nil {
        panic(err)
    }

    // Create state for tracking keyboard
    state := keymap.NewState()

    // Update modifier state (from wl_keyboard.modifiers)
    state.UpdateMask(shiftPressed, 0, capsLockOn, 0, 0, 0)

    // Translate keycode to keysym
    sym := state.KeyGetOneSym(keycode)

    // Get Unicode character
    char := state.KeyGetUTF32(keycode)
    fmt.Printf("Key: %s (%c)\n", xkb.KeysymGetName(sym), char)
}
```

## Features

- [x] Keysym utilities (name lookup, Unicode conversion)
- [ ] Keymap parsing (XKB text format v1)
- [ ] Keyboard state tracking
- [ ] Modifier handling
- [ ] Compose/dead key support

## Documentation

- [Architecture](docs/architecture.md) - Design and data structures
- [Roadmap](docs/roadmap.md) - Implementation phases and progress
- [References](docs/references.md) - Specifications and resources

## Compatibility

API is designed to be similar to libxkbcommon for easy migration:

| libxkbcommon | xkb-go |
|--------------|--------|
| `xkb_context_new()` | `xkb.NewContext()` |
| `xkb_keymap_new_from_string()` | `ctx.NewKeymapFromString()` |
| `xkb_state_new()` | `keymap.NewState()` |
| `xkb_state_key_get_one_sym()` | `state.KeyGetOneSym()` |
| `xkb_state_key_get_utf32()` | `state.KeyGetUTF32()` |

## License

MIT License - see [LICENSE](LICENSE)

## Related Projects

- [libxkbcommon](https://xkbcommon.org/) - The C reference implementation
- [wlui](https://github.com/thegrumpylion/wlui) - Go HTML/CSS UI library
- [wl-polkit-agent](https://github.com/greatliontech/wl-polkit-agent) - Wayland polkit agent (uses this library)
