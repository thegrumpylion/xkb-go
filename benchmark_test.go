package xkb

import (
	"context"
	"os"
	"testing"
)

// Benchmarks for xkb-go components

// BenchmarkLexer benchmarks the lexer performance
func BenchmarkLexer(b *testing.B) {
	// Load test data
	data, err := os.ReadFile("testdata/us_intl.xkb")
	if err != nil {
		b.Skip("No system keymap available")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(data)
		for {
			tok := lexer.NextToken()
			if tok.Type == TokenEOF || tok.Type == TokenError {
				break
			}
		}
	}
}

// BenchmarkLexerTokenize benchmarks the Tokenize method
func BenchmarkLexerTokenize(b *testing.B) {
	data, err := os.ReadFile("testdata/us_intl.xkb")
	if err != nil {
		b.Skip("No system keymap available")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(data)
		_ = lexer.Tokenize()
	}
}

// BenchmarkParser benchmarks the parser performance
func BenchmarkParser(b *testing.B) {
	data, err := os.ReadFile("testdata/us_intl.xkb")
	if err != nil {
		b.Skip("No system keymap available")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser(data)
		_, _ = p.Parse()
	}
}

// BenchmarkNewKeymapFromString benchmarks keymap creation via Context
func BenchmarkNewKeymapFromString(b *testing.B) {
	data, err := os.ReadFile("testdata/us_intl.xkb")
	if err != nil {
		b.Skip("No system keymap available")
	}

	ctx := NewContext(context.Background(), ContextNoFlags)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctx.NewKeymapFromString(data, KeymapFormatTextV1)
	}
}

// BenchmarkStateKeyGetOneSym benchmarks key translation
func BenchmarkStateKeyGetOneSym(b *testing.B) {
	km := TestKeymap()
	state := km.NewState()
	keycode := Keycode(38) // 'a' key

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = state.KeyGetOneSym(keycode)
	}
}

// BenchmarkStateKeyGetOneSym_WithModifiers benchmarks key translation with modifiers
func BenchmarkStateKeyGetOneSym_WithModifiers(b *testing.B) {
	km := TestKeymap()
	state := km.NewState()
	keycode := Keycode(38) // 'a' key
	state.UpdateMask(ModShift, 0, 0, 0, 0, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = state.KeyGetOneSym(keycode)
	}
}

// BenchmarkStateKeyGetUTF8 benchmarks UTF-8 conversion
func BenchmarkStateKeyGetUTF8(b *testing.B) {
	km := TestKeymap()
	state := km.NewState()
	keycode := Keycode(38) // 'a' key

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = state.KeyGetUTF8(keycode)
	}
}

// BenchmarkStateUpdateMask benchmarks modifier updates
func BenchmarkStateUpdateMask(b *testing.B) {
	km := TestKeymap()
	state := km.NewState()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Toggle shift on/off
		if i%2 == 0 {
			state.UpdateMask(ModShift, 0, 0, 0, 0, 0)
		} else {
			state.UpdateMask(0, 0, 0, 0, 0, 0)
		}
	}
}

// BenchmarkKeysymToUTF32 benchmarks keysym to unicode conversion
func BenchmarkKeysymToUTF32(b *testing.B) {
	ks := Keysym('a')

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = KeysymToUTF32(ks)
	}
}

// BenchmarkKeysymToUTF32_Unicode benchmarks unicode keysym conversion
func BenchmarkKeysymToUTF32_Unicode(b *testing.B) {
	ks := Keysym(0x01002603) // Unicode snowman

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = KeysymToUTF32(ks)
	}
}

// BenchmarkKeysymToUTF32_TableLookup benchmarks table lookup conversion
func BenchmarkKeysymToUTF32_TableLookup(b *testing.B) {
	ks := Keysym(0x01a1) // Aogonek (requires table lookup)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = KeysymToUTF32(ks)
	}
}

// BenchmarkKeysymToUTF8 benchmarks UTF-8 string conversion
func BenchmarkKeysymToUTF8(b *testing.B) {
	ks := Keysym('a')

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = KeysymToUTF8(ks)
	}
}

// BenchmarkKeysymGetName benchmarks keysym name lookup
func BenchmarkKeysymGetName(b *testing.B) {
	ks := KeyReturn

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = KeysymGetName(ks)
	}
}

// BenchmarkKeysymGetName_Generated benchmarks generated table lookup
func BenchmarkKeysymGetName_Generated(b *testing.B) {
	ks := Keysym(0xfe51) // dead_acute (from generated table)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = KeysymGetName(ks)
	}
}

// BenchmarkKeysymFromName benchmarks name to keysym lookup
func BenchmarkKeysymFromName(b *testing.B) {
	name := "Return"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = KeysymFromName(name, KeysymNameNoFlags)
	}
}

