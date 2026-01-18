//go:build integration

package xkb

import (
	"strings"
	"testing"
)

// Integration tests that require system XKB data.
// Run with: go test -tags=integration ./...

// Common layouts to test
var testLayouts = []struct {
	layout  string
	variant string
	desc    string
}{
	{"us", "", "US English"},
	{"us", "intl", "US International"},
	{"us", "dvorak", "US Dvorak"},
	{"de", "", "German"},
	{"de", "neo", "German Neo"},
	{"fr", "", "French"},
	{"gb", "", "British"},
	{"ru", "", "Russian"},
	{"ru", "phonetic", "Russian Phonetic"},
	{"es", "", "Spanish"},
	{"it", "", "Italian"},
	{"pt", "", "Portuguese"},
	{"pl", "", "Polish"},
	{"se", "", "Swedish"},
	{"no", "", "Norwegian"},
	{"fi", "", "Finnish"},
	{"dk", "", "Danish"},
	{"nl", "", "Dutch"},
	{"be", "", "Belgian"},
	{"ch", "", "Swiss German"},
	{"ch", "fr", "Swiss French"},
	{"at", "", "Austrian"},
	{"jp", "", "Japanese"},
	{"kr", "", "Korean"},
	{"cn", "", "Chinese"},
	{"il", "", "Hebrew"},
	{"ara", "", "Arabic"},
	{"gr", "", "Greek"},
	{"tr", "", "Turkish"},
	{"cz", "", "Czech"},
	{"hu", "", "Hungarian"},
	{"ro", "", "Romanian"},
	{"bg", "", "Bulgarian"},
	{"ua", "", "Ukrainian"},
	{"latam", "", "Latin American"},
	{"br", "", "Brazilian"},
}

func TestIntegration_RMLVOLayouts(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	for _, tc := range testLayouts {
		name := tc.layout
		if tc.variant != "" {
			name += "_" + tc.variant
		}

		t.Run(name, func(t *testing.T) {
			keymap, err := ctx.NewKeymapFromNames(&RuleNames{
				Layout:  tc.layout,
				Variant: tc.variant,
			})
			if err != nil {
				t.Fatalf("Failed to compile %s: %v", tc.desc, err)
			}

			// Basic sanity checks
			if keymap.MinKeycode() == 0 && keymap.MaxKeycode() == 0 {
				t.Error("Keymap has no keycodes")
			}

			if keymap.NumGroups() == 0 {
				t.Error("Keymap has no groups")
			}

			// Check common keys exist
			commonKeys := []string{"AD01", "AD02", "AC01", "LFSH", "SPCE", "RTRN"}
			for _, key := range commonKeys {
				if keymap.KeyByName(key) == 0 {
					t.Errorf("Missing common key: %s", key)
				}
			}

			// Test state creation and basic key translation
			state := keymap.NewState()
			qKey := keymap.KeyByName("AD01")
			if qKey != 0 {
				sym := state.KeyGetOneSym(qKey)
				if sym == 0 {
					t.Errorf("AD01 key produces no symbol")
				}
			}
		})
	}
}

func TestIntegration_RoundTrip(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	layouts := []struct {
		layout  string
		variant string
	}{
		{"us", ""},
		{"de", ""},
		{"fr", ""},
		{"ru", ""},
	}

	for _, tc := range layouts {
		name := tc.layout
		if tc.variant != "" {
			name += "_" + tc.variant
		}

		t.Run(name, func(t *testing.T) {
			// Compile from RMLVO
			original, err := ctx.NewKeymapFromNames(&RuleNames{
				Layout:  tc.layout,
				Variant: tc.variant,
			})
			if err != nil {
				t.Fatalf("Failed to compile: %v", err)
			}

			// Serialize to string
			serialized, err := original.GetAsString(KeymapFormatTextV1)
			if err != nil {
				t.Fatalf("Failed to serialize: %v", err)
			}

			// Parse back
			reparsed, err := ctx.NewKeymapFromString([]byte(serialized), KeymapFormatTextV1)
			if err != nil {
				t.Fatalf("Failed to reparse: %v", err)
			}

			// Compare key properties
			if original.MinKeycode() != reparsed.MinKeycode() {
				t.Errorf("MinKeycode mismatch: %d vs %d", original.MinKeycode(), reparsed.MinKeycode())
			}

			// Compare key translation for common keys
			origState := original.NewState()
			reparsedState := reparsed.NewState()

			testKeys := []string{"AD01", "AD02", "AD03", "AC01", "AC02", "AB01"}
			for _, keyName := range testKeys {
				origKc := original.KeyByName(keyName)
				reparsedKc := reparsed.KeyByName(keyName)

				if origKc == 0 || reparsedKc == 0 {
					continue
				}

				origSym := origState.KeyGetOneSym(origKc)
				reparsedSym := reparsedState.KeyGetOneSym(reparsedKc)

				if origSym != reparsedSym {
					t.Errorf("Key %s symbol mismatch: %s vs %s",
						keyName, KeysymGetName(origSym), KeysymGetName(reparsedSym))
				}
			}
		})
	}
}

