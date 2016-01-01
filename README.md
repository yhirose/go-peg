go-peg
======

Go language [PEG](http://en.wikipedia.org/wiki/Parsing_expression_grammar) (Parsing Expression Grammars) library.

This library is ported from [cpp-pegib](https://github.com/yhirose/cpp-peglib).

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
	g["TERM_OPERATOR"].Action = func(sv *SemanticValues, dt Any) (Any, error) { return sv.S, nil }
	g["FACTOR_OPERATOR"].Action = func(sv *SemanticValues, dt Any) (Any, error) { return sv.S, nil }
	g["NUMBER"].Action = func(sv *SemanticValues, dt Any) (Any, error) { return strconv.Atoi(sv.S) }

	// Parse
	input := " 1 + 2 * 3 * (4 - 5 + 6) / 7 - 8 "
	if val, err := parser.ParseAndGetValue(input, nil); err == nil {
		fmt.Println(val) // -3
	} else {
		fmt.Println(err)
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
