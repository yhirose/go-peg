package peg

// Error
type ErrorDetail struct {
	ln  int
	col int
	msg string
}

type Error struct {
	details []ErrorDetail
}

func (e *Error) Error() string {
	return "syntax error..." // TODO: Better error report
}

// Tracer
type TracerBegin func(name string, s string, sv *SemanticValues, c *context, dt Any, p int)
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

	tokenChecker *tokenChecker
}

func (r *Rule) Parse(s string, dt Any) (l int, v Any, err *Error) {
	sv := &SemanticValues{}
	c := &context{
		s:           s,
		errorPos:    -1,
		messagePos:  -1,
		tracerBegin: r.TracerBegin,
		tracerEnd:   r.TracerEnd,
	}

	l = r.parse(s, sv, c, dt)

	if Success(l) && len(sv.Vs) > 0 && sv.Vs[0].V != nil {
		v = sv.Vs[0].V
	}

	if Fail(l) || l != len(s) {
		var pos int
		var msg string
		if Fail(l) {
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
		err.details = append(err.details, ErrorDetail{ln, col, msg})
	}

	return
}

func (r *Rule) parse(s string, sv *SemanticValues, c *context, dt Any) int {
	var v Any
	tok := s[:]

	chldsv := c.stack.push()
	defer c.stack.pop()

	if r.Enter != nil {
		r.Enter(dt)
	}
	if r.Exit != nil {
		defer r.Exit(dt)
	}

	l := r.Ope.parse(s, chldsv, c, dt)

	// TODO: Packrat parser support
	if Success(l) {
		tok = s[:l]

		if chldsv.isValidString {
			tok = chldsv.S
		} else {
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

	if Success(l) {
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
