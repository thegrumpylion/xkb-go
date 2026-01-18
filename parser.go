package xkb

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser parses XKB text format into a Keymap.
type Parser struct {
	lexer   *Lexer
	current Token
	prev    Token
	errors  []error
}

// NewParser creates a new parser for the given input.
func NewParser(input []byte) *Parser {
	p := &Parser{
		lexer: NewLexer(input),
	}
	p.advance() // Prime the parser with the first token
	return p
}

// advance moves to the next token.
func (p *Parser) advance() Token {
	p.prev = p.current
	p.current = p.lexer.NextToken()
	return p.current
}

// check returns true if the current token is of the given type.
func (p *Parser) check(t TokenType) bool {
	return p.current.Type == t
}

// match consumes the current token if it matches any of the given types.
func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

// expect consumes the current token if it matches, otherwise returns an error.
func (p *Parser) expect(t TokenType) error {
	if p.check(t) {
		p.advance()
		return nil
	}
	return p.errorf("expected %s, got %s", t, p.current.Type)
}

// expectIdent consumes an identifier and returns its value.
func (p *Parser) expectIdent() (string, error) {
	if !p.check(TokenIdent) {
		return "", p.errorf("expected identifier, got %s", p.current.Type)
	}
	val := p.current.Value
	p.advance()
	return val, nil
}

// expectString consumes a string and returns its value.
func (p *Parser) expectString() (string, error) {
	if !p.check(TokenString) {
		return "", p.errorf("expected string, got %s", p.current.Type)
	}
	val := p.current.Value
	p.advance()
	return val, nil
}

// expectNumber consumes a number and returns its value.
func (p *Parser) expectNumber() (int, error) {
	if !p.check(TokenNumber) {
		return 0, p.errorf("expected number, got %s", p.current.Type)
	}
	val, err := p.parseNumber(p.current.Value)
	if err != nil {
		return 0, p.errorf("invalid number: %s", p.current.Value)
	}
	p.advance()
	return val, nil
}

// expectKeycode consumes a keycode and returns its name.
func (p *Parser) expectKeycode() (string, error) {
	if !p.check(TokenKeycode) {
		return "", p.errorf("expected keycode, got %s", p.current.Type)
	}
	val := p.current.Value
	p.advance()
	return val, nil
}

// parseNumber parses a number string (decimal, hex, or octal).
func (p *Parser) parseNumber(s string) (int, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		val, err := strconv.ParseInt(s[2:], 16, 64)
		return int(val), err
	}
	if len(s) > 1 && s[0] == '0' {
		val, err := strconv.ParseInt(s, 8, 64)
		return int(val), err
	}
	val, err := strconv.ParseInt(s, 10, 64)
	return int(val), err
}

// errorf creates a parse error with location information.
func (p *Parser) errorf(format string, args ...any) error {
	return &Error{
		Op:   "parse",
		Line: p.current.Line,
		Col:  p.current.Col,
		Err:  fmt.Errorf(format, args...),
	}
}

// addError adds an error to the list.
func (p *Parser) addError(err error) {
	p.errors = append(p.errors, err)
}

// skipToSemicolonOrBrace skips tokens until a semicolon or closing brace is found.
// Used for error recovery.
func (p *Parser) skipToSemicolonOrBrace() {
	for {
		switch p.current.Type {
		case TokenEOF, TokenSemicolon, TokenRBrace:
			return
		}
		p.advance()
	}
}

