package treebank

import (
	"testing"
)

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

var removeNoneCases = []struct{ input, output *ParseTree }{
	{FromString("((S (NP this) (VP (V is) (NP (DT a) (NN test)))))"),
		FromString("((S (NP this) (VP (V is) (NP (DT a) (NN test)))))")},
	{FromString("((S (NP (-NONE- (NP *PRO*)))))"), FromString("(())")},
	{FromString("((S (NP (-NONE- (NP *PRO*)) (-NONE- *T*)) (VP (-NONE- *T*) (V v))))"),
		FromString("((S (VP (V v))))")},
}

func TestRemoveNone(t *testing.T) {
	for _, c := range removeNoneCases {
		tree0 := c.input
		tree1 := c.output
		(&tree0).RemoveNone()
		if !equiv(tree0, tree1) {
			t.Errorf("expected %q; got %q\n", tree1, tree0)
		}
	}
}

var isPreTerminalCases = []struct {
	input  *ParseTree
	output bool
}{
	{&ParseTree{fromParents(0, []NodeId{NoNodeId}), []string{"A"}}, false},
	{FromString("((A B))"), true},
	{FromString("((A (B (C D))))"), false},
	{FromString("((A (B C) (D E)))"), false},
}

func TestIsPreTerminal(t *testing.T) {
	for _, c := range isPreTerminalCases {
		a := c.input.Topology.PreTerminal(c.input.Topology.Root())
		b := c.output
		if a != b {
			t.Errorf("expected %v; got %v for PreTerminal(%q)\n",
				b, a, c.input)
		}
	}
}
