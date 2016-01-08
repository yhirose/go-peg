package peg

import (
	"reflect"
	"sync"
)

func success(l int) bool {
	return l != -1
}

func fail(l int) bool {
	return l == -1
}

// Any
type Any interface {
}

// Token
type Token struct {
	Pos int
	S   string
}

// Semantic values
type Values struct {
	Vs     []Any
	Pos    int
	S      string
	Choice int
	Ts     []Token
}

func (v *Values) Len() int {
	return len(v.Vs)
}

func (v *Values) ToStr(i int) string {
	return v.Vs[i].(string)
}

func (v *Values) ToInt(i int) int {
	return v.Vs[i].(int)
}

func (v *Values) ToOpe(i int) operator {
	return v.Vs[i].(operator)
}

func (v *Values) Token() string {
	if len(v.Ts) > 0 {
		return v.Ts[0].S
	}
	return v.S
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

	keywordOpe operator

	tracerEnter func(name string, s string, v *Values, d Any, p int)
	tracerLeave func(name string, s string, v *Values, d Any, p int, l int)
}

func (c *context) setErrorPos(p int) {
	if c.errorPos < p {
		c.errorPos = p
	}
}

func parse(o operator, s string, p int, v *Values, c *context, d Any) (l int) {
	if c.tracerEnter != nil {
		c.tracerEnter(o.Label(), s, v, d, p)
	}

	l = o.parseCore(s, p, v, c, d)

	if c.tracerLeave != nil {
		c.tracerLeave(o.Label(), s, v, d, p, l)
	}
	return
}

// Operator
type operator interface {
	Label() string
	parse(s string, p int, v *Values, c *context, d Any) int
	parseCore(s string, p int, v *Values, c *context, d Any) int
	accept(v visitor)
}

// Operator base
type opeBase struct {
	derived operator
}

func (o *opeBase) Label() string {
	return reflect.TypeOf(o.derived).String()[5:]
}

func (o *opeBase) parse(s string, p int, v *Values, c *context, d Any) int {
	return parse(o.derived, s, p, v, c, d)
}

// Sequence
type sequence struct {
	opeBase
	opes []operator
}