// Parse parses a complete XKB keymap.
func (p *Parser) Parse() (*Keymap, error) {
	keymap := &Keymap{
		keycodeNames:   make(map[Keycode]string),
		keycodesByName: make(map[string]Keycode),
		types:          make(map[string]*KeyType),
		keys:           make(map[Keycode]*Key),
		virtualMods:    make(map[string]ModMask),
		leds:           make(map[string]*LED),
		groupNames:     make([]string, 4),
		numGroups:      1,
		minKeycode:     8,
		maxKeycode:     255,
	}

	// Expect: xkb_keymap {
	if ident, err := p.expectIdent(); err != nil {
		return nil, err
	} else if ident != "xkb_keymap" {
		return nil, p.errorf("expected xkb_keymap, got %s", ident)
	}

	if err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	// Parse sections until }
	for !p.check(TokenRBrace) && !p.check(TokenEOF) {
		if err := p.parseSection(keymap); err != nil {
			p.addError(err)
			p.skipToSemicolonOrBrace()
			if p.check(TokenSemicolon) {
				p.advance()
			}
		}
	}

	if err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	// Optional trailing semicolon
	p.match(TokenSemicolon)

	if len(p.errors) > 0 {
		return keymap, p.errors[0] // Return first error
	}

	return keymap, nil
}

// parseSection parses a single xkb section (keycodes, types, compat, symbols, geometry).
func (p *Parser) parseSection(keymap *Keymap) error {
	sectionName, err := p.expectIdent()
	if err != nil {
		return err
	}

	switch sectionName {
	case "xkb_keycodes":
		return p.parseKeycodes(keymap)
	case "xkb_types":
		return p.parseTypes(keymap)
	case "xkb_compat", "xkb_compatibility":
		return p.parseCompat(keymap)
	case "xkb_symbols":
		return p.parseSymbols(keymap)
	case "xkb_geometry":
		return p.skipSection() // Ignored
	default:
		return p.errorf("unknown section: %s", sectionName)
	}
}

// skipSection skips an entire section (used for xkb_geometry).
func (p *Parser) skipSection() error {
	// Skip optional name string
	p.match(TokenString)

	// Expect {
	if err := p.expect(TokenLBrace); err != nil {
		return err
	}

	// Skip until matching }
	depth := 1
	for depth > 0 && !p.check(TokenEOF) {
		if p.check(TokenLBrace) {
			depth++
		} else if p.check(TokenRBrace) {
			depth--
		}
		p.advance()
	}

	// Expect ;
	return p.expect(TokenSemicolon)
}

// parseKeycodes parses the xkb_keycodes section.
func (p *Parser) parseKeycodes(keymap *Keymap) error {
	// Optional section name
	p.match(TokenString)

	// Expect {
	if err := p.expect(TokenLBrace); err != nil {
		return err
	}

	// Parse statements until }
	for !p.check(TokenRBrace) && !p.check(TokenEOF) {
		if err := p.parseKeycodesStatement(keymap); err != nil {
			p.addError(err)
			p.skipToSemicolonOrBrace()
		}
		if p.check(TokenSemicolon) {
			p.advance()
		}
	}

	if err := p.expect(TokenRBrace); err != nil {
		return err
	}

	return p.expect(TokenSemicolon)
}

// parseKeycodesStatement parses a single statement in xkb_keycodes.
func (p *Parser) parseKeycodesStatement(keymap *Keymap) error {
	// Check for keycode definition: <NAME> = number
	if p.check(TokenKeycode) {
		return p.parseKeycodeDefinition(keymap)
	}

	// Must be an identifier (minimum, maximum, alias, indicator)
	ident, err := p.expectIdent()
	if err != nil {
		return err
	}

	switch ident {
	case "minimum":
		if err := p.expect(TokenEquals); err != nil {
			return err
		}
		min, err := p.expectNumber()
		if err != nil {
			return err
		}
		keymap.minKeycode = Keycode(min)
		return nil

	case "maximum":
		if err := p.expect(TokenEquals); err != nil {
			return err
		}
		max, err := p.expectNumber()
		if err != nil {
			return err
		}
		keymap.maxKeycode = Keycode(max)
		return nil

	case "alias":
		return p.parseKeycodeAlias(keymap)

	case "indicator":
		return p.parseIndicatorName(keymap)

	case "virtual":
		// virtual indicators - skip for now
		p.skipToSemicolonOrBrace()
		return nil

	default:
		return p.errorf("unexpected identifier in xkb_keycodes: %s", ident)
	}
}

