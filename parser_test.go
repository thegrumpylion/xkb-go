package xkb

import (
	"os"
	"testing"
)

func TestParserMinimalKeymap(t *testing.T) {
	input := `xkb_keymap {
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
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check keycodes
	if keymap.MinKeycode() != 8 {
		t.Errorf("MinKeycode = %d, want 8", keymap.MinKeycode())
	}
	if keymap.MaxKeycode() != 255 {
		t.Errorf("MaxKeycode = %d, want 255", keymap.MaxKeycode())
	}
	if keymap.KeyByName("AD01") != 24 {
		t.Errorf("KeyByName(AD01) = %d, want 24", keymap.KeyByName("AD01"))
	}

	// Check types
	if keymap.NumTypes() == 0 {
		t.Error("Expected at least one type")
	}

	// Check keys
	key := keymap.keys[24]
	if key == nil {
		t.Fatal("Key 24 not found")
	}
	if len(key.groups) == 0 {
		t.Fatal("Key has no groups")
	}
	if len(key.groups[0].levels) == 0 {
		t.Fatal("Key group has no levels")
	}
}

func TestParserKeycodes(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "evdev" {
			minimum = 8;
			maximum = 255;
			<ESC> = 9;
			<AE01> = 10;
			<AD01> = 24;
			<LFSH> = 50;
			<RALT> = 108;
			alias <ALGR> = <RALT>;
			indicator 1 = "Caps Lock";
			indicator 2 = "Num Lock";
		};
		xkb_types "test" {};
		xkb_compat "test" {};
		xkb_symbols "test" {};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	tests := []struct {
		name string
		want Keycode
	}{
		{"ESC", 9},
		{"AE01", 10},
		{"AD01", 24},
		{"LFSH", 50},
		{"RALT", 108},
	}

	for _, tt := range tests {
		got := keymap.KeyByName(tt.name)
		if got != tt.want {
			t.Errorf("KeyByName(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}

	// Check alias
	if keymap.KeyByName("ALGR") != 108 {
		t.Errorf("Alias ALGR = %d, want 108", keymap.KeyByName("ALGR"))
	}

	// Check indicators (LEDGetIndex returns -1 for not found)
	if keymap.LEDGetIndex("Caps Lock") < 0 {
		t.Error("LED 'Caps Lock' not found")
	}
	if keymap.LEDGetIndex("Num Lock") < 0 {
		t.Error("LED 'Num Lock' not found")
	}
}

func TestParserTypes(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
		};
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
		xkb_compat "test" {};
		xkb_symbols "test" {};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check virtual modifiers
	if _, ok := keymap.virtualMods["NumLock"]; !ok {
		t.Error("Virtual modifier NumLock not found")
	}
	if _, ok := keymap.virtualMods["Alt"]; !ok {
		t.Error("Virtual modifier Alt not found")
	}

	// Check types
	if keymap.NumTypes() < 3 {
		t.Errorf("NumTypes = %d, want at least 3", keymap.NumTypes())
	}

	// Check ONE_LEVEL
	oneLevelType := keymap.types["ONE_LEVEL"]
	if oneLevelType == nil {
		t.Fatal("Type ONE_LEVEL not found")
	}
	if oneLevelType.numLevels != 1 {
		t.Errorf("ONE_LEVEL numLevels = %d, want 1", oneLevelType.numLevels)
	}

	// Check TWO_LEVEL
	twoLevelType := keymap.types["TWO_LEVEL"]
	if twoLevelType == nil {
		t.Fatal("Type TWO_LEVEL not found")
	}
	if twoLevelType.numLevels != 2 {
		t.Errorf("TWO_LEVEL numLevels = %d, want 2", twoLevelType.numLevels)
	}
	if twoLevelType.mods != ModShift {
		t.Errorf("TWO_LEVEL mods = %d, want %d (Shift)", twoLevelType.mods, ModShift)
	}

	// Check ALPHABETIC
	alphaType := keymap.types["ALPHABETIC"]
	if alphaType == nil {
		t.Fatal("Type ALPHABETIC not found")
	}
	if alphaType.numLevels != 2 {
		t.Errorf("ALPHABETIC numLevels = %d, want 2", alphaType.numLevels)
	}
	if alphaType.mods != ModShift|ModLock {
		t.Errorf("ALPHABETIC mods = %d, want %d (Shift+Lock)", alphaType.mods, ModShift|ModLock)
	}
}

func TestParserSymbols(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
			<AD01> = 24;
			<AD02> = 25;
			<AE01> = 10;
		};
		xkb_types "test" {
			type "ONE_LEVEL" {
				modifiers = none;
			};
			type "TWO_LEVEL" {
				modifiers = Shift;
				map[Shift] = Level2;
			};
		};
		xkb_compat "test" {};
		xkb_symbols "us" {
			name[Group1] = "English (US)";

			key <AD01> { [ q, Q ] };
			key <AD02> { [ w, W ] };
			key <AE01> { [ 1, exclam ] };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check group name
	if keymap.GroupName(0) != "English (US)" {
		t.Errorf("GroupName(0) = %q, want %q", keymap.GroupName(0), "English (US)")
	}

	// Check keys
	tests := []struct {
		keycode  Keycode
		level0   Keysym
		level1   Keysym
	}{
		{24, Keysym('q'), Keysym('Q')},
		{25, Keysym('w'), Keysym('W')},
		{10, Keysym('1'), Keysym('!')},
	}

	for _, tt := range tests {
		key := keymap.keys[tt.keycode]
		if key == nil {
			t.Errorf("Key %d not found", tt.keycode)
			continue
		}
		if len(key.groups) == 0 || len(key.groups[0].levels) < 2 {
			t.Errorf("Key %d has insufficient levels", tt.keycode)
			continue
		}
		if key.groups[0].levels[0].syms[0] != tt.level0 {
			t.Errorf("Key %d level 0 = %#x, want %#x", tt.keycode, key.groups[0].levels[0].syms[0], tt.level0)
		}
		if key.groups[0].levels[1].syms[0] != tt.level1 {
			t.Errorf("Key %d level 1 = %#x, want %#x", tt.keycode, key.groups[0].levels[1].syms[0], tt.level1)
		}
	}
}

func TestParserComplexKey(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
			<AD01> = 24;
		};
		xkb_types "test" {
			type "ALPHABETIC" {
				modifiers = Shift + Lock;
				map[Shift] = Level2;
				map[Lock] = Level2;
			};
		};
		xkb_compat "test" {};
		xkb_symbols "test" {
			key <AD01> {
				type = "ALPHABETIC",
				symbols[Group1] = [ a, A ]
			};
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	key := keymap.keys[24]
	if key == nil {
		t.Fatal("Key 24 not found")
	}
	if len(key.groups) == 0 {
		t.Fatal("Key has no groups")
	}
	if key.groups[0].keyType == nil {
		t.Fatal("Key group has no type")
	}
	if key.groups[0].keyType.name != "ALPHABETIC" {
		t.Errorf("Key type = %q, want %q", key.groups[0].keyType.name, "ALPHABETIC")
	}
}

func TestParserCompat(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
		};
		xkb_types "test" {};
		xkb_compat "complete" {
			virtual_modifiers NumLock, Alt;

			interpret Shift_L {
				action = SetMods(modifiers=Shift);
			};

			indicator "Caps Lock" {
				modifiers = Lock;
			};
		};
		xkb_symbols "test" {};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check indicator
	capsLed := keymap.leds["Caps Lock"]
	if capsLed == nil {
		t.Fatal("LED 'Caps Lock' not found")
	}
	if capsLed.mods != ModLock {
		t.Errorf("Caps Lock mods = %d, want %d (Lock)", capsLed.mods, ModLock)
	}
}

func TestParserGeometrySkipped(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
		};
		xkb_types "test" {};
		xkb_compat "test" {};
		xkb_symbols "test" {};
		xkb_geometry "pc(pc105)" {
			// Complex geometry that should be skipped
			shape.cornerRadius = 1;
			shape "NORM" { { [ 18,18] }, { [2,1], [ 16,17] } };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should parse without error
	if keymap == nil {
		t.Fatal("Keymap is nil")
	}
}

func TestParserErrorHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"missing xkb_keymap",
			`something_else {}`,
		},
		{
			"missing opening brace",
			`xkb_keymap xkb_keycodes {}`,
		},
		{
			"unterminated string in keycode",
			`xkb_keymap { xkb_keycodes "test { }; };`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.input))
			_, err := p.Parse()
			if err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}

func TestParserNumbers(t *testing.T) {
	p := &Parser{}

	tests := []struct {
		input string
		want  int
	}{
		{"123", 123},
		{"0", 0},
		{"0x1F", 31},
		{"0xFF", 255},
		{"017", 15},  // octal
		{"0755", 493}, // octal
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := p.parseNumber(tt.input)
			if err != nil {
				t.Fatalf("parseNumber(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseNumber(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParserLevelName(t *testing.T) {
	p := &Parser{}

	tests := []struct {
		input string
		want  Level
		err   bool
	}{
		{"Level1", 0, false},
		{"Level2", 1, false},
		{"Level8", 7, false},
		{"Invalid", 0, true},
		{"Levelx", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := p.parseLevelName(tt.input)
			if tt.err {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("parseLevelName(%q) = %d, want %d", tt.input, got, tt.want)
				}
			}
		})
	}
}

func TestParserGroupIdent(t *testing.T) {
	p := &Parser{}

	tests := []struct {
		input string
		want  int
		err   bool
	}{
		{"Group1", 0, false},
		{"Group2", 1, false},
		{"Group4", 3, false},
		{"Invalid", 0, true},
		{"Groupx", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := p.parseGroupIdent(tt.input)
			if tt.err {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("parseGroupIdent(%q) = %d, want %d", tt.input, got, tt.want)
				}
			}
		})
	}
}