// BenchmarkKeysymFromName_CaseInsensitive benchmarks case-insensitive lookup
func BenchmarkKeysymFromName_CaseInsensitive(b *testing.B) {
	name := "return"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = KeysymFromName(name, KeysymNameCaseInsensitive)
	}
}

// BenchmarkComposeFeed benchmarks compose state feeding
func BenchmarkComposeFeed(b *testing.B) {
	table := TestComposeTable()
	if table == nil {
		b.Skip("No compose table available")
	}
	state := table.NewState(0)
	ks := Keysym('a')

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = state.Feed(ks)
		state.Reset()
	}
}

// BenchmarkComposeFeed_Sequence benchmarks a compose sequence
func BenchmarkComposeFeed_Sequence(b *testing.B) {
	table := TestComposeTable()
	if table == nil {
		b.Skip("No compose table available")
	}
	state := table.NewState(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = state.Feed(KeyDeadAcute)
		_ = state.Feed(Keysym('a'))
		state.Reset()
	}
}

// BenchmarkNewContext benchmarks context creation
func BenchmarkNewContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewContext(context.Background(), ContextNoFlags)
	}
}

// BenchmarkNewState benchmarks state creation
func BenchmarkNewState(b *testing.B) {
	km := TestKeymap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = km.NewState()
	}
}

// BenchmarkLexerSmall benchmarks lexer with small input
func BenchmarkLexerSmall(b *testing.B) {
	data := []byte(`xkb_keymap { xkb_keycodes "a" { minimum=8; maximum=255; }; };`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(data)
		for {
			tok := lexer.NextToken()
			if tok.Type == TokenEOF || tok.Type == TokenError {
				break
			}
		}
	}
}

// BenchmarkParserSmall benchmarks parser with small input
func BenchmarkParserSmall(b *testing.B) {
	data := []byte(`xkb_keymap {
		xkb_keycodes "a" { minimum=8; maximum=255; <K>=10; };
		xkb_types "a" { type "ONE_LEVEL" { modifiers=none; }; };
		xkb_compat "a" {};
		xkb_symbols "a" { key <K> { [a] }; };
	};`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser(data)
		_, _ = p.Parse()
	}
}

// BenchmarkKeyGetSyms benchmarks getting all keysyms for a key
func BenchmarkKeyGetSyms(b *testing.B) {
	km := TestKeymap()
	state := km.NewState()
	keycode := Keycode(38)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = state.KeyGetSyms(keycode)
	}
}

// BenchmarkSerializeMods benchmarks modifier serialization
func BenchmarkSerializeMods(b *testing.B) {
	km := TestKeymap()
	state := km.NewState()
	state.UpdateMask(ModShift|ModControl, ModMod1, ModLock, 0, 0, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = state.SerializeMods(StateModEffective)
	}
}

// BenchmarkModNameIsActive benchmarks modifier name checking
func BenchmarkModNameIsActive(b *testing.B) {
	km := TestKeymap()
	state := km.NewState()
	state.UpdateMask(ModShift, 0, 0, 0, 0, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = state.ModNameIsActive("Shift", StateModDepressed)
	}
}

// BenchmarkKeymap_KeyByName benchmarks keycode lookup by name
func BenchmarkKeymap_KeyByName(b *testing.B) {
	km := TestKeymap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = km.KeyByName("AD01")
	}
}

// BenchmarkKeymap_KeyGetName benchmarks key name lookup by keycode
func BenchmarkKeymap_KeyGetName(b *testing.B) {
	km := TestKeymap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = km.KeyGetName(24)
	}
}

// BenchmarkNewKeymapFromFile benchmarks loading keymap from file
func BenchmarkNewKeymapFromFile(b *testing.B) {
	ctx := NewContext(context.Background(), ContextNoFlags)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctx.NewKeymapFromFile("testdata/us_intl.xkb", KeymapFormatTextV1)
	}
}

// BenchmarkNewKeymapFromNames benchmarks RMLVO compilation
func BenchmarkNewKeymapFromNames(b *testing.B) {
	ctx := NewContext(context.Background(), ContextNoFlags)

	// Skip if system XKB data not available
	_, err := ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})
	if err != nil {
		b.Skip("System XKB data not available")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctx.NewKeymapFromNames(&RuleNames{Layout: "us"})
	}
}

// BenchmarkKeymap_GetAsString benchmarks keymap serialization
func BenchmarkKeymap_GetAsString(b *testing.B) {
	km := TestKeymap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = km.GetAsString(KeymapFormatTextV1)
	}
}

// BenchmarkKeymap_GetAsString_Large benchmarks serialization of a large keymap
func BenchmarkKeymap_GetAsString_Large(b *testing.B) {
	ctx := NewContext(context.Background(), ContextNoFlags)
	km, err := ctx.NewKeymapFromFile("testdata/us_intl.xkb", KeymapFormatTextV1)
	if err != nil {
		b.Skip("testdata/us_intl.xkb not available")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = km.GetAsString(KeymapFormatTextV1)
	}
}
