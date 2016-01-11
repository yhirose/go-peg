package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/pprof"

	"github.com/yhirose/go-peg"
)

var usageMessage = `usage: peglint [-ast] [-trace] [grammar file path] [source file path]

peglint checks syntax of a given PEG grammar file and reports errors. If the check is successful and a user gives a source file for the grammar, it will also check syntax of the source file.

The -ast flag prints the AST (abstract syntax tree) of the source file.

The -trace flag can be used with the source file. It prints names of rules and operators that the PEG parser detects on standard error.
`

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(1)
}

var (
	astFlag   = flag.Bool("ast", false, "show ast")
	traceFlag = flag.Bool("trace", false, "show trace message")
	profFlag  = flag.String("prof", "", "write cpu profile to file")
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func pcheck(perr *peg.Error) {
	if perr != nil {
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

	parser, perr := peg.NewParser(string(dat))
	pcheck(perr)

	if len(args) >= 2 {
		path := args[1]

		var err error
		if path == "-" {
			dat, err = ioutil.ReadAll(os.Stdin)
		} else {
			dat, err = ioutil.ReadFile(path)
		}
		check(err)
		source := string(dat)

		if *traceFlag {
			SetupTracer(parser)
		}

		if *astFlag {
			peg.EnableAst(parser)
		}

		if *profFlag != "" {
			f, err := os.Create(*profFlag)
			check(err)
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		val, perr := parser.ParseAndGetValue(source, nil)
		pcheck(perr)

		if *astFlag {
			ast := val.(*peg.Ast)
			fmt.Println(peg.AstToS(ast, "", 0))
		}
	}
}