func TestParserModifierMask(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" { minimum = 8; maximum = 255; };
		xkb_types "test" {
			type "TEST" {
				modifiers = Shift + Control;
			};
		};
		xkb_compat "test" {};
		xkb_symbols "test" {};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	testType := keymap.types["TEST"]
	if testType == nil {
		t.Fatal("Type TEST not found")
	}
	expected := ModShift | ModControl
	if testType.mods != expected {
		t.Errorf("Type mods = %d, want %d (Shift+Control)", testType.mods, expected)
	}
}

func TestParserModifierMap(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
			<LFSH> = 50;
			<RTSH> = 62;
		};
		xkb_types "test" {
			type "ONE_LEVEL" { modifiers = none; };
		};
		xkb_compat "test" {};
		xkb_symbols "test" {
			key <LFSH> { [ Shift_L ] };
			key <RTSH> { [ Shift_R ] };
			modifier_map Shift { <LFSH>, <RTSH> };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check that shift keys have Shift in vmodmap
	lfshKey := keymap.keys[50]
	if lfshKey == nil {
		t.Fatal("Key LFSH not found")
	}
	if lfshKey.vmodmap&ModShift == 0 {
		t.Error("LFSH should have Shift in vmodmap")
	}

	rtshKey := keymap.keys[62]
	if rtshKey == nil {
		t.Fatal("Key RTSH not found")
	}
	if rtshKey.vmodmap&ModShift == 0 {
		t.Error("RTSH should have Shift in vmodmap")
	}
}

