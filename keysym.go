package xkb

import (
	"strings"
	"unicode/utf8"
)

// KeysymToUTF32 converts a [Keysym] to a Unicode codepoint.
//
// Returns 0 if the keysym doesn't represent a printable character
// (e.g., modifier keys, function keys).
// See also [KeysymToUTF8] for the UTF-8 encoded string.
func KeysymToUTF32(keysym Keysym) rune {
	// Unicode keysyms: 0x01000000 + codepoint
	if keysym >= 0x01000000 && keysym <= 0x0110ffff {
		return rune(keysym - 0x01000000)
	}

	// Latin-1 range maps directly to Unicode
	if keysym >= 0x0020 && keysym <= 0x007e {
		return rune(keysym)
	}
	if keysym >= 0x00a0 && keysym <= 0x00ff {
		return rune(keysym)
	}

	// Look up in generated table
	if r, ok := keysymToUnicode[keysym]; ok {
		return r
	}

	return 0
}

// KeysymToUTF8 converts a [Keysym] to a UTF-8 string.
//
// Returns empty string if the keysym doesn't represent a printable character.
// See also [KeysymToUTF32] for the raw codepoint.
func KeysymToUTF8(keysym Keysym) string {
	r := KeysymToUTF32(keysym)
	if r == 0 {
		return ""
	}
	buf := make([]byte, utf8.UTFMax)
	n := utf8.EncodeRune(buf, r)
	return string(buf[:n])
}

// UTF32ToKeysym converts a Unicode codepoint to a [Keysym].
//
// For codepoints outside the basic keysym range (Latin-1), returns
// Unicode keysym format (0x01000000 + codepoint).
// Returns [KeyNoSymbol] if the codepoint is invalid.
func UTF32ToKeysym(r rune) Keysym {
	// Latin-1 characters map directly
	if r >= 0x0020 && r <= 0x007e {
		return Keysym(r)
	}
	if r >= 0x00a0 && r <= 0x00ff {
		return Keysym(r)
	}

	// Check generated table for specific keysym
	for ks, u := range keysymToUnicode {
		if u == r {
			return ks
		}
	}

	// Use Unicode keysym format
	if r >= 0x100 && r <= 0x10ffff {
		return Keysym(r + 0x01000000)
	}

	return KeyNoSymbol
}

// KeysymGetName returns the name of a [Keysym] (e.g., "Return", "a", "Shift_L").
//
// Returns empty string if the keysym is not recognized.
// See also [KeysymFromName] for the reverse lookup.
func KeysymGetName(keysym Keysym) string {
	// Check generated table
	if name, ok := keysymNames[keysym]; ok {
		return name
	}

	// For Unicode keysyms, return the Unicode codepoint name
	if keysym >= 0x01000000 && keysym <= 0x0110ffff {
		r := rune(keysym - 0x01000000)
		return string(r)
	}

	return ""
}

// KeysymFromName returns the [Keysym] for a name.
//
// Returns [KeyNoSymbol] if the name is not recognized.
// Use [KeysymNameCaseInsensitive] flag for case-insensitive matching.
// See also [KeysymGetName] for the reverse lookup.
func KeysymFromName(name string, flags KeysymNameFlags) Keysym {
	// Check generated table
	if ks, ok := keysymsByName[name]; ok {
		return ks
	}

	// Try case-insensitive match if requested
	if flags&KeysymNameCaseInsensitive != 0 {
		nameLower := strings.ToLower(name)
		for n, ks := range keysymsByName {
			if strings.ToLower(n) == nameLower {
				return ks
			}
		}
	}

	return KeyNoSymbol
}

// KeysymNameFlags controls [KeysymFromName] lookup behavior.
type KeysymNameFlags uint32

const (
	// KeysymNameNoFlags is the default (case-sensitive) lookup.
	KeysymNameNoFlags KeysymNameFlags = 0

	// KeysymNameCaseInsensitive performs case-insensitive lookup.
	// This is slower but matches names regardless of case.
	KeysymNameCaseInsensitive KeysymNameFlags = 1 << 0
)

// KeysymIsModifier returns true if the [Keysym] is a modifier key.
//
// Modifier keys include Shift, Control, Alt, Super, Caps Lock, Num Lock, etc.
func KeysymIsModifier(keysym Keysym) bool {
	return (keysym >= KeyShiftL && keysym <= KeyHyperR) ||
		(keysym >= KeyISOLock && keysym <= KeyISOLevel5Lock) ||
		keysym == KeyNumLock ||
		keysym == KeyCapsLock ||
		keysym == KeyScrollLock
}

// KeysymIsKeypad returns true if the [Keysym] is a keypad (numpad) key.
func KeysymIsKeypad(keysym Keysym) bool {
	return keysym >= KeyKPSpace && keysym <= KeyKP9
}

// KeysymIsFunctionKey returns true if the [Keysym] is a function key (F1-F35).
func KeysymIsFunctionKey(keysym Keysym) bool {
	return keysym >= KeyF1 && keysym <= KeyF35
}
