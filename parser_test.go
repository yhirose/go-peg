package peg

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestSimpleSyntax(t *testing.T) {
	_, err := NewParser(`
		ROOT ← _
		_    <- ' '
	`)
	if err != nil {
		t.Error(err)
	}
}

func TestEmptySyntax(t *testing.T) {
	_, err := NewParser("")
	if err == nil {
		t.Error(err)
	}
}

func assert(t *testing.T, ok bool) {
	if ok == false {
		t.Error("error...")
	}
}

func TestStringCapture(t *testing.T) {
	parser, _ := NewParser(`
		ROOT      <-  _ ('[' TAG_NAME ']' _)*
		TAG_NAME  <-  (!']' .)+
		_         <-  [ \t]*
	`)

	var tags []string
	parser.Grammar["TAG_NAME"].Action = func(sv *Values, d Any) (v Any, err error) {
		tags = append(tags, sv.S)
		return
	}

	assert(t, parser.Parse(" [tag1] [tag:2] [tag-3] ", nil) == nil)
	assert(t, len(tags) == 3)
	assert(t, tags[0] == "tag1")
	assert(t, tags[1] == "tag:2")
	assert(t, tags[2] == "tag-3")
}

/*
TEST_CASE("String capture test with match", "[general]")
{
    peg::match m;
    auto ret = peg::peg_match(
        "  ROOT      <-  _ ('[' $< TAG_NAME > ']' _)*  "
        "  TAG_NAME  <-  (!']' .)+                "
        "  _         <-  [ \t]*                   ",
        " [tag1] [tag:2] [tag-3] ",
        m);

    REQUIRE(ret == true);
    REQUIRE(m.size() == 4);
    REQUIRE(m.str(1) == "tag1");
    REQUIRE(m.str(2) == "tag:2");
    REQUIRE(m.str(3) == "tag-3");
}
*/

func TestStringCapture2(t *testing.T) {
	var tags []string

	var ROOT, TAG, TAG_NAME, WS Rule
	ROOT.Ope = Seq(&WS, Zom(&TAG))
	TAG.Ope = Seq(Lit("["), &TAG_NAME, Lit("]"), &WS)
	TAG_NAME.Ope = Oom(Seq(Npd(Lit("]")), Dot()))
	WS.Ope = Zom(Cls(" \t"))

	TAG_NAME.Action = func(sv *Values, d Any) (v Any, err error) {
		tags = append(tags, sv.S)
		return
	}

	_, _, err := ROOT.Parse(" [tag1] [tag:2] [tag-3] ", nil)
	assert(t, err == nil)
	assert(t, len(tags) == 3)
	assert(t, tags[0] == "tag1")
	assert(t, tags[1] == "tag:2")
	assert(t, tags[2] == "tag-3")
}

func TestStringCapture3(t *testing.T) {
	syntax := `
		ROOT  <- _ TOKEN*
		TOKEN <- '[' < (!']' .)+ > ']' _
		_     <- [ \t\r\n]*
	`

	parser, _ := NewParser(syntax)

	var tags []string
	parser.Grammar["TOKEN"].Action = func(sv *Values, d Any) (v Any, err error) {
		tags = append(tags, sv.Token())
		return
	}

	assert(t, parser.Parse(" [tag1] [tag:2] [tag-3] ", nil) == nil)
	assert(t, len(tags) == 3)
	assert(t, tags[0] == "tag1")
	assert(t, tags[1] == "tag:2")
	assert(t, tags[2] == "tag-3")
}

/*
TEST_CASE("Named capture test", "[general]")
{
    peg::match m;

    auto ret = peg::peg_match(
        "  ROOT      <-  _ ('[' $test< TAG_NAME > ']' _)*  "
        "  TAG_NAME  <-  (!']' .)+                "
        "  _         <-  [ \t]*                   ",
        " [tag1] [tag:2] [tag-3] ",
        m);

    auto cap = m.named_capture("test");

    REQUIRE(ret == true);
    REQUIRE(m.size() == 4);
    REQUIRE(cap.size() == 3);
    REQUIRE(m.str(cap[2]) == "tag-3");
}

TEST_CASE("String capture test with embedded match action", "[general]")
{
    Definition ROOT, TAG, TAG_NAME, WS;

    vector<string> tags;

    ROOT     <= seq(WS, zom(TAG));
    TAG      <= seq(chr('['),
                    cap(TAG_NAME, [&](const char* s, size_t n, size_t id, const std::string& name) {
                        tags.push_back(string(s, n));
                    }),
                    chr(']'),
                    WS);
    TAG_NAME <= oom(seq(npd(chr(']')), dot()));
    WS       <= zom(cls(" \t"));

    auto r = ROOT.parse(" [tag1] [tag:2] [tag-3] ");

    REQUIRE(r.ret == true);
    REQUIRE(tags.size() == 3);
    REQUIRE(tags[0] == "tag1");
    REQUIRE(tags[1] == "tag:2");
    REQUIRE(tags[2] == "tag-3");
}
*/

