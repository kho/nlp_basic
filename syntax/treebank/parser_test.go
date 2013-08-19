package treebank

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

type tokenizeCase struct {
	input  string
	tokens []string
}

var tokTestCases = []tokenizeCase{
	{"", nil},
	{"(())", []string{"(", "(", ")", ")"}},
	{"(a)", []string{"(", "a", ")"}},
	{"( a )", []string{"(", "a", ")"}},
	{"(ab cd)", []string{"(", "ab", "cd", ")"}},
	{"ab(cd e ) ", []string{"ab", "(", "cd", "e", ")"}}}

func TestParserPeekTokenOnce(t *testing.T) {
	for _, c := range tokTestCases {
		input := strings.NewReader(c.input)
		tok := NewParser(input)
		b, k, e := tok.peekToken()
		s := string(b)
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

func TestParserNextToken(t *testing.T) {
	for _, c := range tokTestCases {
		input := strings.NewReader(c.input)
		tok := NewParser(input)
		tok_id := 0
		for b, k, e := tok.nextToken(); e == nil; b, k, e = tok.nextToken() {
			s := string(b)
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

func TestParserPeekPeekNext(t *testing.T) {
	for _, c := range tokTestCases {
		input := strings.NewReader(c.input)
		tok := NewParser(input)
		for i := 0; i < len(c.tokens); i++ {
			b0, k0, e0 := tok.peekToken()
			b1, k1, e1 := tok.peekToken()
			s0, s1 := string(b0), string(b1)
			if s0 != s1 || k0 != k1 || e0 != e1 {
				t.Errorf("two Peek gave different results: (%q, %v, %q) vs (%q, %v, %q) at input %q, token %d\n",
					s0, k0, e0, s1, k1, e1, formatInput(c.input, input), i)
			}
			b2, k2, e2 := tok.nextToken()
			s2 := string(b2)
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
	tree  *ParseTree
	err   bool
}

var parseCases = []parserCase{
	{"((a b))", &ParseTree{Topology: fromParents(0, []NodeId{NoNodeId, 0}), Label: []string{"a", "b"}}, false},
	{"((a (b c)))", &ParseTree{Topology: fromParents(0, []NodeId{NoNodeId, 0, 1}), Label: []string{"a", "b", "c"}}, false},
	{"((a(b c)(d (e f))))", &ParseTree{Topology: fromParents(0, []NodeId{NoNodeId, 0, 1, 0, 3, 4}), Label: []string{"a", "b", "c", "d", "e", "f"}}, false},
	{"(())", &ParseTree{Topology: NewEmptyTopology(), Label: nil}, false},
	{"", nil, true},
	{"(", nil, true},
	{")", nil, true},
	{"((a))", nil, true},
	{"((a b)", nil, true},
}

func TestParserSingle(t *testing.T) {
	for _, c := range parseCases {
		input := strings.NewReader(c.input)
		parser := NewParser(input)
		tree, err := parser.Next()
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
	var goodTrees []string
	for _, c := range parseCases {
		if c.err {
			continue
		}
		goodTrees = append(goodTrees, c.input)
	}
	inputString := strings.Join(goodTrees, " ")
	input := strings.NewReader(inputString)
	parser := NewParser(input)
	for _, c := range parseCases {
		if c.err {
			continue
		}
		tree, err := parser.Next()
		if err != nil {
			t.Errorf("expected nil; got %q at input %q\n", err, c.input)
		}
		if err == nil && !equiv(tree, c.tree) {
			t.Errorf("expected %v; got %v at input %q\n", c.tree, tree, c.input)
		}
	}

	input.Seek(0, 0)
	trees, err := ParseAll(input)
	if err != nil {
		t.Errorf("expected nil; got %q at input %q\n", err, formatInput(inputString, input))
	}
	i := 0
	for _, c := range parseCases {
		if c.err {
			continue
		}
		if !equiv(trees[i], c.tree) {
			t.Errorf("expected %v; got %v as the %d-th tree\n", c.tree, trees[i], i)
		}
		i++
	}
}

var noParseCases = []string{"(())", "  (())  ", " ( ( ) ) "}

func TestParseNoParse(t *testing.T) {
	for _, c := range noParseCases {
		input := strings.NewReader(c)
		parser := NewParser(input)
		tree, err := parser.Next()
		if err != nil {
			t.Errorf("expected nil; got %q at input %q\n", err, formatInput(c, input))
		}
		if root := tree.Topology.Root; root != NoNodeId {
			t.Errorf("expected NoNodeId; got %d\n", root)
		}
		labelTreeSanityCheck(tree, t)
	}

	input := strings.NewReader(strings.Join(noParseCases, " "))
	trees, err := ParseAll(input)
	if err != nil {
		t.Errorf("expected nil; got %q\n", err)
	}
	for _, tree := range trees {
		if root := tree.Topology.Root; root != NoNodeId {
			t.Errorf("expected NoNodeId; got %d\n", root)
		}
		labelTreeSanityCheck(tree, t)
	}
}

var mixedCases = []string{"(()) ((a a))", "((a a)) (()) ((a a))"}

func TestParseMixed(t *testing.T) {
	for _, c := range mixedCases {
		input := strings.NewReader(c)
		parser := NewParser(input)
		_, err := parser.Next()
		for err == nil {
			_, err = parser.Next()
		}
		if err != io.EOF {
			t.Errorf("expected EOF; got %q at input %q\n", err, formatInput(c, input))
		}
	}
}

var fromStringCases = []struct {
	input string
	error bool
}{
	{"((A B))", false},
	{"((A (B C) (D E)))    ", false},
	{"((A (B C) (D E))) ((A B))", true},
	{"(())", false},
}

func TestFromString(t *testing.T) {
	for _, c := range fromStringCases {
		func() {
			defer func() {
				err := recover()
				if (err == nil) != (c.error == false) {
					t.Errorf("expected error = %v; got %q\n",
						c.error, err)
				}
			}()
			_ = FromString(c.input)
		}()
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		input := strings.NewReader(benchmarkCases)
		parser := NewParser(input)
		_, err := parser.Next()
		for err == nil {
			_, err = parser.Next()
		}
		if err != io.EOF {
			b.Errorf("unexpected error %q", err)
		}
	}
}

func checkKind(s string, k kind, t *testing.T) {
	if s == "(" && k != kOpen {
		t.Errorf("expected kind %v; got %v\n", kOpen, k)
	}
	if s == ")" && k != kClose {
		t.Errorf("expected kind %v; got %v\n", kClose, k)
	}
	if s != "(" && s != ")" && k != kWord {
		t.Errorf("expected kind %v; got %v\n", kWord, k)
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

func labelTreeSanityCheck(tree *ParseTree, t *testing.T) {
	topologySanityCheck(tree.Topology, t)
	if size1, size2 := tree.Topology.NumNodes(), len(tree.Label); size1 != size2 {
		t.Errorf("num nodes and num labels do not match: %d vs %d\n",
			size1, size2)
	}
}

func equiv(a *ParseTree, b *ParseTree) bool {
	if !a.Topology.Equal(b.Topology) {
		return false
	}
	if len(a.Label) != len(b.Label) {
		return false
	}
	for i := 0; i < len(a.Label); i++ {
		if a.Label[i] != b.Label[i] {
			return false
		}
	}
	if len(a.Id) != len(b.Id) {
		return false
	}
	for i := 0; i < len(a.Id); i++ {
		if a.Id[i] != b.Id[i] {
			return false
		}
	}
	if len(a.Span) != len(b.Span) {
		return false
	}
	for i := 0; i < len(a.Span); i++ {
		if a.Span[i] != b.Span[i] {
			return false
		}
	}
	if len(a.Head) != len(b.Head) {
		return false
	}
	for i := 0; i < len(a.Head); i++ {
		if a.Head[i] != b.Head[i] {
			return false
		}
	}
	if len(a.HeadLeaf) != len(b.HeadLeaf) {
		return false
	}
	for i := 0; i < len(a.HeadLeaf); i++ {
		if a.HeadLeaf[i] != b.HeadLeaf[i] {
			return false
		}
	}
	return true
}

var benchmarkCases = `((aa (aaaaaa (aa aaa)) (aa (aa aaaaaa) (aa aaa) (aa aaa) (aaaaaa (aaaaaa (aaaaa (aa aaaaaa) (aa aaaaaa)) (aa (aaaaaa (aaaaaa aaaa)) (aa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aa aaaaaa) (aaaaaa (aa aaaaaa)))) (aaa aaa))) (aa (aa aaaaaa))) (aa (aa aaaaaa) (aa (aaaaaa (a aaa) (aaa (aa (aaaaa (aa aaaaaa)) (aaaa (aa aaaaaa)) (aa (aa aaaaaaaaaaaaaaa) (aaa (a aaa))) (aaaa (aa aaaaaa)) (aa (aa aaaaaaaaa))) (aa aaa))) (aa aaa) (aaa (aa (aaaa (aa aaa)) (aa (aa aaaaaa))) (aaa aaa)) (aa (aa (aa aaa) (aaaaaa (aa (aa aaaaaa)) (aa (aa aaaaaaaaa) (aa aaaaaa)))) (aa (aa (aa aaaaaa) (aaaaaa (aa aaaaaaaaa))) (aa aaa) (aa (aa aaaaaa) (aaaaaa (aa aaaaaa))))))))) (aa aaa)))
(())
((aa (aa (aaaaaa (aa aaa) (aa (aa aaa) (aaa (a aaa))) (aa (aa aaaaaa))) (aa aaa) (aa (aa aaaaaa) (aa (aa aaa) (aaaaaa (aaaaaa (aaaaaa aaaaa)) (aa (aa aaa) (aaaaaa (aa (aa (aa (aaaaaa (aaaaaa aaaa)) (aa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aaaaaa (aa aaaaaa)) (aaaa (aa aaa)) (aa (aa aaa)))) (aaa aaa))) (aa (aa aaaaa) (aa aaaaaa)))) (aa (aa aaaaaa)))))))) (aa aaa) (aa (aaaaaa (aa (aaaaaa (aaaaaa aaaa)) (aa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aa aaa) (aaaaaa (aaaaaa aaaaa)))) (aaa aaa))) (aa (aa aaaaaa) (aa aaaaaa))) (aa (aaaa (aa aaaaaa)) (aa (aa aaa)))) (aa aaa) (aa aaaaaa)))
((aa (aa (aaaaaa (aa aaa)) (aa (aa aaa) (aaaaaa (aa aaaaaa)) (aa (aa (aaaa (aa aaaaaa)) (aaaa (aa aaa)) (aa (aaa (aa aaa) (aa aaa) (aa aaa)) (aaaaaa (aaaaaa (aaaaaa aaaaa)) (aa (aa (aa aaa) (aaaaaa (aa aaaaaa))) (aa aaa) (aa (aa (a aaa) (aa (aa aaa))) (aa (aa aaa) (aaaa (aa aaaaaa))))))))))) (aa aaa) (aa (aaaaaa (aa aaaaaa)) (aa (aa aaa) (aaaaaaa (aa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aa aaa))) (aa aaa)) (aa aaa) (aa (aaaaaa (aaaaaa aaaaa)) (aa (aaaa (aa aaaaaa)) (aaaa (aa aaaaaa)) (aaaa (aa aaa)) (aa (aa aaaaaa))))))) (aa aaa)))
((aa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aaaaaa (a aaa) (aaa (aa (aaa (aa (aa aaaaaa) (aa aaa) (aa aaaaaa)) (aaa aaa)) (aa (aa aaaaaa))) (aa aaa))) (aa (aa aaa) (aa (aaaa (aa aaaaaa)) (aa (aa aaaaaa) (aaaaaa (aa (aa aaa) (aaa (a aaa))) (aaaa (aa aaaaaa)) (aa (aa aaaaaa)))))))) (aa aaa) (aa (aaaaaa (aa aaa)) (aa (aa aaa) (aaaaaa (aaaaaa (aaaaaa aaaaa)) (aa (aa (aa (aaaa (aa aaaaaa)) (aa (aa aaaaaaaaaaaa))) (aa aaa) (aa (aaaa (aa aaaaaa)) (aa (aa aaaaaa) (aaaaaa (aa (aa aaaaaa)) (aa (aaaaaa (aaaaaa aaaa)) (aa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aaaaaa (aa aaaaaa)) (aa (aa aaaaaa)))) (aaa aaa))) (aaaa (aa aaaaaa)) (aa (aa aaaaaa))))) (aa aaa) (aa (aa aaaaaa) (aaaaaa (aa (aa aaa)) (aa (aa aaaaaaaaa) (aa aaa) (aa aaaaaaaaa)))) (aa aaa) (aa (aa aaaaaa) (aaaaaa (aaa (aa (aa aaaaaa) (aa aaaaaa) (aa aaaaaa)) (aa aaa)) (aa (aaaaaa (aaaaaa aaaa)) (aa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aa (aa aaaaaa) (aaaaaa (aaaaaa aaaaaaa))) (aa aaa) (aa (aa aaaaaa) (aaaaaa (aaaaaa aaaaaaa))) (aa aaa) (aa (aa aaaaaa) (aaaaaaaa (aa aaaaaa))))) (aaa aaa))) (aaaa (aa aaaaaa)) (aa (aa aaaaaa)))) (aa aaa) (aa (aaaa (aa aaaaaa)) (aa (aa aaaaaa) (aaaaaa (aa aaaaaa) (aa aaaaaa)))) (aa aaa) (aa (aa aaaaaa) (aaaaaa (aa aaaaaa) (aa aaaaaa))) (aa aaa) (aa (aa aaaaaa) (aaaaaa (aa aaaaaa) (aa aaaaaa)))) (aa aaa) (aa (aaa aaa) (aa (aa aaa) (aaaaaa (aa aaaaaa)))))))) (aa aaa)))
(())
((aa (aaaaaa (aaaaaa (aaaaaa aaaaa)) (aa (aa aaa) (aa aaa) (aaaaaa (aa aaaaaaaaaaaa)))) (aa aaa) (aaaaaa (aa aaaaaa)) (aa (aaaa (aa aaa)) (aa (aaa (aa aaaaaa) (aa aaa)) (aaaaaa (aa aaaaaaaaaaaaaaaaaa)))) (aa aaa)))
((aa (aaaa (aa aaaaaa)) (aaaaaa (aaaaaa (aa (aaaaaa (aaaaaa aaaa)) (aa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aaaa (aa aaaaaa)) (aa (aa aaa) (aaaaaa (aaaaaa aaaaa))))) (aaa aaa))) (aa (aa aaaaaa) (aa (aa aaa) (aaa (a aaa)))) (aa (aa aaaaaa))) (aa aaa) (aa (aa (aa aaaaaa) (aa aaaaaa)) (aa aaa) (aa (aa aaaaaa) (aa aaaaaa)) (aa aaa) (aa (aa aaaaaa) (aa aaaaaa)))) (aa aaa) (aaaaaaa (aa (aa (aa aaa) (aa (aa aaa) (aaa (a aaa)))) (aa (aa aaaaaa))) (aa aaa)) (aa aaa) (aaa (aa aaa)) (aa (aaaaaa (aa aaaaaa)) (aaaa (aa aaa)) (aa (aa aaaaaa) (aaaaaa (aa aaaaaa)))) (aa aaa)))
((aaaaa (aa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aa aaa) (aa (aa aaa) (aaaaaa (aa aaa))))) (aa aaa)) (aa aaa) (aa (aa (aaaaaa (aa (aaaaa (aa aaaaaaaaa)) (aaa (aaaa (aa aaaaaa)) (aaa aaa)) (aa (aa (aa aaaaaa)) (aa aaa) (aa (aa aaaaaa)))) (aa (aa aaa)) (aaa (aaaa (aa aaaaaa)) (aaa aaa)) (aa (aa aaaaaa))) (aa (aa aaa) (aa (aa aaa) (aaaaaaaaa (aa (aa aaaaaa)) (aa (aa aaaaaa)))))) (aa aaa)) (aa aaa)))
(())
((aaa (aaa aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa) (aa aaa) (aa aaa)))
((aaa (aaa (aa (aa aaa) (aa (aa (aa aaaaaa)) (aa (aa (aa aaa) (aaa (a aaa))) (aa (aa aaa)))) (aa aaa) (aa (aa (aa aaaaaa)) (aa (aa aaa) (aaa (a aaa))) (aa (aa aaa))) (aa aaa)) (aa aaa) (aa (aaaaaa (aaaaaa aaaaa)) (aa (aaaa (aa aaaaaa)) (aa (aa aaa) (aaaaaa (aa (aa aaa) (aaa (a aaa))) (aa (aa aaa))))))) (aa aaa) (aa (aaaaaa (aa aaaaaa)) (aa (aaaa (aa aaa)) (aa (aa aaa) (aaaaaa (aa aaaaaa)) (aa (aaaaaa (aaaaaa aaaaa)) (aa (aa aaa)))))) (aa aaa) (aa (aaaaaa (aaaaaa aaaaa)) (aa (aa aaa) (aaaaaa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aaa (aa aaa) (aa aaa) (aa aaaaaa)))) (aa aaa)))) (aa aaa)))
((aaa (aaa (aa (aaaa (aa aaa)) (aaaaaa (aaa (aa (aa aaaaaa)) (aaa aaa)) (aa (aaaaa (aa aaaaaa)) (aa (aa aaaaaa)))) (aa aaa) (aa (aaaa (aa aaa)) (aa (aa aaa) (aaaaaa (aa (aaaaaa (aaaaaa aaaa)) (aa (aa (aaaaaa (aa (aaaaaa (aaaaaa aaaa)) (aa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aaaa (aa aaaaaa)) (aa (aa aaaaaa)))) (aaa aaa))) (aa (aa aaaaaa))) (aa (aa aaaaaa) (aaaaaa (aaaaaa aaaaa)))) (aaa aaa))) (aa (aa aaaaaa)))))) (aa aaa) (aa (aaaaaa (aa aaaaaa)) (aa (aaaa (aa aaa)) (aa (aa aaa) (aa (aa aaaaaa))))) (aa aaa) (aa (aa (aaaa (aa aaaaaa)) (aaaaaa (aa aaaaaa)) (aa (aaaa (aa aaa)) (aa (aa aaa) (aa (aaaa (aa aaaaaa)) (aa (aa aaaaaa)))))) (aa aaa))) (aa aaa) (aa (aaaaaa (aa aaaaaa)) (aa (aaaa (aa aaa)) (aa (aa aaaaaa) (aa aaa) (aaaaaa (aa (aaaaaa (aaaaa (aa aaaaaaaaa)) (aa (aa aaaaaa))) (aa (aa aaaaaa))) (aa aaa) (aa (aaaaaa (aaaaaa aaaaa)) (aa (aa (aa aaaaaa) (aaaaaa (aa aaaaaa))) (aa aaa) (aa (aa aaaaaa) (aaaaaa (aaaaaa (aa (aaaaaa (aaaaaa aaaaa)) (aa (aa aaaaaa))) (aaa aaa)) (aa (aa aaaaaa)))))) (aa aaa) (aa (aaaa (aa aaa)) (aaaaaa (aaaa (aa aaaaaa)) (aa (aaaaaa (aa aaaaaa)) (aa (aa aaa) (aa aaa)))) (aa aaa) (aaaaaa (aaaaaa aaaaa)) (aa (aa (aaaa (aa aaa)) (aa (aa aaa) (aaaaaa (aaaaaa (aa aaaaaa)) (aa (aaaa (aa aaaaaa)) (aa (aa aaa) (aa (aa aaaaaa) (aa aaa))))))) (aa aaa) (aa (aaaa (aa aaa)) (aaaa (aa aaa)) (aa (aa aaa) (aa (aa aaaaaa) (aaaaaa (aa aaa))))))))))) (aa aaa)))
`
