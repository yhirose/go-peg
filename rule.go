package peg

import "fmt"

// Error detail
type ErrorDetail struct {
	Ln  int
	Col int
	Msg string
}

func (d ErrorDetail) String() string {
	return fmt.Sprintf("%d:%d %s", d.Ln, d.Col, d.Msg)
}

// Error
type Error struct {
	Details []ErrorDetail
}

func (e *Error) Error() string {
	d := e.Details[0]
	return fmt.Sprintf("%d:%d %s", d.Ln, d.Col, d.Msg)
}

// Rule
type Rule struct {
	Name          string
	Ope           operator
	Action        func(v *Values, d Any) (Any, error)
	Enter         func(d Any)
	Leave         func(d Any)
	Message       func() (message string)
	Ignore        bool
	WhitespaceOpe operator
	TracerEnter   func(name string, s string, v *Values, d Any, p int)
	TracerLeave   func(name string, s string, v *Values, d Any, p int, l int)

	tokenChecker *tokenChecker
}

func (r *Rule) Parse(s string, d Any) (l int, val Any, err *Error) {
	v := &Values{}
	c := &context{
		s:             s,
		errorPos:      -1,
		messagePos:    -1,
		whitespaceOpe: r.WhitespaceOpe,
		tracerEnter:   r.TracerEnter,
		tracerLeave:   r.TracerLeave,
	}

	l = r.parse(s, 0, v, c, d)

	if success(l) && len(v.Vs) > 0 && v.Vs[0] != nil {
		val = v.Vs[0]
	}

	if fail(l) || l != len(s) {
		var pos int
		var msg string
		if fail(l) {
			if c.messagePos > -1 {
				pos = c.messagePos
				msg = c.message
			} else {
				msg = "syntax error"
				pos = c.errorPos
			}
		} else {
			msg = "not exact match"
			pos = l
		}
		ln, col := lineInfo(s, pos)
		err = &Error{}
		err.Details = append(err.Details, ErrorDetail{ln, col, msg})
	}

	return
}

func (o *Rule) Label() string {
	return fmt.Sprintf("[%s]", o.Name)
}

func (o *Rule) parse(s string, p int, v *Values, c *context, d Any) int {
	return parse(o, s, p, v, c, d)
}

func (r *Rule) parseCore(s string, p int, v *Values, c *context, d Any) int {
	if r.Enter != nil {
		r.Enter(d)
	}

	c.ruleStack.push(r)
	chv := c.svStack.push()

	// Setup whitespace operator if necessary
	ope := r.Ope
	if !c.inToken && c.whitespaceOpe != nil {
		if c.ruleStack.size() == 1 {
			if r.isToken() && !r.hasTokenBoundary() {
				ope = Seq(c.whitespaceOpe, Tok(r.Ope))
			} else {
				ope = Seq(c.whitespaceOpe, r.Ope)
			}
		} else if r.isToken() {
			if !r.hasTokenBoundary() {
				ope = Seq(Tok(r.Ope), c.whitespaceOpe)
			} else {
				ope = Seq(r.Ope, c.whitespaceOpe)
			}
		}
	}

	// Parse
	var l int
	if !c.inToken && r.isToken() {
		c.inToken = true
		l = ope.parse(s, p, chv, c, d)
		c.inToken = false
	} else {
		l = ope.parse(s, p, chv, c, d)
	}

	// Invoke action
	var val Any

	if success(l) {
		if chv.isValidString {
		} else {
			chv.S = s[p : p+l]
			chv.Pos = p
		}

		if r.Action != nil {
			var err error
			if val, err = r.Action(chv, d); err != nil {
				if c.messagePos < p {
					c.messagePos = p
					c.message = err.Error()
				}
				l = -1
			}
		} else if len(chv.Vs) > 0 {
			val = chv.Vs[0]
		}
	}

	if success(l) {
		if r.Ignore == false {
			v.Vs = append(v.Vs, val)
		}
	} else {
		if r.Message != nil {
			if c.messagePos < p {
				c.messagePos = p
				c.message = r.Message()
			}
		}
	}

	c.svStack.pop()
	c.ruleStack.pop()

	if r.Leave != nil {
		r.Leave(d)
	}

	return l
}

func (r *Rule) accept(v visitor) {
	v.visitRule(r)
}

func (r *Rule) isToken() bool {
	if r.tokenChecker == nil {
		r.tokenChecker = &tokenChecker{}
		r.Ope.accept(r.tokenChecker)
	}
	return r.tokenChecker.isToken()
}

func (r *Rule) hasTokenBoundary() bool {
	if r.tokenChecker == nil {
		r.tokenChecker = &tokenChecker{}
		r.Ope.accept(r.tokenChecker)
	}
	return r.tokenChecker.hasTokenBoundary
}

// lineInfo
func lineInfo(s string, curPos int) (ln int, col int) {
	pos := 0
	colStartPos := 0
	ln = 1

	for pos < curPos {
		if s[pos] == '\n' {
			ln++
			colStartPos = pos + 1
		}
		pos++
	}

	col = pos - colStartPos + 1
	return
}
