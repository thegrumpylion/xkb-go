package xkb

import (
	"os"
	"path/filepath"
	"testing"
)

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

// Tests for compose file parser

func TestComposeParserBasic(t *testing.T) {
	composeData := `# Test compose file
<dead_acute> <a>	: "á"
<dead_acute> <e>	: "é"
<dead_grave> <a>	: "à"
`
	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "Compose")
	if err := os.WriteFile(composePath, []byte(composeData), 0644); err != nil {
		t.Fatalf("Failed to write test compose file: %v", err)
	}

	ctx := NewContext(ContextNoDefaultIncludes)
	table, err := ctx.NewComposeTableFromFile(composePath, "en_US.UTF-8", ComposeCompileNoFlags)
	if err != nil {
		t.Fatalf("NewComposeTableFromFile failed: %v", err)
	}

	if table == nil {
		t.Fatal("Compose table is nil")
	}

	if table.Locale() != "en_US.UTF-8" {
		t.Errorf("Expected locale en_US.UTF-8, got %s", table.Locale())
	}

	// Test that sequences work
	state := table.NewState(ComposeStateNoFlags)
	state.Feed(KeyDeadAcute)
	state.Feed('a')
	if state.GetStatus() != ComposeComposed {
		t.Errorf("Expected ComposeComposed, got %v", state.GetStatus())
	}
	if state.GetUTF8() != "á" {
		t.Errorf("Expected 'á', got %q", state.GetUTF8())
	}
}

func TestComposeParserWithKeysymName(t *testing.T) {
	composeData := `
<dead_acute> <a>	: "á"	aacute
<dead_acute> <A>	: "Á"	Aacute
`
	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "Compose")
	if err := os.WriteFile(composePath, []byte(composeData), 0644); err != nil {
		t.Fatalf("Failed to write test compose file: %v", err)
	}

	ctx := NewContext(ContextNoDefaultIncludes)
	table, err := ctx.NewComposeTableFromFile(composePath, "en_US.UTF-8", ComposeCompileNoFlags)
	if err != nil {
		t.Fatalf("NewComposeTableFromFile failed: %v", err)
	}

	state := table.NewState(ComposeStateNoFlags)
	state.Feed(KeyDeadAcute)
	state.Feed('a')
	if state.GetUTF8() != "á" {
		t.Errorf("Expected 'á', got %q", state.GetUTF8())
	}
}

func TestComposeParserInclude(t *testing.T) {
	tmpDir := t.TempDir()

	// Create base compose file
	baseData := `
<dead_acute> <a>	: "á"
<dead_acute> <e>	: "é"
`
	basePath := filepath.Join(tmpDir, "base.compose")
	if err := os.WriteFile(basePath, []byte(baseData), 0644); err != nil {
		t.Fatalf("Failed to write base compose file: %v", err)
	}

	// Create main compose file that includes base
	mainData := `include "` + basePath + `"
<dead_grave> <a>	: "à"
`
	mainPath := filepath.Join(tmpDir, "main.compose")
	if err := os.WriteFile(mainPath, []byte(mainData), 0644); err != nil {
		t.Fatalf("Failed to write main compose file: %v", err)
	}

	ctx := NewContext(ContextNoDefaultIncludes)
	table, err := ctx.NewComposeTableFromFile(mainPath, "en_US.UTF-8", ComposeCompileNoFlags)
	if err != nil {
		t.Fatalf("NewComposeTableFromFile failed: %v", err)
	}

	state := table.NewState(ComposeStateNoFlags)

	// Test included sequence: dead_acute + a = á
	state.Feed(KeyDeadAcute)
	state.Feed('a')
	if state.GetStatus() != ComposeComposed || state.GetUTF8() != "á" {
		t.Errorf("Expected 'á' from included file, got %q (status: %v)", state.GetUTF8(), state.GetStatus())
	}

	state.Reset()

	// Test sequence from main file: dead_grave + a = à
	state.Feed(KeyDeadGrave)
	state.Feed('a')
	if state.GetStatus() != ComposeComposed || state.GetUTF8() != "à" {
		t.Errorf("Expected 'à' from main file, got %q (status: %v)", state.GetUTF8(), state.GetStatus())
	}
}

