package xkb

import "unicode/utf8"

// KeysymToUTF32 converts a keysym to a Unicode codepoint.
// Returns 0 if the keysym doesn't represent a character.
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

	// Look up in table for other keysyms
	if r, ok := keysymToUnicode[keysym]; ok {
		return r
	}

	return 0
}

// KeysymToUTF8 converts a keysym to a UTF-8 string.
// Returns empty string if the keysym doesn't represent a character.
func KeysymToUTF8(keysym Keysym) string {
	r := KeysymToUTF32(keysym)
	if r == 0 {
		return ""
	}
	buf := make([]byte, utf8.UTFMax)
	n := utf8.EncodeRune(buf, r)
	return string(buf[:n])
}

// UTF32ToKeysym converts a Unicode codepoint to a keysym.
// For codepoints outside the basic keysym range, returns Unicode keysym format.
func UTF32ToKeysym(r rune) Keysym {
	// Latin-1 characters map directly
	if r >= 0x0020 && r <= 0x007e {
		return Keysym(r)
	}
	if r >= 0x00a0 && r <= 0x00ff {
		return Keysym(r)
	}

	// Check if there's a specific keysym for this codepoint
	if ks, ok := unicodeToKeysym[r]; ok {
		return ks
	}

	// Use Unicode keysym format
	if r >= 0x100 && r <= 0x10ffff {
		return Keysym(r + 0x01000000)
	}

	return KeyNoSymbol
}

