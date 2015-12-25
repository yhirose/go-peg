package peg

// Utilities
func Success(l int) bool {
	return l != -1
}

func Fail(l int) bool {
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

func (s *semanticValuesStack) len() int {
	return len(s.vs)
}

func (s *semanticValuesStack) push() *SemanticValues {
	s.vs = append(s.vs, SemanticValues{})
	return &s.vs[len(s.vs)-1]
}

func (s *semanticValuesStack) pop() {
	s.vs = s.vs[:len(s.vs)-1]
}

// context
type context struct {
	s           string
	errorPos    int
	messagePos  int
	message     string
	stack       semanticValuesStack
	tracerBegin TracerBegin
	tracerEnd   TracerEnd
}

func (c *context) setErrorPos(s string) {
	pos := len(c.s) - len(s)
	if c.errorPos < pos {
		c.errorPos = pos
	}
}

func (c *context) trace(name string, s string, sv *SemanticValues, dt Any) {
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

// Sequence
type Sequence struct {
	opes []Ope
}

func (o *Sequence) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("Sequence", s, sv, dt)
	defer c.traceEnd("Sequence", s, sv, dt, &l)
	l = 0
	for _, ope := range o.opes {
		chldl := ope.parse(s[l:], sv, c, dt)
		if Fail(chldl) {
			l = -1
			return
		}
		l += chldl
	}
	return
}

func (o *Sequence) accept(v visitor) {
	v.visitSequence(o)
}

// PrioritizedChoice
type PrioritizedChoice struct {
	opes []Ope
}

func (o *PrioritizedChoice) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("PrioritizedChoice", s, sv, dt)
	defer c.traceEnd("PrioritizedChoice", s, sv, dt, &l)
	id := 0
	for _, ope := range o.opes {
		chldsv := c.stack.push()
		defer c.stack.pop()
		l = ope.parse(s, chldsv, c, dt)
		if Success(l) {
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

func (o *PrioritizedChoice) accept(v visitor) {
	v.visitPrioritizedChoice(o)
}

// ZeroOrMore
type ZeroOrMore struct {
	ope Ope
}

func (o *ZeroOrMore) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("ZeroOrMore", s, sv, dt)
	defer c.traceEnd("ZeroOrMore", s, sv, dt, &l)
	saveErrorPos := c.errorPos
	l = 0
	for len(s)-l > 0 {
		saveVs := sv.Vs
		chldl := o.ope.parse(s[l:], sv, c, dt)
		if Fail(chldl) {
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

func (o *ZeroOrMore) accept(v visitor) {
	v.visitZeroOrMore(o)
}

// OneOrMore
type OneOrMore struct {
	ope Ope
}

func (o *OneOrMore) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("OneOrMore", s, sv, dt)
	defer c.traceEnd("OneOrMore", s, sv, dt, &l)
	l = o.ope.parse(s, sv, c, dt)
	if Fail(l) {
		return
	}
	saveErrorPos := c.errorPos
	for len(s)-l > 0 {
		saveVs := sv.Vs
		chldl := o.ope.parse(s[l:], sv, c, dt)
		if Fail(chldl) {
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

func (o *OneOrMore) accept(v visitor) {
	v.visitOneOrMore(o)
}

// Option
type Option struct {
	ope Ope
}

func (o *Option) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("Option", s, sv, dt)
	defer c.traceEnd("Option", s, sv, dt, &l)
	saveErrorPos := c.errorPos
	saveVs := sv.Vs
	l = o.ope.parse(s, sv, c, dt)
	if Fail(l) {
		if len(sv.Vs) != len(saveVs) {
			sv.Vs = saveVs
		}
		c.errorPos = saveErrorPos
		l = 0
	}
	return
}

func (o *Option) accept(v visitor) {
	v.visitOption(o)
}

// AndPredicate
type AndPredicate struct {
	ope Ope
}

func (o *AndPredicate) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("AndPredicate", s, sv, dt)
	defer c.traceEnd("AndPredicate", s, sv, dt, &l)
	chldsv := c.stack.push()
	defer c.stack.pop()
	chldl := o.ope.parse(s, chldsv, c, dt)
	if Success(chldl) {
		l = 0
	} else {
		l = -1
	}
	return
}

func (o *AndPredicate) accept(v visitor) {
	v.visitAndPredicate(o)
}

// NotPredicate
type NotPredicate struct {
	ope Ope
}

func (o *NotPredicate) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("NotPredicate", s, sv, dt)
	defer c.traceEnd("NotPredicate", s, sv, dt, &l)
	saveErrorPos := c.errorPos
	chldsv := c.stack.push()
	defer c.stack.pop()
	chldl := o.ope.parse(s, chldsv, c, dt)
	if Success(chldl) {
		c.setErrorPos(s)
		l = -1
	} else {
		c.errorPos = saveErrorPos
		l = 0
	}
	return
}

func (o *NotPredicate) accept(v visitor) {
	v.visitNotPredicate(o)
}

// LiteralString
type LiteralString struct {
	lit string
}

func (o *LiteralString) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("LiteralString", s, sv, dt)
	defer c.traceEnd("LiteralString", s, sv, dt, &l)
	l = 0
	for ; l < len(o.lit); l++ {
		if l >= len(s) || s[l] != o.lit[l] {
			c.setErrorPos(s)
			l = -1
			return
		}
	}
	return
}

func (o *LiteralString) accept(v visitor) {
	v.visitLiteralString(o)
}

// CharacterClass
type CharacterClass struct {
	chars string
}

func (o *CharacterClass) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("CharacterClass", s, sv, dt)
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

func (o *CharacterClass) accept(v visitor) {
	v.visitCharacterClass(o)
}

// AnyCharacter
type AnyCharacter struct {
}

func (o *AnyCharacter) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("AnyCharacter", s, sv, dt)
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

func (o *AnyCharacter) accept(v visitor) {
	v.visitAnyCharacter(o)
}

// TokenBoundary
type TokenBoundary struct {
	ope Ope
}

func (o *TokenBoundary) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("TokenBoundary", s, sv, dt)
	defer c.traceEnd("TokenBoundary", s, sv, dt, &l)
	l = o.ope.parse(s, sv, c, dt)
	if Success(l) {
		sv.Pos = len(c.s) - len(s)
		sv.S = s[:l]
		sv.isValidString = true
	}
	return l
}

func (o *TokenBoundary) accept(v visitor) {
	v.visitTokenBoundary(o)
}

// Ignore
type Ignore struct {
	ope Ope
}

func (o *Ignore) parse(s string, sv *SemanticValues, c *context, dt Any) int {
	chldsv := c.stack.push()
	defer c.stack.pop()
	return o.ope.parse(s, chldsv, c, dt)
}

func (o *Ignore) accept(v visitor) {
	v.visitIgnore(o)
}

// User
type User struct {
	fn func(s string, sv *SemanticValues, dt Any) int
}

func (o *User) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	c.trace("User", s, sv, dt)
	defer c.traceEnd("User", s, sv, dt, &l)
	l = o.fn(s, sv, dt)
	return
}

func (o *User) accept(v visitor) {
	v.visitUser(o)
}

// Reference
type Reference struct {
	grammar map[string]*Rule
	name    string
	pos     int
}

func (o *Reference) parse(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	name := "[" + o.name + "]"
	c.trace(name, s, sv, dt)
	defer c.traceEnd(name, s, sv, dt, &l)
	rule := o.getRule()
	l = rule.parse(s, sv, c, dt)
	return
}

func (o *Reference) accept(v visitor) {
	v.visitReference(o)
}

func (o *Reference) getRule() Ope {
	return o.grammar[o.name] // TODO: fixup
}

// Constructors
func Seq(opes ...Ope) *Sequence                                   { return &Sequence{opes} }
func Cho(opes ...Ope) *PrioritizedChoice                          { return &PrioritizedChoice{opes} }
func Zom(ope Ope) *ZeroOrMore                                     { return &ZeroOrMore{ope} }
func Oom(ope Ope) *OneOrMore                                      { return &OneOrMore{ope} }
func Opt(ope Ope) *Option                                         { return &Option{ope} }
func Apd(ope Ope) *AndPredicate                                   { return &AndPredicate{ope} }
func Npd(ope Ope) *NotPredicate                                   { return &NotPredicate{ope} }
func Lit(lit string) *LiteralString                               { return &LiteralString{lit} }
func Cls(chars string) *CharacterClass                            { return &CharacterClass{chars} }
func Dot() *AnyCharacter                                          { return &AnyCharacter{} }
func Tok(ope Ope) *TokenBoundary                                  { return &TokenBoundary{ope} }
func Ign(ope Ope) *Ignore                                         { return &Ignore{ope} }
func Usr(fn func(s string, sv *SemanticValues, dt Any) int) *User { return &User{fn} }
func Ref(g map[string]*Rule, ident string, pos int) *Reference    { return &Reference{g, ident, pos} }