// parseKeycodeDefinition parses: <NAME> = number
func (p *Parser) parseKeycodeDefinition(keymap *Keymap) error {
	name, err := p.expectKeycode()
	if err != nil {
		return err
	}

	if err := p.expect(TokenEquals); err != nil {
		return err
	}

	code, err := p.expectNumber()
	if err != nil {
		return err
	}

	keycode := Keycode(code)
	keymap.keycodeNames[keycode] = name
	keymap.keycodesByName[name] = keycode

	return nil
}

// parseKeycodeAlias parses: alias <NAME> = <TARGET>
func (p *Parser) parseKeycodeAlias(keymap *Keymap) error {
	aliasName, err := p.expectKeycode()
	if err != nil {
		return err
	}

	if err := p.expect(TokenEquals); err != nil {
		return err
	}

	targetName, err := p.expectKeycode()
	if err != nil {
		return err
	}

	// Look up the target keycode
	if targetCode, ok := keymap.keycodesByName[targetName]; ok {
		keymap.keycodesByName[aliasName] = targetCode
	}
	// If target not found, the alias still gets recorded (may be resolved later)

	return nil
}

// parseIndicatorName parses: indicator number = "Name"
func (p *Parser) parseIndicatorName(keymap *Keymap) error {
	index, err := p.expectNumber()
	if err != nil {
		return err
	}

	if err := p.expect(TokenEquals); err != nil {
		return err
	}

	name, err := p.expectString()
	if err != nil {
		return err
	}

	keymap.leds[name] = &LED{
		name:  name,
		index: index,
	}

	return nil
}

// parseTypes parses the xkb_types section.
func (p *Parser) parseTypes(keymap *Keymap) error {
	// Optional section name
	p.match(TokenString)

	// Expect {
	if err := p.expect(TokenLBrace); err != nil {
		return err
	}

	// Parse statements until }
	for !p.check(TokenRBrace) && !p.check(TokenEOF) {
		if err := p.parseTypesStatement(keymap); err != nil {
			p.addError(err)
			p.skipToSemicolonOrBrace()
		}
		if p.check(TokenSemicolon) {
			p.advance()
		}
	}

	if err := p.expect(TokenRBrace); err != nil {
		return err
	}

	return p.expect(TokenSemicolon)
}

// parseTypesStatement parses a single statement in xkb_types.
func (p *Parser) parseTypesStatement(keymap *Keymap) error {
	ident, err := p.expectIdent()
	if err != nil {
		return err
	}

	switch ident {
	case "virtual_modifiers":
		return p.parseVirtualModifiers(keymap)
	case "type":
		return p.parseTypeDefinition(keymap)
	default:
		return p.errorf("unexpected identifier in xkb_types: %s", ident)
	}
}

// parseVirtualModifiers parses: virtual_modifiers Name1, Name2, Name3=0x4000, ...;
func (p *Parser) parseVirtualModifiers(keymap *Keymap) error {
	for {
		name, err := p.expectIdent()
		if err != nil {
			return err
		}

		// Check for optional =value assignment (e.g., Hyper=0x4000)
		var mask ModMask
		if p.match(TokenEquals) {
			num, err := p.expectNumber()
			if err != nil {
				return err
			}
			mask = ModMask(num)
		}

		// Assign a virtual modifier
		if _, exists := keymap.virtualMods[name]; !exists {
			keymap.virtualMods[name] = mask
		}

		if !p.match(TokenComma) {
			break
		}
	}
	return nil
}

// parseTypeDefinition parses: type "NAME" { ... }
func (p *Parser) parseTypeDefinition(keymap *Keymap) error {
	name, err := p.expectString()
	if err != nil {
		return err
	}

	if err := p.expect(TokenLBrace); err != nil {
		return err
	}

	keyType := &KeyType{
		name:      name,
		numLevels: 1,
	}

	// Parse type body
	for !p.check(TokenRBrace) && !p.check(TokenEOF) {
		if err := p.parseTypeStatement(keyType, keymap); err != nil {
			p.addError(err)
			p.skipToSemicolonOrBrace()
		}
		if p.check(TokenSemicolon) {
			p.advance()
		}
	}

	if err := p.expect(TokenRBrace); err != nil {
		return err
	}

	keymap.types[name] = keyType
	keymap.typesList = append(keymap.typesList, keyType)
	return nil
}