// KeysymGetName returns the name of a keysym (e.g., "Return", "a", "Shift_L").
// Returns empty string if the keysym is not recognized.
func KeysymGetName(keysym Keysym) string {
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

// KeysymFromName returns the keysym for a name.
// Returns KeyNoSymbol if the name is not recognized.
//
// Flags can be used to control matching behavior (not yet implemented).
func KeysymFromName(name string, flags KeysymNameFlags) Keysym {
	if ks, ok := keysymsByName[name]; ok {
		return ks
	}

	// Try case-insensitive match if requested
	if flags&KeysymNameCaseInsensitive != 0 {
		// TODO: Implement case-insensitive lookup
	}

	return KeyNoSymbol
}

// KeysymNameFlags controls keysym name lookup behavior.
type KeysymNameFlags uint32

const (
	// KeysymNameNoFlags is the default (case-sensitive) lookup.
	KeysymNameNoFlags KeysymNameFlags = 0

	// KeysymNameCaseInsensitive performs case-insensitive lookup.
	KeysymNameCaseInsensitive KeysymNameFlags = 1 << 0
)

// KeysymIsModifier returns true if the keysym is a modifier key.
func KeysymIsModifier(keysym Keysym) bool {
	return (keysym >= KeyShiftL && keysym <= KeyHyperR) ||
		(keysym >= KeyISOLock && keysym <= KeyISOLevel5Lock) ||
		keysym == KeyNumLock ||
		keysym == KeyCapsLock ||
		keysym == KeyScrollLock
}

// KeysymIsKeypad returns true if the keysym is a keypad key.
func KeysymIsKeypad(keysym Keysym) bool {
	return keysym >= KeyKPSpace && keysym <= KeyKP9
}

// KeysymIsFunctionKey returns true if the keysym is a function key (F1-F35).
func KeysymIsFunctionKey(keysym Keysym) bool {
	return keysym >= KeyF1 && keysym <= 0xffd5 // F1-F35
}

// keysymToUnicode maps non-Latin1 keysyms to Unicode codepoints.
// This is a subset - the full table would be generated from keysymdef.h
var keysymToUnicode = map[Keysym]rune{
	// Latin-2
	0x01a1: 0x0104, // Aogonek
	0x01a2: 0x02d8, // breve
	0x01a3: 0x0141, // Lstroke
	0x01a5: 0x013d, // Lcaron
	0x01a6: 0x015a, // Sacute
	0x01a9: 0x0160, // Scaron
	0x01aa: 0x015e, // Scedilla
	0x01ab: 0x0164, // Tcaron
	0x01ac: 0x0179, // Zacute
	0x01ae: 0x017d, // Zcaron
	0x01af: 0x017b, // Zabovedot
	0x01b1: 0x0105, // aogonek
	0x01b2: 0x02db, // ogonek
	0x01b3: 0x0142, // lstroke
	0x01b5: 0x013e, // lcaron
	0x01b6: 0x015b, // sacute
	0x01b7: 0x02c7, // caron
	0x01b9: 0x0161, // scaron
	0x01ba: 0x015f, // scedilla
	0x01bb: 0x0165, // tcaron
	0x01bc: 0x017a, // zacute
	0x01bd: 0x02dd, // doubleacute
	0x01be: 0x017e, // zcaron
	0x01bf: 0x017c, // zabovedot
	0x01c0: 0x0154, // Racute
	0x01c3: 0x0102, // Abreve
	0x01c5: 0x0139, // Lacute
	0x01c6: 0x0106, // Cacute
	0x01c8: 0x010c, // Ccaron
	0x01ca: 0x0118, // Eogonek
	0x01cc: 0x011a, // Ecaron
	0x01cf: 0x010e, // Dcaron
	0x01d0: 0x0110, // Dstroke
	0x01d1: 0x0143, // Nacute
	0x01d2: 0x0147, // Ncaron
	0x01d5: 0x0150, // Odoubleacute
	0x01d8: 0x0158, // Rcaron
	0x01d9: 0x016e, // Uring
	0x01db: 0x0170, // Udoubleacute
	0x01de: 0x0162, // Tcedilla
	0x01e0: 0x0155, // racute
	0x01e3: 0x0103, // abreve
	0x01e5: 0x013a, // lacute
	0x01e6: 0x0107, // cacute
	0x01e8: 0x010d, // ccaron
	0x01ea: 0x0119, // eogonek
	0x01ec: 0x011b, // ecaron
	0x01ef: 0x010f, // dcaron
	0x01f0: 0x0111, // dstroke
	0x01f1: 0x0144, // nacute
	0x01f2: 0x0148, // ncaron
	0x01f5: 0x0151, // odoubleacute
	0x01f8: 0x0159, // rcaron
	0x01f9: 0x016f, // uring
	0x01fb: 0x0171, // udoubleacute
	0x01fe: 0x0163, // tcedilla
	0x01ff: 0x02d9, // abovedot

	// Special characters
	0x0ad2: 0x2026, // ellipsis ...
	0x0aae: 0x2022, // bullet
	0x0aa1: 0x00b7, // periodcentered
	0x0ae6: 0x2122, // trademark
	0x0ad0: 0x2014, // emdash
	0x0ad1: 0x2013, // endash
	0x0ad4: 0x2018, // leftsinglequotemark
	0x0ad5: 0x2019, // rightsinglequotemark
	0x0ad6: 0x201c, // leftdoublequotemark
	0x0ad7: 0x201d, // rightdoublequotemark

	// Currency
	0x20ac: 0x20ac, // EuroSign €

	// Tab, Return, etc. don't produce characters
}

// unicodeToKeysym maps Unicode codepoints to specific keysyms.
// Only needed for codepoints that have specific keysym values.
var unicodeToKeysym = map[rune]Keysym{
	0x20ac: 0x20ac, // Euro sign
}

// keysymNames maps keysyms to their names.
// This is a subset - the full table would be generated from keysymdef.h
var keysymNames = map[Keysym]string{
	KeyNoSymbol:   "NoSymbol",
	KeyBackSpace:  "BackSpace",
	KeyTab:        "Tab",
	KeyLinefeed:   "Linefeed",
	KeyClear:      "Clear",
	KeyReturn:     "Return",
	KeyPause:      "Pause",
	KeyScrollLock: "Scroll_Lock",
	KeySysReq:     "Sys_Req",
	KeyEscape:     "Escape",
	KeyDelete:     "Delete",
	KeyHome:       "Home",
	KeyLeft:       "Left",
	KeyUp:         "Up",
	KeyRight:      "Right",
	KeyDown:       "Down",
	KeyPageUp:     "Page_Up",
	KeyPageDown:   "Page_Down",
	KeyEnd:        "End",
	KeyBegin:      "Begin",
	KeySelect:     "Select",
	KeyPrint:      "Print",
	KeyExecute:    "Execute",
	KeyInsert:     "Insert",
	KeyUndo:       "Undo",
	KeyRedo:       "Redo",
	KeyMenu:       "Menu",
	KeyFind:       "Find",
	KeyCancel:     "Cancel",
	KeyHelp:       "Help",
	KeyBreak:      "Break",
	KeyNumLock:    "Num_Lock",

	// Keypad
	KeyKPSpace:    "KP_Space",
	KeyKPTab:      "KP_Tab",
	KeyKPEnter:    "KP_Enter",
	KeyKPF1:       "KP_F1",
	KeyKPF2:       "KP_F2",
	KeyKPF3:       "KP_F3",
	KeyKPF4:       "KP_F4",
	KeyKPHome:     "KP_Home",
	KeyKPLeft:     "KP_Left",
	KeyKPUp:       "KP_Up",
	KeyKPRight:    "KP_Right",
	KeyKPDown:     "KP_Down",
	KeyKPPageUp:   "KP_Page_Up",
	KeyKPPageDown: "KP_Page_Down",
	KeyKPEnd:      "KP_End",
	KeyKPBegin:    "KP_Begin",
	KeyKPInsert:   "KP_Insert",
	KeyKPDelete:   "KP_Delete",
	KeyKPEqual:    "KP_Equal",
	KeyKPMultiply: "KP_Multiply",
	KeyKPAdd:      "KP_Add",
	KeyKPSubtract: "KP_Subtract",
	KeyKPDecimal:  "KP_Decimal",
	KeyKPDivide:   "KP_Divide",
	KeyKP0:        "KP_0",
	KeyKP1:        "KP_1",
	KeyKP2:        "KP_2",
	KeyKP3:        "KP_3",
	KeyKP4:        "KP_4",
	KeyKP5:        "KP_5",
	KeyKP6:        "KP_6",
	KeyKP7:        "KP_7",
	KeyKP8:        "KP_8",
	KeyKP9:        "KP_9",

	// Function keys
	KeyF1:  "F1",
	KeyF2:  "F2",
	KeyF3:  "F3",
	KeyF4:  "F4",
	KeyF5:  "F5",
	KeyF6:  "F6",
	KeyF7:  "F7",
	KeyF8:  "F8",
	KeyF9:  "F9",
	KeyF10: "F10",
	KeyF11: "F11",
	KeyF12: "F12",

	// Modifiers
	KeyShiftL:    "Shift_L",
	KeyShiftR:    "Shift_R",
	KeyControlL:  "Control_L",
	KeyControlR:  "Control_R",
	KeyCapsLock:  "Caps_Lock",
	KeyShiftLock: "Shift_Lock",
	KeyMetaL:     "Meta_L",
	KeyMetaR:     "Meta_R",
	KeyAltL:      "Alt_L",
	KeyAltR:      "Alt_R",
	KeySuperL:    "Super_L",
	KeySuperR:    "Super_R",
	KeyHyperL:    "Hyper_L",
	KeyHyperR:    "Hyper_R",

	// ISO keys
	KeyISOLock:        "ISO_Lock",
	KeyISOLevel2Latch: "ISO_Level2_Latch",
	KeyISOLevel3Shift: "ISO_Level3_Shift",
	KeyISOLevel3Latch: "ISO_Level3_Latch",
	KeyISOLevel3Lock:  "ISO_Level3_Lock",
	KeyISOLevel5Shift: "ISO_Level5_Shift",
	KeyISOLevel5Latch: "ISO_Level5_Latch",
	KeyISOLevel5Lock:  "ISO_Level5_Lock",

	// Dead keys
	KeyDeadGrave:           "dead_grave",
	KeyDeadAcute:           "dead_acute",
	KeyDeadCircumflex:      "dead_circumflex",
	KeyDeadTilde:           "dead_tilde",
	KeyDeadMacron:          "dead_macron",
	KeyDeadBreve:           "dead_breve",
	KeyDeadAbovedot:        "dead_abovedot",
	KeyDeadDiaeresis:       "dead_diaeresis",
	KeyDeadAbovering:       "dead_abovering",
	KeyDeadDoubleacute:     "dead_doubleacute",
	KeyDeadCaron:           "dead_caron",
	KeyDeadCedilla:         "dead_cedilla",
	KeyDeadOgonek:          "dead_ogonek",
	KeyDeadIota:            "dead_iota",
	KeyDeadVoicedSound:     "dead_voiced_sound",
	KeyDeadSemivoicedSound: "dead_semivoiced_sound",
	KeyDeadBelowdot:        "dead_belowdot",

	// Multi_key (Compose)
	KeyMultiKey: "Multi_key",
}

// keysymsByName is the inverse of keysymNames.
var keysymsByName map[string]Keysym

func init() {
	keysymsByName = make(map[string]Keysym, len(keysymNames))
	for ks, name := range keysymNames {
		keysymsByName[name] = ks
	}

	// Add common lowercase letter keysyms (they map to ASCII)
	for r := 'a'; r <= 'z'; r++ {
		keysymNames[Keysym(r)] = string(r)
		keysymsByName[string(r)] = Keysym(r)
	}
	for r := 'A'; r <= 'Z'; r++ {
		keysymNames[Keysym(r)] = string(r)
		keysymsByName[string(r)] = Keysym(r)
	}
	for r := '0'; r <= '9'; r++ {
		keysymNames[Keysym(r)] = string(r)
		keysymsByName[string(r)] = Keysym(r)
	}

	// Common punctuation
	punct := map[rune]string{
		' ':  "space",
		'!':  "exclam",
		'"':  "quotedbl",
		'#':  "numbersign",
		'$':  "dollar",
		'%':  "percent",
		'&':  "ampersand",
		'\'': "apostrophe",
		'(':  "parenleft",
		')':  "parenright",
		'*':  "asterisk",
		'+':  "plus",
		',':  "comma",
		'-':  "minus",
		'.':  "period",
		'/':  "slash",
		':':  "colon",
		';':  "semicolon",
		'<':  "less",
		'=':  "equal",
		'>':  "greater",
		'?':  "question",
		'@':  "at",
		'[':  "bracketleft",
		'\\': "backslash",
		']':  "bracketright",
		'^':  "asciicircum",
		'_':  "underscore",
		'`':  "grave",
		'{':  "braceleft",
		'|':  "bar",
		'}':  "braceright",
		'~':  "asciitilde",
	}
	for r, name := range punct {
		keysymNames[Keysym(r)] = name
		keysymsByName[name] = Keysym(r)
	}
}
