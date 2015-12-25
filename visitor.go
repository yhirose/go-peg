package peg

// visitor
type visitor interface {
	visitSequence(ope *Sequence)
	visitPrioritizedChoice(ope *PrioritizedChoice)
	visitZeroOrMore(ope *ZeroOrMore)
	visitOneOrMore(ope *OneOrMore)
	visitOption(ope *Option)
	visitAndPredicate(ope *AndPredicate)
	visitNotPredicate(ope *NotPredicate)
	visitLiteralString(ope *LiteralString)
	visitCharacterClass(ope *CharacterClass)
	visitAnyCharacter(ope *AnyCharacter)
	visitTokenBoundary(ope *TokenBoundary)
	visitIgnore(ope *Ignore)
	visitUser(ope *User)
	visitReference(ope *Reference)
	visitRule(ope *Rule)
}

// abstractVisitor
type abstractVisitor struct {
}

func (v *abstractVisitor) visitSequence(ope *Sequence)                   {}
func (v *abstractVisitor) visitPrioritizedChoice(ope *PrioritizedChoice) {}
func (v *abstractVisitor) visitZeroOrMore(ope *ZeroOrMore)               {}
func (v *abstractVisitor) visitOneOrMore(ope *OneOrMore)                 {}
func (v *abstractVisitor) visitOption(ope *Option)                       {}
func (v *abstractVisitor) visitAndPredicate(ope *AndPredicate)           {}
func (v *abstractVisitor) visitNotPredicate(ope *NotPredicate)           {}
func (v *abstractVisitor) visitLiteralString(ope *LiteralString)         {}
func (v *abstractVisitor) visitCharacterClass(ope *CharacterClass)       {}
func (v *abstractVisitor) visitAnyCharacter(ope *AnyCharacter)           {}
func (v *abstractVisitor) visitTokenBoundary(ope *TokenBoundary)         {}
func (v *abstractVisitor) visitIgnore(ope *Ignore)                       {}
func (v *abstractVisitor) visitUser(ope *User)                           {}
func (v *abstractVisitor) visitReference(ope *Reference)                 {}
func (v *abstractVisitor) visitRule(ope *Rule)                           {}

// tokenChecker
type tokenChecker struct {
	*abstractVisitor
	hasTokenBoundary bool
	hasRule          bool
}

func (v *tokenChecker) visitSequence(ope *Sequence) {
	for _, o := range ope.opes {
		o.accept(v)
	}
}

func (v *tokenChecker) visitPrioritizedChoice(ope *PrioritizedChoice) {
	for _, o := range ope.opes {
		o.accept(v)
	}
}

func (v *tokenChecker) visitZeroOrMore(ope *ZeroOrMore)       { ope.ope.accept(v) }
func (v *tokenChecker) visitOneOrMore(ope *OneOrMore)         { ope.ope.accept(v) }
func (v *tokenChecker) visitOption(ope *Option)               { ope.ope.accept(v) }
func (v *tokenChecker) visitTokenBoundary(ope *TokenBoundary) { v.hasTokenBoundary = true }
func (v *tokenChecker) visitIgnore(ope *Ignore)               { ope.ope.accept(v) }
func (v *tokenChecker) visitReference(ope *Reference)         { v.hasRule = true }
func (v *tokenChecker) visitRule(ope *Rule)                   { v.hasRule = true }

func (v *tokenChecker) isToken() bool {
	return v.hasTokenBoundary || !v.hasRule
}

// detectLeftRecursion
type detectLeftRecursion struct {
	*abstractVisitor
	pos  int
	name string
	refs map[string]bool
	done bool
}

func (v *detectLeftRecursion) visitSequence(ope *Sequence) {
	for _, o := range ope.opes {
		o.accept(v)
		if v.done {
			break
		} else if v.pos != -1 {
			v.done = true
			break
		}
	}
}

func (v *detectLeftRecursion) visitPrioritizedChoice(ope *PrioritizedChoice) {
	for _, o := range ope.opes {
		o.accept(v)
		if v.pos != -1 {
			v.done = true
			break
		}
	}
}

func (v *detectLeftRecursion) visitZeroOrMore(ope *ZeroOrMore)         { ope.ope.accept(v); v.done = false }
func (v *detectLeftRecursion) visitOneOrMore(ope *OneOrMore)           { ope.ope.accept(v); v.done = true }
func (v *detectLeftRecursion) visitOption(ope *Option)                 { ope.ope.accept(v); v.done = false }
func (v *detectLeftRecursion) visitAndPredicate(ope *AndPredicate)     { ope.ope.accept(v); v.done = false }
func (v *detectLeftRecursion) visitNotPredicate(ope *NotPredicate)     { ope.ope.accept(v); v.done = false }
func (v *detectLeftRecursion) visitLiteralString(ope *LiteralString)   { v.done = len(ope.lit) > 0 }
func (v *detectLeftRecursion) visitCharacterClass(ope *CharacterClass) { v.done = true }
func (v *detectLeftRecursion) visitAnyCharacter(ope *AnyCharacter)     { v.done = true }
func (v *detectLeftRecursion) visitTokenBoundary(ope *TokenBoundary)   { ope.ope.accept(v) }
func (v *detectLeftRecursion) visitIgnore(ope *Ignore)                 { ope.ope.accept(v) }

func (v *detectLeftRecursion) visitReference(ope *Reference) {
	if ope.name == v.name {
		v.pos = ope.pos
	} else if _, ok := v.refs[ope.name]; ok {

	} else {
		v.refs[ope.name] = true
		ope.getRule().accept(v)
	}
	v.done = true
}

func (v *detectLeftRecursion) visitRule(ope *Rule) { ope.Ope.accept(v) }
