go-peg
======

Yet another [PEG](http://en.wikipedia.org/wiki/Parsing_expression_grammar) (Parsing Expression Grammars) parser generator for Go.

### Extended features

 * Token operator: `<` `>`
 * Automatic whitespace skipping: `%whitespace`
 * Keyword boundary assertion: `%keyword`
 * Expression parsing (precedence climbing)
 * AST generation

### Usage

```go
// Create a PEG parser
parser, _ := NewParser(`
    # Grammar for simple calculator...
    EXPR         <-  ATOM (BINOP ATOM)*     # Use expression parsing option
    ATOM         <-  NUMBER / '(' EXPR ')'
    BINOP        <-  < [-+/*] >
    NUMBER       <-  < [0-9]+ >
    %whitespace  <-  [ \t]*
    ---
    # Expression parsing option
    %expr  = EXPR   # Rule for expression parsing
    %binop = L + -  # Precedence level 1
    %binop = L * /  # Precedence level 2
`)

// Setup semantic actions
g := parser.Grammar
g["EXPR"].Action = func(v *Values, d Any) (Any, error) {
    val := v.ToInt(0)
    if v.Len() > 1 {
        ope := v.ToStr(1)
        rhs := v.ToInt(2)
        switch ope {
        case "+": val += rhs
        case "-": val -= rhs
        case "*": val *= rhs
        case "/": val /= rhs
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

fmt.Println(val) // Output: -3
```

TODO
----
 * Memoization (Packrat parsing)

License
-------

MIT license (© 2016 Yuji Hirose)