// parseTypeStatement parses a single statement in a type definition.
func (p *Parser) parseTypeStatement(keyType *KeyType, keymap *Keymap) error {
	ident, err := p.expectIdent()
	if err != nil {
		return err
	}

	switch ident {
	case "modifiers":
		if err := p.expect(TokenEquals); err != nil {
			return err
		}
		mods, err := p.parseModifierMask(keymap)
		if err != nil {
			return err
		}
		keyType.mods = mods
		return nil

	case "map":
		return p.parseTypeMapEntry(keyType, keymap)

	case "level_name":
		return p.parseTypeLevelName(keyType)

	case "preserve":
		// Skip preserve statements for now
		p.skipToSemicolonOrBrace()
		return nil

	default:
		return p.errorf("unexpected identifier in type: %s", ident)
	}
}

// parseModifierMask parses a modifier expression like: Shift + Lock or none
func (p *Parser) parseModifierMask(keymap *Keymap) (ModMask, error) {
	var mask ModMask

	for {
		if p.check(TokenIdent) {
			name := p.current.Value
			p.advance()

			if name == "none" || name == "None" {
				return 0, nil
			}

			modMask := p.modifierNameToMask(name, keymap)
			mask |= modMask
		} else {
			return 0, p.errorf("expected modifier name")
		}

		if !p.match(TokenPlus) {
			break
		}
	}

	return mask, nil
}

// modifierNameToMask converts a modifier name to its mask.
func (p *Parser) modifierNameToMask(name string, keymap *Keymap) ModMask {
	switch name {
	case "Shift":
		return ModShift
	case "Lock":
		return ModLock
	case "Control":
		return ModControl
	case "Mod1":
		return ModMod1
	case "Mod2":
		return ModMod2
	case "Mod3":
		return ModMod3
	case "Mod4":
		return ModMod4
	case "Mod5":
		return ModMod5
	default:
		// Check virtual modifiers
		if mask, ok := keymap.virtualMods[name]; ok {
			return mask
		}
		return 0
	}
}

// parseTypeMapEntry parses: map[MODS] = LevelN or map[MODS] = N
func (p *Parser) parseTypeMapEntry(keyType *KeyType, keymap *Keymap) error {
	if err := p.expect(TokenLBracket); err != nil {
		return err
	}

	mods, err := p.parseModifierMask(keymap)
	if err != nil {
		return err
	}

	if err := p.expect(TokenRBracket); err != nil {
		return err
	}

	if err := p.expect(TokenEquals); err != nil {
		return err
	}

	// Handle both "Level2" and numeric "2"
	var level Level
	if p.check(TokenNumber) {
		num, err := p.expectNumber()
		if err != nil {
			return err
		}
		level = Level(num - 1) // Convert 1-based to 0-based
	} else {
		levelName, err := p.expectIdent()
		if err != nil {
			return err
		}
		level, err = p.parseLevelName(levelName)
		if err != nil {
			return err
		}
	}

	keyType.entries = append(keyType.entries, KeyTypeEntry{
		mods:  mods,
		level: level,
	})

	if int(level)+1 > keyType.numLevels {
		keyType.numLevels = int(level) + 1
	}

	return nil
}

// parseLevelName parses a level name like "Level1" or "Level2".
func (p *Parser) parseLevelName(name string) (Level, error) {
	if !strings.HasPrefix(name, "Level") {
		return 0, fmt.Errorf("invalid level name: %s", name)
	}
	numStr := name[5:]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid level number: %s", name)
	}
	return Level(num - 1), nil // Convert 1-based to 0-based
}

