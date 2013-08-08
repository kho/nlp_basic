package treebank

import (
	"bytes"
	"fmt"
	"io"
	"strings"
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

func TestTokenizerPeekOnce(t *testing.T) {
	for _, c := range tokTestCases {
		input := strings.NewReader(c.input)
		tok := newTokenizer(input)
		s, k, e := tok.Peek()
		if len(c.tokens) > 0 {
			if e != nil {
				t.Errorf("expected nil error; got %q at first peek of %q\n", e, formatInput(c.input, input))
			}
			if s != c.tokens[0] {
				t.Errorf("expected token %q; got %q at first peek of %q\n", c.tokens[0], s, formatInput(c.input, input))
			}
			checkKind(s, k, t)
		} else {
			if e != io.EOF {
				t.Errorf("expected EOF; got (%q, %v, %q) at first peek of %q\n", s, k, e, formatInput(c.input, input))
			}
		}
	}
}

func TestTokenizerNext(t *testing.T) {
	for _, c := range tokTestCases {
		input := strings.NewReader(c.input)
		tok := newTokenizer(input)
		tok_id := 0
		for s, k, e := tok.Next(); e == nil; s, k, e = tok.Next() {
			ss := c.tokens[tok_id]
			if s != ss {
				t.Errorf("expected %q; got %q at %q\n", s, ss, formatInput(c.input, input))
			}
			checkKind(s, k, t)
			tok_id++
		}
		if tok_id != len(c.tokens) {
			t.Errorf("expected %d tokens; got %d at %q\n", len(c.tokens), tok_id, formatInput(c.input, input))
		}
	}
}

func TestTokenizerPeekPeekNext(t *testing.T) {
	for _, c := range tokTestCases {
		input := strings.NewReader(c.input)
		tok := newTokenizer(input)
		for i := 0; i < len(c.tokens); i++ {
			s0, k0, e0 := tok.Peek()
			s1, k1, e1 := tok.Peek()
			if s0 != s1 || k0 != k1 || e0 != e1 {
				t.Errorf("two Peek gave different results: (%q, %v, %q) vs (%q, %v, %q) at input %q, token %d\n",
					s0, k0, e0, s1, k1, e1, formatInput(c.input, input), i)
			}
			s2, k2, e2 := tok.Next()
			if s1 != s2 || k1 != k2 || e1 != e2 {
				t.Errorf("Peek and Next gave different results: (%q, %v, %q) vs (%q, %v, %q) at input %q, token %d\n",
					s1, k1, e1, 2, k2, e2, formatInput(c.input, input), i)
			}
			s := c.tokens[i]
			if s2 != s {
				t.Errorf("expected %q; got %q at input %q, token %d\n",
					s, s2, formatInput(c.input, input), i)
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

func TestParseSingle(t *testing.T) {
	for _, c := range parseCases {
		input := strings.NewReader(c.input)
		tree, err := Parse(input)
		if (err != nil) != c.err {
			s := "no error"
			if c.err {
				s = "error"
			}
			t.Errorf("expected %s; got %q at input %q\n", s, err, formatInput(c.input, input))
		}
		if err == nil && !equiv(tree, c.tree) {
			t.Errorf("expected %v; got %v at input %q\n", c.tree, tree, formatInput(c.input, input))
		}
	}
}

func TestParseMultiple(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	for _, c := range parseCases {
		if c.err {
			continue
		}
		n, e := buf.WriteString(c.input)
		if n != len(c.input) || e != nil {
			t.Fatal("error in creating test case!")
		}
		n, e = buf.WriteString("\n")
		if n != 1 || e != nil {
			t.Fatal("error in creating test case!")
		}
	}
	for _, c := range parseCases {
		if c.err {
			continue
		}
		tree, err := Parse(buf)
		if err != nil {
			t.Errorf("expected nil; got %q at input %q\n", err, c.input)
		}
		if err == nil && !equiv(tree, c.tree) {
			t.Errorf("expected %v; got %v at input %q\n", c.tree, tree, c.input)
		}
	}
}

var noParseCases = []string{"(())", "  (())  ", " ( ( ) ) "}

func TestParseNoParse(t *testing.T) {
	for _, c := range noParseCases {
		input := strings.NewReader(c)
		_, err := Parse(input)
		if err != NoParse {
			t.Errorf("expected NoParse; got %q at input %q\n", err, formatInput(c, input))
		}
	}
}

var mixedCases = []string{"(()) ((a a))", "((a a)) (()) ((a a))"}

func TestParseMixed(t *testing.T) {
	for _, c := range mixedCases {
		input := strings.NewReader(c)
		_, err := Parse(input)
		for err == nil || err == NoParse {
			_, err = Parse(input)
		}
		if err != io.EOF {
			t.Errorf("expected EOF; got %q at input %q\n", err, formatInput(c, input))
		}
	}
}

func checkKind(s string, k kind, t *testing.T) {
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

type unreader interface {
	// Len returns the number of unread bytes (e.g. strings.Reader)
	Len() int
}

func formatInput(s string, r unreader) string {
	unread := r.Len()
	return fmt.Sprintf("%s.%s", s[:len(s)-unread], s[len(s)-unread:])
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
