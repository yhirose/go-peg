package peg

import (
	"strconv"
	"strings"
)

const (
	assocNone = iota
	assocLeft
	assocRight
)

type Ast struct {
	//Path  string
	Ln     int
	Col    int
	Name   string
	Token  string
	Nodes  []*Ast
	Parent *Ast
}

func AstToS(ast *Ast, s string, level int) string {
	for i := 0; i < level; i++ {
		s = s + "  "
	}
	if len(ast.Token) > 0 {
		s = s + "- " + ast.Name + " (" + strconv.Quote(ast.Token) + ")\n"
	} else {
		s = s + "+ " + ast.Name + "\n"
	}
	for _, node := range ast.Nodes {
		s = AstToS(node, s, level+1)
	}
	return s
}

type BinOpeInfo map[string]struct {
	level int
	assoc int
}

func parseExpr(v *Values, i int, minPrec int, ln int, col int, nm string, bopinf BinOpeInfo) (*Ast, int) {
	ast := v.Vs[i].(*Ast)
	i++

	for i < len(v.Vs) {
		ope := v.Vs[i].(*Ast)
		inf, ok := bopinf[ope.Token]
		if !ok || inf.level < minPrec {
			break
		}

		i++

		nextMinPrec := inf.level
		if inf.assoc == assocLeft {
			nextMinPrec = inf.level + 1
		}

		var rhs *Ast
		rhs, i = parseExpr(v, i, nextMinPrec, ln, col, nm, bopinf)

		nodes := []*Ast{ast, ope, rhs}
		ast = &Ast{Ln: ln, Col: col, Name: nm, Nodes: nodes}
		for _, node := range nodes {
			node.Parent = ast
		}
	}

	return ast, i
}

func EnableAst(p *Parser) (err error) {
	// Expression rule
	exprRule := ""
	if vs, ok := p.Options["%expr"]; ok {
		exprRule = vs[0]
		// TODO: error handling
	}

	// Binary operator info
	binOpeInfo := make(BinOpeInfo)
	if vs, ok := p.Options["%binop"]; ok {
		level := len(vs)
		for _, s := range vs {
			flds := strings.Split(s, " ")
			// TODO: error handling
			assoc := assocNone
			for i, fld := range flds {
				switch i {
				case 0:
					switch fld {
					case "L":
						assoc = assocLeft
					case "R":
						assoc = assocRight
					default:
						// TODO: error handling
					}
				default:
					binOpeInfo[fld] = struct {
						level int
						assoc int
					}{level, assoc}
				}
			}
			level--
		}
	}

	// Setup actions for generating AST
	for name, rule := range p.Grammar {
		nm := name
		if exprRule == name {
			rule.Action = func(v *Values, d Any) (Any, error) {
				ln, col := lineInfo(v.SS, v.Pos)
				ast, _ := parseExpr(v, 0, 0, ln, col, nm, binOpeInfo)
				return ast, nil
			}
		} else if rule.isToken() {
			rule.Action = func(v *Values, d Any) (Any, error) {
				ln, col := lineInfo(v.SS, v.Pos)
				ast := &Ast{Ln: ln, Col: col, Name: nm, Token: v.Token()}
				return ast, nil
			}
		} else {
			rule.Action = func(v *Values, d Any) (Any, error) {
				ln, col := lineInfo(v.SS, v.Pos)

				var nodes []*Ast
				for _, val := range v.Vs {
					nodes = append(nodes, val.(*Ast))
				}

				ast := &Ast{Ln: ln, Col: col, Name: nm, Nodes: nodes}
				for _, node := range nodes {
					node.Parent = ast
				}

				return ast, nil
			}
		}
	}

	return err
}