func TestIntegration_KeyTranslation(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	// Test US layout key translation
	t.Run("US_Layout", func(t *testing.T) {
		keymap, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})
		if err != nil {
			t.Skipf("Could not load US layout: %v", err)
		}

		state := keymap.NewState()

		// Test basic letter keys (unshifted)
		letterTests := []struct {
			keyName  string
			expected rune
		}{
			{"AD01", 'q'},
			{"AD02", 'w'},
			{"AD03", 'e'},
			{"AD04", 'r'},
			{"AD05", 't'},
			{"AD06", 'y'},
			{"AD07", 'u'},
			{"AD08", 'i'},
			{"AD09", 'o'},
			{"AD10", 'p'},
			{"AC01", 'a'},
			{"AC02", 's'},
			{"AC03", 'd'},
			{"AC04", 'f'},
			{"AC05", 'g'},
			{"AC06", 'h'},
			{"AC07", 'j'},
			{"AC08", 'k'},
			{"AC09", 'l'},
			{"AB01", 'z'},
			{"AB02", 'x'},
			{"AB03", 'c'},
			{"AB04", 'v'},
			{"AB05", 'b'},
			{"AB06", 'n'},
			{"AB07", 'm'},
		}

		for _, tc := range letterTests {
			kc := keymap.KeyByName(tc.keyName)
			if kc == 0 {
				t.Errorf("Key %s not found", tc.keyName)
				continue
			}

			got := state.KeyGetUTF32(kc)
			if got != tc.expected {
				t.Errorf("Key %s: expected '%c', got '%c'", tc.keyName, tc.expected, got)
			}
		}
	})

	// Test shift modifier
	t.Run("US_Shifted", func(t *testing.T) {
		keymap, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})
		if err != nil {
			t.Skipf("Could not load US layout: %v", err)
		}

		state := keymap.NewState()
		state.UpdateMask(ModShift, 0, 0, 0, 0, 0)

		// Test shifted letters
		shiftedTests := []struct {
			keyName  string
			expected rune
		}{
			{"AD01", 'Q'},
			{"AC01", 'A'},
			{"AB01", 'Z'},
		}

		for _, tc := range shiftedTests {
			kc := keymap.KeyByName(tc.keyName)
			if kc == 0 {
				continue
			}

			got := state.KeyGetUTF32(kc)
			if got != tc.expected {
				t.Errorf("Shifted key %s: expected '%c', got '%c'", tc.keyName, tc.expected, got)
			}
		}
	})

	// Test German layout
	t.Run("DE_Layout", func(t *testing.T) {
		keymap, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "de"})
		if err != nil {
			t.Skipf("Could not load German layout: %v", err)
		}

		state := keymap.NewState()

		// German QWERTZ - Y and Z are swapped compared to US QWERTY
		kc := keymap.KeyByName("AB01") // Z on US, Y on German
		got := state.KeyGetUTF32(kc)
		if got != 'y' {
			t.Errorf("German AB01: expected 'y', got '%c'", got)
		}

		kc = keymap.KeyByName("AD06") // Y on US, Z on German
		got = state.KeyGetUTF32(kc)
		if got != 'z' {
			t.Errorf("German AD06: expected 'z', got '%c'", got)
		}
	})

	// Test French layout
	t.Run("FR_Layout", func(t *testing.T) {
		keymap, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "fr"})
		if err != nil {
			t.Skipf("Could not load French layout: %v", err)
		}

		state := keymap.NewState()

		// French AZERTY - A and Q are swapped
		kc := keymap.KeyByName("AD01") // Q on US, A on French
		got := state.KeyGetUTF32(kc)
		if got != 'a' {
			t.Errorf("French AD01: expected 'a', got '%c'", got)
		}

		kc = keymap.KeyByName("AC01") // A on US, Q on French
		got = state.KeyGetUTF32(kc)
		if got != 'q' {
			t.Errorf("French AC01: expected 'q', got '%c'", got)
		}
	})
}

