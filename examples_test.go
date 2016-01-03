package peg_test

import (
	"fmt"

	. "github.com/yhirose/go-peg"
)

func Example() {
	// Create a PEG parser
	parser, _ := NewParser(`
        # Grammar for simple calculator...
        EXPRESSION       <-  _ TERM (TERM_OPERATOR TERM)*
        TERM             <-  FACTOR (FACTOR_OPERATOR FACTOR)*
        FACTOR           <-  NUMBER / '(' _ EXPRESSION ')' _
        TERM_OPERATOR    <-  < [-+] > _
        FACTOR_OPERATOR  <-  < [/*] > _
        NUMBER           <-  < [0-9]+ > _
		~_               <-  [ \t]*
    `)

	// Setup actions
	reduce := func(sv *SemanticValues, dt Any) (Any, error) {
		val := sv.ToInt(0)
		for i := 1; i < len(sv.Vs); i += 2 {
			num := sv.ToInt(i + 1)
			switch sv.ToStr(i) {
			case "+":
				val += num
			case "-":
				val -= num
			case "*":
				val *= num
			case "/":
				val /= num
			}
		}
		return val, nil
	}

	g := parser.Grammar
	g["EXPRESSION"].Action = reduce
	g["TERM"].Action = reduce
	g["TERM_OPERATOR"].Action = ActionToStr
	g["FACTOR_OPERATOR"].Action = ActionToStr
	g["NUMBER"].Action = ActionToInt

	// Parse
	input := " 1 + 2 * 3 * (4 - 5 + 6) / 7 - 8 "
	val, _ := parser.ParseAndGetValue(input, nil)

	fmt.Println(val)
	// Output: -3
}

func Example_combinators() {
	// Grammar
	var EXPRESSION, TERM, FACTOR, TERM_OPERATOR, FACTOR_OPERATOR, NUMBER Rule

	EXPRESSION.Ope = Seq(&TERM, Zom(Seq(&TERM_OPERATOR, &TERM)))
	TERM.Ope = Seq(&FACTOR, Zom(Seq(&FACTOR_OPERATOR, &FACTOR)))
	FACTOR.Ope = Cho(&NUMBER, Seq(Lit("("), &EXPRESSION, Lit(")")))
	TERM_OPERATOR.Ope = Seq(Tok(Cls("-+")))
	FACTOR_OPERATOR.Ope = Seq(Tok(Cls("/*")))
	NUMBER.Ope = Seq(Tok(Oom(Cls("0-9"))))

	EXPRESSION.WhitespaceOpe = Zom(Cls(" \t"))

	// Actions
	reduce := func(sv *SemanticValues, dt Any) (Any, error) {
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

	EXPRESSION.Action = reduce
	TERM.Action = reduce
	TERM_OPERATOR.Action = ActionToStr
	FACTOR_OPERATOR.Action = ActionToStr
	NUMBER.Action = ActionToInt

	// Parse
	l, v, _ := EXPRESSION.Parse(" (1 + 2 * (3 + 4)) / 5 - 6 ", nil)

	fmt.Println(l)
	fmt.Println(v)
	// Output:
	// 27
	// -3
}

func Example_whitespace() {
	// Create a PEG parser
	parser, _ := NewParser(`
        # Grammar for simple calculator...
        EXPRESSION       <-  TERM (TERM_OPERATOR TERM)*
        TERM             <-  FACTOR (FACTOR_OPERATOR FACTOR)*
        FACTOR           <-  NUMBER / '(' EXPRESSION ')'
        TERM_OPERATOR    <-  [-+]
        FACTOR_OPERATOR  <-  [/*]
        NUMBER           <-  [0-9]+
		%whitespace      <-  [ \t]*
    `)

	// Setup actions
	reduce := func(sv *SemanticValues, dt Any) (Any, error) {
		val := sv.ToInt(0)
		for i := 1; i < len(sv.Vs); i += 2 {
			num := sv.ToInt(i + 1)
			switch sv.ToStr(i) {
			case "+":
				val += num
			case "-":
				val -= num
			case "*":
				val *= num
			case "/":
				val /= num
			}
		}
		return val, nil
	}

	g := parser.Grammar
	g["EXPRESSION"].Action = reduce
	g["TERM"].Action = reduce
	g["TERM_OPERATOR"].Action = ActionToStr
	g["FACTOR_OPERATOR"].Action = ActionToStr
	g["NUMBER"].Action = ActionToInt

	// Parse
	input := " 1 + 2 * 3 * (4 - 5 + 6) / 7 - 8 "
	val, _ := parser.ParseAndGetValue(input, nil)

	fmt.Println(val)
	// Output: -3
}
