package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/yhirose/go-peg"
)

var usageMessage = `usage: peglint [-trace] grammar [source]
`

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(1)
}

var (
	traceFlag = flag.Bool("trace", false, "trace mode")
)

func check(err error) {
	if err != nil {
		fmt.Println(err)
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
	p.TracerBegin = func(name string, s string, v *peg.Values, d peg.Any, p int) {
		var backtrack string
		if p < prevPos {
			backtrack = "*"
		}
		fmt.Printf("%d:%d%s\t%s%s\n", p, level, backtrack, indent(level), name)
		prevPos = p
		level++
	}
	p.TracerEnd = func(name string, s string, v *peg.Values, d peg.Any, l int) {
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
		dat, err := ioutil.ReadFile(args[1])
		check(err)

		if *traceFlag {
			SetupTracer(parser)
		}

		perr = parser.Parse(string(dat), nil)
		pcheck(perr)
	}
}