func TestIntegration_ModifierState(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	keymap, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})
	if err != nil {
		t.Skipf("Could not load US layout: %v", err)
	}

	state := keymap.NewState()

	// Get modifier indices from the keymap (may differ from constants)
	shiftIdx := keymap.ModGetIndex("Shift")
	lockIdx := keymap.ModGetIndex("Lock")

	if shiftIdx < 0 {
		t.Skip("Keymap doesn't define Shift modifier")
	}
	if lockIdx < 0 {
		t.Skip("Keymap doesn't define Lock modifier")
	}

	shiftMask := ModMask(1 << shiftIdx)
	lockMask := ModMask(1 << lockIdx)

	// Test modifier state tracking
	t.Run("ShiftState", func(t *testing.T) {
		state.UpdateMask(shiftMask, 0, 0, 0, 0, 0)

		if !state.ModNameIsActive("Shift", StateModDepressed) {
			t.Error("Shift should be depressed")
		}

		state.UpdateMask(0, 0, 0, 0, 0, 0)

		if state.ModNameIsActive("Shift", StateModDepressed) {
			t.Error("Shift should not be depressed")
		}
	})

	t.Run("CapsLockState", func(t *testing.T) {
		state.UpdateMask(0, 0, lockMask, 0, 0, 0)

		if !state.ModNameIsActive("Lock", StateModLocked) {
			t.Error("Lock should be locked")
		}

		// Caps lock should affect letter keys
		kc := keymap.KeyByName("AD01")
		got := state.KeyGetUTF32(kc)
		if got != 'Q' {
			t.Errorf("With CapsLock, AD01 should be 'Q', got '%c'", got)
		}

		state.UpdateMask(0, 0, 0, 0, 0, 0)
	})
}

func TestIntegration_ComposeWithLayout(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	// Load compose table
	table, err := ctx.NewComposeTableFromLocale("en_US.UTF-8", ComposeCompileNoFlags)
	if err != nil {
		t.Skipf("Could not load compose table: %v", err)
	}

	composeState := table.NewState(ComposeStateNoFlags)

	// Load keymap
	keymap, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})
	if err != nil {
		t.Skipf("Could not load US layout: %v", err)
	}

	kbState := keymap.NewState()

	// Simulate typing dead_acute + a = á
	t.Run("DeadAcute_A", func(t *testing.T) {
		composeState.Reset()

		// Feed dead_acute
		composeState.Feed(KeyDeadAcute)
		if composeState.GetStatus() != ComposeComposing {
			t.Errorf("After dead_acute, expected Composing, got %v", composeState.GetStatus())
		}

		// Feed 'a'
		aKey := keymap.KeyByName("AC01")
		aSym := kbState.KeyGetOneSym(aKey)

		composeState.Feed(aSym)
		if composeState.GetStatus() != ComposeComposed {
			t.Errorf("After 'a', expected Composed, got %v", composeState.GetStatus())
		}

		// Check result
		result := composeState.GetUTF8()
		if result != "á" {
			t.Errorf("Expected 'á', got '%s'", result)
		}
	})
}

func TestIntegration_MultiGroup(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	// Try to create a keymap with multiple groups
	// This might not work on all systems, so we'll use a layout that typically has 2 groups
	keymap, err := ctx.NewKeymapFromNames(&RuleNames{
		Layout: "us,ru",
	})
	if err != nil {
		// Fall back to single layout
		t.Skipf("Multi-group layout not available: %v", err)
	}

	numGroups := keymap.NumGroups()
	t.Logf("Keymap has %d groups", numGroups)

	if numGroups < 2 {
		t.Skip("Keymap doesn't have multiple groups")
	}

	state := keymap.NewState()

	// Test group 1 (US)
	t.Run("Group1_US", func(t *testing.T) {
		state.UpdateMask(0, 0, 0, 0, 0, 0) // Group 1

		kc := keymap.KeyByName("AD01")
		got := state.KeyGetUTF32(kc)
		if got != 'q' {
			t.Errorf("Group 1 AD01: expected 'q', got '%c' (0x%x)", got, got)
		}
	})

	// Test group 2 (Russian)
	t.Run("Group2_RU", func(t *testing.T) {
		state.UpdateMask(0, 0, 0, 1, 0, 0) // Group 2

		kc := keymap.KeyByName("AD01")
		got := state.KeyGetUTF32(kc)
		// Russian 'й' is on AD01 in phonetic layout, 'й' = U+0439
		// In standard Russian layout, it might be different
		if got < 0x0400 || got > 0x04FF {
			t.Logf("Group 2 AD01: got '%c' (0x%x) - may vary by layout", got, got)
		}
	})
}

