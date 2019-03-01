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

// Action
type Action func(v *Values, d Any) (Any, error)

// Rule
type Rule struct {
	Name          string
	SS            string
	Pos           int
	Ope           operator
	Action        Action
	Enter         func(d Any)
	Leave         func(d Any)
	Message       func() (message string)
	Ignore        bool
	WhitespaceOpe operator
	WordOpe       operator

	Parameters []string

	TracerEnter func(name string, s string, v *Values, d Any, p int)
	TracerLeave func(name string, s string, v *Values, d Any, p int, l int)

	tokenChecker  *tokenChecker
	disableAction bool
}

func (r *Rule) Parse(s string, d Any) (l int, val Any, err error) {
	v := &Values{}
	c := &context{
		s:             s,
		errorPos:      -1,
		messagePos:    -1,
		whitespaceOpe: r.WhitespaceOpe,
		wordOpe:       r.WordOpe,
		tracerEnter:   r.TracerEnter,
		tracerLeave:   r.TracerLeave,
	}

	var ope operator = r
	if r.WhitespaceOpe != nil {
		ope = Seq(r.WhitespaceOpe, r) // Skip whitespace at beginning
	}

	l, err = ope.parse(s, 0, v, c, d)

	if err == nil && len(v.Vs) > 0 && v.Vs[0] != nil {
		val = v.Vs[0]
	}

	if l != len(s) && err == nil {
		msg := "not exact match"
		ln, col := lineInfo(s, l)
		err = &Error{}
		err.(*Error).Details = append(err.(*Error).Details, ErrorDetail{ln, col, msg})
	}

	return
}

func (o *Rule) Label() string {
	return fmt.Sprintf("[%s]", o.Name)
}

func (o *Rule) parse(s string, p int, v *Values, c *context, d Any) (int, error) {
	return parse(o, s, p, v, c, d)
}

func (r *Rule) parseCore(s string, p int, v *Values, c *context, d Any) (int, error) {
	// Macro reference
	if r.Parameters != nil {
		return r.Ope.parse(s, p, v, c, d)
	}

	if r.Enter != nil {
		r.Enter(d)
	}

	chv := c.push()

	l, err := r.Ope.parse(s, p, chv, c, d)

	// Invoke action
	var val Any
	var intErr error

	if err == nil {
		if r.Action != nil && !r.disableAction {
			chv.S = s[p : p+l]
			chv.Pos = p

			if val, intErr = r.Action(chv, d); intErr != nil {
				ln, col := lineInfo(s, p)

				err = &Error{}
				err.(*Error).Details = append(err.(*Error).Details,
					ErrorDetail{ln, col, intErr.Error()})

				l = 0
			}
		} else if len(chv.Vs) > 0 {
			val = chv.Vs[0]
		}

		if r.Ignore == false {
			v.Vs = append(v.Vs, val)
		}
	} else if r.Message != nil && c.messagePos < p {
		ln, col := lineInfo(s, p)

		err = &Error{}
		err.(*Error).Details = append(err.(*Error).Details, ErrorDetail{ln, col, r.Message()})
	}

	c.pop()

	if r.Leave != nil {
		r.Leave(d)
	}

	return l, err
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

// printLine
func printLine(s string, line int) (int, int) {
	currentLine := 1
	currentOffset := 0
	lineLength := 0

	for currentLine < line && currentOffset < len(s) {
		if s[currentOffset] == '\n' {
			currentLine++
		}
		currentOffset++
	}

	for currentOffset+lineLength < len(s) && s[currentOffset+lineLength] != '\n' {
		lineLength++
	}

	return currentOffset, currentOffset + lineLength
}
