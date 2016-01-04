package peg

const (
	WhitespceRuleName = "%whitespace"
)

// PEG parser generator
type duplicate struct {
	name string
	pos  int
}

type data struct {
	grammar    map[string]*Rule
	start      string
	references map[string]int
	duplicates []duplicate
}

func newData() *data {
	return &data{
		grammar:    make(map[string]*Rule),
		references: make(map[string]int),
	}
}

var rStart, rDefinition, rExpression,
	rSequence, rPrefix, rSuffix, rPrimary,
	rIdentifier, rIdentCont, rIdentStart, rIdentRest,
	rLiteral, rClass, rRange, rChar,
	rLEFTARROW, rSLASH, rAND, rNOT, rQUESTION, rSTAR, rPLUS, rOPEN, rCLOSE, rDOT,
	rSpacing, rComment, rSpace, rEndOfLine, rEndOfFile, rBeginTok, rEndTok,
	rIGNORE Rule

func init() {
	// Setup PEG syntax parser
	rStart.Ope = Seq(&rSpacing, Oom(&rDefinition), &rEndOfFile)
	rDefinition.Ope = Seq(Opt(&rIGNORE), &rIdentifier, &rLEFTARROW, &rExpression)

	rExpression.Ope = Seq(&rSequence, Zom(Seq(&rSLASH, &rSequence)))
	rSequence.Ope = Zom(&rPrefix)
	rPrefix.Ope = Seq(Opt(Cho(&rAND, &rNOT)), &rSuffix)
	rSuffix.Ope = Seq(&rPrimary, Opt(Cho(&rQUESTION, &rSTAR, &rPLUS)))
	rPrimary.Ope = Cho(
		Seq(Opt(&rIGNORE), &rIdentifier, Npd(&rLEFTARROW)),
		Seq(&rOPEN, &rExpression, &rCLOSE),
		Seq(&rBeginTok, &rExpression, &rEndTok),
		&rLiteral, &rClass, &rDOT)

	rIdentifier.Ope = Seq(&rIdentCont, &rSpacing)
	rIdentCont.Ope = Seq(&rIdentStart, Zom(&rIdentRest))
	rIdentStart.Ope = Cls("a-zA-Z_\x80-\xff%")
	rIdentRest.Ope = Cho(&rIdentStart, Cls("0-9"))

	rLiteral.Ope = Cho(
		Seq(Lit("'"), Tok(Zom(Seq(Npd(Lit("'")), &rChar))), Lit("'"), &rSpacing),
		Seq(Lit("\""), Tok(Zom(Seq(Npd(Lit("\"")), &rChar))), Lit("\""), &rSpacing))

	rClass.Ope = Seq(Lit("["), Tok(Zom(Seq(Npd(Lit("]")), &rRange))), Lit("]"), &rSpacing)

	rRange.Ope = Cho(Seq(&rChar, Lit("-"), &rChar), &rChar)
	rChar.Ope = Cho(
		Seq(Lit("\\"), Cls("nrt'\"[]\\")),
		Seq(Lit("\\"), Cls("0-3"), Cls("0-7"), Cls("0-7")),
		Seq(Lit("\\"), Cls("0-7"), Opt(Cls("0-7"))),
		Seq(Lit("\\x"), Cls("0-9a-fA-F"), Opt(Cls("0-9a-fA-F"))),
		Seq(Npd(Lit("\\")), Dot()))

	rLEFTARROW.Ope = Seq(Lit("<-"), &rSpacing)
	rSLASH.Ope = Seq(Lit("/"), &rSpacing)
	rSLASH.Ignore = true
	rAND.Ope = Seq(Lit("&"), &rSpacing)
	rNOT.Ope = Seq(Lit("!"), &rSpacing)
	rQUESTION.Ope = Seq(Lit("?"), &rSpacing)
	rSTAR.Ope = Seq(Lit("*"), &rSpacing)
	rPLUS.Ope = Seq(Lit("+"), &rSpacing)
	rOPEN.Ope = Seq(Lit("("), &rSpacing)
	rCLOSE.Ope = Seq(Lit(")"), &rSpacing)
	rDOT.Ope = Seq(Lit("."), &rSpacing)

	rSpacing.Ope = Zom(Cho(&rSpace, &rComment))
	rComment.Ope = Seq(Lit("#"), Zom(Seq(Npd(&rEndOfLine), Dot())), &rEndOfLine)
	rSpace.Ope = Cho(Lit(" "), Lit("\t"), &rEndOfLine)
	rEndOfLine.Ope = Cho(Lit("\r\n"), Lit("\n"), Lit("\r"))
	rEndOfFile.Ope = Npd(Dot())

	rBeginTok.Ope = Seq(Lit("<"), &rSpacing)
	rEndTok.Ope = Seq(Lit(">"), &rSpacing)

	rIGNORE.Ope = Lit("~")

	// Setup actions
	rDefinition.Action = func(v *Values, d Any) (val Any, err error) {
		data := d.(*data)

		ignore := len(v.Vs) == 4

		baseId := 0
		if ignore {
			baseId = 1
		}

		name := v.ToStr(baseId)
		ope := v.ToOpe(baseId + 2)

		_, ok := data.grammar[name]
		if ok {
			data.duplicates = append(data.duplicates, duplicate{name, v.Pos})
		} else {
			data.grammar[name] = &Rule{
				Ope:    ope,
				Name:   name,
				Ignore: ignore,
			}
			if len(data.start) == 0 {
				data.start = name
			}
		}
		return
	}

	rExpression.Action = func(v *Values, d Any) (val Any, err error) {
		if len(v.Vs) == 1 {
			val = v.ToOpe(0)
		} else {
			var opes []operator
			for i := 0; i < len(v.Vs); i++ {
				opes = append(opes, v.ToOpe(i))
			}
			val = Cho(opes...)
		}
		return
	}

	rSequence.Action = func(v *Values, d Any) (val Any, err error) {
		if len(v.Vs) == 1 {
			val = v.ToOpe(0)
		} else {
			var opes []operator
			for i := 0; i < len(v.Vs); i++ {
				opes = append(opes, v.ToOpe(i))
			}
			val = Seq(opes...)
		}
		return
	}

	rPrefix.Action = func(v *Values, d Any) (val Any, err error) {
		if len(v.Vs) == 1 {
			val = v.ToOpe(0)
		} else {
			tok := v.ToStr(0)
			ope := v.ToOpe(1)
			switch tok {
			case "&":
				val = Apd(ope)
			case "!":
				val = Npd(ope)
			}
		}
		return
	}

	rSuffix.Action = func(v *Values, d Any) (val Any, err error) {
		ope := v.ToOpe(0)
		if len(v.Vs) == 1 {
			val = ope
		} else {
			tok := v.ToStr(1)
			switch tok {
			case "?":
				val = Opt(ope)
			case "*":
				val = Zom(ope)
			case "+":
				val = Oom(ope)
			}
		}
		return
	}

	rPrimary.Action = func(v *Values, d Any) (val Any, err error) {
		data := d.(*data)

		switch v.Choice {
		case 0: // Reference
			ignore := len(v.Vs) == 2
			baseId := 0
			if ignore {
				baseId = 1
			}

			ident := v.ToStr(baseId)

			if _, ok := data.references[ident]; !ok {
				data.references[ident] = v.Pos // for error handling
			}

			if ignore {
				val = Ign(Ref(data.grammar, ident, v.Pos))
			} else {
				val = Ref(data.grammar, ident, v.Pos)
			}
		case 1: // (Expression)
			val = v.ToOpe(1)
		case 2: // TokenBoundary
			val = Tok(v.ToOpe(1))
		default:
			val = v.ToOpe(0)
		}
		return
	}

	rIdentCont.Action = func(v *Values, d Any) (Any, error) {
		return v.S, nil
	}

	rLiteral.Action = func(v *Values, d Any) (Any, error) {
		return Lit(resolveEscapeSequence(v.S)), nil
	}

	rClass.Action = func(v *Values, d Any) (Any, error) {
		return Cls(resolveEscapeSequence(v.S)), nil
	}

	rAND.Action = func(v *Values, d Any) (Any, error) {
		return v.S[:1], nil
	}
	rNOT.Action = func(v *Values, d Any) (Any, error) {
		return v.S[:1], nil
	}
	rQUESTION.Action = func(v *Values, d Any) (Any, error) {
		return v.S[:1], nil
	}
	rSTAR.Action = func(v *Values, d Any) (Any, error) {
		return v.S[:1], nil
	}
	rPLUS.Action = func(v *Values, d Any) (Any, error) {
		return v.S[:1], nil
	}

	rDOT.Action = func(v *Values, d Any) (Any, error) {
		return Dot(), nil
	}
}

