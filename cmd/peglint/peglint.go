package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/pprof"

	"github.com/yhirose/go-peg"
)

var usageMessage = `usage: peglint [-ast] [-opt] [-trace] [-f path] [-s string] [grammar path]

peglint checks syntax of a given PEG grammar file and reports errors. If the check is successful and a user gives a source file for the grammar, it will also check syntax of the source file.

The -ast flag prints the AST (abstract syntax tree) of the source file.

The -opt flag prints the optimized AST (abstract syntax tree) of the source file.

The -trace flag can be used with the source file. It prints names of rules and operators that the PEG parser detects on standard error.

The -f 'path' specifies a file path to the source text.

The -s 'string' specifies the source text.
`

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(1)
}

var (
	astFlag        = flag.Bool("ast", false, "show ast")
	optFlag        = flag.Bool("opt", false, "show optimized ast")
	traceFlag      = flag.Bool("trace", false, "show trace message")
	sourceFilePath = flag.String("f", "", "source file path")
	sourceString   = flag.String("s", "", "source string")
	profPath       = flag.String("prof", "", "write cpu profile to file")
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func pcheck(err error) {
	if perr, ok := err.(*peg.Error); ok {
		for _, d := range perr.Details {
			fmt.Println(d)
		}
		os.Exit(1)
	}
}

func SetupTracer(p *peg.Parser) {
	indent := func(level int) string {
		s := ""
		for level > 0 {
			s = s + "  "
			level--
		}
		return s
	}

	fmt.Println("pos:lev\trule/ope")
	fmt.Println("-------\t--------")

	level := 0
	prevPos := 0

	p.TracerEnter = func(name string, s string, v *peg.Values, d peg.Any, p int) {
		var backtrack string
		if p < prevPos {
			backtrack = "*"
		}
		fmt.Printf("%d:%d%s\t%s%s\n", p, level, backtrack, indent(level), name)
		prevPos = p
		level++
	}

	p.TracerLeave = func(name string, s string, v *peg.Values, d peg.Any, p int, l int) {
		level--
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		usage()
	}

	dat, err := ioutil.ReadFile(args[0])
	check(err)

	parser, err := peg.NewParser(string(dat))
	pcheck(err)

	var source string

	if *sourceFilePath != "" {
		if *sourceFilePath == "-" {
			dat, err := ioutil.ReadAll(os.Stdin)
			check(err)
			source = string(dat)
		} else {
			dat, err := ioutil.ReadFile(*sourceFilePath)
			check(err)
			source = string(dat)
		}
	}

	if *sourceString != "" {
		source = *sourceString
	}

	if len(source) > 0 {
		if *traceFlag {
			SetupTracer(parser)
		}

		if *astFlag || *optFlag {
			parser.EnableAst()
		}

		if *profPath != "" {
			f, err := os.Create(*profPath)
			check(err)
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		val, err := parser.ParseAndGetValue(source, nil)
		pcheck(err)

		if *astFlag || *optFlag {
			ast := val.(*peg.Ast)
			if *optFlag {
				opt := peg.NewAstOptimizer(nil)
				ast = opt.Optimize(ast, nil)
			}
			fmt.Println(ast)
		}
	}
}
