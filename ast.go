package peg

import (
	"fmt"
	"strconv"
)

type Ast struct {
	//Path  string
	Ln     int
	Col    int
	S      string
	Name   string
	Token  string
	Nodes  []*Ast
	Parent *Ast
	Data   interface{}
}

func (ast *Ast) String() string {
	return astToS(ast, "", 0)
}

func astToS(ast *Ast, s string, level int) string {
	for i := 0; i < level; i++ {
		s = s + "  "
	}
	if len(ast.Token) > 0 {
		if ast.Data != nil {
			s = fmt.Sprintf("%s- %s (%s) [%v]\n", s, ast.Name, strconv.Quote(ast.Token), ast.Data)
		} else {
			s = fmt.Sprintf("%s- %s (%s)\n", s, ast.Name, strconv.Quote(ast.Token))
		}
	} else {
		if ast.Data != nil {
			s = fmt.Sprintf("%s+ %s [%v]\n", s, ast.Name, ast.Data)
		} else {
			s = fmt.Sprintf("%s+ %s\n", s, ast.Name)
		}
	}
	for _, node := range ast.Nodes {
		s = astToS(node, s, level+1)
	}
	return s
}

func EnableAst(p *Parser) (err error) {
	for name, rule := range p.Grammar {
		nm := name
		if rule.isToken() {
			rule.Action = func(v *Values, d Any) (Any, error) {
				ln, col := lineInfo(v.SS, v.Pos)
				ast := &Ast{Ln: ln, Col: col, S: v.S, Name: nm, Token: v.Token()}
				return ast, nil
			}
		} else {
			rule.Action = func(v *Values, d Any) (Any, error) {
				ln, col := lineInfo(v.SS, v.Pos)

				var nodes []*Ast
				for _, val := range v.Vs {
					nodes = append(nodes, val.(*Ast))
				}

				ast := &Ast{Ln: ln, Col: col, S: v.S, Name: nm, Nodes: nodes}
				for _, node := range nodes {
					node.Parent = ast
				}

				return ast, nil
			}
		}
	}

	return err
}