func isHex(c byte) (v int, ok bool) {
	if '0' <= c && c <= '9' {
		v = int(c - '0')
		ok = true
	} else if 'a' <= c && c <= 'f' {
		v = int(c - 'a' + 10)
		ok = true
	} else if 'A' <= c && c <= 'F' {
		v = int(c - 'A' + 10)
		ok = true
	}
	return
}

func isDigit(c byte) (v int, ok bool) {
	if '0' <= c && c <= '9' {
		v = int(c - '0')
		ok = true
	}
	return
}

func parseHexNumber(s string, i int) (byte, int) {
	ret := 0
	for i < len(s) {
		val, ok := isHex(s[i])
		if !ok {
			break
		}
		ret = ret*16 + val
		i++
	}
	return byte(ret), i
}

func parseOctNumber(s string, i int) (byte, int) {
	ret := 0
	for i < len(s) {
		val, ok := isDigit(s[i])
		if !ok {
			break
		}
		ret = ret*8 + val
		i++
	}
	return byte(ret), i
}

func resolveEscapeSequence(s string) string {
	n := len(s)
	b := make([]byte, 0, n)

	i := 0
	for i < n {
		ch := s[i]
		if ch == '\\' {
			i++
			switch s[i] {
			case 'n':
				b = append(b, '\n')
				i++
			case 'r':
				b = append(b, '\r')
				i++
			case 't':
				b = append(b, '\t')
				i++
			case '\'':
				b = append(b, '\'')
				i++
			case '"':
				b = append(b, '"')
				i++
			case '[':
				b = append(b, '[')
				i++
			case ']':
				b = append(b, ']')
				i++
			case '\\':
				b = append(b, '\\')
				i++
			case 'x':
				ch, i = parseHexNumber(s, i+1)
				b = append(b, ch)
			default:
				ch, i = parseOctNumber(s, i)
				b = append(b, ch)
			}
		} else {
			b = append(b, ch)
			i++
		}
	}

	return string(b)
}

