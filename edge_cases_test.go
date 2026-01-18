package xkb

import (
	"strings"
	"testing"
)

// Edge case tests for comprehensive coverage

// TestLexerEdgeCases tests lexer edge cases
func TestLexerEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty input", ""},
		{"only whitespace", "   \t\n\r   "},
		{"only comments", "// comment\n/* block */"},
		{"deeply nested braces", "{{{{{{}}}}}}"},
		{"many semicolons", ";;;;;;;"},
		{"mixed operators", "=+[]{}();,"},
		{"long identifier", strings.Repeat("a", 1000)},
		{"long string", `"` + strings.Repeat("x", 1000) + `"`},
		{"string with all escapes", `"\n\t\r\\\"\000\xff"`},
		{"invalid escape", `"\z"`},
		{"unterminated string", `"unterminated`},
		{"unterminated block comment", "/* never closed"},
		{"keycode at end", "<ABC"},
		{"keycode variations", "<A> <AB> <ABC> <ABCD> <AB01>"},
		{"numbers", "0 1 255 0x0 0xff 0xFF 0777 00"},
		{"hex upper lower", "0xABCDEF 0xabcdef"},
		{"large hex number", "0xFFFFFFFF"},
		{"negative looking", "-1"},
		{"special chars in string", `"\x00\x01\x1f"`},
		{"unicode in input", "// 日本語コメント"},
		{"consecutive strings", `"a""b""c"`},
		{"newline in weird places", "key\n<\nAD01\n>\n{\n}\n;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer([]byte(tt.input))
			// Should not panic, collect all tokens
			var tokens []Token
			for {
				tok := lexer.NextToken()
				tokens = append(tokens, tok)
				if tok.Type == TokenEOF || tok.Type == TokenError {
					break
				}
				if len(tokens) > 10000 {
					t.Fatal("Too many tokens, possible infinite loop")
				}
			}
		})
	}
}

// TestParserEdgeCases tests parser edge cases
func TestParserEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			"minimal valid",
			`xkb_keymap { xkb_keycodes "a" { minimum=8; maximum=255; }; xkb_types "a" {}; xkb_compat "a" {}; xkb_symbols "a" {}; };`,
			false,
		},
		{
			"empty sections",
			`xkb_keymap { xkb_keycodes "" { }; xkb_types "" {}; xkb_compat "" {}; xkb_symbols "" {}; };`,
			false, // Parser is permissive
		},
		{
			"min equals max",
			`xkb_keymap { xkb_keycodes "a" { minimum=8; maximum=8; }; xkb_types "a" {}; xkb_compat "a" {}; xkb_symbols "a" {}; };`,
			false,
		},
		{
			"min greater than max",
			`xkb_keymap { xkb_keycodes "a" { minimum=255; maximum=8; }; xkb_types "a" {}; xkb_compat "a" {}; xkb_symbols "a" {}; };`,
			false, // Parser accepts, logic may be weird
		},
		{
			"zero keycodes",
			`xkb_keymap { xkb_keycodes "a" { minimum=0; maximum=0; }; xkb_types "a" {}; xkb_compat "a" {}; xkb_symbols "a" {}; };`,
			false,
		},
		{
			"very large keycode",
			`xkb_keymap { xkb_keycodes "a" { minimum=0; maximum=65535; <MAX> = 65535; }; xkb_types "a" {}; xkb_compat "a" {}; xkb_symbols "a" {}; };`,
			false,
		},
		{
			"duplicate keycode name",
			`xkb_keymap { xkb_keycodes "a" { minimum=8; maximum=255; <DUP> = 10; <DUP> = 11; }; xkb_types "a" {}; xkb_compat "a" {}; xkb_symbols "a" {}; };`,
			false, // Second one overwrites
		},
		{
			"type with many levels",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; };
				xkb_types "a" {
					type "EIGHT" {
						modifiers = Shift+Lock+Control+Mod1+Mod2+Mod3+Mod4+Mod5;
						map[Shift] = Level2;
						map[Lock] = Level3;
						map[Control] = Level4;
						map[Mod1] = Level5;
						map[Mod2] = Level6;
						map[Mod3] = Level7;
						map[Mod4] = Level8;
					};
				};
				xkb_compat "a" {};
				xkb_symbols "a" {};
			};`,
			false,
		},
		{
			"key with many symbols",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; <K> = 10; };
				xkb_types "a" { type "ONE_LEVEL" { modifiers = none; }; };
				xkb_compat "a" {};
				xkb_symbols "a" {
					key <K> { [ a, b, c, d, e, f, g, h ] };
				};
			};`,
			false,
		},
		{
			"multiple groups",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; <K> = 10; };
				xkb_types "a" { type "ONE_LEVEL" { modifiers = none; }; };
				xkb_compat "a" {};
				xkb_symbols "a" {
					name[Group1] = "English";
					name[Group2] = "Russian";
					key <K> {
						symbols[Group1] = [ a ],
						symbols[Group2] = [ Cyrillic_a ]
					};
				};
			};`,
			false,
		},
		{
			"virtual modifier chain",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; };
				xkb_types "a" { virtual_modifiers A, B, C, D, E, F, G, H; };
				xkb_compat "a" {};
				xkb_symbols "a" {};
			};`,
			false,
		},
		{
			"indicator with all options",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; indicator 1 = "Test"; };
				xkb_types "a" {};
				xkb_compat "a" {
					indicator "Test" {
						modifiers = Shift+Lock;
						groups = Group1+Group2;
					};
				};
				xkb_symbols "a" {};
			};`,
			false,
		},
		{
			"action in interpret",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; };
				xkb_types "a" {};
				xkb_compat "a" {
					interpret Shift_L {
						action = SetMods(modifiers=Shift, clearLocks);
					};
					interpret Control_L {
						action = LockMods(modifiers=Control);
					};
				};
				xkb_symbols "a" {};
			};`,
			false,
		},
		{
			"no closing brace",
			`xkb_keymap { xkb_keycodes "a" { minimum=8; maximum=255;`,
			true,
		},
		{
			"wrong section order",
			`xkb_keymap { xkb_symbols "a" {}; xkb_keycodes "a" { minimum=8; maximum=255; }; xkb_types "a" {}; xkb_compat "a" {}; };`,
			false, // Parser does not enforce section order
		},
		{
			"missing required section",
			`xkb_keymap { xkb_keycodes "a" { minimum=8; maximum=255; }; xkb_types "a" {}; };`,
			false, // Parser is permissive about missing sections
		},
		{
			"geometry section",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; };
				xkb_types "a" {};
				xkb_compat "a" {};
				xkb_symbols "a" {};
				xkb_geometry "test" {
					width = 470;
					height = 180;
				};
			};`,
			false, // Geometry is skipped
		},
		{
			"preserve in type",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; };
				xkb_types "a" {
					type "PRESERVE" {
						modifiers = Shift+Control;
						map[Shift] = Level2;
						preserve[Shift] = Shift;
					};
				};
				xkb_compat "a" {};
				xkb_symbols "a" {};
			};`,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			keymap, err := p.Parse()
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if keymap == nil && err == nil {
					t.Error("Got nil keymap without error")
				}
			}
		})
	}
}

