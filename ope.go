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

// Semantic value
type Value struct {
	V Any
	S string
}

// Semantic values
type Values struct {
	Vs     []Value
	Pos    int
	S      string
	Choice int

	isValidString bool
}

func (v *Values) Len() int {
	return len(v.Vs)
}

func (v *Values) ToStr(i int) string {
	return v.Vs[i].V.(string)
}

func (v *Values) ToInt(i int) int {
	return v.Vs[i].V.(int)
}

func (v *Values) ToOpe(i int) operator {
	return v.Vs[i].V.(operator)
}

// Semantic values stack
type semanticValuesStack struct {
	vs []Values
}

func (s *semanticValuesStack) push() *Values {
	s.vs = append(s.vs, Values{})
	return &s.vs[len(s.vs)-1]
}

func (s *semanticValuesStack) pop() {
	s.vs = s.vs[:len(s.vs)-1]
}

// Rule stack
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

// Context
type context struct {
	s string

	errorPos   int
	messagePos int
	message    string

	svStack   semanticValuesStack
	ruleStack ruleStack

	whitespaceOpe operator
	inWhitespace  bool
	inToken       bool

	tracerEnter TracerEnter
	tracerLeave TracerLeave
}

func (c *context) setErrorPos(s string) {
	pos := len(c.s) - len(s)
	if c.errorPos < pos {
		c.errorPos = pos
	}
}

func parse(o operator, s string, v *Values, c *context, d Any) (l int) {
	if c.tracerEnter != nil {
		pos := len(c.s) - len(s)
		c.tracerEnter(o.Label(), s, v, d, pos)
	}

	l = o.parseCore(s, v, c, d)

	if c.tracerLeave != nil {
		c.tracerLeave(o.Label(), s, v, d, l)
	}
	return
}

// Operator
type operator interface {
	Label() string
	parse(s string, v *Values, c *context, d Any) int
	parseCore(s string, v *Values, c *context, d Any) int
	accept(v visitor)
}

// Operator base
type opeBase struct {
	derived operator
}

func (o *opeBase) Label() string {
	return reflect.TypeOf(o.derived).String()[5:]
}

func (o *opeBase) parse(s string, v *Values, c *context, d Any) int {
	return parse(o.derived, s, v, c, d)
}

// Sequence
type sequence struct {
	opeBase
	opes []operator
}

