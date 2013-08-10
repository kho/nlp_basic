package treebank

import (
	"testing"
)

var copyCases = []Node{
	{"a", nil},
	{"a", []Node{{"b", nil}, {"c", nil}}},
	{"a", []Node{{"b", []Node{{"c", nil}}}}},
}

var modifyFuncs = []func(*Node){clearLabels, addChild}

func TestCopy(t *testing.T) {
	for _, tree := range copyCases {
		for _, f := range modifyFuncs {
			treeCopy := Copy(tree)
			if !equiv(tree, treeCopy) {
				t.Errorf("expected %v; got %v after copy\n", tree, treeCopy)
			}
			f(&treeCopy)
			if equiv(tree, treeCopy) {
				t.Errorf("modification is propagated between copies: %v\n", tree)
			}
		}
	}
}

func BenchmarkCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := range copyCases {
			_ = Copy(copyCases[j])
		}
	}
}

func clearLabels(node *Node) {
	node.Label = ""
	for i := range node.Children {
		clearLabels(&node.Children[i])
	}
}

func addChild(node *Node) {
	for i := range node.Children {
		addChild(&node.Children[i])
	}
	node.Children = append(node.Children, Node{"x", nil})
}

var stripAnnotationCases = []struct{ input, output string }{
	{"((S (NP this) (VP (V is) (NP (DT a) (NN test)))))",
		"((S (NP this) (VP (V is) (NP (DT a) (NN test)))))"},
	{"((S (NP-1 this-this) (VP-2 (V-3-4 is) (NP-NONE (DT=2 a) (NN test)))))",
		"((S (NP this-this) (VP (V is) (NP (DT a) (NN test)))))"},
	{"((S (NP this) (-NONE- (NP-1 *PRO*-2)) (VP (V is) (NP (DT a) (NN test)))))",
		"((S (NP this) (-NONE- (NP *PRO*)) (VP (V is) (NP (DT a) (NN test)))))"},
}

func TestStripAnnotation(t *testing.T) {
	for _, c := range stripAnnotationCases {
		tree0 := FromString(c.input)
		tree1 := FromString(c.output)
		(&tree0).StripAnnotation()
		if !equiv(tree0, tree1) {
			t.Errorf("expected %q; got %q\n")
		}
	}
}

var removeNoneCases = []struct{ input, output Node }{
	{FromString("((S (NP this) (VP (V is) (NP (DT a) (NN test)))))"),
		FromString("((S (NP this) (VP (V is) (NP (DT a) (NN test)))))")},
	{FromString("((S (NP (-NONE- (NP *PRO*)))))"), Node{"-NONE-", nil}},
	{FromString("((S (NP (-NONE- (NP *PRO*)) (-NONE- *T*)) (VP (-NONE- *T*) (V v))))"),
		FromString("((S (VP (V v))))")},
}

func TestRemoveNone(t *testing.T) {
	for _, c := range removeNoneCases {
		tree0 := c.input
		tree1 := c.output
		(&tree0).RemoveNone()
		if !equiv(tree0, tree1) {
			t.Errorf("expected %q; got %q\n")
		}
	}
}

var isPreTerminalCases = []struct {
	input  Node
	output bool
}{
	{Node{"A", nil}, false},
	{FromString("((A B))"), true},
	{FromString("((A (B (C D))))"), false},
	{FromString("((A (B C) (D E)))"), false},
}

func TestIsPreTerminal(t *testing.T) {
	for _, c := range isPreTerminalCases {
		a := IsPreTerminal(c.input)
		b := c.output
		if a != b {
			t.Errorf("expected %v; got %v for PreTerminal(%q)\n",
				b, a, c.input)
		}
	}
}