// TestKeysymEdgeCases tests keysym edge cases
func TestKeysymEdgeCases(t *testing.T) {
	t.Run("boundary keysyms", func(t *testing.T) {
		// Test boundary values
		boundaries := []Keysym{
			0,                     // NoSymbol
			0x0020,                // space (lowest printable)
			0x007e,                // tilde (highest ASCII printable)
			0x007f,                // DEL (not printable)
			0x00a0,                // NBSP (start of Latin-1 supplement)
			0x00ff,                // ÿ (end of Latin-1)
			0x0100,                // Start of Latin Extended
			0xff08,                // BackSpace
			0xffff,                // End of legacy keysyms
			0x01000000,            // Start of Unicode keysyms
			0x01000041,            // Unicode 'A'
			0x0110ffff,            // End of valid Unicode keysyms
			0x01110000,            // Invalid (beyond Unicode)
			0xffffffff,            // Max uint32
		}

		for _, ks := range boundaries {
			// Should not panic
			name := KeysymGetName(ks)
			utf32 := KeysymToUTF32(ks)
			utf8 := KeysymToUTF8(ks)
			_ = name
			_ = utf32
			_ = utf8
		}
	})

	t.Run("unicode keysym range", func(t *testing.T) {
		// Unicode keysyms are 0x01000000 + codepoint
		tests := []struct {
			keysym Keysym
			want   rune
		}{
			{0x01000041, 'A'},      // ASCII
			{0x010000e4, 'ä'},      // Latin-1 Supplement
			{0x01002603, '☃'},      // Snowman
			{0x0101f600, 0x1f600},  // Emoji (grinning face)
			{0x0110ffff, 0x10ffff}, // Max valid Unicode
		}

		for _, tt := range tests {
			got := KeysymToUTF32(tt.keysym)
			if got != tt.want {
				t.Errorf("KeysymToUTF32(%#x) = %#x, want %#x", tt.keysym, got, tt.want)
			}
		}
	})

	t.Run("name lookup edge cases", func(t *testing.T) {
		// Empty and unusual names
		names := []string{
			"",
			"a",
			"A",
			"space",
			"SPACE",                            // Case sensitivity
			"dead_acute",
			"Dead_Acute",                       // Case variation
			strings.Repeat("x", 100),           // Long name
			"non_existent_keysym",
			"0x0020",                           // Hex format (not a name)
			"Return\n",                         // With newline
			" space ",                          // With surrounding whitespace
		}

		for _, name := range names {
			ks := KeysymFromName(name, KeysymNameNoFlags)
			ksCaseInsensitive := KeysymFromName(name, KeysymNameCaseInsensitive)
			_ = ks
			_ = ksCaseInsensitive
		}
	})
}