func TestSyclicGrammar(t *testing.T) {
	var PARENT, CHILD Rule
	PARENT.Ope = Seq(&CHILD)
	CHILD.Ope = Seq(&PARENT)
}

/*
TEST_CASE("Visit test", "[general]")
{
    Definition ROOT, TAG, TAG_NAME, WS;

    ROOT     <= seq(WS, zom(TAG));
    TAG      <= seq(chr('['), TAG_NAME, chr(']'), WS);
    TAG_NAME <= oom(seq(npd(chr(']')), dot()));
    WS       <= zom(cls(" \t"));

    AssignIDToDefinition defIds;
    ROOT.accept(defIds);

    REQUIRE(defIds.ids.size() == 4);
}
*/

func TestTokenCheckTest(t *testing.T) {
	parser, _ := NewParser(`
        EXPRESSION       <-  _ TERM (TERM_OPERATOR TERM)*
        TERM             <-  FACTOR (FACTOR_OPERATOR FACTOR)*
        FACTOR           <-  NUMBER / '(' _ EXPRESSION ')' _
        TERM_OPERATOR    <-  < [-+] > _
        FACTOR_OPERATOR  <-  < [/*] > _
        NUMBER           <-  < [0-9]+ > _
        _                <-  [ \t\r\n]*
	`)

	assert(t, parser.Grammar["EXPRESSION"].isToken() == false)
	assert(t, parser.Grammar["FACTOR"].isToken() == false)
	assert(t, parser.Grammar["FACTOR_OPERATOR"].isToken() == true)
	assert(t, parser.Grammar["NUMBER"].isToken() == true)
	assert(t, parser.Grammar["_"].isToken() == true)
}

func TestLambdaAction(t *testing.T) {
	parser, _ := NewParser(`
       START <- (CHAR)*
       CHAR  <- .
	`)

	var ss string
	parser.Grammar["CHAR"].Action = func(sv *Values, d Any) (v Any, err error) {
		ss += sv.S
		return
	}

	assert(t, parser.Parse("hello", nil) == nil)
	assert(t, ss == "hello")
}

func TestEnterExitHandlers(t *testing.T) {
	parser, _ := NewParser(`
        START  <- LTOKEN '=' RTOKEN
        LTOKEN <- TOKEN
        RTOKEN <- TOKEN
        TOKEN  <- [A-Za-z]+
	`)

	parser.Grammar["LTOKEN"].Enter = func(d Any) {
		*d.(*bool) = false
	}
	parser.Grammar["LTOKEN"].Leave = func(d Any) {
		*d.(*bool) = true
	}

	msg := "should be upper case string..."

	parser.Grammar["TOKEN"].Action = func(sv *Values, d Any) (v Any, err error) {
		if *d.(*bool) {
			if sv.S != strings.ToUpper(sv.S) {
				err = errors.New(msg)
			}
		}
		return
	}

	requireUpperCase := false
	var d Any = &requireUpperCase
	assert(t, parser.Parse("hello=world", d) != nil)
	assert(t, parser.Parse("HELLO=world", d) != nil)
	assert(t, parser.Parse("hello=WORLD", d) == nil)
	assert(t, parser.Parse("HELLO=WORLD", d) == nil)

	err := parser.Parse("hello=world", d)
	assert(t, err.Details[0].Ln == 1)
	assert(t, err.Details[0].Col == 7)
	assert(t, err.Details[0].Msg == msg)
}

func TestWhitespace(t *testing.T) {
	parser, _ := NewParser(`
        # Rules
        ROOT         <-  ITEM (',' ITEM)*
        ITEM         <-  WORD / PHRASE

        # Tokens
        WORD         <-  < [a-zA-Z0-9_]+ >
        PHRASE       <-  < '"' (!'"' .)* '"' >

        %whitespace  <-  [ \t\r\n]*
	`)

	err := parser.Parse(`  one, 	 "two, three",   four  `, nil)
	assert(t, err == nil)
}

