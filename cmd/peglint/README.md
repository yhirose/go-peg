peglint
-------

The lint utility for PEG.

```
usage: peglint [-ast] [-opt] [-trace] [-f path] [-s string] [grammar path]
```

peglint checks syntax of a given PEG grammar file and reports errors. If the check is successful and a user gives a source file for the grammar, it will also check syntax of the source file.

The -ast flag prints the AST (abstract syntax tree) of the source file.

The -opt flag prints the optimized AST (abstract syntax tree) of the source file.

The -trace flag can be used with the source file. It prints names of rules and operators that the PEG parser detects on standard error.

The -f 'path' specifies a file path to the source text.

The -s 'string' specifies the source text.