// TestStateEdgeCases tests state edge cases
func TestStateEdgeCases(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	t.Run("all modifiers combined", func(t *testing.T) {
		allMods := ModShift | ModLock | ModControl | ModMod1 | ModMod2 | ModMod3 | ModMod4 | ModMod5
		state.UpdateMask(allMods, allMods, allMods, 0, 0, 0)
		effective := state.SerializeMods(StateModEffective)
		if effective != allMods {
			t.Errorf("Effective mods = %d, want %d", effective, allMods)
		}
	})

	t.Run("group wraparound", func(t *testing.T) {
		// TestKeymap has 1 group, so all groups should wrap to 0
		for g := Group(0); g < 10; g++ {
			state.UpdateMask(0, 0, 0, g, 0, 0)
			effective := state.SerializeGroup(StateGroupEffective)
			if effective != 0 {
				t.Errorf("Group %d wrapped to %d, want 0", g, effective)
			}
		}
	})

	t.Run("keycode boundaries", func(t *testing.T) {
		// Test keycode boundaries
		keycodes := []Keycode{
			0,                    // Below min
			8,                    // Min (typical)
			255,                  // Max (typical)
			256,                  // Above max
			0xFFFFFFFF,           // Max uint32
		}

		for _, kc := range keycodes {
			// Should not panic
			_ = state.KeyGetOneSym(kc)
			_ = state.KeyGetSyms(kc)
			_ = state.KeyGetUTF32(kc)
			_ = state.KeyGetUTF8(kc)
		}
	})

	t.Run("repeated updates", func(t *testing.T) {
		// Rapidly toggle modifiers
		for i := 0; i < 1000; i++ {
			if i%2 == 0 {
				state.UpdateMask(ModShift, 0, 0, 0, 0, 0)
			} else {
				state.UpdateMask(0, 0, 0, 0, 0, 0)
			}
		}
	})
}

// TestComposeEdgeCases tests compose edge cases
func TestComposeEdgeCases(t *testing.T) {
	table := TestComposeTable()
	if table == nil {
		t.Skip("No test compose table available")
	}
	state := table.NewState(0)

	t.Run("empty sequence", func(t *testing.T) {
		state.Reset()
		if state.GetStatus() != ComposeNothing {
			t.Error("Expected ComposeNothing after reset")
		}
	})

	t.Run("very long sequence", func(t *testing.T) {
		state.Reset()
		// Feed many keysyms
		for i := 0; i < 100; i++ {
			_ = state.Feed(Keysym('a' + (i % 26)))
		}
		// Should eventually cancel or complete
	})

	t.Run("special keysyms", func(t *testing.T) {
		state.Reset()
		specialKeysyms := []Keysym{
			KeyNoSymbol,
			KeyReturn,
			KeyEscape,
			KeyBackSpace,
			0xFFFFFFFF,    // Max
		}

		for _, ks := range specialKeysyms {
			_ = state.Feed(ks)
		}
	})

	t.Run("concurrent feed and query", func(t *testing.T) {
		// Note: ComposeState is not thread-safe, but shouldn't crash
		state.Reset()
		for i := 0; i < 100; i++ {
			_ = state.Feed(Keysym('a'))
			_ = state.GetStatus()
			_ = state.GetOneSym()
			_ = state.GetUTF8()
		}
	})
}

// TestKeyTypeEdgeCases tests key type edge cases
func TestKeyTypeEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"level name variations",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; };
				xkb_types "a" {
					type "TEST" {
						modifiers = Shift;
						map[Shift] = Level2;
						level_name[Level1] = "Base";
						level_name[Level2] = "";
					};
				};
				xkb_compat "a" {};
				xkb_symbols "a" {};
			};`,
		},
		{
			"none modifier",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; };
				xkb_types "a" {
					type "NONE" {
						modifiers = none;
					};
				};
				xkb_compat "a" {};
				xkb_symbols "a" {};
			};`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			keymap, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if keymap == nil {
				t.Fatal("Keymap is nil")
			}
		})
	}
}

// TestModifierMapEdgeCases tests modifier_map edge cases
func TestModifierMapEdgeCases(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "a" {
			minimum = 8;
			maximum = 255;
			<K1> = 10;
			<K2> = 11;
			<K3> = 12;
		};
		xkb_types "a" {
			type "ONE_LEVEL" { modifiers = none; };
		};
		xkb_compat "a" {};
		xkb_symbols "a" {
			key <K1> { [ a ] };
			key <K2> { [ b ] };
			key <K3> { [ c ] };
			modifier_map Shift { <K1> };
			modifier_map Control { <K2> };
			modifier_map Mod1 { <K3> };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check modifier mappings
	if key := keymap.keys[10]; key != nil {
		if key.vmodmap&ModShift == 0 {
			t.Error("Key 10 should have Shift modifier")
		}
	}
	if key := keymap.keys[11]; key != nil {
		if key.vmodmap&ModControl == 0 {
			t.Error("Key 11 should have Control modifier")
		}
	}
	if key := keymap.keys[12]; key != nil {
		if key.vmodmap&ModMod1 == 0 {
			t.Error("Key 12 should have Mod1 modifier")
		}
	}
}
