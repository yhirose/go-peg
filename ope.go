package peg

func success(l int) bool {
	return l != -1
}

func fail(l int) bool {
	return l == -1
}

// Any
type Any interface {
}

// SemanticValue
type SemanticValue struct {
	V Any
	S string
}

// SemanticValues
type SemanticValues struct {
	Vs     []SemanticValue
	Pos    int
	S      string
	Choice int

	isValidString bool
}

func (sv *SemanticValues) ToStr(i int) string {
	return sv.Vs[i].V.(string)
}

func (sv *SemanticValues) ToInt(i int) int {
	return sv.Vs[i].V.(int)
}

func (sv *SemanticValues) ToByte(i int) byte {
	return sv.Vs[i].V.(byte)
}

func (sv *SemanticValues) ToOpe(i int) Ope {
	return sv.Vs[i].V.(Ope)
}

// semanticValuesStack
type semanticValuesStack struct {
	vs []SemanticValues
}

func (s *semanticValuesStack) push() *SemanticValues {
	s.vs = append(s.vs, SemanticValues{})
	return &s.vs[len(s.vs)-1]
}

func (s *semanticValuesStack) pop() {
	s.vs = s.vs[:len(s.vs)-1]
}

// ruleStack
type ruleStack struct {
	rules []*Rule
}

func (s *ruleStack) push(r *Rule) {
	s.rules = append(s.rules, r)
}

func (s *ruleStack) pop() {
	s.rules = s.rules[:len(s.rules)-1]
}

func (s *ruleStack) size() int {
	return len(s.rules)
}

func (s *ruleStack) top() *Rule {
	return s.rules[len(s.rules)-1]
}

// context
type context struct {
	s string

	errorPos   int
	messagePos int
	message    string

	svStack   semanticValuesStack
	ruleStack ruleStack

	whitespaceOpe Ope
	inWhitespace  bool
	inToken       bool

	tracerBegin TracerBegin
	tracerEnd   TracerEnd
}

func (c *context) setErrorPos(s string) {
	pos := len(c.s) - len(s)
	if c.errorPos < pos {
		c.errorPos = pos
	}
}

func (c *context) traceBegin(name string, s string, sv *SemanticValues, dt Any) {
	if c.tracerBegin != nil {
		pos := len(c.s) - len(s)
		c.tracerBegin(name, s, sv, c, dt, pos)
	}
}

func (c *context) traceEnd(name string, s string, sv *SemanticValues, dt Any, l *int) {
	if c.tracerEnd != nil {
		c.tracerEnd(name, s, sv, c, dt, *l)
	}
}

// Ope
type Ope interface {
	parse(s string, sv *SemanticValues, c *context, dt Any) int
	accept(v visitor)
}

// sequence
type sequence struct {
	opes []Ope
}

func (o *sequence) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("Sequence", s, sv, dt)
	defer c.traceEnd("Sequence", s, sv, dt, &l)

	l = 0
	for _, ope := range o.opes {
		chldl := ope.parse(s[l:], sv, c, dt)
		if fail(chldl) {
			l = -1
			return
		}
		l += chldl
	}
	return
}

func (o *sequence) accept(v visitor) {
	v.visitSequence(o)
}

// prioritizedChoice
type prioritizedChoice struct {
	opes []Ope
}

func (o *prioritizedChoice) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("PrioritizedChoice", s, sv, dt)
	defer c.traceEnd("PrioritizedChoice", s, sv, dt, &l)

	id := 0
	for _, ope := range o.opes {
		chldsv := c.svStack.push()
		defer c.svStack.pop()
		l = ope.parse(s, chldsv, c, dt)
		if success(l) {
			if len(chldsv.Vs) > 0 {
				sv.Vs = append(sv.Vs, chldsv.Vs...)
			}
			sv.Pos = chldsv.Pos
			sv.S = chldsv.S
			sv.isValidString = chldsv.isValidString
			sv.Choice = id
			return
		}
		id++
	}
	l = -1
	return
}

func (o *prioritizedChoice) accept(v visitor) {
	v.visitPrioritizedChoice(o)
}

// zeroOrMore
type zeroOrMore struct {
	ope Ope
}

