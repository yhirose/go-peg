package peg

import "testing"

type Cases []struct {
	input  string
	want   int
	expErr bool
}

func run(name string, t *testing.T, ope operator, cases Cases) {
	for _, cs := range cases {
		v := &Values{}
		c := &context{}
		if got, err := ope.parse(cs.input, 0, v, c, nil); got != cs.want && (cs.expErr && err == nil) {
			t.Errorf("[%s] input:%q want:%d got:%d err:%s", name, cs.input, cs.want, got, err)
		}
	}
}

func TestSequence(t *testing.T) {
	ope := Seq(
		Lit("日本語"),
		Lit("も"),
		Lit("OK"),
		Lit("です。"),
	)
	cases := Cases{
		{"日本語もOKです。", 23, false},
		{"日本語OKです。", -1, true},
	}
	run("Sequence", t, ope, cases)
}

func TestPrioritizedChoice(t *testing.T) {
	ope := Cho(
		Lit("English"),
		Lit("日本語"),
	)
	cases := Cases{
		{"日本語", 9, false},
		{"English", 7, false},
		{"Go", -1, true},
	}
	run("PrioritizedChoice", t, ope, cases)
}

func TestZeroOrMore(t *testing.T) {
	ope := Zom(
		Lit("abc"),
	)
	cases := Cases{
		{"", 0, false},
		{"a", 0, false},
		{"b", 0, false},
		{"ab", 0, false},
		{"abc", 3, false},
		{"abca", 3, false},
		{"abcabc", 6, false},
	}
	run("ZeroOrMore", t, ope, cases)
}

func TestOneOrMore(t *testing.T) {
	ope := Oom(
		Lit("abc"),
	)
	cases := Cases{
		{"", -1, true},
		{"a", -1, true},
		{"b", -1, true},
		{"ab", -1, true},
		{"abc", 3, false},
		{"abca", 3, false},
		{"abcabc", 6, false},
	}
	run("OneOrMore", t, ope, cases)
}

func TestOption(t *testing.T) {
	ope := Opt(
		Lit("abc"),
	)
	cases := Cases{
		{"", 0, false},
		{"a", 0, false},
		{"b", 0, false},
		{"ab", 0, false},
		{"abc", 3, false},
		{"abca", 3, false},
		{"abcabc", 3, false},
	}
	run("Option", t, ope, cases)
}

func TestAndPredicate(t *testing.T) {
	ope := Apd(
		Lit("abc"),
	)
	cases := Cases{
		{"", -1, true},
		{"a", -1, true},
		{"b", -1, true},
		{"ab", -1, true},
		{"abc", 0, false},
		{"abca", 0, false},
		{"abcabc", 0, false},
	}
	run("AndPredicate", t, ope, cases)
}

func TestNotPredicate(t *testing.T) {
	ope := Npd(
		Lit("abc"),
	)
	cases := Cases{
		{"", 0, false},
		{"a", 0, false},
		{"b", 0, false},
		{"ab", 0, false},
		{"abc", -1, true},
		{"abca", -1, true},
		{"abcabc", -1, true},
	}
	run("NotPredicate", t, ope, cases)
}

func TestLiteralString(t *testing.T) {
	ope := Lit("日本語")
	cases := Cases{
		{"", -1, true},
		{"日", -1, true},
		{"日本語", 9, false},
		{"日本語です。", 9, false},
		{"English", -1, true},
	}
	run("LiteralString", t, ope, cases)
}

func TestCharacterClass(t *testing.T) {
	ope := Cls("a-zA-Z0-9_")
	cases := Cases{
		{"", -1, true},
		{"a", 1, false},
		{"b", 1, false},
		{"z", 1, false},
		{"A", 1, false},
		{"B", 1, false},
		{"Z", 1, false},
		{"0", 1, false},
		{"1", 1, false},
		{"9", 1, false},
		{"_", 1, false},
		{"-", -1, true},
		{" ", -1, true},
	}
	run("CharacterClass", t, ope, cases)
}

func TestTokenBoundary(t *testing.T) {
	ope := Seq(Tok(Lit("hello")), Lit(" "))
	v := &Values{}
	c := &context{}
	input := "hello "

	want := len(input)
	if got, err := ope.parse(input, 0, v, c, nil); got != want || err != nil {
		t.Errorf("[%s] input:%q want:%d got:%d err:%s", "TokenBoundary", input, want, got, err)
	}

	tok := "hello"
	if len(v.Ts) == 0 || v.Ts[0].S != tok {
		t.Errorf("[%s] input:%q want:%s got:%s", "TokenBoundary", input, tok, v.Ts[0].S)
	}
}

func TestIgnore(t *testing.T) {
	var NUMBER, WS Rule
	NUMBER.Ope = Seq(Tok(Oom(Cls("0-9"))), Ign(&WS))
	WS.Ope = Zom(Cls(" \t"))

	input := "123 "

	NUMBER.Action = func(v *Values, d Any) (Any, error) {
		n := 0
		if len(v.Vs) != n {
			t.Errorf("[%s] input:%q want:%d got:%d", "Ignore", input, n, len(v.Vs))
		}
		return nil, nil
	}

	want := len(input)
	if l, _, _ := NUMBER.Parse(input, nil); l != want {
		t.Errorf("[%s] input:%q want:%d got:%d", "Ignore", input, want, l)
	}
}
