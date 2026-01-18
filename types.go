package xkb

// Keysym is an XKB key symbol.
// Keysyms are 32-bit values that represent the symbols on keyboard keys.
// Values below 0x01000000 are defined in X11 keysym headers.
// Values 0x01000000-0x0110FFFF are Unicode codepoints + 0x01000000.
type Keysym uint32

// Common keysym constants.
const (
	KeyNoSymbol Keysym = 0

	// TTY function keys
	KeyBackSpace Keysym = 0xff08
	KeyTab       Keysym = 0xff09
	KeyLinefeed  Keysym = 0xff0a
	KeyClear     Keysym = 0xff0b
	KeyReturn    Keysym = 0xff0d
	KeyPause     Keysym = 0xff13
	KeyScrollLock Keysym = 0xff14
	KeySysReq    Keysym = 0xff15
	KeyEscape    Keysym = 0xff1b
	KeyDelete    Keysym = 0xffff

	// Cursor control & motion
	KeyHome  Keysym = 0xff50
	KeyLeft  Keysym = 0xff51
	KeyUp    Keysym = 0xff52
	KeyRight Keysym = 0xff53
	KeyDown  Keysym = 0xff54
	KeyPrior Keysym = 0xff55 // Page Up
	KeyPageUp Keysym = 0xff55
	KeyNext  Keysym = 0xff56 // Page Down
	KeyPageDown Keysym = 0xff56
	KeyEnd   Keysym = 0xff57
	KeyBegin Keysym = 0xff58

	// Misc functions
	KeySelect Keysym = 0xff60
	KeyPrint  Keysym = 0xff61
	KeyExecute Keysym = 0xff62
	KeyInsert Keysym = 0xff63
	KeyUndo   Keysym = 0xff65
	KeyRedo   Keysym = 0xff66
	KeyMenu   Keysym = 0xff67
	KeyFind   Keysym = 0xff68
	KeyCancel Keysym = 0xff69
	KeyHelp   Keysym = 0xff6a
	KeyBreak  Keysym = 0xff6b
	KeyNumLock Keysym = 0xff7f

	// Keypad
	KeyKPSpace     Keysym = 0xff80
	KeyKPTab       Keysym = 0xff89
	KeyKPEnter     Keysym = 0xff8d
	KeyKPF1        Keysym = 0xff91
	KeyKPF2        Keysym = 0xff92
	KeyKPF3        Keysym = 0xff93
	KeyKPF4        Keysym = 0xff94
	KeyKPHome      Keysym = 0xff95
	KeyKPLeft      Keysym = 0xff96
	KeyKPUp        Keysym = 0xff97
	KeyKPRight     Keysym = 0xff98
	KeyKPDown      Keysym = 0xff99
	KeyKPPrior     Keysym = 0xff9a
	KeyKPPageUp    Keysym = 0xff9a
	KeyKPNext      Keysym = 0xff9b
	KeyKPPageDown  Keysym = 0xff9b
	KeyKPEnd       Keysym = 0xff9c
	KeyKPBegin     Keysym = 0xff9d
	KeyKPInsert    Keysym = 0xff9e
	KeyKPDelete    Keysym = 0xff9f
	KeyKPEqual     Keysym = 0xffbd
	KeyKPMultiply  Keysym = 0xffaa
	KeyKPAdd       Keysym = 0xffab
	KeyKPSeparator Keysym = 0xffac
	KeyKPSubtract  Keysym = 0xffad
	KeyKPDecimal   Keysym = 0xffae
	KeyKPDivide    Keysym = 0xffaf
	KeyKP0         Keysym = 0xffb0
	KeyKP1         Keysym = 0xffb1
	KeyKP2         Keysym = 0xffb2
	KeyKP3         Keysym = 0xffb3
	KeyKP4         Keysym = 0xffb4
	KeyKP5         Keysym = 0xffb5
	KeyKP6         Keysym = 0xffb6
	KeyKP7         Keysym = 0xffb7
	KeyKP8         Keysym = 0xffb8
	KeyKP9         Keysym = 0xffb9

	// Function keys
	KeyF1  Keysym = 0xffbe
	KeyF2  Keysym = 0xffbf
	KeyF3  Keysym = 0xffc0
	KeyF4  Keysym = 0xffc1
	KeyF5  Keysym = 0xffc2
	KeyF6  Keysym = 0xffc3
	KeyF7  Keysym = 0xffc4
	KeyF8  Keysym = 0xffc5
	KeyF9  Keysym = 0xffc6
	KeyF10 Keysym = 0xffc7
	KeyF11 Keysym = 0xffc8
	KeyF12 Keysym = 0xffc9

	// Modifiers
	KeyShiftL   Keysym = 0xffe1
	KeyShiftR   Keysym = 0xffe2
	KeyControlL Keysym = 0xffe3
	KeyControlR Keysym = 0xffe4
	KeyCapsLock Keysym = 0xffe5
	KeyShiftLock Keysym = 0xffe6
	KeyMetaL    Keysym = 0xffe7
	KeyMetaR    Keysym = 0xffe8
	KeyAltL     Keysym = 0xffe9
	KeyAltR     Keysym = 0xffea
	KeySuperL   Keysym = 0xffeb
	KeySuperR   Keysym = 0xffec
	KeyHyperL   Keysym = 0xffed
	KeyHyperR   Keysym = 0xffee

	// ISO 9995 Function and Modifier Keys
	KeyISOLock         Keysym = 0xfe01
	KeyISOLevel2Latch  Keysym = 0xfe02
	KeyISOLevel3Shift  Keysym = 0xfe03
	KeyISOLevel3Latch  Keysym = 0xfe04
	KeyISOLevel3Lock   Keysym = 0xfe05
	KeyISOLevel5Shift  Keysym = 0xfe11
	KeyISOLevel5Latch  Keysym = 0xfe12
	KeyISOLevel5Lock   Keysym = 0xfe13

	// Dead keys
	KeyDeadGrave           Keysym = 0xfe50
	KeyDeadAcute           Keysym = 0xfe51
	KeyDeadCircumflex      Keysym = 0xfe52
	KeyDeadTilde           Keysym = 0xfe53
	KeyDeadMacron          Keysym = 0xfe54
	KeyDeadBreve           Keysym = 0xfe55
	KeyDeadAbovedot        Keysym = 0xfe56
	KeyDeadDiaeresis       Keysym = 0xfe57
	KeyDeadAbovering       Keysym = 0xfe58
	KeyDeadDoubleacute     Keysym = 0xfe59
	KeyDeadCaron           Keysym = 0xfe5a
	KeyDeadCedilla         Keysym = 0xfe5b
	KeyDeadOgonek          Keysym = 0xfe5c
	KeyDeadIota            Keysym = 0xfe5d
	KeyDeadVoicedSound     Keysym = 0xfe5e
	KeyDeadSemivoicedSound Keysym = 0xfe5f
	KeyDeadBelowdot        Keysym = 0xfe60

	// Special
	KeyMultiKey Keysym = 0xff20 // Compose key
)

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