func (o *zeroOrMore) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("ZeroOrMore", s, sv, dt)
	defer c.traceEnd("ZeroOrMore", s, sv, dt, &l)

	saveErrorPos := c.errorPos
	l = 0
	for len(s)-l > 0 {
		saveVs := sv.Vs
		chldl := o.ope.parse(s[l:], sv, c, dt)
		if fail(chldl) {
			if len(sv.Vs) != len(saveVs) {
				sv.Vs = saveVs
			}
			c.errorPos = saveErrorPos
			break
		}
		l += chldl
	}
	return
}

func (o *zeroOrMore) accept(v visitor) {
	v.visitZeroOrMore(o)
}

// oneOrMore
type oneOrMore struct {
	ope Ope
}

func (o *oneOrMore) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("OneOrMore", s, sv, dt)
	defer c.traceEnd("OneOrMore", s, sv, dt, &l)

	l = o.ope.parse(s, sv, c, dt)
	if fail(l) {
		return
	}
	saveErrorPos := c.errorPos
	for len(s)-l > 0 {
		saveVs := sv.Vs
		chldl := o.ope.parse(s[l:], sv, c, dt)
		if fail(chldl) {
			if len(sv.Vs) != len(saveVs) {
				sv.Vs = saveVs
			}
			c.errorPos = saveErrorPos
			break
		}
		l += chldl
	}
	return
}

func (o *oneOrMore) accept(v visitor) {
	v.visitOneOrMore(o)
}

// option
type option struct {
	ope Ope
}

func (o *option) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("Option", s, sv, dt)
	defer c.traceEnd("Option", s, sv, dt, &l)

	saveErrorPos := c.errorPos
	saveVs := sv.Vs
	l = o.ope.parse(s, sv, c, dt)
	if fail(l) {
		if len(sv.Vs) != len(saveVs) {
			sv.Vs = saveVs
		}
		c.errorPos = saveErrorPos
		l = 0
	}
	return
}

func (o *option) accept(v visitor) {
	v.visitOption(o)
}

// andPredicate
type andPredicate struct {
	ope Ope
}

func (o *andPredicate) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("AndPredicate", s, sv, dt)
	defer c.traceEnd("AndPredicate", s, sv, dt, &l)

	chldsv := c.svStack.push()
	defer c.svStack.pop()
	chldl := o.ope.parse(s, chldsv, c, dt)
	if success(chldl) {
		l = 0
	} else {
		l = -1
	}
	return
}

func (o *andPredicate) accept(v visitor) {
	v.visitAndPredicate(o)
}

// notPredicate
type notPredicate struct {
	ope Ope
}

func (o *notPredicate) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("NotPredicate", s, sv, dt)
	defer c.traceEnd("NotPredicate", s, sv, dt, &l)

	saveErrorPos := c.errorPos
	chldsv := c.svStack.push()
	defer c.svStack.pop()
	chldl := o.ope.parse(s, chldsv, c, dt)
	if success(chldl) {
		c.setErrorPos(s)
		l = -1
	} else {
		c.errorPos = saveErrorPos
		l = 0
	}
	return
}

func (o *notPredicate) accept(v visitor) {
	v.visitNotPredicate(o)
}

// literalString
type literalString struct {
	lit string
}

func (o *literalString) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("LiteralString", s, sv, dt)
	defer c.traceEnd("LiteralString", s, sv, dt, &l)

	l = 0
	for ; l < len(o.lit); l++ {
		if l >= len(s) || s[l] != o.lit[l] {
			c.setErrorPos(s)
			l = -1
			return
		}
	}

	// Skip whiltespace
	if c.whitespaceOpe != nil && c.ruleStack.size() > 0 {
		r := c.ruleStack.top()
		if !r.isToken() {
			len := c.whitespaceOpe.parse(s[l:], sv, c, dt)
			if fail(len) {
				l = -1
				return
			}
			l += len
		}
	}
	return
}

func (o *literalString) accept(v visitor) {
	v.visitLiteralString(o)
}

// characterClass
type characterClass struct {
	chars string
}

func (o *characterClass) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("CharacterClass", s, sv, dt)
	defer c.traceEnd("CharacterClass", s, sv, dt, &l)

	// TODO: UTF8 support
	if len(s) < 1 {
		c.setErrorPos(s)
		l = -1
		return
	}
	ch := s[0]
	i := 0
	for i < len(o.chars) {
		if i+2 < len(o.chars) && o.chars[i+1] == '-' {
			if o.chars[i] <= ch && ch <= o.chars[i+2] {
				l = 1
				return
			}
			i += 3
		} else {
			if o.chars[i] == ch {
				l = 1
				return
			}
			i += 1
		}
	}
	c.setErrorPos(s)
	l = -1
	return
}