func TestWhitespace2(t *testing.T) {
	parser, _ := NewParser(`
        # Rules
        ROOT         <-  ITEM (',' ITEM)*
        ITEM         <-  '[' < [a-zA-Z0-9_]+ > ']'

        %whitespace  <-  (SPACE / TAB)*
        SPACE        <-  ' '
        TAB          <-  '\t'
	`)

	var items []string
	parser.Grammar["ITEM"].Action = func(sv *Values, d Any) (v Any, err error) {
		items = append(items, sv.Token())
		return
	}

	err := parser.Parse(`[one], 	[two] ,[three] `, nil)
	assert(t, err == nil)
	assert(t, len(items) == 3)
	assert(t, items[0] == "one")
	assert(t, items[1] == "two")
	assert(t, items[2] == "three")
}

func TestKeywordBounary(t *testing.T) {
	parser, _ := NewParser(`
        ROOT         <-  'hello' ','? 'world'
        %whitespace  <-  [ \t\r\n]*
        %keyword     <-  [a-z]+
	`)

	assert(t, parser.Parse(`helloworld`, nil) != nil)
	assert(t, parser.Parse(`hello world`, nil) == nil)
	assert(t, parser.Parse(`hello,world`, nil) == nil)
	assert(t, parser.Parse(`hello, world`, nil) == nil)
	assert(t, parser.Parse(`hello , world`, nil) == nil)
}

func TestSkipToken(t *testing.T) {
	parser, _ := NewParser(`
        ROOT  <-  _ ITEM (',' _ ITEM _)*
        ITEM  <-  ([a-z0-9])+
        ~_    <-  [ \t]*
	`)

	parser.Grammar["ROOT"].Action = func(sv *Values, d Any) (v Any, err error) {
		assert(t, len(sv.Vs) == 2)
		return
	}

	assert(t, parser.Parse(" item1, item2 ", nil) == nil)
}

func TestSkipToken2(t *testing.T) {
	parser, _ := NewParser(`
        ROOT        <-  ITEM (',' ITEM)*
        ITEM        <-  < ([a-z0-9])+ >
        %whitespace <-  [ \t]*
	`)

	parser.Grammar["ROOT"].Action = func(sv *Values, d Any) (v Any, err error) {
		assert(t, len(sv.Vs) == 2)
		return
	}

	assert(t, parser.Parse(" item1, item2 ", nil) == nil)
}

/*
TEST_CASE("Backtracking test", "[general]")
{
    parser parser(
       "  START <- PAT1 / PAT2  "
       "  PAT1  <- HELLO ' One' "
       "  PAT2  <- HELLO ' Two' "
       "  HELLO <- 'Hello'      "
    );

    size_t count = 0;
    parser["HELLO"] = [&](const SemanticValues& sv) {
        count++;
    };

    parser.enable_packrat_parsing();

    bool ret = parser.parse("Hello Two");
    REQUIRE(ret == true);
    REQUIRE(count == 1); // Skip second time
}
*/

func TestBacktrackingWithAst(t *testing.T) {
	parser, _ := NewParser(`
        S <- A? B (A B)* A
        A <- 'a'
        B <- 'b'
    `)

	parser.EnableAst()
	val, err := parser.ParseAndGetValue("ba", nil)
	ast := val.(*Ast)

	assert(t, err == nil)
	assert(t, len(ast.Nodes) == 2)
}

func TestOctalHexValue(t *testing.T) {
	parser, _ := NewParser(`
        ROOT <- '\132\x7a'
    `)

	assert(t, parser.Parse("Zz", nil) == nil)
}

func TestSimpleCalculator(t *testing.T) {
	parser, _ := NewParser(`
        Additive  <- Multitive '+' Additive / Multitive
        Multitive <- Primary '*' Multitive / Primary
        Primary   <- '(' Additive ')' / Number
        Number    <- [0-9]+
    `)

	parser.Grammar["Additive"].Action = func(sv *Values, d Any) (v Any, err error) {
		switch sv.Choice {
		case 0:
			v = sv.ToInt(0) + sv.ToInt(1)
		default:
			v = sv.ToInt(0)
		}
		return
	}

	parser.Grammar["Multitive"].Action = func(sv *Values, d Any) (v Any, err error) {
		switch sv.Choice {
		case 0:
			v = sv.ToInt(0) * sv.ToInt(1)
		default:
			v = sv.ToInt(0)
		}
		return
	}

	parser.Grammar["Number"].Action = func(sv *Values, d Any) (v Any, err error) {
		return strconv.Atoi(sv.S)
	}

	val, err := parser.ParseAndGetValue("(1+2)*3", nil)

	assert(t, err == nil)
	assert(t, val == 9)
}

