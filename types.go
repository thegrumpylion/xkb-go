package xkb

// Keysym is an XKB key symbol.
//
// Keysyms are 32-bit values that represent the symbols on keyboard keys.
// Values below 0x01000000 are defined in X11 keysym headers.
// Values 0x01000000-0x0110FFFF are Unicode codepoints + 0x01000000.
//
// Use [KeysymGetName] to get the name of a keysym, [KeysymFromName] to look up
// by name, and [KeysymToUTF32] to convert to a Unicode codepoint.
//
// See keysym_generated.go for keysym constants (KeyBackSpace, KeyReturn, etc.)
type Keysym uint32

// Keycode is a physical key identifier.
//
// In evdev (Linux), keycode = scancode + 8. Keycodes are translated to
// keysyms using [State.KeyGetOneSym] or [State.KeyGetSyms].
type Keycode uint32

// ModMask is a bitmask of active modifiers.
//
// Use the [ModShift], [ModLock], [ModControl], and ModMod1-5 constants
// to construct masks. Pass masks to [State.UpdateMask] to update state.
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
//
// Use [State.ModIndexIsActive] to check if a modifier at a given index is active.
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
//
// Level 0 is the base level, Level 1 is typically Shift, etc.
// The active level is determined by the key type and current modifier state.
type Level uint8

// Group represents a keyboard layout index.
//
// Most keyboards have 1-4 groups (e.g., Group 0 = English, Group 1 = Russian).
// Use [State.SerializeGroup] to get the current group and [Keymap.NumGroups]
// to query the number of groups.
type Group uint8

// KeyDirection indicates whether a key was pressed or released.
//
// Used with [State.UpdateKey] to update keyboard state based on key events.
type KeyDirection int

const (
	// KeyReleased indicates a key release event.
	KeyReleased KeyDirection = 0
	// KeyPressed indicates a key press event.
	KeyPressed KeyDirection = 1
)

// StateComponent identifies which parts of keyboard state changed.
//
// Returned by [State.UpdateMask] and [State.UpdateKey] to indicate what changed.
// Use with [State.ModIndexIsActive] and [State.SerializeMods] to query specific
// state components.
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
//
// Used with [Context.NewKeymapFromString], [Context.NewKeymapFromFile],
// and [Keymap.GetAsString].
type KeymapFormat uint32

const (
	// KeymapFormatTextV1 is the XKB text format, version 1.
	// This is the format used by Wayland compositors in wl_keyboard.keymap events.
	KeymapFormatTextV1 KeymapFormat = 1
)

// KeymapCompileFlags controls keymap compilation.
//
// Currently no flags are defined; pass [KeymapCompileNoFlags].
type KeymapCompileFlags uint32

const (
	// KeymapCompileNoFlags is the default keymap compilation flags.
	KeymapCompileNoFlags KeymapCompileFlags = 0
)

// ComposeCompileFlags controls compose table compilation.
//
// Used with [Context.NewComposeTableFromLocale] and [Context.NewComposeTableFromFile].
type ComposeCompileFlags uint32

const (
	// ComposeCompileNoFlags is the default compose table compilation flags.
	ComposeCompileNoFlags ComposeCompileFlags = 0
)

// ComposeStateFlags controls compose state creation.
//
// Used with [ComposeTable.NewState].
type ComposeStateFlags uint32

const (
	// ComposeStateNoFlags is the default compose state flags.
	ComposeStateNoFlags ComposeStateFlags = 0
)

// ComposeStatus indicates the state of a compose sequence.
//
// Retrieved via [ComposeState.GetStatus] after calling [ComposeState.Feed].
type ComposeStatus int

const (
	// ComposeNothing means no compose sequence is in progress.
	ComposeNothing ComposeStatus = iota
	// ComposeComposing means a compose sequence is in progress, waiting for more input.
	ComposeComposing
	// ComposeComposed means a compose sequence completed successfully.
	// Call [ComposeState.GetOneSym] or [ComposeState.GetUTF8] to get the result.
	ComposeComposed
	// ComposeCancelled means a compose sequence was cancelled by an invalid keysym.
	ComposeCancelled
)

// ComposeFeedResult indicates how [ComposeState.Feed] handled a keysym.
type ComposeFeedResult int

const (
	// ComposeFeedIgnored means the keysym was ignored (not a valid compose starter).
	ComposeFeedIgnored ComposeFeedResult = iota
	// ComposeFeedAccepted means the keysym was consumed by the compose sequence.
	ComposeFeedAccepted
)

// RuleNames specifies RMLVO (Rules, Model, Layout, Variant, Options) for keymap compilation.
//
// Pass to [Context.NewKeymapFromNames] to build a keymap from system XKB data.
// Empty fields fall back to defaults: Rules="evdev", Model="pc105", Layout="us".
type RuleNames struct {
	// Rules is the rules file to use, usually "evdev".
	Rules string
	// Model is the keyboard model, e.g., "pc105", "pc104", "macbook".
	Model string
	// Layout is the keyboard layout, e.g., "us", "de", "us,ru" for multiple.
	Layout string
	// Variant is the layout variant, e.g., "intl", "dvorak", ",phonetic" for multiple.
	Variant string
	// Options are extra options, e.g., "ctrl:nocaps,compose:ralt".
	Options string
}