func TestComposeParserUnicodeKeysym(t *testing.T) {
	composeData := `
<dead_acute> <a>	: "á"	U00E1
<Multi_key> <e> <u> <r>	: "€"	U20AC
`
	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "Compose")
	if err := os.WriteFile(composePath, []byte(composeData), 0644); err != nil {
		t.Fatalf("Failed to write test compose file: %v", err)
	}

	ctx := NewContext(ContextNoDefaultIncludes)
	table, err := ctx.NewComposeTableFromFile(composePath, "en_US.UTF-8", ComposeCompileNoFlags)
	if err != nil {
		t.Fatalf("NewComposeTableFromFile failed: %v", err)
	}

	state := table.NewState(ComposeStateNoFlags)

	// Test Unicode keysym: dead_acute + a = á (U+00E1)
	state.Feed(KeyDeadAcute)
	state.Feed('a')
	if state.GetStatus() != ComposeComposed {
		t.Errorf("Expected ComposeComposed, got %v", state.GetStatus())
	}
	if state.GetUTF8() != "á" {
		t.Errorf("Expected 'á', got %q", state.GetUTF8())
	}
}

func TestComposeParserMultiKey(t *testing.T) {
	composeData := `
<Multi_key> <c> <o>		: "©"	copyright
<Multi_key> <o> <r>		: "®"	registered
`
	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "Compose")
	if err := os.WriteFile(composePath, []byte(composeData), 0644); err != nil {
		t.Fatalf("Failed to write test compose file: %v", err)
	}

	ctx := NewContext(ContextNoDefaultIncludes)
	table, err := ctx.NewComposeTableFromFile(composePath, "en_US.UTF-8", ComposeCompileNoFlags)
	if err != nil {
		t.Fatalf("NewComposeTableFromFile failed: %v", err)
	}

	state := table.NewState(ComposeStateNoFlags)

	// Test: Multi_key + c + o = ©
	state.Feed(KeyMultiKey)
	state.Feed('c')
	state.Feed('o')
	if state.GetStatus() != ComposeComposed {
		t.Errorf("Expected ComposeComposed, got %v", state.GetStatus())
	}
	if state.GetUTF8() != "©" {
		t.Errorf("Expected '©', got %q", state.GetUTF8())
	}
}

func TestComposeParserEscapeSequences(t *testing.T) {
	composeData := `
<Multi_key> <n> <l>		: "\n"
<Multi_key> <t> <a> <b>	: "\t"
<Multi_key> <b> <s>		: "\\"
<Multi_key> <q> <t>		: "\""
`
	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "Compose")
	if err := os.WriteFile(composePath, []byte(composeData), 0644); err != nil {
		t.Fatalf("Failed to write test compose file: %v", err)
	}

	ctx := NewContext(ContextNoDefaultIncludes)
	table, err := ctx.NewComposeTableFromFile(composePath, "en_US.UTF-8", ComposeCompileNoFlags)
	if err != nil {
		t.Fatalf("NewComposeTableFromFile failed: %v", err)
	}

	state := table.NewState(ComposeStateNoFlags)

	// Test newline escape
	state.Feed(KeyMultiKey)
	state.Feed('n')
	state.Feed('l')
	if state.GetUTF8() != "\n" {
		t.Errorf("Expected newline, got %q", state.GetUTF8())
	}

	state.Reset()

	// Test backslash escape
	state.Feed(KeyMultiKey)
	state.Feed('b')
	state.Feed('s')
	if state.GetUTF8() != "\\" {
		t.Errorf("Expected backslash, got %q", state.GetUTF8())
	}
}

func TestComposeTableNotFound(t *testing.T) {
	ctx := NewContext(ContextNoDefaultIncludes)

	_, err := ctx.NewComposeTableFromFile("/nonexistent/path/Compose", "en_US.UTF-8", ComposeCompileNoFlags)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestComposeParserOverride(t *testing.T) {
	// Later definitions should override earlier ones
	composeData := `
<dead_acute> <a>	: "first"
<dead_acute> <a>	: "second"
`
	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "Compose")
	if err := os.WriteFile(composePath, []byte(composeData), 0644); err != nil {
		t.Fatalf("Failed to write test compose file: %v", err)
	}

	ctx := NewContext(ContextNoDefaultIncludes)
	table, err := ctx.NewComposeTableFromFile(composePath, "en_US.UTF-8", ComposeCompileNoFlags)
	if err != nil {
		t.Fatalf("NewComposeTableFromFile failed: %v", err)
	}

	state := table.NewState(ComposeStateNoFlags)
	state.Feed(KeyDeadAcute)
	state.Feed('a')
	if state.GetUTF8() != "second" {
		t.Errorf("Expected 'second' (override), got %q", state.GetUTF8())
	}
}

