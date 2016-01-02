package peg

// ErrorDetail
type ErrorDetail struct {
	Ln  int
	Col int
	Msg string
}

// Error
type Error struct {
	Details []ErrorDetail
}

func (e *Error) Error() string {
	return "syntax error..." // TODO: Better error report
}

// TracerBegin
type TracerBegin func(name string, s string, sv *SemanticValues, c *context, dt Any, p int)

// TracerEnd
type TracerEnd func(name string, s string, sv *SemanticValues, c *context, dt Any, l int)

// Rule
type Rule struct {
	Name        string
	Ope         Ope
	Action      func(sv *SemanticValues, dt Any) (Any, error)
	Enter       func(dt Any)
	Exit        func(dt Any)
	Message     func() (message string)
	Ignore      bool
	TracerBegin TracerBegin
	TracerEnd   TracerEnd

	tokenChecker  *tokenChecker
	whitespaceOpe Ope
}

func (r *Rule) Parse(s string, dt Any) (l int, v Any, err *Error) {
	sv := &SemanticValues{}
	c := &context{
		s:             s,
		errorPos:      -1,
		messagePos:    -1,
		whitespaceOpe: r.whitespaceOpe,
		tracerBegin:   r.TracerBegin,
		tracerEnd:     r.TracerEnd,
	}

	l = r.parse(s, sv, c, dt)

	if success(l) && len(sv.Vs) > 0 && sv.Vs[0].V != nil {
		v = sv.Vs[0].V
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

func (r *Rule) parse(s string, sv *SemanticValues, c *context, dt Any) int {
	var v Any
	tok := s[:]

	// TODO: Packrat parser support
	c.ruleStack.push(r)
	chldsv := c.svStack.push()
	if r.Enter != nil {
		r.Enter(dt)
	}

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

	//l := r.Ope.parse(s, chldsv, c, dt)
	var l int
	if !c.inToken && r.isToken() {
		c.inToken = true
		l = ope.parse(s, chldsv, c, dt)
		c.inToken = false
	} else {
		l = ope.parse(s, chldsv, c, dt)
	}

	// Invoke action
	if success(l) {
		if chldsv.isValidString {
			tok = chldsv.S
		} else {
			tok = s[:l]
			chldsv.S = s[:l]
		}

		if r.Action != nil {
			var err error
			if v, err = r.Action(chldsv, dt); err != nil {
				pos := len(c.s) - len(s)
				if c.messagePos < pos {
					c.messagePos = pos
					c.message = err.Error()
				}
				l = -1
			}
		} else if len(chldsv.Vs) > 0 {
			v = chldsv.Vs[0].V
		}
	}

	if success(l) {
		if r.Ignore == false {
			sv.Vs = append(sv.Vs, SemanticValue{v, tok})
		}
	} else {
		if r.Message != nil {
			pos := len(c.s) - len(s)
			if c.messagePos < pos {
				c.messagePos = pos
				c.message = r.Message()
			}
		}
	}

	c.ruleStack.pop()
	c.svStack.pop()
	if r.Exit != nil {
		r.Exit(dt)
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