// parseTypeLevelName parses: level_name[LevelN] = "Name" or level_name[N] = "Name"
func (p *Parser) parseTypeLevelName(keyType *KeyType) error {
	if err := p.expect(TokenLBracket); err != nil {
		return err
	}

	// Handle both "Level1" and numeric "1" indices
	var level Level
	if p.check(TokenNumber) {
		num, err := p.expectNumber()
		if err != nil {
			return err
		}
		level = Level(num - 1) // Convert 1-based to 0-based
	} else {
		levelName, err := p.expectIdent()
		if err != nil {
			return err
		}
		var err2 error
		level, err2 = p.parseLevelName(levelName)
		if err2 != nil {
			return err2
		}
	}

	if err := p.expect(TokenRBracket); err != nil {
		return err
	}

	if err := p.expect(TokenEquals); err != nil {
		return err
	}

	_, err := p.expectString()
	if err != nil {
		return err
	}

	// Level names are informational, we don't store them
	// But we do update numLevels
	if int(level)+1 > keyType.numLevels {
		keyType.numLevels = int(level) + 1
	}

	return nil
}

// parseCompat parses the xkb_compat section.
func (p *Parser) parseCompat(keymap *Keymap) error {
	// Optional section name
	p.match(TokenString)

	// Expect {
	if err := p.expect(TokenLBrace); err != nil {
		return err
	}

	// Parse statements until }
	for !p.check(TokenRBrace) && !p.check(TokenEOF) {
		if err := p.parseCompatStatement(keymap); err != nil {
			p.addError(err)
			p.skipToSemicolonOrBrace()
		}
		if p.check(TokenSemicolon) {
			p.advance()
		}
	}

	if err := p.expect(TokenRBrace); err != nil {
		return err
	}

	return p.expect(TokenSemicolon)
}

// parseCompatStatement parses a single statement in xkb_compat.
func (p *Parser) parseCompatStatement(keymap *Keymap) error {
	ident, err := p.expectIdent()
	if err != nil {
		return err
	}

	switch ident {
	case "virtual_modifiers":
		return p.parseVirtualModifiers(keymap)
	case "interpret":
		return p.parseInterpret(keymap)
	case "indicator":
		return p.parseCompatIndicator(keymap)
	case "group":
		// Skip group statements
		p.skipToSemicolonOrBrace()
		return nil
	default:
		return p.errorf("unexpected identifier in xkb_compat: %s", ident)
	}
}

// parseInterpret parses: interpret KeysymName { ... }
func (p *Parser) parseInterpret(keymap *Keymap) error {
	// Parse keysym name or expression
	// Can be: interpret Shift_L { ... } or interpret Any+Exactly(Shift) { ... }
	// For now, skip to the body and parse it minimally
	for !p.check(TokenLBrace) && !p.check(TokenEOF) && !p.check(TokenSemicolon) {
		p.advance()
	}

	if !p.check(TokenLBrace) {
		return nil // No body
	}

	p.advance() // consume {

	// Parse interpret body - looking for action = ...
	for !p.check(TokenRBrace) && !p.check(TokenEOF) {
		p.skipToSemicolonOrBrace()
		if p.check(TokenSemicolon) {
			p.advance()
		}
	}

	if err := p.expect(TokenRBrace); err != nil {
		return err
	}

	return nil
}

// parseCompatIndicator parses: indicator "Name" { ... }
func (p *Parser) parseCompatIndicator(keymap *Keymap) error {
	name, err := p.expectString()
	if err != nil {
		return err
	}

	if err := p.expect(TokenLBrace); err != nil {
		return err
	}

	led, exists := keymap.leds[name]
	if !exists {
		led = &LED{name: name}
		keymap.leds[name] = led
	}

	// Parse indicator body
	for !p.check(TokenRBrace) && !p.check(TokenEOF) {
		if err := p.parseCompatIndicatorStatement(led, keymap); err != nil {
			p.addError(err)
			p.skipToSemicolonOrBrace()
		}
		if p.check(TokenSemicolon) {
			p.advance()
		}
	}

	if err := p.expect(TokenRBrace); err != nil {
		return err
	}

	return nil
}

