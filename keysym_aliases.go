package xkb

// Keysym aliases for common names.
// The X11 headers use historical names (Prior/Next) that may be less familiar.
// These aliases provide more intuitive names.

const (
	// Navigation key aliases
	KeyPageUp   = KeyPrior // 0xff55
	KeyPageDown = KeyNext  // 0xff56

	// Keypad navigation aliases
	KeyKPPageUp   = KeyKPPrior // 0xff9a
	KeyKPPageDown = KeyKPNext  // 0xff9b
)
