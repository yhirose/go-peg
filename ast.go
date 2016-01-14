package peg

import "strconv"

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

func EnableAst(p *Parser) (err error) {
	for name, rule := range p.Grammar {
		nm := name
		if rule.isToken() {
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