func (o *sequence) parseCore(s string, p int, v *Values, c *context, d Any) (l int) {
	l = 0
	for _, ope := range o.opes {
		chl := ope.parse(s, p+l, v, c, d)
		if fail(chl) {
			l = -1
			return
		}
		l += chl
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

func (o *prioritizedChoice) parseCore(s string, p int, v *Values, c *context, d Any) (l int) {
	id := 0
	for _, ope := range o.opes {
		chv := c.svStack.push()
		l = ope.parse(s, p, chv, c, d)
		c.svStack.pop()
		if success(l) {
			if len(chv.Vs) > 0 {
				v.Vs = append(v.Vs, chv.Vs...)
			}
			v.Pos = chv.Pos
			v.S = chv.S
			v.Choice = id
			if len(chv.Ts) > 0 {
				v.Ts = append(v.Ts, chv.Ts...)
			}
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

func (o *zeroOrMore) parseCore(s string, p int, v *Values, c *context, d Any) (l int) {
	saveErrorPos := c.errorPos
	l = 0
	for p+l < len(s) {
		saveVs := v.Vs
		saveTs := v.Ts
		chl := o.ope.parse(s, p+l, v, c, d)
		if fail(chl) {
			v.Vs = saveVs
			v.Ts = saveTs
			c.errorPos = saveErrorPos
			break
		}
		l += chl
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

func (o *oneOrMore) parseCore(s string, p int, v *Values, c *context, d Any) (l int) {
	l = o.ope.parse(s, p, v, c, d)
	if fail(l) {
		return
	}
	saveErrorPos := c.errorPos
	for p+l < len(s) {
		saveVs := v.Vs
		saveTs := v.Ts
		chl := o.ope.parse(s, p+l, v, c, d)
		if fail(chl) {
			v.Vs = saveVs
			v.Ts = saveTs
			c.errorPos = saveErrorPos
			break
		}
		l += chl
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

func (o *option) parseCore(s string, p int, v *Values, c *context, d Any) (l int) {
	saveErrorPos := c.errorPos
	saveVs := v.Vs
	saveTs := v.Ts
	l = o.ope.parse(s, p, v, c, d)
	if fail(l) {
		v.Vs = saveVs
		v.Ts = saveTs
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

func (o *andPredicate) parseCore(s string, p int, v *Values, c *context, d Any) (l int) {
	chv := c.svStack.push()
	chl := o.ope.parse(s, p, chv, c, d)
	c.svStack.pop()

	if success(chl) {
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

func (o *notPredicate) parseCore(s string, p int, v *Values, c *context, d Any) (l int) {
	saveErrorPos := c.errorPos

	chv := c.svStack.push()
	chl := o.ope.parse(s, p, chv, c, d)
	c.svStack.pop()

	if success(chl) {
		c.setErrorPos(p)
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
	lit           string
	initIsKeyword sync.Once
	isKeyword     bool
}

func (o *literalString) parseCore(s string, p int, v *Values, c *context, d Any) int {
	l := 0
	for ; l < len(o.lit); l++ {
		if p+l == len(s) || s[p+l] != o.lit[l] {
			c.setErrorPos(p)
			return -1
		}
	}

	// Keyword boundary check
	o.initIsKeyword.Do(func() {
		if c.keywordOpe != nil {
			len := c.keywordOpe.parse(o.lit, 0, &Values{}, &context{}, nil)
			o.isKeyword = success(len)
		}
	})
	if o.isKeyword {
		len := Npd(c.keywordOpe).parse(s, p+l, v, &context{}, nil)
		if fail(len) {
			return -1
		}
		l += len
	}

	// Skip whiltespace
	if c.whitespaceOpe != nil {
		len := c.whitespaceOpe.parse(s, p+l, v, c, d)
		if fail(len) {
			return -1
		}
		l += len
	}
	return l
}

func (o *literalString) accept(v visitor) {
	v.visitLiteralString(o)
}

// Character Class
type characterClass struct {
	opeBase
	chars string
}

func (o *characterClass) parseCore(s string, p int, v *Values, c *context, d Any) (l int) {
	// TODO: UTF8 support
	if len(s)-p < 1 {
		c.setErrorPos(p)
		l = -1
		return
	}
	ch := s[p]
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
			i++
		}
	}
	c.setErrorPos(p)
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

func (o *anyCharacter) parseCore(s string, p int, v *Values, c *context, d Any) (l int) {
	// TODO: UTF8 support
	if len(s)-p < 1 {
		c.setErrorPos(p)
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

func (o *tokenBoundary) parseCore(s string, p int, v *Values, c *context, d Any) int {
	l := o.ope.parse(s, p, v, c, d)
	if success(l) {
		v.Ts = append(v.Ts, Token{p, s[p : p+l]})

		// Skip whiltespace
		if c.whitespaceOpe != nil {
			len := c.whitespaceOpe.parse(s, p+l, v, c, d)
			if fail(len) {
				return -1
			}
			l += len
		}
	}
	return l
}

func (o *tokenBoundary) accept(v visitor) {
	v.visitTokenBoundary(o)
}

// Ignore
type ignore struct {
	opeBase
	ope operator
}

func (o *ignore) parseCore(s string, p int, v *Values, c *context, d Any) int {
	chv := c.svStack.push()
	l := o.ope.parse(s, p, chv, c, d)
	c.svStack.pop()
	return l
}

func (o *ignore) accept(v visitor) {
	v.visitIgnore(o)
}

// User
type user struct {
	opeBase
	fn func(s string, p int, v *Values, d Any) int
}

func (o *user) parseCore(s string, p int, v *Values, c *context, d Any) int {
	return o.fn(s, p, v, d)
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

func (o *reference) parseCore(s string, p int, v *Values, c *context, d Any) (l int) {
	rule := o.getRule()
	l = rule.parse(s, p, v, c, d)
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

func (o *whitespace) parseCore(s string, p int, v *Values, c *context, d Any) int {
	if c.inWhitespace {
		return 0
	} else {
		c.inWhitespace = true
		l := o.ope.parse(s, p, v, c, d)
		c.inWhitespace = false
		return l
	}
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
func Usr(fn func(s string, p int, v *Values, d Any) int) operator {
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