func (o *sequence) parseCore(s string, v *Values, c *context, d Any) (l int) {
	l = 0
	for _, ope := range o.opes {
		chldl := ope.parse(s[l:], v, c, d)
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

// Prioritized Choice
type prioritizedChoice struct {
	opeBase
	opes []operator
}

func (o *prioritizedChoice) parseCore(s string, v *Values, c *context, d Any) (l int) {
	id := 0
	for _, ope := range o.opes {
		chldsv := c.svStack.push()
		l = ope.parse(s, chldsv, c, d)
		c.svStack.pop()
		if success(l) {
			if len(chldsv.Vs) > 0 {
				v.Vs = append(v.Vs, chldsv.Vs...)
			}
			v.Pos = chldsv.Pos
			v.S = chldsv.S
			v.isValidString = chldsv.isValidString
			v.Choice = id
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

// Zero or More
type zeroOrMore struct {
	opeBase
	ope operator
}

func (o *zeroOrMore) parseCore(s string, v *Values, c *context, d Any) (l int) {
	saveErrorPos := c.errorPos
	l = 0
	for len(s)-l > 0 {
		saveVs := v.Vs
		chldl := o.ope.parse(s[l:], v, c, d)
		if fail(chldl) {
			if len(v.Vs) != len(saveVs) {
				v.Vs = saveVs
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

// One or More
type oneOrMore struct {
	opeBase
	ope operator
}

func (o *oneOrMore) parseCore(s string, v *Values, c *context, d Any) (l int) {
	l = o.ope.parse(s, v, c, d)
	if fail(l) {
		return
	}
	saveErrorPos := c.errorPos
	for len(s)-l > 0 {
		saveVs := v.Vs
		chldl := o.ope.parse(s[l:], v, c, d)
		if fail(chldl) {
			if len(v.Vs) != len(saveVs) {
				v.Vs = saveVs
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

// Option
type option struct {
	opeBase
	ope operator
}

func (o *option) parseCore(s string, v *Values, c *context, d Any) (l int) {
	saveErrorPos := c.errorPos
	saveVs := v.Vs
	l = o.ope.parse(s, v, c, d)
	if fail(l) {
		if len(v.Vs) != len(saveVs) {
			v.Vs = saveVs
		}
		c.errorPos = saveErrorPos
		l = 0
	}
	return
}

func (o *option) accept(v visitor) {
	v.visitOption(o)
}

// And Predicate
type andPredicate struct {
	opeBase
	ope operator
}

func (o *andPredicate) parseCore(s string, v *Values, c *context, d Any) (l int) {
	chldsv := c.svStack.push()
	chldl := o.ope.parse(s, chldsv, c, d)
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

// Not Predicate
type notPredicate struct {
	opeBase
	ope operator
}

func (o *notPredicate) parseCore(s string, v *Values, c *context, d Any) (l int) {
	saveErrorPos := c.errorPos

	chldsv := c.svStack.push()
	chldl := o.ope.parse(s, chldsv, c, d)
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

// Literal String
type literalString struct {
	opeBase
	lit string
}

func (o *literalString) parseCore(s string, v *Values, c *context, d Any) (l int) {
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
			len := c.whitespaceOpe.parse(s[l:], v, c, d)
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

// Character Class
type characterClass struct {
	opeBase
	chars string
}

func (o *characterClass) parseCore(s string, v *Values, c *context, d Any) (l int) {
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

// Any Character
type anyCharacter struct {
	opeBase
}

func (o *anyCharacter) parseCore(s string, v *Values, c *context, d Any) (l int) {
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

// Token Boundary
type tokenBoundary struct {
	opeBase
	ope operator
}

func (o *tokenBoundary) parseCore(s string, v *Values, c *context, d Any) (l int) {
	l = o.ope.parse(s, v, c, d)
	if success(l) {
		v.Pos = len(c.s) - len(s)
		v.S = s[:l]
		v.isValidString = true
	}
	return
}

func (o *tokenBoundary) accept(v visitor) {
	v.visitTokenBoundary(o)
}

// Ignore
type ignore struct {
	opeBase
	ope operator
}

func (o *ignore) parseCore(s string, v *Values, c *context, d Any) int {
	chldsv := c.svStack.push()
	l := o.ope.parse(s, chldsv, c, d)
	c.svStack.pop()
	return l
}

func (o *ignore) accept(v visitor) {
	v.visitIgnore(o)
}

// User
type user struct {
	opeBase
	fn func(s string, v *Values, d Any) int
}

func (o *user) parseCore(s string, v *Values, c *context, d Any) int {
	return o.fn(s, v, d)
}

func (o *user) accept(v visitor) {
	v.visitUser(o)
}

// Reference
type reference struct {
	opeBase
	grammar map[string]*Rule
	name    string
	pos     int
}

func (o *reference) parseCore(s string, v *Values, c *context, d Any) (l int) {
	rule := o.getRule()
	l = rule.parse(s, v, c, d)
	return
}

func (o *reference) accept(v visitor) {
	v.visitReference(o)
}

func (o *reference) getRule() operator {
	return o.grammar[o.name] // TODO: fixup
}

// Whitespace
type whitespace struct {
	opeBase
	ope operator
}

func (o *whitespace) parseCore(s string, v *Values, c *context, d Any) (l int) {
	if c.inWhitespace {
		return 0
	}
	c.inWhitespace = true
	l = o.ope.parse(s, v, c, d)
	c.inWhitespace = false
	return
}

func (o *whitespace) accept(v visitor) {
	v.visitWhitespace(o)
}

func Seq(opes ...operator) operator {
	o := &sequence{opes: opes}
	o.derived = o
	return o
}
func Cho(opes ...operator) operator {
	o := &prioritizedChoice{opes: opes}
	o.derived = o
	return o
}
func Zom(ope operator) operator {
	o := &zeroOrMore{ope: ope}
	o.derived = o
	return o
}
func Oom(ope operator) operator {
	o := &oneOrMore{ope: ope}
	o.derived = o
	return o
}
func Opt(ope operator) operator {
	o := &option{ope: ope}
	o.derived = o
	return o
}
func Apd(ope operator) operator {
	o := &andPredicate{ope: ope}
	o.derived = o
	return o
}
func Npd(ope operator) operator {
	o := &notPredicate{ope: ope}
	o.derived = o
	return o
}
func Lit(lit string) operator {
	o := &literalString{lit: lit}
	o.derived = o
	return o
}
func Cls(chars string) operator {
	o := &characterClass{chars: chars}
	o.derived = o
	return o
}
func Dot() operator {
	o := &anyCharacter{}
	o.derived = o
	return o
}
func Tok(ope operator) operator {
	o := &tokenBoundary{ope: ope}
	o.derived = o
	return o
}
func Ign(ope operator) operator {
	o := &ignore{ope: ope}
	o.derived = o
	return o
}
func Usr(fn func(s string, v *Values, d Any) int) operator {
	o := &user{fn: fn}
	o.derived = o
	return o
}
func Ref(g map[string]*Rule, ident string, pos int) operator {
	o := &reference{grammar: g, name: ident, pos: pos}
	o.derived = o
	return o
}
func Wsp(ope operator) operator {
	o := &whitespace{ope: Ign(ope)}
	o.derived = o
	return o
}
