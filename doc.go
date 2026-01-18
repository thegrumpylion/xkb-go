// Package xkb provides a pure Go implementation of the X Keyboard Extension (XKB).
//
// This package enables keyboard handling for Wayland compositors and other applications
// that need XKB functionality without CGO dependencies. It supports:
//
//   - Loading keymaps from RMLVO names (rules, model, layout, variant, options)
//   - Parsing XKB text format keymaps
//   - Key translation with modifier state tracking
//   - Compose sequence handling
//   - LED indicator state
//
// Basic usage:
//
//	ctx := xkb.NewContext(context.Background(), xkb.ContextNoFlags)
//	keymap, err := ctx.NewKeymapFromNames(&xkb.RuleNames{
//	    Layout: "us",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	state := keymap.NewState()
//
//	// Translate a key press
//	sym := state.KeyGetOneSym(keycode)
//	utf8 := state.KeyGetUTF8(keycode)
package xkb
