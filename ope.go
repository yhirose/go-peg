package peg

import "reflect"

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

func (sv *SemanticValues) Len() int {
	return len(sv.Vs)
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

// parse
func parse(o Ope, s string, sv *SemanticValues, c *context, dt Any) (l int) {
	if c.tracerBegin != nil {
		pos := len(c.s) - len(s)
		c.tracerBegin(o, s, sv, c, dt, pos)
	}

	l = o.parseCore(s, sv, c, dt)

	if c.tracerEnd != nil {
		c.tracerEnd(o, s, sv, c, dt, l)
	}
	return
}

// Ope
type Ope interface {
	Label() string
	parse(s string, sv *SemanticValues, c *context, dt Any) int
	parseCore(s string, sv *SemanticValues, c *context, dt Any) int
	accept(v visitor)
}

// opeBase
type opeBase struct {
	derived Ope
}

func (o *opeBase) Label() string {
	return reflect.TypeOf(o.derived).String()
}

func (o *opeBase) parse(s string, sv *SemanticValues, c *context, dt Any) int {
	return parse(o.derived, s, sv, c, dt)
}

// sequence
type sequence struct {
	opeBase
	opes []Ope
}

func (o *sequence) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
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
	opeBase
	opes []Ope
}

func (o *prioritizedChoice) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	id := 0
	for _, ope := range o.opes {
		chldsv := c.svStack.push()
		l = ope.parse(s, chldsv, c, dt)
		c.svStack.pop()
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
	opeBase
	ope Ope
}

func (o *zeroOrMore) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
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
	opeBase
	ope Ope
}

func (o *oneOrMore) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
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
	opeBase
	ope Ope
}

func (o *option) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
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
	opeBase
	ope Ope
}

func (o *andPredicate) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	chldsv := c.svStack.push()
	chldl := o.ope.parse(s, chldsv, c, dt)
	c.svStack.pop()

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
	opeBase
	ope Ope
}

func (o *notPredicate) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	saveErrorPos := c.errorPos

	chldsv := c.svStack.push()
	chldl := o.ope.parse(s, chldsv, c, dt)
	c.svStack.pop()

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
	opeBase
	lit string
}

func (o *literalString) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
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
	opeBase
	chars string
}

func (o *characterClass) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
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
	opeBase
}

func (o *anyCharacter) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
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
	opeBase
	ope Ope
}

func (o *tokenBoundary) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	l = o.ope.parse(s, sv, c, dt)
	if success(l) {
		sv.Pos = len(c.s) - len(s)
		sv.S = s[:l]
		sv.isValidString = true
	}
	return
}

func (o *tokenBoundary) accept(v visitor) {
	v.visitTokenBoundary(o)
}

// ignore
type ignore struct {
	opeBase
	ope Ope
}

func (o *ignore) parseCore(s string, sv *SemanticValues, c *context, dt Any) int {
	chldsv := c.svStack.push()
	l := o.ope.parse(s, chldsv, c, dt)
	c.svStack.pop()
	return l
}

func (o *ignore) accept(v visitor) {
	v.visitIgnore(o)
}

// user
type user struct {
	opeBase
	fn func(s string, sv *SemanticValues, dt Any) int
}

func (o *user) parseCore(s string, sv *SemanticValues, c *context, dt Any) int {
	return o.fn(s, sv, dt)
}

func (o *user) accept(v visitor) {
	v.visitUser(o)
}

// reference
type reference struct {
	opeBase
	grammar map[string]*Rule
	name    string
	pos     int
}

func (o *reference) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
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
	opeBase
	ope Ope
}

func (o *whitespace) parseCore(s string, sv *SemanticValues, c *context, dt Any) (l int) {
	if c.inWhitespace {
		return 0
	}
	c.inWhitespace = true
	l = o.ope.parse(s, sv, c, dt)
	c.inWhitespace = false
	return
}

func (o *whitespace) accept(v visitor) {
	v.visitWhitespace(o)
}

func Seq(opes ...Ope) Ope {
	o := &sequence{opes: opes}
	o.derived = o
	return o
}
func Cho(opes ...Ope) Ope {
	o := &prioritizedChoice{opes: opes}
	o.derived = o
	return o
}
func Zom(ope Ope) Ope {
	o := &zeroOrMore{ope: ope}
	o.derived = o
	return o
}
func Oom(ope Ope) Ope {
	o := &oneOrMore{ope: ope}
	o.derived = o
	return o
}
func Opt(ope Ope) Ope {
	o := &option{ope: ope}
	o.derived = o
	return o
}
func Apd(ope Ope) Ope {
	o := &andPredicate{ope: ope}
	o.derived = o
	return o
}
func Npd(ope Ope) Ope {
	o := &notPredicate{ope: ope}
	o.derived = o
	return o
}
func Lit(lit string) Ope {
	o := &literalString{lit: lit}
	o.derived = o
	return o
}
func Cls(chars string) Ope {
	o := &characterClass{chars: chars}
	o.derived = o
	return o
}
func Dot() Ope {
	o := &anyCharacter{}
	o.derived = o
	return o
}
func Tok(ope Ope) Ope {
	o := &tokenBoundary{ope: ope}
	o.derived = o
	return o
}
func Ign(ope Ope) Ope {
	o := &ignore{ope: ope}
	o.derived = o
	return o
}
func Usr(fn func(s string, sv *SemanticValues, dt Any) int) Ope {
	o := &user{fn: fn}
	o.derived = o
	return o
}
func Ref(g map[string]*Rule, ident string, pos int) Ope {
	o := &reference{grammar: g, name: ident, pos: pos}
	o.derived = o
	return o
}
func Wsp(ope Ope) Ope {
	o := &whitespace{ope: Ign(ope)}
	o.derived = o
	return o
}
