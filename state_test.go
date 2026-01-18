package xkb

import "testing"

func TestStateKeymap(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	if state.Keymap() != km {
		t.Error("State.Keymap() should return the keymap")
	}
}

func TestStateUpdateMask(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	// Update with Shift pressed
	changed := state.UpdateMask(ModShift, 0, 0, 0, 0, 0)
	if changed&StateModDepressed == 0 {
		t.Error("Expected StateModDepressed in changed")
	}
	if changed&StateModEffective == 0 {
		t.Error("Expected StateModEffective in changed")
	}

	// Update with same state - no change
	changed = state.UpdateMask(ModShift, 0, 0, 0, 0, 0)
	if changed != 0 {
		t.Errorf("Expected no change, got %d", changed)
	}

	// Update with Caps Lock locked
	changed = state.UpdateMask(0, 0, ModLock, 0, 0, 0)
	if changed&StateModLocked == 0 {
		t.Error("Expected StateModLocked in changed")
	}

	// Update group
	changed = state.UpdateMask(0, 0, 0, 0, 0, 1)
	if changed&StateGroupLocked == 0 {
		t.Error("Expected StateGroupLocked in changed")
	}
}

func TestStateKeyGetOneSym(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	tests := []struct {
		name    string
		keycode Keycode
		mods    ModMask
		want    Keysym
	}{
		{"a without mods", 38, 0, 'a'},
		{"a with Shift", 38, ModShift, 'A'},
		{"a with CapsLock", 38, ModLock, 'A'},
		{"a with Shift+CapsLock", 38, ModShift | ModLock, 'a'}, // Shift cancels Caps
		{"q without mods", 24, 0, 'q'},
		{"q with Shift", 24, ModShift, 'Q'},
		{"1 without mods", 10, 0, '1'},
		{"1 with Shift", 10, ModShift, '!'},
		{"Return", 36, 0, KeyReturn},
		{"Space", 65, 0, ' '},
		{"Escape", 9, 0, KeyEscape},
		{"invalid keycode", 999, 0, KeyNoSymbol},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state.UpdateMask(tt.mods, 0, 0, 0, 0, 0)
			got := state.KeyGetOneSym(tt.keycode)
			if got != tt.want {
				t.Errorf("KeyGetOneSym(%d) with mods %d = %#x (%s), want %#x (%s)",
					tt.keycode, tt.mods, got, KeysymGetName(got), tt.want, KeysymGetName(tt.want))
			}
		})
	}
}

func TestStateKeyGetSyms(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	syms := state.KeyGetSyms(38) // 'a' key
	if len(syms) != 1 {
		t.Fatalf("Expected 1 keysym, got %d", len(syms))
	}
	if syms[0] != 'a' {
		t.Errorf("Expected 'a', got %#x", syms[0])
	}

	// Invalid keycode
	syms = state.KeyGetSyms(999)
	if syms != nil {
		t.Errorf("Expected nil for invalid keycode, got %v", syms)
	}
}

func TestStateKeyGetUTF32(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	tests := []struct {
		keycode Keycode
		mods    ModMask
		want    rune
	}{
		{38, 0, 'a'},
		{38, ModShift, 'A'},
		{10, 0, '1'},
		{65, 0, ' '},
		{36, 0, 0}, // Return doesn't produce a character
	}

	for _, tt := range tests {
		state.UpdateMask(tt.mods, 0, 0, 0, 0, 0)
		got := state.KeyGetUTF32(tt.keycode)
		if got != tt.want {
			t.Errorf("KeyGetUTF32(%d) with mods %d = %#x, want %#x", tt.keycode, tt.mods, got, tt.want)
		}
	}
}

func TestStateKeyGetUTF8(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	got := state.KeyGetUTF8(38) // 'a' key
	if got != "a" {
		t.Errorf("KeyGetUTF8(38) = %q, want %q", got, "a")
	}

	state.UpdateMask(ModShift, 0, 0, 0, 0, 0)
	got = state.KeyGetUTF8(38)
	if got != "A" {
		t.Errorf("KeyGetUTF8(38) with Shift = %q, want %q", got, "A")
	}
}

func TestStateModNameIsActive(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	state.UpdateMask(ModShift, 0, ModLock, 0, 0, 0)

	if !state.ModNameIsActive("Shift", StateModDepressed) {
		t.Error("Shift should be depressed")
	}
	if !state.ModNameIsActive("Shift", StateModEffective) {
		t.Error("Shift should be effective")
	}
	if !state.ModNameIsActive("Lock", StateModLocked) {
		t.Error("Lock should be locked")
	}
	if state.ModNameIsActive("Control", StateModEffective) {
		t.Error("Control should not be active")
	}
	if state.ModNameIsActive("NonExistent", StateModEffective) {
		t.Error("NonExistent modifier should not be active")
	}
}