func TestParserKeyRepeat(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
			<AD01> = 24;
			<LFSH> = 50;
		};
		xkb_types "test" {
			type "ONE_LEVEL" { modifiers = none; };
		};
		xkb_compat "test" {};
		xkb_symbols "test" {
			key <AD01> { [ q ], repeat = yes };
			key <LFSH> { [ Shift_L ], repeat = no };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check repeat settings
	qKey := keymap.keys[24]
	if qKey == nil {
		t.Fatal("Key AD01 not found")
	}
	if !qKey.repeats {
		t.Error("Key AD01 should repeat")
	}

	shiftKey := keymap.keys[50]
	if shiftKey == nil {
		t.Fatal("Key LFSH not found")
	}
	if shiftKey.repeats {
		t.Error("Key LFSH should not repeat")
	}
}

func TestParserKeysymLookup(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
			<RTRN> = 36;
			<ESC> = 9;
		};
		xkb_types "test" {
			type "ONE_LEVEL" { modifiers = none; };
		};
		xkb_compat "test" {};
		xkb_symbols "test" {
			key <RTRN> { [ Return ] };
			key <ESC> { [ Escape ] };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check Return key
	rtnKey := keymap.keys[36]
	if rtnKey == nil {
		t.Fatal("Key RTRN not found")
	}
	if len(rtnKey.groups) == 0 || len(rtnKey.groups[0].levels) == 0 {
		t.Fatal("Key RTRN has no levels")
	}
	if rtnKey.groups[0].levels[0].syms[0] != KeyReturn {
		t.Errorf("RTRN sym = %#x, want %#x (Return)", rtnKey.groups[0].levels[0].syms[0], KeyReturn)
	}

	// Check Escape key
	escKey := keymap.keys[9]
	if escKey == nil {
		t.Fatal("Key ESC not found")
	}
	if len(escKey.groups) == 0 || len(escKey.groups[0].levels) == 0 {
		t.Fatal("Key ESC has no levels")
	}
	if escKey.groups[0].levels[0].syms[0] != KeyEscape {
		t.Errorf("ESC sym = %#x, want %#x (Escape)", escKey.groups[0].levels[0].syms[0], KeyEscape)
	}
}