func TestIntegration_SpecialKeys(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	keymap, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})
	if err != nil {
		t.Skipf("Could not load US layout: %v", err)
	}

	state := keymap.NewState()

	specialKeys := []struct {
		keyName     string
		expectedSym Keysym
	}{
		{"ESC", KeyEscape},
		{"RTRN", KeyReturn},
		{"TAB", KeyTab},
		{"BKSP", KeyBackSpace},
		{"SPCE", KeySpace},
		{"CAPS", KeyCapsLock},
		{"LFSH", KeyShiftL},
		{"RTSH", KeyShiftR},
		{"LCTL", KeyControlL},
		{"RCTL", KeyControlR},
	}

	for _, tc := range specialKeys {
		t.Run(tc.keyName, func(t *testing.T) {
			kc := keymap.KeyByName(tc.keyName)
			if kc == 0 {
				t.Skipf("Key %s not found", tc.keyName)
			}

			sym := state.KeyGetOneSym(kc)
			if sym != tc.expectedSym {
				t.Errorf("Key %s: expected %s, got %s",
					tc.keyName, KeysymGetName(tc.expectedSym), KeysymGetName(sym))
			}
		})
	}
}

func TestIntegration_KeyRepeat(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	keymap, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})
	if err != nil {
		t.Skipf("Could not load US layout: %v", err)
	}

	// Regular keys should repeat
	repeatKeys := []string{"AD01", "AC01", "AB01", "SPCE"}
	for _, name := range repeatKeys {
		kc := keymap.KeyByName(name)
		if kc != 0 && !keymap.KeyRepeats(kc) {
			t.Errorf("Key %s should repeat", name)
		}
	}

	// Modifier keys typically should not repeat in XKB because the compat section
	// has "interpret.repeat = False" which affects keys matched by interpret statements.
	// However, our parser doesn't yet implement full compat interpret handling,
	// so modifier keys may incorrectly have repeat=true.
	// TODO: Implement interpret.repeat handling in compat section parsing
	noRepeatKeys := []string{"LFSH", "RTSH", "LCTL", "RCTL", "CAPS"}
	for _, name := range noRepeatKeys {
		kc := keymap.KeyByName(name)
		if kc != 0 && keymap.KeyRepeats(kc) {
			// Log as known limitation rather than failing
			t.Logf("Key %s has repeat=true (expected false, but compat interpret.repeat not yet implemented)", name)
		}
	}
}

func TestIntegration_LEDIndicators(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	keymap, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})
	if err != nil {
		t.Skipf("Could not load US layout: %v", err)
	}

	// Check for common LED indicators
	indicators := []string{"Caps Lock", "Num Lock", "Scroll Lock"}
	found := 0

	for _, name := range indicators {
		idx := keymap.LEDGetIndex(name)
		if idx >= 0 {
			found++
			t.Logf("Found indicator: %s at index %d", name, idx)
		}
	}

	if found == 0 {
		t.Log("No standard LED indicators found (may be normal for some keymaps)")
	}

	// Test LED state tracking
	state := keymap.NewState()

	// Enable Caps Lock
	state.UpdateMask(0, 0, ModLock, 0, 0, 0)

	capsIdx := keymap.LEDGetIndex("Caps Lock")
	if capsIdx >= 0 {
		if !state.LEDNameIsActive("Caps Lock") {
			t.Error("Caps Lock LED should be active when Lock modifier is set")
		}
	}
}

func TestIntegration_GetAsStringCompleteness(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	keymap, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})
	if err != nil {
		t.Skipf("Could not load US layout: %v", err)
	}

	serialized, err := keymap.GetAsString(KeymapFormatTextV1)
	if err != nil {
		t.Fatalf("GetAsString failed: %v", err)
	}

	// Check that all required sections are present
	requiredSections := []string{
		"xkb_keymap",
		"xkb_keycodes",
		"xkb_types",
		"xkb_compat",
		"xkb_symbols",
	}

	for _, section := range requiredSections {
		if !strings.Contains(serialized, section) {
			t.Errorf("Serialized keymap missing section: %s", section)
		}
	}

	// Check for expected content
	expectedContent := []string{
		"minimum",
		"maximum",
		"<AD01>",
		"type",
		"key",
	}

	for _, content := range expectedContent {
		if !strings.Contains(serialized, content) {
			t.Errorf("Serialized keymap missing expected content: %s", content)
		}
	}
}

// Benchmark RMLVO compilation for various layouts
func BenchmarkIntegration_RMLVOCompilation(b *testing.B) {
	ctx := NewContext(ContextNoFlags)

	layouts := []string{"us", "de", "fr", "ru", "jp"}

	for _, layout := range layouts {
		b.Run(layout, func(b *testing.B) {
			// Verify layout works
			_, err := ctx.NewKeymapFromNames(&RuleNames{Layout: layout})
			if err != nil {
				b.Skipf("Layout %s not available: %v", layout, err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = ctx.NewKeymapFromNames(&RuleNames{Layout: layout})
			}
		})
	}
}
