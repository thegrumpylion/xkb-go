package xkb

// Keymap is an immutable compiled keyboard mapping.
// It contains all information about keys, layouts, types, and modifiers.
//
// Keymap is safe to share across goroutines after creation.
type Keymap struct {
	ctx *Context

	// Keycode ↔ name mapping
	keycodeNames   map[Keycode]string
	keycodesByName map[string]Keycode
	minKeycode     Keycode
	maxKeycode     Keycode

	// Key type definitions
	types     map[string]*KeyType
	typesList []*KeyType

	// Per-key information
	keys map[Keycode]*Key

	// Modifier definitions
	modNames    [8]string          // Real modifier names (Shift, Lock, Control, Mod1-5)
	virtualMods map[string]ModMask // Virtual modifier → real modifier mask

	// LED/indicator definitions
	leds map[string]*LED

	// Group (layout) names
	groupNames []string
	numGroups  int
}

// KeyType defines how modifiers affect the shift level of a key.
type KeyType struct {
	name      string
	mods      ModMask // Modifiers this type considers
	numLevels int
	entries   []KeyTypeEntry
}

// KeyTypeEntry maps a modifier combination to a level.
type KeyTypeEntry struct {
	mods     ModMask // Modifier combination
	level    Level   // Resulting level
	preserve ModMask // Modifiers to preserve (not consume)
}

// Key holds per-key information.
type Key struct {
	keycode Keycode
	name    string
	groups  []KeyGroup
	repeats bool
	vmodmap ModMask // Virtual modifiers this key activates
}

// KeyGroup holds key symbols for one group (layout).
type KeyGroup struct {
	keyType *KeyType
	levels  []KeyLevel
}

// KeyLevel holds keysyms for one shift level.
type KeyLevel struct {
	syms []Keysym // Usually 1, rarely more (e.g., for key aliases)
}

// LED represents a keyboard LED indicator.
type LED struct {
	name  string
	mods  ModMask // Modifiers that activate this LED
	group Group   // Group that activates this LED
}

// MinKeycode returns the minimum keycode in the keymap.
func (km *Keymap) MinKeycode() Keycode {
	return km.minKeycode
}

// MaxKeycode returns the maximum keycode in the keymap.
func (km *Keymap) MaxKeycode() Keycode {
	return km.maxKeycode
}

// KeyGetName returns the symbolic name for a keycode (e.g., "AD01" for Q).
// Returns empty string if keycode is not found.
func (km *Keymap) KeyGetName(keycode Keycode) string {
	return km.keycodeNames[keycode]
}

// KeyByName returns the keycode for a symbolic name.
// Returns 0 if name is not found.
func (km *Keymap) KeyByName(name string) Keycode {
	return km.keycodesByName[name]
}

// NumGroups returns the number of groups (layouts) in the keymap.
func (km *Keymap) NumGroups() int {
	return km.numGroups
}

// GroupName returns the name of a group (layout).
// Returns empty string if group index is out of range.
func (km *Keymap) GroupName(group Group) string {
	if int(group) >= len(km.groupNames) {
		return ""
	}
	return km.groupNames[group]
}

// NumTypes returns the number of key types in the keymap.
func (km *Keymap) NumTypes() int {
	return len(km.typesList)
}

// ModGetIndex returns the index of a modifier by name.
// Returns -1 if not found.
//
// Works for both real modifiers (Shift, Lock, Control, Mod1-5)
// and virtual modifiers (Alt, Super, etc.).
func (km *Keymap) ModGetIndex(name string) int {
	// Check real modifiers first
	for i, n := range km.modNames {
		if n == name {
			return i
		}
	}
	// Virtual modifiers are indexed after real modifiers
	// but we return the real modifier index they map to
	if mask, ok := km.virtualMods[name]; ok {
		// Find the first set bit
		for i := 0; i < 8; i++ {
			if mask&(1<<i) != 0 {
				return i
			}
		}
	}
	return -1
}

// NumLEDs returns the number of LED indicators in the keymap.
func (km *Keymap) NumLEDs() int {
	return len(km.leds)
}

// LEDGetIndex returns the index of an LED by name.
// Returns -1 if not found.
func (km *Keymap) LEDGetIndex(name string) int {
	i := 0
	for n := range km.leds {
		if n == name {
			return i
		}
		i++
	}
	return -1
}

// LEDGetName returns the name of an LED by index.
// Returns empty string if index is out of range.
func (km *Keymap) LEDGetName(index int) string {
	i := 0
	for n := range km.leds {
		if i == index {
			return n
		}
		i++
	}
	return ""
}

// KeyRepeats returns whether a key should repeat when held.
func (km *Keymap) KeyRepeats(keycode Keycode) bool {
	if key, ok := km.keys[keycode]; ok {
		return key.repeats
	}
	return false
}

// NewState creates a new keyboard state for this keymap.
func (km *Keymap) NewState() *State {
	return &State{
		keymap: km,
	}
}

// Context returns the context this keymap was created from.
func (km *Keymap) Context() *Context {
	return km.ctx
}
