package xkb

import "testing"

func TestComposeTableLocale(t *testing.T) {
	ct := TestComposeTable()
	if ct.Locale() != "en_US.UTF-8" {
		t.Errorf("Locale() = %q, want %q", ct.Locale(), "en_US.UTF-8")
	}
}

func TestComposeStateTable(t *testing.T) {
	ct := TestComposeTable()
	cs := ct.NewState(ComposeStateNoFlags)

	if cs.Table() != ct {
		t.Error("ComposeState.Table() should return the table")
	}
}

func TestComposeStateFeedDeadAcute(t *testing.T) {
	ct := TestComposeTable()
	cs := ct.NewState(ComposeStateNoFlags)

	// Initial state
	if cs.GetStatus() != ComposeNothing {
		t.Error("Initial status should be ComposeNothing")
	}

	// Feed dead_acute - starts sequence
	result := cs.Feed(KeyDeadAcute)
	if result != ComposeFeedAccepted {
		t.Error("dead_acute should be accepted")
	}
	if cs.GetStatus() != ComposeComposing {
		t.Error("Status should be ComposeComposing after dead_acute")
	}

	// Feed 'a' - completes sequence
	result = cs.Feed('a')
	if result != ComposeFeedAccepted {
		t.Error("'a' should be accepted")
	}
	if cs.GetStatus() != ComposeComposed {
		t.Error("Status should be ComposeComposed after dead_acute + a")
	}

	// Check result
	if cs.GetOneSym() != 0x00e1 { // á
		t.Errorf("GetOneSym() = %#x, want %#x", cs.GetOneSym(), 0x00e1)
	}
	if cs.GetUTF8() != "á" {
		t.Errorf("GetUTF8() = %q, want %q", cs.GetUTF8(), "á")
	}
}

func TestComposeStateFeedDeadAcuteUppercase(t *testing.T) {
	ct := TestComposeTable()
	cs := ct.NewState(ComposeStateNoFlags)

	cs.Feed(KeyDeadAcute)
	cs.Feed('A')

	if cs.GetStatus() != ComposeComposed {
		t.Error("Status should be ComposeComposed")
	}
	if cs.GetUTF8() != "Á" {
		t.Errorf("GetUTF8() = %q, want %q", cs.GetUTF8(), "Á")
	}
}

func TestComposeStateFeedDeadGrave(t *testing.T) {
	ct := TestComposeTable()
	cs := ct.NewState(ComposeStateNoFlags)

	cs.Feed(KeyDeadGrave)
	cs.Feed('a')

	if cs.GetStatus() != ComposeComposed {
		t.Error("Status should be ComposeComposed")
	}
	if cs.GetUTF8() != "à" {
		t.Errorf("GetUTF8() = %q, want %q", cs.GetUTF8(), "à")
	}
}

func TestComposeStateFeedDeadDiaeresis(t *testing.T) {
	ct := TestComposeTable()
	cs := ct.NewState(ComposeStateNoFlags)

	cs.Feed(KeyDeadDiaeresis)
	cs.Feed('u')

	if cs.GetStatus() != ComposeComposed {
		t.Error("Status should be ComposeComposed")
	}
	if cs.GetUTF8() != "ü" {
		t.Errorf("GetUTF8() = %q, want %q", cs.GetUTF8(), "ü")
	}
}

func TestComposeStateMultiKey(t *testing.T) {
	ct := TestComposeTable()
	cs := ct.NewState(ComposeStateNoFlags)

	// Multi_key + a + e = æ
	cs.Feed(KeyMultiKey)
	if cs.GetStatus() != ComposeComposing {
		t.Error("Status should be ComposeComposing after Multi_key")
	}

	cs.Feed('a')
	if cs.GetStatus() != ComposeComposing {
		t.Error("Status should still be ComposeComposing after Multi_key + a")
	}

	cs.Feed('e')
	if cs.GetStatus() != ComposeComposed {
		t.Error("Status should be ComposeComposed after Multi_key + a + e")
	}
	if cs.GetUTF8() != "æ" {
		t.Errorf("GetUTF8() = %q, want %q", cs.GetUTF8(), "æ")
	}
}

func TestComposeStateCancelled(t *testing.T) {
	ct := TestComposeTable()
	cs := ct.NewState(ComposeStateNoFlags)

	// Start sequence
	cs.Feed(KeyDeadAcute)
	if cs.GetStatus() != ComposeComposing {
		t.Error("Status should be ComposeComposing")
	}

	// Feed invalid continuation
	cs.Feed('x') // No dead_acute + x sequence
	if cs.GetStatus() != ComposeCancelled {
		t.Error("Status should be ComposeCancelled")
	}

	// GetOneSym and GetUTF8 should return empty
	if cs.GetOneSym() != KeyNoSymbol {
		t.Error("GetOneSym should return KeyNoSymbol when cancelled")
	}
	if cs.GetUTF8() != "" {
		t.Error("GetUTF8 should return empty string when cancelled")
	}
}

func TestComposeStateIgnored(t *testing.T) {
	ct := TestComposeTable()
	cs := ct.NewState(ComposeStateNoFlags)

	// Regular key that doesn't start a sequence
	result := cs.Feed('a')
	if result != ComposeFeedIgnored {
		t.Error("Regular 'a' should be ignored (not start a sequence)")
	}
	if cs.GetStatus() != ComposeNothing {
		t.Error("Status should remain ComposeNothing")
	}
}

func TestComposeStateReset(t *testing.T) {
	ct := TestComposeTable()
	cs := ct.NewState(ComposeStateNoFlags)

	// Start a sequence
	cs.Feed(KeyDeadAcute)
	if cs.GetStatus() != ComposeComposing {
		t.Error("Status should be ComposeComposing")
	}

	// Reset
	cs.Reset()
	if cs.GetStatus() != ComposeNothing {
		t.Error("Status should be ComposeNothing after Reset")
	}
}

func TestComposeStateNilTable(t *testing.T) {
	cs := &ComposeState{table: nil}

	result := cs.Feed(KeyDeadAcute)
	if result != ComposeFeedIgnored {
		t.Error("Feed with nil table should return ComposeFeedIgnored")
	}
	if cs.GetStatus() != ComposeNothing {
		t.Error("Status should be ComposeNothing with nil table")
	}
}

func TestComposeStateEmptyTable(t *testing.T) {
	ct := &ComposeTable{locale: "test", root: nil}
	cs := ct.NewState(ComposeStateNoFlags)

	result := cs.Feed(KeyDeadAcute)
	if result != ComposeFeedIgnored {
		t.Error("Feed with empty table should return ComposeFeedIgnored")
	}
}

func TestComposeStateGetResultsWhenNotComposed(t *testing.T) {
	ct := TestComposeTable()
	cs := ct.NewState(ComposeStateNoFlags)

	// Not composed yet
	if cs.GetOneSym() != KeyNoSymbol {
		t.Error("GetOneSym should return KeyNoSymbol when not composed")
	}
	if cs.GetUTF8() != "" {
		t.Error("GetUTF8 should return empty when not composed")
	}

	// Start composing
	cs.Feed(KeyDeadAcute)
	if cs.GetOneSym() != KeyNoSymbol {
		t.Error("GetOneSym should return KeyNoSymbol while composing")
	}
	if cs.GetUTF8() != "" {
		t.Error("GetUTF8 should return empty while composing")
	}
}
