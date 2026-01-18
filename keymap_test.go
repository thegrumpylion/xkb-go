package xkb

import "testing"

func TestKeymapMinMaxKeycode(t *testing.T) {
	km := TestKeymap()

	if km.MinKeycode() != 8 {
		t.Errorf("MinKeycode() = %d, want 8", km.MinKeycode())
	}
	if km.MaxKeycode() != 255 {
		t.Errorf("MaxKeycode() = %d, want 255", km.MaxKeycode())
	}
}

func TestKeymapKeyGetName(t *testing.T) {
	km := TestKeymap()

	tests := []struct {
		keycode Keycode
		want    string
	}{
		{24, "AD01"}, // Q
		{38, "AC01"}, // A
		{52, "AB01"}, // Z
		{36, "RTRN"}, // Return
		{999, ""},    // Invalid
	}

	for _, tt := range tests {
		got := km.KeyGetName(tt.keycode)
		if got != tt.want {
			t.Errorf("KeyGetName(%d) = %q, want %q", tt.keycode, got, tt.want)
		}
	}
}

func TestKeymapKeyByName(t *testing.T) {
	km := TestKeymap()

	tests := []struct {
		name string
		want Keycode
	}{
		{"AD01", 24}, // Q
		{"AC01", 38}, // A
		{"RTRN", 36}, // Return
		{"NONE", 0},  // Invalid
	}

	for _, tt := range tests {
		got := km.KeyByName(tt.name)
		if got != tt.want {
			t.Errorf("KeyByName(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestKeymapNumGroups(t *testing.T) {
	km := TestKeymap()

	if km.NumGroups() != 1 {
		t.Errorf("NumGroups() = %d, want 1", km.NumGroups())
	}
}

func TestKeymapGroupName(t *testing.T) {
	km := TestKeymap()

	if km.GroupName(0) != "English (US)" {
		t.Errorf("GroupName(0) = %q, want %q", km.GroupName(0), "English (US)")
	}
	if km.GroupName(1) != "" {
		t.Errorf("GroupName(1) = %q, want empty", km.GroupName(1))
	}
}

func TestKeymapNumTypes(t *testing.T) {
	km := TestKeymap()

	if km.NumTypes() != 4 {
		t.Errorf("NumTypes() = %d, want 4", km.NumTypes())
	}
}

func TestKeymapModGetIndex(t *testing.T) {
	km := TestKeymap()

	tests := []struct {
		name string
		want int
	}{
		{"Shift", 0},
		{"Lock", 1},
		{"Control", 2},
		{"Mod1", 3},
		{"Mod4", 6},
		{"Alt", 3},   // Virtual mod -> Mod1
		{"Super", 6}, // Virtual mod -> Mod4
		{"None", -1}, // Invalid
	}

	for _, tt := range tests {
		got := km.ModGetIndex(tt.name)
		if got != tt.want {
			t.Errorf("ModGetIndex(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestKeymapNumLEDs(t *testing.T) {
	km := TestKeymap()

	if km.NumLEDs() != 2 {
		t.Errorf("NumLEDs() = %d, want 2", km.NumLEDs())
	}
}

func TestKeymapLEDGetIndex(t *testing.T) {
	km := TestKeymap()

	// Note: map iteration order is not guaranteed, so we just check valid/invalid
	idx := km.LEDGetIndex("Caps Lock")
	if idx < 0 || idx > 1 {
		t.Errorf("LEDGetIndex(\"Caps Lock\") = %d, want 0 or 1", idx)
	}

	idx = km.LEDGetIndex("NonExistent")
	if idx != -1 {
		t.Errorf("LEDGetIndex(\"NonExistent\") = %d, want -1", idx)
	}
}

func TestKeymapLEDGetName(t *testing.T) {
	km := TestKeymap()

	// Get valid LED names
	name0 := km.LEDGetName(0)
	name1 := km.LEDGetName(1)

	// Both should be non-empty
	if name0 == "" || name1 == "" {
		t.Error("LED names should not be empty")
	}

	// Out of range
	if km.LEDGetName(99) != "" {
		t.Error("LEDGetName(99) should return empty")
	}
}

func TestKeymapKeyRepeats(t *testing.T) {
	km := TestKeymap()

	// Letters repeat
	if !km.KeyRepeats(38) { // 'a'
		t.Error("Letter 'a' should repeat")
	}

	// Space repeats
	if !km.KeyRepeats(65) {
		t.Error("Space should repeat")
	}

	// Return doesn't repeat
	if km.KeyRepeats(36) {
		t.Error("Return should not repeat")
	}

	// Escape doesn't repeat
	if km.KeyRepeats(9) {
		t.Error("Escape should not repeat")
	}

	// Shift doesn't repeat
	if km.KeyRepeats(50) {
		t.Error("Shift should not repeat")
	}

	// Invalid keycode
	if km.KeyRepeats(999) {
		t.Error("Invalid keycode should not repeat")
	}
}

func TestKeymapNewState(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	if state == nil {
		t.Fatal("NewState() returned nil")
	}
	if state.Keymap() != km {
		t.Error("State.Keymap() should return the keymap")
	}
}

func TestKeymapContext(t *testing.T) {
	km := TestKeymap()
	ctx := km.Context()

	if ctx == nil {
		t.Error("Context() should not return nil")
	}
}

func TestKeymapGetAsString(t *testing.T) {
	km := TestKeymap()

	output, err := km.GetAsString(KeymapFormatTextV1)
	if err != nil {
		t.Fatalf("GetAsString failed: %v", err)
	}

	if output == "" {
		t.Fatal("GetAsString returned empty string")
	}

	// Verify it contains expected sections
	if !contains(output, "xkb_keymap") {
		t.Error("Output should contain xkb_keymap")
	}
	if !contains(output, "xkb_keycodes") {
		t.Error("Output should contain xkb_keycodes")
	}
	if !contains(output, "xkb_types") {
		t.Error("Output should contain xkb_types")
	}
	if !contains(output, "xkb_compatibility") {
		t.Error("Output should contain xkb_compatibility")
	}
	if !contains(output, "xkb_symbols") {
		t.Error("Output should contain xkb_symbols")
	}

	// Verify it can be re-parsed
	ctx := NewContext(ContextNoFlags)
	reparsed, err := ctx.NewKeymapFromString([]byte(output), KeymapFormatTextV1)
	if err != nil {
		t.Fatalf("Failed to re-parse GetAsString output: %v", err)
	}

	// Verify reparsed keymap has same basic properties
	if reparsed.MinKeycode() != km.MinKeycode() {
		t.Errorf("Reparsed MinKeycode = %d, want %d", reparsed.MinKeycode(), km.MinKeycode())
	}
	if reparsed.MaxKeycode() != km.MaxKeycode() {
		t.Errorf("Reparsed MaxKeycode = %d, want %d", reparsed.MaxKeycode(), km.MaxKeycode())
	}
	if reparsed.NumGroups() != km.NumGroups() {
		t.Errorf("Reparsed NumGroups = %d, want %d", reparsed.NumGroups(), km.NumGroups())
	}
}

func TestKeymapGetAsStringUnsupportedFormat(t *testing.T) {
	km := TestKeymap()

	_, err := km.GetAsString(KeymapFormat(99))
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestKeymapGetAsStringRoundTrip(t *testing.T) {
	// Load real keymap
	ctx := NewContext(ContextNoFlags)
	original, err := ctx.NewKeymapFromFile("testdata/us.xkb", KeymapFormatTextV1)
	if err != nil {
		t.Fatalf("Failed to load keymap: %v", err)
	}

	// Serialize
	output, err := original.GetAsString(KeymapFormatTextV1)
	if err != nil {
		t.Fatalf("GetAsString failed: %v", err)
	}

	// Re-parse
	reparsed, err := ctx.NewKeymapFromString([]byte(output), KeymapFormatTextV1)
	if err != nil {
		t.Fatalf("Failed to re-parse: %v", err)
	}

	// Verify key functionality is preserved
	state1 := original.NewState()
	state2 := reparsed.NewState()

	// Test a few keys
	testKeys := []Keycode{38, 24, 10, 36} // a, q, 1, Return
	for _, kc := range testKeys {
		sym1 := state1.KeyGetOneSym(kc)
		sym2 := state2.KeyGetOneSym(kc)
		if sym1 != sym2 {
			t.Errorf("Key %d: original sym %#x != reparsed sym %#x", kc, sym1, sym2)
		}
	}

	// Test with Shift
	state1.UpdateMask(ModShift, 0, 0, 0, 0, 0)
	state2.UpdateMask(ModShift, 0, 0, 0, 0, 0)

	for _, kc := range testKeys {
		sym1 := state1.KeyGetOneSym(kc)
		sym2 := state2.KeyGetOneSym(kc)
		if sym1 != sym2 {
			t.Errorf("Key %d with Shift: original sym %#x != reparsed sym %#x", kc, sym1, sym2)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
