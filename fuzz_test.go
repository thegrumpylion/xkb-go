package xkb

import (
	"os"
	"testing"
)

// FuzzLexer fuzzes the XKB lexer to find panics or crashes.
func FuzzLexer(f *testing.F) {
	// Seed corpus from test cases
	seeds := []string{
		`xkb_keymap { }`,
		`xkb_keycodes "test" { minimum = 8; maximum = 255; }`,
		`<AD01> = 24;`,
		`key <AD01> { [ q, Q ] };`,
		`type "ONE_LEVEL" { modifiers = none; }`,
		`modifier_map Shift { <LFSH> };`,
		`"string with \n escape"`,
		`0xff08`,
		`0755`,
		`// comment`,
		`/* multi
		   line
		   comment */`,
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	// Add system keymap as seed if available
	if data, err := os.ReadFile("testdata/us_intl.xkb"); err == nil {
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// We just want to ensure the lexer doesn't panic
		lexer := NewLexer(data)
		for {
			tok := lexer.NextToken()
			if tok.Type == TokenEOF || tok.Type == TokenError {
				break
			}
		}
	})
}

// FuzzParser fuzzes the XKB parser to find panics or crashes.
func FuzzParser(f *testing.F) {
	// Seed corpus with various valid and edge-case keymaps
	seeds := []string{
		// Minimal valid keymap
		`xkb_keymap {
			xkb_keycodes "test" { minimum = 8; maximum = 255; };
			xkb_types "test" {};
			xkb_compat "test" {};
			xkb_symbols "test" {};
		};`,

		// Keymap with keycodes
		`xkb_keymap {
			xkb_keycodes "test" {
				minimum = 8;
				maximum = 255;
				<AD01> = 24;
				<ESC> = 9;
				alias <ALGR> = <RALT>;
			};
			xkb_types "test" {};
			xkb_compat "test" {};
			xkb_symbols "test" {};
		};`,

		// Keymap with types
		`xkb_keymap {
			xkb_keycodes "test" { minimum = 8; maximum = 255; };
			xkb_types "test" {
				virtual_modifiers NumLock, Alt;
				type "TWO_LEVEL" {
					modifiers = Shift;
					map[Shift] = Level2;
				};
			};
			xkb_compat "test" {};
			xkb_symbols "test" {};
		};`,

		// Keymap with symbols
		`xkb_keymap {
			xkb_keycodes "test" { minimum = 8; maximum = 255; <AD01> = 24; };
			xkb_types "test" {
				type "ONE_LEVEL" { modifiers = none; };
			};
			xkb_compat "test" {};
			xkb_symbols "test" {
				key <AD01> { [ q, Q ] };
			};
		};`,

		// Empty input
		``,

		// Just braces
		`{}`,

		// Invalid but shouldn't crash
		`xkb_keymap`,
		`xkb_keymap {`,
		`random garbage`,
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	// Add system keymap as seed if available
	if data, err := os.ReadFile("testdata/us_intl.xkb"); err == nil {
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// We just want to ensure the parser doesn't panic
		// Errors are expected for malformed input
		p := NewParser(data)
		_, _ = p.Parse()
	})
}

// FuzzKeymapFromString fuzzes the full keymap parsing path via Context.
func FuzzKeymapFromString(f *testing.F) {
	// Add seeds similar to FuzzParser
	seeds := []string{
		`xkb_keymap {
			xkb_keycodes "test" { minimum = 8; maximum = 255; };
			xkb_types "test" {};
			xkb_compat "test" {};
			xkb_symbols "test" {};
		};`,
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	if data, err := os.ReadFile("testdata/us_intl.xkb"); err == nil {
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := NewContext(ContextNoFlags)
		_, _ = ctx.NewKeymapFromString(data, KeymapFormatTextV1)
	})
}

// FuzzComposeState fuzzes the compose state machine with arbitrary keysym sequences.
func FuzzComposeState(f *testing.F) {
	// Seed corpus with common keysym sequences
	seeds := [][]byte{
		// dead_acute (0xfe51) + a (0x61) → á
		{0xfe, 0x51, 0x00, 0x61},
		// Multi_key (0xff20) + a (0x61) + e (0x65) → æ
		{0xff, 0x20, 0x00, 0x61, 0x00, 0x65},
		// Random keysyms
		{0x00, 0x61, 0x00, 0x62, 0x00, 0x63},
		// Empty
		{},
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Create a test compose table
		table := TestComposeTable()
		if table == nil {
			return
		}
		state := table.NewState(0)

		// Feed keysyms from the data (2 bytes per keysym, big-endian)
		for i := 0; i+1 < len(data); i += 2 {
			ks := Keysym(uint32(data[i])<<8 | uint32(data[i+1]))
			_ = state.Feed(ks)
			_ = state.GetStatus()
			_ = state.GetOneSym()
			_ = state.GetUTF8()
		}

		// Reset and try again
		state.Reset()
		for i := 0; i+1 < len(data); i += 2 {
			ks := Keysym(uint32(data[i])<<8 | uint32(data[i+1]))
			_ = state.Feed(ks)
		}
	})
}

// FuzzKeysymFromName fuzzes keysym name lookup.
func FuzzKeysymFromName(f *testing.F) {
	seeds := []string{
		"a", "A", "Return", "Escape", "Shift_L",
		"dead_acute", "Multi_key", "BackSpace",
		"space", "exclam", "at",
		"", "NotAKeysym", "UPPERCASE",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, name string) {
		// Should not panic
		_ = KeysymFromName(name, KeysymNameNoFlags)
		_ = KeysymFromName(name, KeysymNameCaseInsensitive)
	})
}

// FuzzKeysymToUTF32 fuzzes keysym to unicode conversion.
func FuzzKeysymToUTF32(f *testing.F) {
	seeds := []uint32{
		0x0000, 0x0020, 0x0041, 0x0061, 0x007e, 0x00ff,
		0xff08, 0xff0d, 0xffe1,
		0x01000000, 0x0100ffff, 0x0110ffff,
		0xffffffff,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, ks uint32) {
		// Should not panic
		_ = KeysymToUTF32(Keysym(ks))
		_ = KeysymToUTF8(Keysym(ks))
	})
}