func TestCalculator(t *testing.T) {
	// Construct grammer
	var EXPRESSION, TERM, FACTOR, TERM_OPERATOR, FACTOR_OPERATOR, NUMBER Rule

	EXPRESSION.Ope = Seq(&TERM, Zom(Seq(&TERM_OPERATOR, &TERM)))
	TERM.Ope = Seq(&FACTOR, Zom(Seq(&FACTOR_OPERATOR, &FACTOR)))
	FACTOR.Ope = Cho(&NUMBER, Seq(Lit("("), &EXPRESSION, Lit(")")))
	TERM_OPERATOR.Ope = Cls("+-")
	FACTOR_OPERATOR.Ope = Cls("/*")
	NUMBER.Ope = Oom(Cls("0-9"))

	// Setup actions
	reduce := func(sv *Values, d Any) (Any, error) {
		ret := sv.ToInt(0)
		for i := 1; i < len(sv.Vs); i += 2 {
			num := sv.ToInt(i + 1)
			ope := sv.ToStr(i)
			switch ope {
			case "+":
				ret += num
			case "-":
				ret -= num
			case "*":
				ret *= num
			case "/":
				ret /= num
			}
		}
		return ret, nil
	}

	EXPRESSION.Action = reduce
	TERM.Action = reduce
	TERM_OPERATOR.Action = func(sv *Values, d Any) (v Any, err error) { return sv.S, nil }
	FACTOR_OPERATOR.Action = func(sv *Values, d Any) (v Any, err error) { return sv.S, nil }
	NUMBER.Action = func(sv *Values, d Any) (v Any, err error) { return strconv.Atoi(sv.S) }

	// Parse
	_, val, err := EXPRESSION.Parse("1+2*3*(4-5+6)/7-8", nil)

	assert(t, err == nil)
	assert(t, val == -3)
}

func TestCalculator2(t *testing.T) {
	parser, _ := NewParser(`
        # Grammar for Calculator...
        EXPRESSION       <-  TERM (TERM_OPERATOR TERM)*
        TERM             <-  FACTOR (FACTOR_OPERATOR FACTOR)*
        FACTOR           <-  NUMBER / '(' EXPRESSION ')'
        TERM_OPERATOR    <-  [-+]
        FACTOR_OPERATOR  <-  [/*]
        NUMBER           <-  [0-9]+
    `)

	// Setup actions
	reduce := func(sv *Values, d Any) (Any, error) {
		ret := sv.ToInt(0)
		for i := 1; i < len(sv.Vs); i += 2 {
			num := sv.ToInt(i + 1)
			ope := sv.ToStr(i)
			switch ope {
			case "+":
				ret += num
			case "-":
				ret -= num
			case "*":
				ret *= num
			case "/":
				ret /= num
			}
		}
		return ret, nil
	}

	g := parser.Grammar
	g["EXPRESSION"].Action = reduce
	g["TERM"].Action = reduce
	g["TERM_OPERATOR"].Action = func(sv *Values, d Any) (Any, error) { return sv.S, nil }
	g["FACTOR_OPERATOR"].Action = func(sv *Values, d Any) (Any, error) { return sv.S, nil }
	g["NUMBER"].Action = func(sv *Values, d Any) (Any, error) { return strconv.Atoi(sv.S) }

	// Parse
	val, err := parser.ParseAndGetValue("1+2*3*(4-5+6)/7-8", nil)

	assert(t, err == nil)
	assert(t, val == -3)
}

