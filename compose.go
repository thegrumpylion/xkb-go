package xkb

// ComposeTable holds compose sequence definitions for dead key handling.
//
// Create a ComposeTable with [Context.NewComposeTableFromLocale] or
// [Context.NewComposeTableFromFile]. It is immutable after creation
// and safe to share across goroutines.
//
// Use [ComposeTable.NewState] to create a [ComposeState] for processing input.
type ComposeTable struct {
	locale string
	root   *composeNode
}

// composeNode is a node in the compose sequence trie.
type composeNode struct {
	children map[Keysym]*composeNode
	result   *composeResult // non-nil if this is a terminal node
}

// composeResult holds the result of a completed compose sequence.
type composeResult struct {
	keysym Keysym
	utf8   string
}

// Locale returns the locale this compose table was created for.
func (ct *ComposeTable) Locale() string {
	return ct.locale
}

// NewState creates a new [ComposeState] for this table.
//
// Each input context should have its own ComposeState.
func (ct *ComposeTable) NewState(flags ComposeStateFlags) *ComposeState {
	return &ComposeState{
		table:  ct,
		status: ComposeNothing,
	}
}

// ComposeState tracks the state of an ongoing compose sequence.
//
// Feed keysyms with [ComposeState.Feed], then check [ComposeState.GetStatus].
// If status is [ComposeComposed], retrieve the result with [ComposeState.GetUTF8]
// or [ComposeState.GetOneSym].
//
// ComposeState is NOT safe for concurrent use.
type ComposeState struct {
	table   *ComposeTable
	current *composeNode
	status  ComposeStatus
	result  *composeResult
}

// Table returns the [ComposeTable] this state is for.
func (cs *ComposeState) Table() *ComposeTable {
	return cs.table
}

// Feed processes a [Keysym] through the compose state machine.
//
// Returns [ComposeFeedAccepted] if the keysym was consumed, or
// [ComposeFeedIgnored] if it doesn't start a compose sequence.
//
// After calling Feed, check [ComposeState.GetStatus] to determine the state:
//   - [ComposeNothing]: No compose sequence in progress
//   - [ComposeComposing]: Sequence in progress, waiting for more input
//   - [ComposeComposed]: Sequence completed, call [ComposeState.GetUTF8] for result
//   - [ComposeCancelled]: Sequence was cancelled (invalid keysym)
func (cs *ComposeState) Feed(keysym Keysym) ComposeFeedResult {
	if cs.table == nil || cs.table.root == nil {
		cs.status = ComposeNothing
		return ComposeFeedIgnored
	}

	// Start from root if not currently composing
	if cs.current == nil {
		cs.current = cs.table.root
	}

	// Look for this keysym in children
	if next, ok := cs.current.children[keysym]; ok {
		cs.current = next
		if next.result != nil {
			// Sequence complete
			cs.status = ComposeComposed
			cs.result = next.result
		} else {
			// Sequence continues
			cs.status = ComposeComposing
			cs.result = nil
		}
		return ComposeFeedAccepted
	}

	// Keysym not found in current state
	if cs.status == ComposeComposing {
		// Was in a sequence, now cancelled
		cs.status = ComposeCancelled
		cs.current = nil
		cs.result = nil
		return ComposeFeedAccepted
	}

	// Not in a sequence, check if this starts one
	if next, ok := cs.table.root.children[keysym]; ok {
		cs.current = next
		if next.result != nil {
			cs.status = ComposeComposed
			cs.result = next.result
		} else {
			cs.status = ComposeComposing
			cs.result = nil
		}
		return ComposeFeedAccepted
	}

	// Doesn't start a sequence
	cs.status = ComposeNothing
	cs.current = nil
	cs.result = nil
	return ComposeFeedIgnored
}

// GetStatus returns the current [ComposeStatus].
func (cs *ComposeState) GetStatus() ComposeStatus {
	return cs.status
}

// GetOneSym returns the composed [Keysym] after a successful sequence.
//
// Returns [KeyNoSymbol] if status is not [ComposeComposed].
func (cs *ComposeState) GetOneSym() Keysym {
	if cs.status != ComposeComposed || cs.result == nil {
		return KeyNoSymbol
	}
	return cs.result.keysym
}

// GetUTF8 returns the composed UTF-8 string after a successful sequence.
//
// Returns empty string if status is not [ComposeComposed].
func (cs *ComposeState) GetUTF8() string {
	if cs.status != ComposeComposed || cs.result == nil {
		return ""
	}
	return cs.result.utf8
}

// Reset resets the compose state to the initial state.
//
// Call this to cancel an in-progress sequence (e.g., when Escape is pressed).
func (cs *ComposeState) Reset() {
	cs.current = nil
	cs.status = ComposeNothing
	cs.result = nil
}
