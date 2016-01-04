go-peg
======

Yet another [PEG](http://en.wikipedia.org/wiki/Parsing_expression_grammar) (Parsing Expression Grammars) parser generator for Go.

```go
package main

import (
	"fmt"
	"strconv"
	. "github.com/yhirose/go-peg"
)

func main() {
	// Create a PEG parser
	parser, _ := NewParser(`
		EXPRESSION       <-  TERM (TERM_OPERATOR TERM)*
		TERM             <-  FACTOR (FACTOR_OPERATOR FACTOR)*
		FACTOR           <-  NUMBER / '(' EXPRESSION ')'
		TERM_OPERATOR    <-  [-+]
		FACTOR_OPERATOR  <-  [/*]
		NUMBER           <-  [0-9]+
		%whitespace      <-  [ \t]*
	`)

	// Setup actions
	reduce := func(v *Values, d Any) (Any, error) {
		val := v.ToInt(0)
		for i := 1; i < v.Len(); i += 2 {
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
	g["TERM_OPERATOR"].Action = func(v *Values, d Any) (Any, error) { return v.S, nil }
	g["FACTOR_OPERATOR"].Action = func(v *Values, d Any) (Any, error) { return v.S, nil }
	g["NUMBER"].Action = func(v *Values, d Any) (Any, error) { return strconv.Atoi(v.S) }

	// Parse
	val, err := parser.ParseAndGetValue(" 1 + 2 * 3 * (4 - 5 + 6) / 7 - 8 ", nil)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(val)
	}
}
```

TODO
----
 * Memoization (Packrat parsing)
 * AST generation

License
-------

MIT license (© 2016 Yuji Hirose)
