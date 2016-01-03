peglint
-------

The lint utility for PEG.

```
usage: peglint [--trace] [grammar file path] [source file path]

peglint checks syntax of a given PEG grammar file and reports errors. If the check is successful and a user gives a source file for the grammar, it will also check syntax of the source file.

The -trace flag can be used with the source file. It prints names of rules and operators that the PEG parser detects on standard error.
```