func TestComposeWithSystemFile(t *testing.T) {
	// Test loading the real system compose file
	composePath := "/usr/share/X11/locale/en_US.UTF-8/Compose"
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		t.Skipf("System compose file not found: %s", composePath)
	}

	ctx := NewContext(ContextNoFlags)
	table, err := ctx.NewComposeTableFromFile(composePath, "en_US.UTF-8", ComposeCompileNoFlags)
	if err != nil {
		t.Fatalf("Failed to load system compose file: %v", err)
	}

	if table == nil {
		t.Fatal("Compose table is nil")
	}

	state := table.NewState(ComposeStateNoFlags)

	// Test some common compose sequences from the system file
	tests := []struct {
		sequence []Keysym
		expected string
		desc     string
	}{
		{[]Keysym{KeyDeadAcute, 'a'}, "á", "dead_acute + a = á"},
		{[]Keysym{KeyDeadAcute, 'e'}, "é", "dead_acute + e = é"},
		{[]Keysym{KeyDeadGrave, 'a'}, "à", "dead_grave + a = à"},
		{[]Keysym{KeyDeadDiaeresis, 'u'}, "ü", "dead_diaeresis + u = ü"},
		{[]Keysym{KeyDeadTilde, 'n'}, "ñ", "dead_tilde + n = ñ"},
		{[]Keysym{KeyMultiKey, 'o', 'o'}, "°", "Multi_key + o + o = ° (degree)"},
		{[]Keysym{KeyMultiKey, 'o', 'c'}, "©", "Multi_key + o + c = © (copyright)"},
	}

	for _, tt := range tests {
		state.Reset()
		for _, ks := range tt.sequence {
			state.Feed(ks)
		}
		if state.GetStatus() != ComposeComposed {
			t.Errorf("%s: expected ComposeComposed, got %v", tt.desc, state.GetStatus())
			continue
		}
		if state.GetUTF8() != tt.expected {
			t.Errorf("%s: expected %q, got %q", tt.desc, tt.expected, state.GetUTF8())
		}
	}

	t.Logf("Successfully loaded and tested system compose file")
}

func TestNewComposeTableFromLocale(t *testing.T) {
	ctx := NewContext(ContextNoFlags)

	// Test loading by locale
	table, err := ctx.NewComposeTableFromLocale("en_US.UTF-8", ComposeCompileNoFlags)
	if err != nil {
		t.Skipf("Could not load compose table for en_US.UTF-8: %v", err)
	}

	if table == nil {
		t.Fatal("Compose table is nil")
	}

	state := table.NewState(ComposeStateNoFlags)

	// Test a simple sequence
	state.Feed(KeyDeadAcute)
	state.Feed('a')
	if state.GetStatus() != ComposeComposed || state.GetUTF8() != "á" {
		t.Errorf("Expected 'á', got %q (status: %v)", state.GetUTF8(), state.GetStatus())
	}

	t.Logf("Successfully loaded compose table for locale: %s", table.Locale())
}

func TestComposeEdgeCases(t *testing.T) {
	table := TestComposeTable()
	if table == nil {
		t.Skip("No test compose table available")
	}
	state := table.NewState(0)

	t.Run("empty sequence", func(t *testing.T) {
		state.Reset()
		if state.GetStatus() != ComposeNothing {
			t.Error("Expected ComposeNothing after reset")
		}
	})

	t.Run("very long sequence", func(t *testing.T) {
		state.Reset()
		for i := 0; i < 100; i++ {
			_ = state.Feed(Keysym('a' + (i % 26)))
		}
	})

	t.Run("special keysyms", func(t *testing.T) {
		state.Reset()
		specialKeysyms := []Keysym{
			KeyNoSymbol,
			KeyReturn,
			KeyEscape,
			KeyBackSpace,
			0xFFFFFFFF,
		}

		for _, ks := range specialKeysyms {
			_ = state.Feed(ks)
		}
	})

	t.Run("concurrent feed and query", func(t *testing.T) {
		state.Reset()
		for i := 0; i < 100; i++ {
			_ = state.Feed(Keysym('a'))
			_ = state.GetStatus()
			_ = state.GetOneSym()
			_ = state.GetUTF8()
		}
	})
}