func TestParserIncludeSkipped(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
		};
		xkb_types "test" {};
		xkb_compat "test" {};
		xkb_symbols "test" {
			include "us(basic)"
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should parse without error (include is skipped)
	if keymap == nil {
		t.Fatal("Keymap is nil")
	}
}

// TestParserIntegration tests parsing a more complete keymap
func TestParserIntegration(t *testing.T) {
	// A more realistic keymap with multiple key definitions
	input := `xkb_keymap {
	xkb_keycodes "evdev+aliases(qwerty)" {
		minimum = 8;
		maximum = 255;
		<ESC>  = 9;
		<AE01> = 10;
		<AE02> = 11;
		<AE03> = 12;
		<TAB>  = 23;
		<AD01> = 24;
		<AD02> = 25;
		<AD03> = 26;
		<RTRN> = 36;
		<LFSH> = 50;
		<RTSH> = 62;
		<CAPS> = 66;
		<SPCE> = 65;
		<BKSP> = 22;
		indicator 1 = "Caps Lock";
		indicator 2 = "Num Lock";
		indicator 3 = "Scroll Lock";
	};

	xkb_types "complete" {
		virtual_modifiers NumLock,Alt,LevelThree,LAlt,RAlt,RControl,LControl,ScrollLock,LevelFive,AltGr,Meta,Super,Hyper;

		type "ONE_LEVEL" {
			modifiers= none;
			level_name[Level1]= "Any";
		};

		type "TWO_LEVEL" {
			modifiers= Shift;
			map[Shift]= Level2;
			level_name[Level1]= "Base";
			level_name[Level2]= "Shift";
		};

		type "ALPHABETIC" {
			modifiers= Shift+Lock;
			map[Shift]= Level2;
			map[Lock]= Level2;
			map[Shift+Lock]= Level1;
			level_name[Level1]= "Base";
			level_name[Level2]= "Caps";
		};

		type "KEYPAD" {
			modifiers= Shift+NumLock;
			map[Shift]= Level2;
			map[NumLock]= Level2;
			level_name[Level1]= "Base";
			level_name[Level2]= "Number";
		};
	};

	xkb_compat "complete" {
		virtual_modifiers NumLock,Alt,LevelThree;

		interpret Shift_L+AnyOfOrNone(all) {
			action= SetMods(modifiers=Shift,clearLocks);
		};

		interpret Caps_Lock+AnyOfOrNone(all) {
			action= LockMods(modifiers=Lock);
		};

		indicator "Caps Lock" {
			modifiers= Lock;
		};

		indicator "Num Lock" {
			modifiers= NumLock;
		};
	};

	xkb_symbols "pc+us" {
		name[Group1]= "English (US)";

		key <ESC>  { [ Escape ] };
		key <AE01> { [ 1, exclam ] };
		key <AE02> { [ 2, at ] };
		key <AE03> { [ 3, numbersign ] };
		key <TAB>  { [ Tab, ISO_Left_Tab ] };
		key <AD01> { [ q, Q ] };
		key <AD02> { [ w, W ] };
		key <AD03> { [ e, E ] };
		key <RTRN> { [ Return ] };
		key <LFSH> {
			type= "ONE_LEVEL",
			symbols[Group1]= [ Shift_L ]
		};
		key <RTSH> {
			type= "ONE_LEVEL",
			symbols[Group1]= [ Shift_R ]
		};
		key <CAPS> {
			type= "ONE_LEVEL",
			symbols[Group1]= [ Caps_Lock ]
		};
		key <SPCE> { [ space ] };
		key <BKSP> { [ BackSpace, BackSpace ] };

		modifier_map Shift { <LFSH>, <RTSH> };
		modifier_map Lock { <CAPS> };
	};
};`

	ctx := NewContext(ContextNoFlags)
	keymap, err := ctx.NewKeymapFromString([]byte(input), KeymapFormatTextV1)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify keycodes
	tests := []struct {
		name    string
		keycode Keycode
	}{
		{"ESC", 9},
		{"AD01", 24},
		{"RTRN", 36},
		{"LFSH", 50},
		{"CAPS", 66},
	}

	for _, tt := range tests {
		if keymap.KeyByName(tt.name) != tt.keycode {
			t.Errorf("KeyByName(%q) = %d, want %d", tt.name, keymap.KeyByName(tt.name), tt.keycode)
		}
	}

	// Verify types
	if keymap.NumTypes() < 4 {
		t.Errorf("NumTypes = %d, want at least 4", keymap.NumTypes())
	}

	// Verify group name
	if keymap.GroupName(0) != "English (US)" {
		t.Errorf("GroupName(0) = %q, want %q", keymap.GroupName(0), "English (US)")
	}

	// Test key translation using State
	state := keymap.NewState()

	// Test 'q' key without modifiers -> q
	sym := state.KeyGetOneSym(24)
	if sym != Keysym('q') {
		t.Errorf("KeyGetOneSym(24) = %#x, want %#x ('q')", sym, Keysym('q'))
	}

	// Test 'q' key with Shift -> Q
	state.UpdateMask(ModShift, 0, 0, 0, 0, 0)
	sym = state.KeyGetOneSym(24)
	if sym != Keysym('Q') {
		t.Errorf("KeyGetOneSym(24) with Shift = %#x, want %#x ('Q')", sym, Keysym('Q'))
	}

	// Test '1' key -> 1
	state.UpdateMask(0, 0, 0, 0, 0, 0)
	sym = state.KeyGetOneSym(10)
	if sym != Keysym('1') {
		t.Errorf("KeyGetOneSym(10) = %#x, want %#x ('1')", sym, Keysym('1'))
	}

	// Test '1' key with Shift -> !
	state.UpdateMask(ModShift, 0, 0, 0, 0, 0)
	sym = state.KeyGetOneSym(10)
	if sym != Keysym('!') {
		t.Errorf("KeyGetOneSym(10) with Shift = %#x, want %#x ('!')", sym, Keysym('!'))
	}

	// Test Return key
	state.UpdateMask(0, 0, 0, 0, 0, 0)
	sym = state.KeyGetOneSym(36)
	if sym != KeyReturn {
		t.Errorf("KeyGetOneSym(36) = %#x, want %#x (Return)", sym, KeyReturn)
	}

	// Test Escape key
	sym = state.KeyGetOneSym(9)
	if sym != KeyEscape {
		t.Errorf("KeyGetOneSym(9) = %#x, want %#x (Escape)", sym, KeyEscape)
	}

	// Test LEDs
	if keymap.NumLEDs() < 3 {
		t.Errorf("NumLEDs = %d, want at least 3", keymap.NumLEDs())
	}
}

