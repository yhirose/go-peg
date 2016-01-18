package peg_test

import (
	"fmt"
	"strconv"

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
	reduce := func(v *Values, d Any) (Any, error) {
		val := v.ToInt(0)
		for i := 1; i < len(v.Vs); i += 2 {
			num := v.ToInt(i + 1)
			switch v.ToStr(i) {
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
	g["TERM_OPERATOR"].Action = func(v *Values, d Any) (Any, error) { return v.Token(), nil }
	g["FACTOR_OPERATOR"].Action = func(v *Values, d Any) (Any, error) { return v.Token(), nil }
	g["NUMBER"].Action = func(v *Values, d Any) (Any, error) { return strconv.Atoi(v.Token()) }

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
	reduce := func(v *Values, d Any) (Any, error) {
		ret := v.ToInt(0)
		for i := 1; i < len(v.Vs); i += 2 {
			ope := v.ToStr(i)
			n := v.ToInt(i + 1)
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
	TERM_OPERATOR.Action = func(v *Values, d Any) (Any, error) { return v.Token(), nil }
	FACTOR_OPERATOR.Action = func(v *Values, d Any) (Any, error) { return v.Token(), nil }
	NUMBER.Action = func(v *Values, d Any) (Any, error) { return strconv.Atoi(v.Token()) }

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
        TERM_OPERATOR    <-  < [-+] >
        FACTOR_OPERATOR  <-  < [/*] >
        NUMBER           <-  < [0-9]+ >
		%whitespace      <-  [ \t]*
    `)

	// Setup actions
	reduce := func(v *Values, d Any) (Any, error) {
		val := v.ToInt(0)
		for i := 1; i < len(v.Vs); i += 2 {
			num := v.ToInt(i + 1)
			switch v.ToStr(i) {
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
	g["TERM_OPERATOR"].Action = func(v *Values, d Any) (Any, error) { return v.Token(), nil }
	g["FACTOR_OPERATOR"].Action = func(v *Values, d Any) (Any, error) { return v.Token(), nil }
	g["NUMBER"].Action = func(v *Values, d Any) (Any, error) { return strconv.Atoi(v.Token()) }

	// Parse
	input := " 1 + 2 * 3 * (4 - 5 + 6) / 7 - 8 "
	val, _ := parser.ParseAndGetValue(input, nil)

	fmt.Println(val)
	// Output: -3
}

func Example_expressionParsing() {
	// Create a PEG parser
	parser, _ := NewParser(`
        # Grammar for simple calculator...
        EXPRESSION   <-  ATOM (BINOP ATOM)*
        ATOM         <-  NUMBER / '(' EXPRESSION ')'
        BINOP        <-  < [-+/*] >
        NUMBER       <-  < [0-9]+ >
		%whitespace  <-  [ \t]*
		---
        # Expression parsing
		%expr  = EXPRESSION
		%binop = L + -  # level 1
		%binop = L * /  # level 2
    `)

	// Setup actions
	g := parser.Grammar
	g["EXPRESSION"].Action = func(v *Values, d Any) (Any, error) {
		val := v.ToInt(0)
		if v.Len() > 1 {
			rhs := v.ToInt(2)
			ope := v.ToStr(1)
			switch ope {
			case "+":
				val += rhs
			case "-":
				val -= rhs
			case "*":
				val *= rhs
			case "/":
				val /= rhs
			}
		}
		return val, nil
	}
	g["BINOP"].Action = func(v *Values, d Any) (Any, error) {
		return v.Token(), nil
	}
	g["NUMBER"].Action = func(v *Values, d Any) (Any, error) {
		return strconv.Atoi(v.Token())
	}

	// Parse
	input := " 1 + 2 * 3 * (4 - 5 + 6) / 7 - 8 "
	val, _ := parser.ParseAndGetValue(input, nil)

	fmt.Println(val)
	// Output: -3
}
