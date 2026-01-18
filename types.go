package xkb

// Keysym is an XKB key symbol.
// Keysyms are 32-bit values that represent the symbols on keyboard keys.
// Values below 0x01000000 are defined in X11 keysym headers.
// Values 0x01000000-0x0110FFFF are Unicode codepoints + 0x01000000.
// See keysym_generated.go for keysym constants (KeyBackSpace, KeyReturn, etc.)
type Keysym uint32

// Keycode is a physical key identifier.
// In evdev (Linux), keycode = scancode + 8.
type Keycode uint32

// ModMask is a bitmask of active modifiers.
type ModMask uint32

// Real modifier indices (0-7).
const (
	ModShift   ModMask = 1 << 0
	ModLock    ModMask = 1 << 1 // Caps Lock
	ModControl ModMask = 1 << 2
	ModMod1    ModMask = 1 << 3 // Usually Alt
	ModMod2    ModMask = 1 << 4 // Usually Num Lock
	ModMod3    ModMask = 1 << 5
	ModMod4    ModMask = 1 << 6 // Usually Super/Win
	ModMod5    ModMask = 1 << 7 // Usually ISO_Level3_Shift (AltGr)
)

// ModIndex is an index into the modifier array (0-7 for real modifiers).
type ModIndex uint8

// Real modifier indices.
const (
	ModIndexShift   ModIndex = 0
	ModIndexLock    ModIndex = 1
	ModIndexControl ModIndex = 2
	ModIndexMod1    ModIndex = 3
	ModIndexMod2    ModIndex = 4
	ModIndexMod3    ModIndex = 5
	ModIndexMod4    ModIndex = 6
	ModIndexMod5    ModIndex = 7
)

// Level is a shift level within a key group.
// Level 0 is the base level, Level 1 is typically Shift, etc.
type Level uint8

// Group (layout) index.
// Most keyboards have 1-4 groups (e.g., English, Russian).
type Group uint8

// KeyDirection indicates whether a key was pressed or released.
type KeyDirection int

const (
	KeyReleased KeyDirection = 0
	KeyPressed  KeyDirection = 1
)

// StateComponent identifies which parts of keyboard state changed.
type StateComponent uint32

const (
	StateModDepressed   StateComponent = 1 << 0
	StateModLatched     StateComponent = 1 << 1
	StateModLocked      StateComponent = 1 << 2
	StateModEffective   StateComponent = 1 << 3
	StateGroupDepressed StateComponent = 1 << 4
	StateGroupLatched   StateComponent = 1 << 5
	StateGroupLocked    StateComponent = 1 << 6
	StateGroupEffective StateComponent = 1 << 7
	StateLEDs           StateComponent = 1 << 8
)

// StateMods combines all modifier state components.
const StateMods = StateModDepressed | StateModLatched | StateModLocked | StateModEffective

// StateGroups combines all group state components.
const StateGroups = StateGroupDepressed | StateGroupLatched | StateGroupLocked | StateGroupEffective

// KeymapFormat specifies the format of a keymap string.
type KeymapFormat uint32

const (
	// KeymapFormatTextV1 is the XKB text format, version 1.
	// This is the format used by Wayland compositors.
	KeymapFormatTextV1 KeymapFormat = 1
)

// KeymapCompileFlags controls keymap compilation.
type KeymapCompileFlags uint32

const (
	KeymapCompileNoFlags KeymapCompileFlags = 0
)

// ComposeCompileFlags controls compose table compilation.
type ComposeCompileFlags uint32

const (
	ComposeCompileNoFlags ComposeCompileFlags = 0
)

// ComposeStateFlags controls compose state creation.
type ComposeStateFlags uint32

const (
	ComposeStateNoFlags ComposeStateFlags = 0
)

// ComposeStatus indicates the state of a compose sequence.
type ComposeStatus int

const (
	// ComposeNothing means no compose sequence is in progress.
	ComposeNothing ComposeStatus = iota
	// ComposeComposing means a compose sequence is in progress.
	ComposeComposing
	// ComposeComposed means a compose sequence completed successfully.
	ComposeComposed
	// ComposeCancelled means a compose sequence was cancelled.
	ComposeCancelled
)

// ComposeResult indicates how the compose state should be used.
type ComposeFeedResult int

const (
	// ComposeFeedIgnored means the keysym was ignored (not a valid compose starter).
	ComposeFeedIgnored ComposeFeedResult = iota
	// ComposeFeedAccepted means the keysym was consumed by the compose sequence.
	ComposeFeedAccepted
)

// RuleNames specifies RMLVO (Rules, Model, Layout, Variant, Options) for keymap compilation.
type RuleNames struct {
	Rules   string // The rules file to use (usually "evdev")
	Model   string // The keyboard model (e.g., "pc105")
	Layout  string // The keyboard layout (e.g., "us", "de", "us,ru")
	Variant string // The layout variant (e.g., "intl", "dvorak")
	Options string // Extra options (e.g., "ctrl:nocaps,compose:ralt")
}