func TestStateModIndexIsActive(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	state.UpdateMask(ModShift|ModControl, 0, 0, 0, 0, 0)

	if !state.ModIndexIsActive(ModIndexShift, StateModDepressed) {
		t.Error("Shift index should be active")
	}
	if !state.ModIndexIsActive(ModIndexControl, StateModDepressed) {
		t.Error("Control index should be active")
	}
	if state.ModIndexIsActive(ModIndexLock, StateModDepressed) {
		t.Error("Lock index should not be active")
	}
}

func TestStateSerializeMods(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	state.UpdateMask(ModShift, ModMod1, ModLock, 0, 0, 0)

	depressed := state.SerializeMods(StateModDepressed)
	if depressed != ModShift {
		t.Errorf("Depressed mods = %d, want %d", depressed, ModShift)
	}

	latched := state.SerializeMods(StateModLatched)
	if latched != ModMod1 {
		t.Errorf("Latched mods = %d, want %d", latched, ModMod1)
	}

	locked := state.SerializeMods(StateModLocked)
	if locked != ModLock {
		t.Errorf("Locked mods = %d, want %d", locked, ModLock)
	}

	effective := state.SerializeMods(StateModEffective)
	want := ModShift | ModMod1 | ModLock
	if effective != want {
		t.Errorf("Effective mods = %d, want %d", effective, want)
	}
}

func TestStateSerializeGroup(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	state.UpdateMask(0, 0, 0, 0, 0, 0)
	if state.SerializeGroup(StateGroupEffective) != 0 {
		t.Error("Expected group 0")
	}

	// Note: TestKeymap has only 1 group, so group 1 wraps to 0
	state.UpdateMask(0, 0, 0, 1, 0, 0)
	if state.SerializeGroup(StateGroupDepressed) != 0 {
		t.Errorf("Expected depressed group 0 (wrapped from 1), got %d", state.SerializeGroup(StateGroupDepressed))
	}

	// Test latched group
	state.UpdateMask(0, 0, 0, 0, 1, 0)
	if state.SerializeGroup(StateGroupLatched) != 0 {
		t.Errorf("Expected latched group 0 (wrapped from 1), got %d", state.SerializeGroup(StateGroupLatched))
	}

	// Test locked group
	state.UpdateMask(0, 0, 0, 0, 0, 1)
	if state.SerializeGroup(StateGroupLocked) != 0 {
		t.Errorf("Expected locked group 0 (wrapped from 1), got %d", state.SerializeGroup(StateGroupLocked))
	}
}

func TestStateLEDNameIsActive(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	// No modifiers - no LEDs
	if state.LEDNameIsActive("Caps Lock") {
		t.Error("Caps Lock LED should be off")
	}

	// Lock Caps Lock
	state.UpdateMask(0, 0, ModLock, 0, 0, 0)
	if !state.LEDNameIsActive("Caps Lock") {
		t.Error("Caps Lock LED should be on")
	}

	// Num Lock
	state.UpdateMask(0, 0, ModMod2, 0, 0, 0)
	if !state.LEDNameIsActive("Num Lock") {
		t.Error("Num Lock LED should be on")
	}

	// Non-existent LED
	if state.LEDNameIsActive("NonExistent") {
		t.Error("Non-existent LED should be off")
	}
}

func TestStateUpdateKey(t *testing.T) {
	km := TestKeymap()
	state := km.NewState()

	// UpdateKey is not fully implemented yet, but should not panic
	changed := state.UpdateKey(38, KeyPressed)
	_ = changed // Currently returns 0
}

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
		for g := Group(0); g < 10; g++ {
			state.UpdateMask(0, 0, 0, g, 0, 0)
			effective := state.SerializeGroup(StateGroupEffective)
			if effective != 0 {
				t.Errorf("Group %d wrapped to %d, want 0", g, effective)
			}
		}
	})

	t.Run("keycode boundaries", func(t *testing.T) {
		keycodes := []Keycode{
			0,
			8,
			255,
			256,
			0xFFFFFFFF,
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
		for i := 0; i < 1000; i++ {
			if i%2 == 0 {
				state.UpdateMask(ModShift, 0, 0, 0, 0, 0)
			} else {
				state.UpdateMask(0, 0, 0, 0, 0, 0)
			}
		}
	})
}