// parseCompatIndicatorStatement parses a statement in an indicator definition.
func (p *Parser) parseCompatIndicatorStatement(led *LED, keymap *Keymap) error {
	ident, err := p.expectIdent()
	if err != nil {
		return err
	}

	switch ident {
	case "modifiers":
		if err := p.expect(TokenEquals); err != nil {
			return err
		}
		mods, err := p.parseModifierMask(keymap)
		if err != nil {
			return err
		}
		led.mods = mods
		return nil

	case "whichModState", "groups", "whichGroupState", "controls":
		// Skip these for now
		p.skipToSemicolonOrBrace()
		return nil

	default:
		// Skip unknown
		p.skipToSemicolonOrBrace()
		return nil
	}
}

// parseSymbols parses the xkb_symbols section.
func (p *Parser) parseSymbols(keymap *Keymap) error {
	// Optional section name
	if p.check(TokenString) {
		p.advance()
	}

	// Expect {
	if err := p.expect(TokenLBrace); err != nil {
		return err
	}

	// Parse statements until }
	for !p.check(TokenRBrace) && !p.check(TokenEOF) {
		if err := p.parseSymbolsStatement(keymap); err != nil {
			p.addError(err)
			p.skipToSemicolonOrBrace()
		}
		if p.check(TokenSemicolon) {
			p.advance()
		}
	}

	if err := p.expect(TokenRBrace); err != nil {
		return err
	}

	return p.expect(TokenSemicolon)
}

// parseSymbolsStatement parses a single statement in xkb_symbols.
func (p *Parser) parseSymbolsStatement(keymap *Keymap) error {
	// Check for key definition: key <NAME> { ... }
	if p.check(TokenIdent) && p.current.Value == "key" {
		p.advance()
		return p.parseKeyDefinition(keymap)
	}

	ident, err := p.expectIdent()
	if err != nil {
		return err
	}

	switch ident {
	case "name":
		return p.parseGroupName(keymap)
	case "virtual_modifiers":
		return p.parseVirtualModifiers(keymap)
	case "modifier_map":
		return p.parseModifierMap(keymap)
	case "include":
		// Skip include statements (keymap should be self-contained)
		_, _ = p.expectString()
		return nil
	default:
		return p.errorf("unexpected identifier in xkb_symbols: %s", ident)
	}
}

// parseGroupName parses: name[GroupN] = "Name" or name[N] = "Name"
func (p *Parser) parseGroupName(keymap *Keymap) error {
	if err := p.expect(TokenLBracket); err != nil {
		return err
	}

	// Handle both "Group1" and numeric "1" indices
	var group int
	if p.check(TokenNumber) {
		num, err := p.expectNumber()
		if err != nil {
			return err
		}
		group = num - 1 // Convert 1-based to 0-based
	} else {
		groupIdent, err := p.expectIdent()
		if err != nil {
			return err
		}
		group, err = p.parseGroupIdent(groupIdent)
		if err != nil {
			return err
		}
	}

	if err := p.expect(TokenRBracket); err != nil {
		return err
	}

	if err := p.expect(TokenEquals); err != nil {
		return err
	}

	name, err := p.expectString()
	if err != nil {
		return err
	}

	if group < len(keymap.groupNames) {
		keymap.groupNames[group] = name
	}
	if group+1 > keymap.numGroups {
		keymap.numGroups = group + 1
	}

	return nil
}

// parseGroupIdent parses a group identifier like "Group1".
func (p *Parser) parseGroupIdent(name string) (int, error) {
	if !strings.HasPrefix(name, "Group") {
		return 0, fmt.Errorf("invalid group name: %s", name)
	}
	numStr := name[5:]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid group number: %s", name)
	}
	return num - 1, nil // Convert 1-based to 0-based
}