func (o *characterClass) accept(v visitor) {
	v.visitCharacterClass(o)
}

// anyCharacter
type anyCharacter struct {
}

func (o *anyCharacter) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("AnyCharacter", s, sv, dt)
	defer c.traceEnd("AnyCharacter", s, sv, dt, &l)

	// TODO: UTF8 support
	if len(s) < 1 {
		c.setErrorPos(s)
		l = -1
		return
	}
	l = 1
	return
}

func (o *anyCharacter) accept(v visitor) {
	v.visitAnyCharacter(o)
}

// tokenBoundary
type tokenBoundary struct {
	ope Ope
}

func (o *tokenBoundary) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("TokenBoundary", s, sv, dt)
	defer c.traceEnd("TokenBoundary", s, sv, dt, &l)

	l = o.ope.parse(s, sv, c, dt)
	if success(l) {
		sv.Pos = len(c.s) - len(s)
		sv.S = s[:l]
		sv.isValidString = true
	}
	return l
}

func (o *tokenBoundary) accept(v visitor) {
	v.visitTokenBoundary(o)
}

// ignore
type ignore struct {
	ope Ope
}

func (o *ignore) parse(s string, sv *SemanticValues, c *context, dt Any) int {
	chldsv := c.svStack.push()
	defer c.svStack.pop()
	return o.ope.parse(s, chldsv, c, dt)
}

func (o *ignore) accept(v visitor) {
	v.visitIgnore(o)
}

// user
type user struct {
	fn func(s string, sv *SemanticValues, dt Any) int
}

func (o *user) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.traceBegin("User", s, sv, dt)
	defer c.traceEnd("User", s, sv, dt, &l)

	l = o.fn(s, sv, dt)
	return
}

func (o *user) accept(v visitor) {
	v.visitUser(o)
}

// reference
type reference struct {
	grammar map[string]*Rule
	name    string
	pos     int
}

func (o *reference) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	name := "[" + o.name + "]"
	c.traceBegin(name, s, sv, dt)
	defer c.traceEnd(name, s, sv, dt, &l)

	rule := o.getRule()
	l = rule.parse(s, sv, c, dt)
	return
}

func (o *reference) accept(v visitor) {
	v.visitReference(o)
}

func (o *reference) getRule() Ope {
	return o.grammar[o.name] // TODO: fixup
}

// whitespace
type whitespace struct {
	ope Ope
}

func (o *whitespace) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	if c.inWhitespace {
		return 0
	}
	c.inWhitespace = true
	defer func() { c.inWhitespace = false }()
	return o.ope.parse(s, sv, c, dt)
}

func (o *whitespace) accept(v visitor) {
	v.visitWhitespace(o)
}

func Seq(opes ...Ope) *sequence                                   { return &sequence{opes} }
func Cho(opes ...Ope) *prioritizedChoice                          { return &prioritizedChoice{opes} }
func Zom(ope Ope) *zeroOrMore                                     { return &zeroOrMore{ope} }
func Oom(ope Ope) *oneOrMore                                      { return &oneOrMore{ope} }
func Opt(ope Ope) *option                                         { return &option{ope} }
func Apd(ope Ope) *andPredicate                                   { return &andPredicate{ope} }
func Npd(ope Ope) *notPredicate                                   { return &notPredicate{ope} }
func Lit(lit string) *literalString                               { return &literalString{lit} }
func Cls(chars string) *characterClass                            { return &characterClass{chars} }
func Dot() *anyCharacter                                          { return &anyCharacter{} }
func Tok(ope Ope) *tokenBoundary                                  { return &tokenBoundary{ope} }
func Ign(ope Ope) *ignore                                         { return &ignore{ope} }
func Usr(fn func(s string, sv *SemanticValues, dt Any) int) *user { return &user{fn} }
func Ref(g map[string]*Rule, ident string, pos int) *reference    { return &reference{g, ident, pos} }
func Wsp(ope Ope) *whitespace                                     { return &whitespace{&ignore{ope}} }