func TestCalculator3(t *testing.T) {
	parser, _ := NewParser(`
        # Grammar for simple calculator...
        EXPRESSION   <-  ATOM (BINOP ATOM)*
        ATOM         <-  NUMBER / '(' EXPRESSION ')'
        BINOP        <-  < [-+/*] >
        NUMBER       <-  < [0-9]+ >
		%whitespace  <-  [ \t]*
		---
        # Expression parsing
		%expr  = EXPRESSION # rule
		%binop = L + -      # level 1
		%binop = L * /      # level 2
    `)

	// Setup actions
	g := parser.Grammar
	g["EXPRESSION"].Action = func(v *Values, d Any) (Any, error) {
		val := v.ToInt(0)
		if v.Len() > 1 {
			rhs := v.ToInt(2)
			ope := v.ToStr(1)
			switch ope {
			case "+":
				val += rhs
			case "-":
				val -= rhs
			case "*":
				val *= rhs
			case "/":
				val /= rhs
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
	val, err := parser.ParseAndGetValue("1+2*3*(4-5+6)/7-8", nil)

	assert(t, err == nil)
	assert(t, val == -3)

	val, err = parser.ParseAndGetValue(" 1 + 1 + 1 ", nil)

	assert(t, err == nil)
	assert(t, val == 3)
}

func TestCalculatorTestWithAST(t *testing.T) {
	parser, _ := NewParser(`
        EXPRESSION       <-  _ TERM (TERM_OPERATOR TERM)*
        TERM             <-  FACTOR (FACTOR_OPERATOR FACTOR)*
        FACTOR           <-  NUMBER / '(' _ EXPRESSION ')' _
        TERM_OPERATOR    <-  < [-+] > _
        FACTOR_OPERATOR  <-  < [/*] > _
        NUMBER           <-  < [0-9]+ > _
        ~_               <-  [ \t\r\n]*
    `)

	var eval func(ast *Ast) int
	eval = func(ast *Ast) int {
		if ast.Name == "NUMBER" {
			val, _ := strconv.Atoi(ast.Token)
			return val
		} else {
			nodes := ast.Nodes
			result := eval(nodes[0])
			for i := 1; i < len(nodes); i += 2 {
				num := eval(nodes[i+1])
				ope := nodes[i].Token[0]
				switch ope {
				case '+':
					result += num
					break
				case '-':
					result -= num
					break
				case '*':
					result *= num
					break
				case '/':
					result /= num
					break
				}
			}
			return result
		}
	}

	parser.EnableAst()
	val, err := parser.ParseAndGetValue("1+2*3*(4-5+6)/7-8", nil)

	ast := val.(*Ast)
	opt := NewAstOptimizer(nil)
	ast = opt.Optimize(ast, nil)
	ret := eval(ast)

	assert(t, err == nil)
	assert(t, ret == -3)
}

func TestIgnoreSemanticValue(t *testing.T) {
	parser, _ := NewParser(`
		START <-  ~HELLO WORLD
		HELLO <- 'Hello' _
		WORLD <- 'World' _
		_     <- [ \t\r\n]*
    `)

	parser.EnableAst()
	ast, err := parser.ParseAndGetAst("Hello World", nil)

	assert(t, err == nil)
	assert(t, len(ast.Nodes) == 1)
	assert(t, ast.Nodes[0].Name == "WORLD")
}

func TestIgnoreSemanticValueOfORPredicate(t *testing.T) {
	parser, _ := NewParser(`
		START       <- _ !DUMMY HELLO_WORLD '.'
		HELLO_WORLD <- HELLO 'World' _
		HELLO       <- 'Hello' _
		DUMMY       <- 'dummy' _
		~_          <- [ \t\r\n]*
    `)

	parser.EnableAst()
	ast, err := parser.ParseAndGetAst("Hello World.", nil)

	assert(t, err == nil)
	assert(t, len(ast.Nodes) == 1)
	assert(t, ast.Nodes[0].Name == "HELLO_WORLD")
}

func TestIgnoreSemanticValueOfANDPredicate(t *testing.T) {
	parser, _ := NewParser(`
		START       <- _ &HELLO HELLO_WORLD '.'
		HELLO_WORLD <- HELLO 'World' _
		HELLO       <- 'Hello' _
		~_          <- [ \t\r\n]*
    `)

	parser.EnableAst()
	ast, err := parser.ParseAndGetAst("Hello World.", nil)

	assert(t, err == nil)
	assert(t, len(ast.Nodes) == 1)
	assert(t, ast.Nodes[0].Name == "HELLO_WORLD")
}

func TestLiteralTokenOnAst1(t *testing.T) {
	parser, _ := NewParser(`
        STRING_LITERAL  <- '"' (('\\"' / '\\t' / '\\n') / (!["] .))* '"'
    `)

	parser.EnableAst()
	ast, err := parser.ParseAndGetAst(`"a\tb"`, nil)

	assert(t, err == nil)
	assert(t, ast.Token == `"a\tb"`)
	assert(t, len(ast.Nodes) == 0)
}

func TestLiteralTokenOnAst2(t *testing.T) {
	parser, _ := NewParser(`
        STRING_LITERAL  <-  '"' (ESC / CHAR)* '"'
        ESC             <-  ('\\"' / '\\t' / '\\n')
        CHAR            <-  (!["] .)
    `)

	parser.EnableAst()
	ast, err := parser.ParseAndGetAst(`"a\tb"`, nil)

	assert(t, err == nil)
	assert(t, ast.Token == "")
	assert(t, len(ast.Nodes) == 3)
}

func TestLiteralTokenOnAst3(t *testing.T) {
	parser, _ := NewParser(`
        STRING_LITERAL  <-  < '"' (ESC / CHAR)* '"' >
        ESC             <-  ('\\"' / '\\t' / '\\n')
        CHAR            <-  (!["] .)
    `)

	parser.EnableAst()
	ast, err := parser.ParseAndGetAst(`"a\tb"`, nil)

	assert(t, err == nil)
	assert(t, ast.Token == `"a\tb"`)
	assert(t, len(ast.Nodes) == 0)
}

func TestMissingDefinitions(t *testing.T) {
	parser, err := NewParser(`
        A <- B C
    `)

	assert(t, parser == nil)
	assert(t, err != nil)
}

func TestDefinitionDuplicates(t *testing.T) {
	parser, err := NewParser(`
        A <- ''
        A <- ''
    `)

	assert(t, parser == nil)
	assert(t, err != nil)
}

func TestLeftRecursive(t *testing.T) {
	parser, err := NewParser(`
        A <- A 'a'
        B <- A 'a'
    `)

	assert(t, parser == nil)
	assert(t, err != nil)
}

func TestLeftRecursiveWithOption(t *testing.T) {
	parser, err := NewParser(`
        A  <- 'a' / 'b'? B 'c'
        B  <- A
    `)

	assert(t, parser == nil)
	assert(t, err != nil)
}

func TestLeftRecursiveWithZom(t *testing.T) {
	parser, err := NewParser(`
		A <- 'a'* A*
	`)

	assert(t, parser == nil)
	assert(t, err != nil)
}

func TestLeftRecursiveWithEmptyString(t *testing.T) {
	parser, err := NewParser(`
        " A <- '' A"
    `)

	assert(t, parser == nil)
	assert(t, err != nil)
}

func TestUserRule(t *testing.T) {
	syntax := " ROOT <- _ 'Hello' _ NAME '!' _ "

	rules := map[string]operator{
		"NAME": Usr(func(s string, p int, sv *Values, d Any) int {
			names := []string{"PEG", "BNF"}
			for _, name := range names {
				if len(name) <= len(s)-p && name == s[p:p+len(name)] {
					return len(name)
				}
			}
			return -1
		}),
		"~_": Zom(Cls(" \t\r\n")),
	}

	parser, err := NewParserWithUserRules(syntax, rules)
	assert(t, err == nil)
	assert(t, parser.Parse(" Hello BNF! ", nil) == nil)
}

func TestSemanticPredicate(t *testing.T) {
	parser, _ := NewParser("NUMBER  <-  [0-9]+")

	parser.Grammar["NUMBER"].Action = func(sv *Values, d Any) (val Any, err error) {
		val, _ = strconv.Atoi(sv.S)
		if val != 100 {
			err = errors.New("value error!!")
		}
		return
	}

	val, err := parser.ParseAndGetValue("100", nil)
	assert(t, err == nil)
	assert(t, val == 100)

	val, err = parser.ParseAndGetValue("200", nil)
	assert(t, err != nil)
}

func TestJapaneseCharacter(t *testing.T) {
	parser, _ := NewParser(`
        文 <- 修飾語? 主語 述語 '。'
        主語 <- 名詞 助詞
        述語 <- 動詞 助詞
        修飾語 <- 形容詞
        名詞 <- 'サーバー' / 'クライアント'
        形容詞 <- '古い' / '新しい'
        動詞 <- '落ち' / '復旧し'
        助詞 <- 'が' / 'を' / 'た' / 'ます' / 'に'
	`)

	assert(t, parser.Parse("サーバーを復旧します。", nil) == nil)
}

func match(t *testing.T, r *Rule, s string, want bool) {
	l, _, err := r.Parse(s, newData())
	ok := err == nil
	if ok != want {
		t.Errorf("syntax error: %d", l)
	}
}

func TestPegGrammar(t *testing.T) {
	match(t, &rStart, " Definition <- a / ( b c ) / d \n rule2 <- [a-zA-Z][a-z0-9-]+ ", true)
}

func TestPegDefinition(t *testing.T) {
	match(t, &rDefinition, "Definition <- a / (b c) / d ", true)
	match(t, &rDefinition, "Definition <- a / b c / d ", true)
	match(t, &rDefinition, "Definition ", false)
	match(t, &rDefinition, " ", false)
	match(t, &rDefinition, "", false)
	match(t, &rDefinition, "Definition = a / (b c) / d ", false)
}

func TestPegExpression(t *testing.T) {
	match(t, &rExpression, "a / (b c) / d ", true)
	match(t, &rExpression, "a / b c / d ", true)
	match(t, &rExpression, "a b ", true)
	match(t, &rExpression, "", true)
	match(t, &rExpression, " ", false)
	match(t, &rExpression, " a b ", false)
}

func TestPegSequence(t *testing.T) {
	match(t, &rSequence, "a b c d ", true)
	match(t, &rSequence, "", true)
	match(t, &rSequence, "!", false)
	match(t, &rSequence, "<-", false)
	match(t, &rSequence, " a", false)
}

func TestPegPrefix(t *testing.T) {
	match(t, &rPrefix, "&[a]", true)
	match(t, &rPrefix, "![']", true)
	match(t, &rPrefix, "-[']", false)
	match(t, &rPrefix, "", false)
	match(t, &rSequence, " a", false)
}

func TestPegSuffix(t *testing.T) {
	match(t, &rSuffix, "aaa ", true)
	match(t, &rSuffix, "aaa? ", true)
	match(t, &rSuffix, "aaa* ", true)
	match(t, &rSuffix, "aaa+ ", true)
	match(t, &rSuffix, ". + ", true)
	match(t, &rSuffix, "?", false)
	match(t, &rSuffix, "", false)
	match(t, &rSequence, " a", false)
}

func TestPegPrimary(t *testing.T) {
	match(t, &rPrimary, "_Identifier0_ ", true)
	match(t, &rPrimary, "_Identifier0_<-", false)
	match(t, &rPrimary, "( _Identifier0_ _Identifier1_ )", true)
	match(t, &rPrimary, "'Literal String'", true)
	match(t, &rPrimary, "\"Literal String\"", true)
	match(t, &rPrimary, "[a-zA-Z]", true)
	match(t, &rPrimary, ".", true)
	match(t, &rPrimary, "", false)
	match(t, &rPrimary, " ", false)
	match(t, &rPrimary, " a", false)
	match(t, &rPrimary, "", false)
}

func TestPegIdentifier(t *testing.T) {
	match(t, &rIdentifier, "_Identifier0_ ", true)
	match(t, &rIdentifier, "0Identifier_ ", false)
	match(t, &rIdentifier, "Iden|t ", false)
	match(t, &rIdentifier, " ", false)
	match(t, &rIdentifier, " a", false)
	match(t, &rIdentifier, "", false)
}

func TestPegIdentStart(t *testing.T) {
	match(t, &rIdentStart, "_", true)
	match(t, &rIdentStart, "a", true)
	match(t, &rIdentStart, "Z", true)
	match(t, &rIdentStart, "", false)
	match(t, &rIdentStart, " ", false)
	match(t, &rIdentStart, "0", false)
}

func TestPegIdentRest(t *testing.T) {
	match(t, &rIdentRest, "_", true)
	match(t, &rIdentRest, "a", true)
	match(t, &rIdentRest, "Z", true)
	match(t, &rIdentRest, "", false)
	match(t, &rIdentRest, " ", false)
	match(t, &rIdentRest, "0", true)
}

func TestPegLiteral(t *testing.T) {
	match(t, &rLiteral, "'abc' ", true)
	match(t, &rLiteral, "'a\\nb\\tc' ", true)
	match(t, &rLiteral, "'a\\277\tc' ", true)
	match(t, &rLiteral, "'a\\77\tc' ", true)
	match(t, &rLiteral, "'a\\80\tc' ", false)
	match(t, &rLiteral, "'\n' ", true)
	match(t, &rLiteral, "'a\\'b' ", true)
	match(t, &rLiteral, "'a'b' ", false)
	match(t, &rLiteral, "'a\"'b' ", false)
	match(t, &rLiteral, "\"'\\\"abc\\\"'\" ", true)
	match(t, &rLiteral, "\"'\"abc\"'\" ", false)
	match(t, &rLiteral, "abc", false)
	match(t, &rLiteral, "", false)
	match(t, &rLiteral, "日本語", false)
}

func TestPegClass(t *testing.T) {
	match(t, &rClass, "[]", true)
	match(t, &rClass, "[a]", true)
	match(t, &rClass, "[a-z]", true)
	match(t, &rClass, "[az]", true)
	match(t, &rClass, "[a-zA-Z-]", true)
	match(t, &rClass, "[a-zA-Z-0-9]", true)
	match(t, &rClass, "[a-]", false)
	match(t, &rClass, "[-a]", true)
	match(t, &rClass, "[", false)
	match(t, &rClass, "[a", false)
	match(t, &rClass, "]", false)
	match(t, &rClass, "a]", false)
	match(t, &rClass, "あ-ん", false)
	match(t, &rClass, "[-+]", true)
	match(t, &rClass, "[+-]", false)
}

func TestPegRange(t *testing.T) {
	match(t, &rRange, "a", true)
	match(t, &rRange, "a-z", true)
	match(t, &rRange, "az", false)
	match(t, &rRange, "", false)
	match(t, &rRange, "a-", false)
	match(t, &rRange, "-a", false)
}

func TestPegChar(t *testing.T) {
	match(t, &rChar, "\\n", true)
	match(t, &rChar, "\\r", true)
	match(t, &rChar, "\\t", true)
	match(t, &rChar, "\\'", true)
	match(t, &rChar, "\\\"", true)
	match(t, &rChar, "\\[", true)
	match(t, &rChar, "\\]", true)
	match(t, &rChar, "\\\\", true)
	match(t, &rChar, "\\000", true)
	match(t, &rChar, "\\377", true)
	match(t, &rChar, "\\477", false)
	match(t, &rChar, "\\087", false)
	match(t, &rChar, "\\079", false)
	match(t, &rChar, "\\00", true)
	match(t, &rChar, "\\77", true)
	match(t, &rChar, "\\80", false)
	match(t, &rChar, "\\08", false)
	match(t, &rChar, "\\0", true)
	match(t, &rChar, "\\7", true)
	match(t, &rChar, "\\8", false)
	match(t, &rChar, "a", true)
	match(t, &rChar, ".", true)
	match(t, &rChar, "0", true)
	match(t, &rChar, "\\", false)
	match(t, &rChar, " ", true)
	match(t, &rChar, "  ", false)
	match(t, &rChar, "", false)
	match(t, &rChar, "あ", false)
}

func TestPegOperators(t *testing.T) {
	match(t, &rLEFTARROW, "<-", true)
	match(t, &rSLASH, "/ ", true)
	match(t, &rAND, "& ", true)
	match(t, &rNOT, "! ", true)
	match(t, &rQUESTION, "? ", true)
	match(t, &rSTAR, "* ", true)
	match(t, &rPLUS, "+ ", true)
	match(t, &rOPEN, "( ", true)
	match(t, &rCLOSE, ") ", true)
	match(t, &rDOT, ". ", true)
}

func TestPegComment(t *testing.T) {
	match(t, &rComment, "# Comment.\n", true)
	match(t, &rComment, "# Comment.", false)
	match(t, &rComment, " ", false)
	match(t, &rComment, "a", false)
}

func TestPegSpace(t *testing.T) {
	match(t, &rSpace, " ", true)
	match(t, &rSpace, "\t", true)
	match(t, &rSpace, "\n", true)
	match(t, &rSpace, "", false)
	match(t, &rSpace, "a", false)
}

func TestPegEndOfLine(t *testing.T) {
	match(t, &rEndOfLine, "\r\n", true)
	match(t, &rEndOfLine, "\n", true)
	match(t, &rEndOfLine, "\r", true)
	match(t, &rEndOfLine, " ", false)
	match(t, &rEndOfLine, "", false)
	match(t, &rEndOfLine, "a", false)
}

func TestPegEndOfFile(t *testing.T) {
	match(t, &rEndOfFile, "", true)
	match(t, &rEndOfFile, " ", false)
}
