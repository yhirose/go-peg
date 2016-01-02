go-peg
======

Yet another [PEG](http://en.wikipedia.org/wiki/Parsing_expression_grammar) (Parsing Expression Grammars) parser generator for Go.

```go
package main

import (
	"fmt"
	. "github.com/yhirose/go-peg"
	"strconv"
)

func main() {
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
	g["TERM_OPERATOR"].Action = func(sv *SemanticValues, dt Any) (Any, error) { return sv.S, nil }
	g["FACTOR_OPERATOR"].Action = func(sv *SemanticValues, dt Any) (Any, error) { return sv.S, nil }
	g["NUMBER"].Action = func(sv *SemanticValues, dt Any) (Any, error) { return strconv.Atoi(sv.S) }

	// Parse
	input := " 1 + 2 * 3 * (4 - 5 + 6) / 7 - 8 "
	if val, err := parser.ParseAndGetValue(input, nil); err == nil {
		fmt.Println(val) // -3
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