// parseKeyDefinition parses: key <NAME> { ... }
func (p *Parser) parseKeyDefinition(keymap *Keymap) error {
	keycodeName, err := p.expectKeycode()
	if err != nil {
		return err
	}

	keycode, exists := keymap.keycodesByName[keycodeName]
	if !exists {
		// Create a keycode for unknown names (happens with aliases)
		keycode = Keycode(len(keymap.keycodesByName) + int(keymap.minKeycode))
		keymap.keycodeNames[keycode] = keycodeName
		keymap.keycodesByName[keycodeName] = keycode
	}

	if err := p.expect(TokenLBrace); err != nil {
		return err
	}

	key := &Key{
		keycode: keycode,
		name:    keycodeName,
		repeats: true,
	}

	// Parse key body - can be simple [ syms ] or complex { type = ..., symbols[GroupN] = ... }
	// Also handles mixed form: { [ syms ], repeat = no }
	if err := p.parseKeyBody(key, keymap); err != nil {
		return err
	}

	if err := p.expect(TokenRBrace); err != nil {
		return err
	}

	keymap.keys[keycode] = key
	return nil
}

// parseKeyBody parses the body of a complex key definition.
func (p *Parser) parseKeyBody(key *Key, keymap *Keymap) error {
	var currentTypeName string
	groups := make(map[int][]Keysym)

	for !p.check(TokenRBrace) && !p.check(TokenEOF) {
		if p.check(TokenLBracket) {
			// Simple symbols list
			syms, err := p.parseSymbolList(keymap)
			if err != nil {
				return err
			}
			groups[0] = syms
		} else {
			ident, err := p.expectIdent()
			if err != nil {
				return err
			}

			switch ident {
			case "type":
				if p.match(TokenLBracket) {
					// type[GroupN] = "TypeName"
					_, _ = p.expectIdent() // group
					_ = p.expect(TokenRBracket)
				}
				_ = p.expect(TokenEquals)
				typeName, err := p.expectString()
				if err != nil {
					return err
				}
				currentTypeName = typeName

			case "symbols":
				if err := p.expect(TokenLBracket); err != nil {
					return err
				}
				// Handle both "Group1" and numeric "1" indices
				var group int
				if p.check(TokenNumber) {
					num, err := p.expectNumber()
					if err != nil {
						return err
					}
					group = num - 1 // Convert 1-based to 0-based
				} else {
					groupIdent, err := p.expectIdent()
					if err != nil {
						return err
					}
					group, err = p.parseGroupIdent(groupIdent)
					if err != nil {
						return err
					}
				}
				if err := p.expect(TokenRBracket); err != nil {
					return err
				}
				if err := p.expect(TokenEquals); err != nil {
					return err
				}
				syms, err := p.parseSymbolList(keymap)
				if err != nil {
					return err
				}
				groups[group] = syms

			case "actions":
				// Skip action definitions for now
				p.skipToSemicolonOrBrace()

			case "repeat":
				if err := p.expect(TokenEquals); err != nil {
					return err
				}
				repeatIdent, err := p.expectIdent()
				if err != nil {
					return err
				}
				key.repeats = repeatIdent == "yes" || repeatIdent == "true" || repeatIdent == "True"

			case "vmods", "virtualMods":
				// Virtual modifier mapping
				p.skipToSemicolonOrBrace()

			default:
				// Skip unknown
				p.skipToSemicolonOrBrace()
			}
		}

		// Consume comma or semicolon between elements
		if p.match(TokenComma, TokenSemicolon) {
			continue
		}
	}

	// Build key groups from parsed data
	maxGroup := 0
	for g := range groups {
		if g > maxGroup {
			maxGroup = g
		}
	}

	key.groups = make([]KeyGroup, maxGroup+1)
	for g := 0; g <= maxGroup; g++ {
		syms := groups[g]
		var keyType *KeyType
		if currentTypeName != "" {
			keyType = keymap.types[currentTypeName]
		}
		if keyType == nil {
			keyType = p.guessKeyType(keymap, len(syms))
		}
		key.groups[g] = KeyGroup{
			keyType: keyType,
			levels:  p.symsToLevels(syms),
		}
	}

	return nil
}

