package peg

import (
	"errors"
	"strings"
)

const (
	assocNone = iota
	assocLeft
	assocRight
)

type BinOpeInfo map[string]struct {
	level int
	assoc int
}

// Expression parsing
type expression struct {
	opeBase
	atom   operator
	binop  operator
	bopinf BinOpeInfo
	action *Action
}

func (o *expression) parseExpr(s string, p int, v *Values, c *context, d Any, minPrec int) (l int) {
	l = o.atom.parse(s, p, v, c, d)
	if fail(l) {
		return
	}

	var tok string
	r := o.binop.(*reference).getRule().(*Rule)
	action := r.Action
	o.binop.(*reference).getRule().(*Rule).Action = func(v *Values, d Any) (Any, error) {
		if action != nil {
			tok = v.Token()
			return action(v, d)
		}
		return v.Vs[0], nil
	}
	defer func() { r.Action = action }()

	saveErrorPos := c.errorPos

	for p+l < len(s) {
		saveVs := v.Vs
		saveTs := v.Ts

		chv := c.push()
		chl := o.binop.parse(s, p+l, chv, c, d)
		c.pop()

		if fail(chl) {
			c.errorPos = saveErrorPos
			break
		}

		inf, ok := o.bopinf[tok]
		if !ok || inf.level < minPrec {
			break
		}

		v.Vs = append(v.Vs, chv.Vs[0])
		l += chl

		nextMinPrec := inf.level
		if inf.assoc == assocLeft {
			nextMinPrec = inf.level + 1
		}

		chv = c.push()
		chl = o.parseExpr(s, p+l, chv, c, d, nextMinPrec)
		c.pop()

		if fail(chl) {
			v.Vs = saveVs
			v.Ts = saveTs
			c.errorPos = saveErrorPos
			break
		}

		v.Vs = append(v.Vs, chv.Vs[0])
		l += chl

		var val Any
		if *o.action != nil {
			var err error
			if val, err = (*o.action)(v, d); err != nil {
				if c.messagePos < p {
					c.messagePos = p
					c.message = err.Error()
				}
				l = -1
				v.Vs = saveVs
				v.Ts = saveTs
				c.errorPos = saveErrorPos
				break
			}
		} else if len(v.Vs) > 0 {
			val = v.Vs[0]
		}

		v.Vs = []Any{val}
	}

	return
}

func (o *expression) parseCore(s string, p int, v *Values, c *context, d Any) (l int) {
	l = o.parseExpr(s, p, v, c, d, 0)
	return
}

func (o *expression) accept(v visitor) {
	v.visitExpression(o)
}

func Exp(atom operator, binop operator, bopinf BinOpeInfo, action *Action) operator {
	o := &expression{atom: atom, binop: binop, bopinf: bopinf, action: action}
	o.derived = o
	return o
}

func EnableExpressionParsing(p *Parser, opt map[string][]string) error {
	// Expression rule
	exprRule := ""
	if vs, ok := opt["%expr"]; ok {
		exprRule = vs[0]
		// TODO: error handling
	}

	// Binary operator info
	binOpeInfo := make(BinOpeInfo)
	if vs, ok := opt["%binop"]; ok {
		level := len(vs)
		for _, s := range vs {
			flds := strings.Split(s, " ")
			// TODO: error handling
			assoc := assocNone
			for i, fld := range flds {
				switch i {
				case 0:
					switch fld {
					case "L":
						assoc = assocLeft
					case "R":
						assoc = assocRight
					default:
						// TODO: error handling
					}
				default:
					binOpeInfo[fld] = struct {
						level int
						assoc int
					}{level, assoc}
				}
			}
			level--
		}
	}

	if r, ok := p.Grammar[exprRule]; ok {
		seq := r.Ope.(*sequence)
		atom := seq.opes[0].(*reference)
		opes := seq.opes[1].(*zeroOrMore).ope.(*sequence).opes
		atom1 := opes[1].(*reference)
		binop := opes[0].(*reference)

		if atom.name != atom1.name {
			return errors.New("expression syntax error")
		}

		r.Ope = Exp(atom, binop, binOpeInfo, &r.Action)
	}

	return nil
}
