package peg

import (
	"strconv"
	"testing"
)

func reduce(sv *SemanticValues, dt Any) (Any, error) {
	ret := sv.ToInt(0)
	for i := 1; i < len(sv.Vs); i += 2 {
		ope := sv.ToStr(i)
		n := sv.ToInt(i + 1)
		switch ope {
		case "+":
			ret += n
		case "-":
			ret -= n
		case "*":
			ret *= n
		case "/":
			ret /= n
		}
	}
	return ret, nil
}

func TestCalc(t *testing.T) {
	// Rules
	var EXPRESSION, TERM, FACTOR, TERM_OPERATOR, FACTOR_OPERATOR, NUMBER, WS Rule

	// Grammar
	EXPRESSION.Ope = Seq(&WS, &TERM, Zom(Seq(&TERM_OPERATOR, &TERM)))
	TERM.Ope = Seq(&FACTOR, Zom(Seq(&FACTOR_OPERATOR, &FACTOR)))
	FACTOR.Ope = Cho(&NUMBER, Seq(Lit("("), &WS, &EXPRESSION, Lit(")"), &WS))
	TERM_OPERATOR.Ope = Seq(Tok(Cls("-+")), &WS)
	FACTOR_OPERATOR.Ope = Seq(Tok(Cls("*/")), &WS)
	NUMBER.Ope = Seq(Tok(Oom(Cls("0-9"))), &WS)
	WS.Ope = Zom(Cls(" \t"))

	WS.Ignore = true

	// Actions
	EXPRESSION.Action = reduce
	TERM.Action = reduce
	TERM_OPERATOR.Action = func(sv *SemanticValues, dt Any) (Any, error) {
		return sv.S, nil
	}
	FACTOR_OPERATOR.Action = func(sv *SemanticValues, dt Any) (Any, error) {
		return sv.S, nil
	}
	NUMBER.Action = func(sv *SemanticValues, dt Any) (Any, error) {
		return strconv.Atoi(sv.S)
	}

	// Parse
	expr := " (1 + 2 * (3 + 4)) / 5 - 6 "
	l, v, err := EXPRESSION.Parse(expr, nil)
	if err != nil {
		t.Errorf("syntax error: pos:%d", l)
	}
	if v != -3 {
		t.Errorf("action error: pos:%d", l)
	}
}