// parseSymbolList parses: [ sym1, sym2, ... ]
func (p *Parser) parseSymbolList(keymap *Keymap) ([]Keysym, error) {
	if err := p.expect(TokenLBracket); err != nil {
		return nil, err
	}

	var syms []Keysym
	for !p.check(TokenRBracket) && !p.check(TokenEOF) {
		sym, err := p.parseKeysym()
		if err != nil {
			return nil, err
		}
		syms = append(syms, sym)

		if !p.match(TokenComma) {
			break
		}
	}

	if err := p.expect(TokenRBracket); err != nil {
		return nil, err
	}

	return syms, nil
}

// parseKeysym parses a single keysym (identifier or number).
func (p *Parser) parseKeysym() (Keysym, error) {
	if p.check(TokenNumber) {
		numStr := p.current.Value
		p.advance()

		// Single digit numbers are keysym names (0-9), not hex values
		// e.g., "1" means XK_1 (0x31), not Keysym(1)
		if len(numStr) == 1 && numStr[0] >= '0' && numStr[0] <= '9' {
			return Keysym(numStr[0]), nil
		}

		// Hex numbers are literal keysym values
		num, err := p.parseNumber(numStr)
		if err != nil {
			return 0, err
		}
		return Keysym(num), nil
	}

	if p.check(TokenIdent) {
		name := p.current.Value
		p.advance()

		// Look up keysym by name
		ks := KeysymFromName(name, KeysymNameNoFlags)
		if ks == KeyNoSymbol {
			// Try as a single character
			if len(name) == 1 {
				return Keysym(name[0]), nil
			}
		}
		return ks, nil
	}

	return 0, p.errorf("expected keysym")
}

// guessKeyType guesses the appropriate key type based on number of symbols.
func (p *Parser) guessKeyType(keymap *Keymap, numSyms int) *KeyType {
	switch numSyms {
	case 1:
		if t, ok := keymap.types["ONE_LEVEL"]; ok {
			return t
		}
	case 2:
		if t, ok := keymap.types["TWO_LEVEL"]; ok {
			return t
		}
		if t, ok := keymap.types["ALPHABETIC"]; ok {
			return t
		}
	case 4:
		if t, ok := keymap.types["FOUR_LEVEL"]; ok {
			return t
		}
	}

	// Fallback: create a simple type
	return &KeyType{
		name:      "default",
		numLevels: numSyms,
	}
}

// symsToLevels converts a slice of keysyms to KeyLevel slices.
func (p *Parser) symsToLevels(syms []Keysym) []KeyLevel {
	levels := make([]KeyLevel, len(syms))
	for i, sym := range syms {
		levels[i] = KeyLevel{syms: []Keysym{sym}}
	}
	return levels
}

// parseModifierMap parses: modifier_map ModName { <KEY1>, <KEY2>, ... }
func (p *Parser) parseModifierMap(keymap *Keymap) error {
	modName, err := p.expectIdent()
	if err != nil {
		return err
	}

	modMask := p.modifierNameToMask(modName, keymap)

	if err := p.expect(TokenLBrace); err != nil {
		return err
	}

	for !p.check(TokenRBrace) && !p.check(TokenEOF) {
		if p.check(TokenKeycode) {
			keycodeName := p.current.Value
			p.advance()

			if keycode, ok := keymap.keycodesByName[keycodeName]; ok {
				if key, ok := keymap.keys[keycode]; ok {
					key.vmodmap |= modMask
				}
			}
		} else if p.check(TokenIdent) {
			// Skip keysym names
			p.advance()
		}

		if !p.match(TokenComma) {
			break
		}
	}

	if err := p.expect(TokenRBrace); err != nil {
		return err
	}

	return nil
}