// TestParserRealKeymap tests parsing a real system keymap
func TestParserRealKeymap(t *testing.T) {
	data, err := os.ReadFile("testdata/us_intl.xkb")
	if err != nil {
		t.Skipf("Skipping real keymap test: %v", err)
	}

	ctx := NewContext(ContextNoFlags)
	keymap, err := ctx.NewKeymapFromString(data, KeymapFormatTextV1)
	if err != nil {
		t.Fatalf("Failed to parse system keymap: %v", err)
	}

	t.Logf("Parsed system keymap successfully:")
	t.Logf("  Min keycode: %d", keymap.MinKeycode())
	t.Logf("  Max keycode: %d", keymap.MaxKeycode())
	t.Logf("  Num types: %d", keymap.NumTypes())
	t.Logf("  Num groups: %d", keymap.NumGroups())
	t.Logf("  Num LEDs: %d", keymap.NumLEDs())
	t.Logf("  Group 1 name: %q", keymap.GroupName(0))

	// Verify some common keycodes exist
	commonKeys := []struct {
		name    string
		keycode Keycode
	}{
		{"ESC", 9},
		{"AD01", 24},   // Q
		{"RTRN", 36},   // Return
		{"SPCE", 65},   // Space
		{"LFSH", 50},   // Left Shift
		{"CAPS", 66},   // Caps Lock
	}

	for _, kk := range commonKeys {
		if keymap.KeyByName(kk.name) != kk.keycode {
			t.Errorf("KeyByName(%q) = %d, want %d", kk.name, keymap.KeyByName(kk.name), kk.keycode)
		}
	}

	// Test State with the real keymap
	state := keymap.NewState()

	// Test 'q' key (AD01 = keycode 24)
	sym := state.KeyGetOneSym(24)
	if sym != Keysym('q') {
		t.Errorf("KeyGetOneSym(24) = %#x (%s), want 'q'", sym, KeysymGetName(sym))
	}

	// Test with Shift
	state.UpdateMask(ModShift, 0, 0, 0, 0, 0)
	sym = state.KeyGetOneSym(24)
	if sym != Keysym('Q') {
		t.Errorf("KeyGetOneSym(24) with Shift = %#x (%s), want 'Q'", sym, KeysymGetName(sym))
	}

	// Test 'a' key (AC01 = keycode 38)
	state.UpdateMask(0, 0, 0, 0, 0, 0)
	sym = state.KeyGetOneSym(38)
	if sym != Keysym('a') {
		t.Errorf("KeyGetOneSym(38) = %#x (%s), want 'a'", sym, KeysymGetName(sym))
	}

	// Test space (SPCE = keycode 65)
	sym = state.KeyGetOneSym(65)
	if sym != Keysym(' ') {
		t.Errorf("KeyGetOneSym(65) = %#x (%s), want space", sym, KeysymGetName(sym))
	}

	// Test Return (RTRN = keycode 36)
	sym = state.KeyGetOneSym(36)
	if sym != KeyReturn {
		t.Errorf("KeyGetOneSym(36) = %#x (%s), want Return", sym, KeysymGetName(sym))
	}

	// Test Escape (ESC = keycode 9)
	sym = state.KeyGetOneSym(9)
	if sym != KeyEscape {
		t.Errorf("KeyGetOneSym(9) = %#x (%s), want Escape", sym, KeysymGetName(sym))
	}

	// Test number row
	state.UpdateMask(0, 0, 0, 0, 0, 0)
	sym = state.KeyGetOneSym(10) // AE01 = '1'
	if sym != Keysym('1') {
		t.Errorf("KeyGetOneSym(10) = %#x (%s), want '1'", sym, KeysymGetName(sym))
	}

	state.UpdateMask(ModShift, 0, 0, 0, 0, 0)
	sym = state.KeyGetOneSym(10) // AE01 with Shift = '!'
	if sym != Keysym('!') {
		t.Errorf("KeyGetOneSym(10) with Shift = %#x (%s), want '!'", sym, KeysymGetName(sym))
	}

	// Test UTF-8 conversion
	state.UpdateMask(0, 0, 0, 0, 0, 0)
	utf8 := state.KeyGetUTF8(24) // 'q'
	if utf8 != "q" {
		t.Errorf("KeyGetUTF8(24) = %q, want %q", utf8, "q")
	}

	t.Log("Real keymap test passed!")
}

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
			false,
		},
		{
			"min equals max",
			`xkb_keymap { xkb_keycodes "a" { minimum=8; maximum=8; }; xkb_types "a" {}; xkb_compat "a" {}; xkb_symbols "a" {}; };`,
			false,
		},
		{
			"min greater than max",
			`xkb_keymap { xkb_keycodes "a" { minimum=255; maximum=8; }; xkb_types "a" {}; xkb_compat "a" {}; xkb_symbols "a" {}; };`,
			false,
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
			false,
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
				xkb_types "a" {
					type "ONE_LEVEL" { modifiers = none; };
					type "EIGHT_LEVEL" {
						modifiers = Shift+Lock+Control;
						map[Shift] = Level2;
						map[Lock] = Level3;
						map[Control] = Level4;
						map[Shift+Lock] = Level5;
						map[Shift+Control] = Level6;
						map[Lock+Control] = Level7;
						map[Shift+Lock+Control] = Level8;
					};
				};
				xkb_compat "a" {};
				xkb_symbols "a" {
					key <K> { type = "EIGHT_LEVEL", [ a, b, c, d, e, f, g, h ] };
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
			"duplicate keycodes section",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; };
				xkb_keycodes "b" { minimum=8; maximum=255; };
				xkb_types "a" {};
				xkb_compat "a" {};
				xkb_symbols "a" {};
			};`,
			true,
		},
		{
			"duplicate types section",
			`xkb_keymap {
				xkb_keycodes "a" { minimum=8; maximum=255; };
				xkb_types "a" {};
				xkb_types "b" {};
				xkb_compat "a" {};
				xkb_symbols "a" {};
			};`,
			true,
		},
		{
			"wrong section order",
			`xkb_keymap { xkb_symbols "a" {}; xkb_keycodes "a" { minimum=8; maximum=255; }; xkb_types "a" {}; xkb_compat "a" {}; };`,
			false,
		},
		{
			"missing required section",
			`xkb_keymap { xkb_keycodes "a" { minimum=8; maximum=255; }; xkb_types "a" {}; };`,
			true,
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
			false,
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

// TestParserReplaceKey tests the "replace key" modifier.
func TestParserReplaceKey(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
			<AD01> = 24;
		};
		xkb_types "test" {
			type "ONE_LEVEL" { modifiers = none; };
			type "TWO_LEVEL" { modifiers = Shift; map[Shift] = Level2; };
		};
		xkb_compat "test" {};
		xkb_symbols "test" {
			key <AD01> { [ q, Q ] };
			replace key <AD01> { [ a ] };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	key := keymap.keys[24]
	if key == nil {
		t.Fatal("Key AD01 not found")
	}

	// After replace, the key should have 'a' (the replaced value)
	if len(key.groups) == 0 || len(key.groups[0].levels) == 0 {
		t.Fatal("Key has no levels")
	}
	if key.groups[0].levels[0].syms[0] != Keysym('a') {
		t.Errorf("Expected 'a', got %#x", key.groups[0].levels[0].syms[0])
	}
}

// TestParserOverrideKey tests the "override key" modifier.
func TestParserOverrideKey(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
			<AD01> = 24;
		};
		xkb_types "test" {
			type "ONE_LEVEL" { modifiers = none; };
			type "TWO_LEVEL" { modifiers = Shift; map[Shift] = Level2; };
		};
		xkb_compat "test" {};
		xkb_symbols "test" {
			key <AD01> { [ q, Q ] };
			override key <AD01> { [ x, X ] };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	key := keymap.keys[24]
	if key == nil {
		t.Fatal("Key AD01 not found")
	}

	// After override, the key should have 'x', 'X'
	if len(key.groups) == 0 || len(key.groups[0].levels) < 2 {
		t.Fatal("Key has insufficient levels")
	}
	if key.groups[0].levels[0].syms[0] != Keysym('x') {
		t.Errorf("Level 0: expected 'x', got %#x", key.groups[0].levels[0].syms[0])
	}
	if key.groups[0].levels[1].syms[0] != Keysym('X') {
		t.Errorf("Level 1: expected 'X', got %#x", key.groups[0].levels[1].syms[0])
	}
}

// TestParserAugmentKey tests the "augment key" modifier.
func TestParserAugmentKey(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
			<AD01> = 24;
		};
		xkb_types "test" {
			type "ONE_LEVEL" { modifiers = none; };
			type "TWO_LEVEL" { modifiers = Shift; map[Shift] = Level2; };
		};
		xkb_compat "test" {};
		xkb_symbols "test" {
			key <AD01> { [ q, Q ] };
			augment key <AD01> { [ z, Z ] };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	key := keymap.keys[24]
	if key == nil {
		t.Fatal("Key AD01 not found")
	}

	// After augment, key should have 'z', 'Z' (augment replaces like override in our implementation)
	if len(key.groups) == 0 || len(key.groups[0].levels) < 2 {
		t.Fatal("Key has insufficient levels")
	}
	if key.groups[0].levels[0].syms[0] != Keysym('z') {
		t.Errorf("Level 0: expected 'z', got %#x", key.groups[0].levels[0].syms[0])
	}
}

// TestParserAnyKeyword tests the "any" keyword for preserving existing symbols.
func TestParserAnyKeyword(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
			<AD01> = 24;
		};
		xkb_types "test" {
			type "ONE_LEVEL" { modifiers = none; };
			type "TWO_LEVEL" { modifiers = Shift; map[Shift] = Level2; };
			type "FOUR_LEVEL" { modifiers = Shift+Control; map[Shift] = Level2; map[Control] = Level3; map[Shift+Control] = Level4; };
		};
		xkb_compat "test" {};
		xkb_symbols "test" {
			key <AD01> { [ q, Q ] };
			key <AD01> { [ any, any, dollar ] };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	key := keymap.keys[24]
	if key == nil {
		t.Fatal("Key AD01 not found")
	}

	if len(key.groups) == 0 || len(key.groups[0].levels) < 3 {
		t.Fatalf("Key has insufficient levels: %d", len(key.groups[0].levels))
	}

	// Level 0: should preserve 'q' (from first definition)
	if key.groups[0].levels[0].syms[0] != Keysym('q') {
		t.Errorf("Level 0: expected 'q' (preserved), got %#x (%s)",
			key.groups[0].levels[0].syms[0], KeysymGetName(key.groups[0].levels[0].syms[0]))
	}

	// Level 1: should preserve 'Q' (from first definition)
	if key.groups[0].levels[1].syms[0] != Keysym('Q') {
		t.Errorf("Level 1: expected 'Q' (preserved), got %#x (%s)",
			key.groups[0].levels[1].syms[0], KeysymGetName(key.groups[0].levels[1].syms[0]))
	}

	// Level 2: should be 'dollar' (from second definition)
	dollarKeysym := KeysymFromName("dollar", KeysymNameNoFlags)
	if key.groups[0].levels[2].syms[0] != dollarKeysym {
		t.Errorf("Level 2: expected 'dollar' (%#x), got %#x (%s)",
			dollarKeysym, key.groups[0].levels[2].syms[0], KeysymGetName(key.groups[0].levels[2].syms[0]))
	}
}

// TestParserKeyMerging tests that multiple key definitions for the same keycode merge properly.
func TestParserKeyMerging(t *testing.T) {
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
			<AD01> = 24;
		};
		xkb_types "test" {
			type "ONE_LEVEL" { modifiers = none; };
			type "TWO_LEVEL" { modifiers = Shift; map[Shift] = Level2; };
		};
		xkb_compat "test" {};
		xkb_symbols "test" {
			key <AD01> { [ a, A ], repeat = no };
			key <AD01> { [ b, B ] };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	key := keymap.keys[24]
	if key == nil {
		t.Fatal("Key AD01 not found")
	}

	// Second definition should override symbols
	if len(key.groups) == 0 || len(key.groups[0].levels) < 2 {
		t.Fatal("Key has insufficient levels")
	}
	if key.groups[0].levels[0].syms[0] != Keysym('b') {
		t.Errorf("Level 0: expected 'b', got %#x", key.groups[0].levels[0].syms[0])
	}

	// First definition's repeat=no should be preserved
	if key.repeats {
		t.Error("Key should not repeat (repeat=no from first definition should be preserved)")
	}
}

// TestParserSkipStatementWithBraces tests that complex statements with nested braces are properly skipped.
func TestParserSkipStatementWithBraces(t *testing.T) {
	// This tests the skipStatementWithBraces function by having complex nested structures
	input := `xkb_keymap {
		xkb_keycodes "test" {
			minimum = 8;
			maximum = 255;
			<AD01> = 24;
			<AD02> = 25;
		};
		xkb_types "test" {
			type "ONE_LEVEL" { modifiers = none; };
		};
		xkb_compat "test" {};
		xkb_symbols "test" {
			replace key <AD01> {
				type = "ONE_LEVEL",
				symbols[Group1] = [ a ],
				actions[Group1] = [ SetMods(modifiers=Shift) ]
			};
			key <AD02> { [ q ] };
		};
	};`

	p := NewParser([]byte(input))
	keymap, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Both keys should be parsed correctly
	key1 := keymap.keys[24]
	if key1 == nil {
		t.Fatal("Key AD01 not found")
	}
	if key1.groups[0].levels[0].syms[0] != Keysym('a') {
		t.Errorf("AD01: expected 'a', got %#x", key1.groups[0].levels[0].syms[0])
	}

	key2 := keymap.keys[25]
	if key2 == nil {
		t.Fatal("Key AD02 not found")
	}
	if key2.groups[0].levels[0].syms[0] != Keysym('q') {
		t.Errorf("AD02: expected 'q', got %#x", key2.groups[0].levels[0].syms[0])
	}
}
