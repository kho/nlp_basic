package treebank

import (
	"testing"
)

type tokenizerCase struct {
	input  string
	tokens []string
}

var tokTestCases = []tokenizerCase{
	{"", nil},
	{"(())", []string{"(", "(", ")", ")"}},
	{"(a)", []string{"(", "a", ")"}},
	{"( a )", []string{"(", "a", ")"}},
	{"(ab cd)", []string{"(", "ab", "cd", ")"}},
	{"ab(cd e ) ", []string{"ab", "(", "cd", "e", ")"}}}

func checkPart(p Part, s string, t *testing.T) {
	if p.Start < 0 || p.Start >= len(s) || p.End < p.Start || p.End > len(s) {
		t.Errorf("invalid part %v; string len is %d\n", p, len(s))
	}
}

func checkKind(s string, k Kind, t *testing.T) {
	if s == "(" && k != OPEN {
		t.Errorf("expected kind %v; got %v\n", OPEN, k)
	}
	if s == ")" && k != CLOSE {
		t.Errorf("expected kind %v; got %v\n", CLOSE, k)
	}
	if s != "(" && s != ")" && k != WORD {
		t.Errorf("expected kind %v; got %v\n", WORD, k)
	}
}

func TestTokenizerPeekOnce(t *testing.T) {
	for _, c := range tokTestCases {
		tok := NewTokenizer(c.input)
		p, k, e := tok.Peek()
		if len(c.tokens) > 0 {
			if e != nil {
				t.Errorf("expected nil error; got %v\n", e)
			}
			checkPart(p, c.input, t)
			a := c.input[p.Start:p.End]
			b := tok.Token(p)
			if a != b {
				t.Errorf("expected token %s; got %s from %v\n", a, b, p)
			}
			checkKind(tok.Token(p), k, t)
		} else {
			if e != EndOfInput {
				t.Errorf("expected %v; got (%v, %v, %v)\n", EndOfInput, p, k, e)
			}
		}
	}
}

func TestTokenizerNext(t *testing.T) {
	for _, c := range tokTestCases {
		tok := NewTokenizer(c.input)
		tok_id := 0
		for p, k, e := tok.Next(); e == nil; p, k, e = tok.Next() {
			checkPart(p, c.input, t)
			a := c.tokens[tok_id]
			b := tok.Token(p)
			if a != b {
				t.Errorf("expected %s; got %s\n", a, b)
			}
			checkKind(b, k, t)
			tok_id++
		}
		if tok_id != len(c.tokens) {
			t.Errorf("expected %d tokens; got %s\n", len(c.tokens), tok_id)
		}
	}
}

func TestTokenizerPeekPeekNext(t *testing.T) {
	for _, c := range tokTestCases {
		tok := NewTokenizer(c.input)
		for i := 0; i < len(c.tokens); i++ {
			p0, k0, e0 := tok.Peek()
			p1, k1, e1 := tok.Peek()
			if p0 != p1 || k0 != k1 || e0 != e1 {
				t.Errorf("two Peek gave different results: (%v, %v, %v) vs (%v, %v, %v) at input %s, token %d\n",
					p0, k0, e0, p1, k1, e1, c.input, i)
			}
			p2, k2, e2 := tok.Next()
			if p1 != p2 || k1 != k2 || e1 != e2 {
				t.Errorf("Peek and Next gave different results: (%v, %v, %v) vs (%v, %v, %v) at input %s, token %d\n",
					p1, k1, e1, p2, k2, e2, c.input, i)
			}
			a := c.tokens[i]
			b := tok.Token(p2)
			if a != b {
				t.Errorf("expected %s; got %s at input %s, token %d\n",
					a, b, c.input, i)
			}
		}
	}
}

type parserCase struct {
	input string
	tree  Node
	err   bool
}

var parseCases = []parserCase{
	{"((a b))", Node{"a", []Node{{"b", nil}}}, false},
	{"((a (b c)))", Node{"a", []Node{{"b", []Node{{"c", nil}}}}}, false},
	{"((a(b c)(d (e f))))", Node{"a", []Node{{"b", []Node{{"c", nil}}}, {"d", []Node{{"e", []Node{{"f", nil}}}}}}}, false},
	{"", Node{}, true},
	{"(", Node{}, true},
	{")", Node{}, true},
	{"((a))", Node{}, true},
	{"((a b)", Node{}, true},
}

func equiv(a Node, b Node) bool {
	if a.Label != b.Label {
		return false
	}
	if len(a.Children) != len(b.Children) {
		return false
	}
	for i := 0; i < len(a.Children); i++ {
		if !equiv(a.Children[i], b.Children[i]) {
			return false
		}
	}
	return true
}

func TestParse(t *testing.T) {
	for _, c := range parseCases {
		tok := NewTokenizer(c.input)
		tree, err := Parse(tok)
		if (err != nil) != c.err {
			s := "no error"
			if c.err {
				s = "error"
			}
			t.Errorf("expected %s; got %v at input %s.%s\n", s, err, c.input[0:tok.Pos()], c.input[tok.Pos():len(c.input)])
		}
		if err == nil && !equiv(tree, c.tree) {
			t.Fatalf("expected %v; got %v at input %s.%s\n", c.tree, tree, c.input[0:tok.Pos()], c.input[tok.Pos():len(c.input)])
		}
	}
}