// Parser
type Parser struct {
	Grammar     map[string]*Rule
	start       string
	TracerEnter func(name string, s string, v *Values, d Any, p int)
	TracerLeave func(name string, s string, v *Values, d Any, p int, l int)
}

func NewParser(s string) (p *Parser, err *Error) {
	return NewParserWithUserRules(s, nil)
}

func NewParserWithUserRules(s string, rules map[string]operator) (p *Parser, err *Error) {
	data := newData()

	_, _, err = rStart.Parse(s, data)
	if err != nil {
		return nil, err
	}

	// User provided rules
	for name, ope := range rules {
		ignore := false

		if len(name) > 0 && name[0] == '~' {
			ignore = true
			name = name[1:]
		}

		if len(name) > 0 {
			data.grammar[name] = &Rule{
				Ope:    ope,
				Name:   name,
				Ignore: ignore,
			}
		}
	}

	// Check duplicated definitions
	if len(data.duplicates) > 0 {
		err = &Error{}
		for _, dup := range data.duplicates {
			ln, col := lineInfo(s, dup.pos)
			msg := "'" + dup.name + "' is already defined."
			err.Details = append(err.Details, ErrorDetail{ln, col, msg})
		}
	}

	// Check missing definitions
	for name, pos := range data.references {
		if _, ok := data.grammar[name]; !ok {
			if err == nil {
				err = &Error{}
			}
			ln, col := lineInfo(s, pos)
			msg := "'" + name + "' is not defined."
			err.Details = append(err.Details, ErrorDetail{ln, col, msg})
		}
	}

	if err != nil {
		return nil, err
	}

	// Check left recursion
	for name, r := range data.grammar {
		lr := &detectLeftRecursion{
			pos:  -1,
			name: name,
			refs: make(map[string]bool),
			done: false,
		}
		r.accept(lr)
		if lr.pos != -1 {
			if err == nil {
				err = &Error{}
			}
			ln, col := lineInfo(s, lr.pos)
			msg := "'" + name + "' is left recursive."
			err.Details = append(err.Details, ErrorDetail{ln, col, msg})
		}
	}

	if err != nil {
		return nil, err
	}

	// Automatic whitespace skipping
	if r, ok := data.grammar[WhitespceRuleName]; ok {
		data.grammar[data.start].WhitespaceOpe = Wsp(r)
	}

	p = &Parser{Grammar: data.grammar, start: data.start}
	return
}

func (p *Parser) Parse(s string, d Any) (err *Error) {
	_, err = p.ParseAndGetValue(s, d)
	return
}

func (p *Parser) ParseAndGetValue(s string, d Any) (val Any, err *Error) {
	r := p.Grammar[p.start]
	r.TracerEnter = p.TracerEnter
	r.TracerLeave = p.TracerLeave
	_, val, err = r.Parse(s, d)
	return
}
