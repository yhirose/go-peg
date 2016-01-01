package peg

// visitor
type visitor interface {
	visitSequence(ope *sequence)
	visitPrioritizedChoice(ope *prioritizedChoice)
	visitZeroOrMore(ope *zeroOrMore)
	visitOneOrMore(ope *oneOrMore)
	visitOption(ope *option)
	visitAndPredicate(ope *andPredicate)
	visitNotPredicate(ope *notPredicate)
	visitLiteralString(ope *LiteralString)
	visitCharacterClass(ope *characterClass)
	visitAnyCharacter(ope *anyCharacter)
	visitTokenBoundary(ope *tokenBoundary)
	visitIgnore(ope *ignore)
	visitUser(ope *user)
	visitReference(ope *reference)
	visitRule(ope *Rule)
}

// abstractVisitor
type abstractVisitor struct {
}

func (v *abstractVisitor) visitSequence(ope *sequence)                   {}
func (v *abstractVisitor) visitPrioritizedChoice(ope *prioritizedChoice) {}
func (v *abstractVisitor) visitZeroOrMore(ope *zeroOrMore)               {}
func (v *abstractVisitor) visitOneOrMore(ope *oneOrMore)                 {}
func (v *abstractVisitor) visitOption(ope *option)                       {}
func (v *abstractVisitor) visitAndPredicate(ope *andPredicate)           {}
func (v *abstractVisitor) visitNotPredicate(ope *notPredicate)           {}
func (v *abstractVisitor) visitLiteralString(ope *LiteralString)         {}
func (v *abstractVisitor) visitCharacterClass(ope *characterClass)       {}
func (v *abstractVisitor) visitAnyCharacter(ope *anyCharacter)           {}
func (v *abstractVisitor) visitTokenBoundary(ope *tokenBoundary)         {}
func (v *abstractVisitor) visitIgnore(ope *ignore)                       {}
func (v *abstractVisitor) visitUser(ope *user)                           {}
func (v *abstractVisitor) visitReference(ope *reference)                 {}
func (v *abstractVisitor) visitRule(ope *Rule)                           {}

// tokenChecker
type tokenChecker struct {
	*abstractVisitor
	hasTokenBoundary bool
	hasRule          bool
}

func (v *tokenChecker) visitSequence(ope *sequence) {
	for _, o := range ope.opes {
		o.accept(v)
	}
}

func (v *tokenChecker) visitPrioritizedChoice(ope *prioritizedChoice) {
	for _, o := range ope.opes {
		o.accept(v)
	}
}

func (v *tokenChecker) visitZeroOrMore(ope *zeroOrMore)       { ope.ope.accept(v) }
func (v *tokenChecker) visitOneOrMore(ope *oneOrMore)         { ope.ope.accept(v) }
func (v *tokenChecker) visitOption(ope *option)               { ope.ope.accept(v) }
func (v *tokenChecker) visitTokenBoundary(ope *tokenBoundary) { v.hasTokenBoundary = true }
func (v *tokenChecker) visitIgnore(ope *ignore)               { ope.ope.accept(v) }
func (v *tokenChecker) visitReference(ope *reference)         { v.hasRule = true }
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

func (v *detectLeftRecursion) visitSequence(ope *sequence) {
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

func (v *detectLeftRecursion) visitPrioritizedChoice(ope *prioritizedChoice) {
	for _, o := range ope.opes {
		o.accept(v)
		if v.pos != -1 {
			v.done = true
			break
		}
	}
}

func (v *detectLeftRecursion) visitZeroOrMore(ope *zeroOrMore)         { ope.ope.accept(v); v.done = false }
func (v *detectLeftRecursion) visitOneOrMore(ope *oneOrMore)           { ope.ope.accept(v); v.done = true }
func (v *detectLeftRecursion) visitOption(ope *option)                 { ope.ope.accept(v); v.done = false }
func (v *detectLeftRecursion) visitAndPredicate(ope *andPredicate)     { ope.ope.accept(v); v.done = false }
func (v *detectLeftRecursion) visitNotPredicate(ope *notPredicate)     { ope.ope.accept(v); v.done = false }
func (v *detectLeftRecursion) visitLiteralString(ope *LiteralString)   { v.done = len(ope.lit) > 0 }
func (v *detectLeftRecursion) visitCharacterClass(ope *characterClass) { v.done = true }
func (v *detectLeftRecursion) visitAnyCharacter(ope *anyCharacter)     { v.done = true }
func (v *detectLeftRecursion) visitTokenBoundary(ope *tokenBoundary)   { ope.ope.accept(v) }
func (v *detectLeftRecursion) visitIgnore(ope *ignore)                 { ope.ope.accept(v) }

func (v *detectLeftRecursion) visitReference(ope *reference) {
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
