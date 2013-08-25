package treebank

import (
	"github.com/kho/nlp_basic/bimap"
	"github.com/kho/nlp_basic/syntax/heads"
	"reflect"
	"testing"
)

var labelIdRemapCases = []string{"((S (NP this) (VP (V is) (NP (DT a) (NN test)))))", "(())"}

func TestParseTreeRemapByLabel(t *testing.T) {
	m := bimap.New()
	for _, c := range labelIdRemapCases {
		tree := FromString(c)
		if tree.Topology.NumNodes() != len(tree.Label) {
			t.Errorf("Topology has %d nodes; Lable has %d labels", tree.Topology.NumNodes(), len(tree.Label))
		}
		tree.RemapByLabel(m)
		checkLabelId(tree.Label, tree.Id, m, t)
		tree.Id = nil
		tree.RemapByLabel(nil)
		checkLabelId(tree.Label, tree.Id, m, t)
	}
}

func TestParseTreeRemapById(t *testing.T) {
	m := bimap.New()
	for _, c := range labelIdRemapCases {
		tree := FromString(c)
		tree.RemapByLabel(m)
		tree.Label = nil
		tree.Map = nil
		tree.RemapById(m)
		checkLabelId(tree.Label, tree.Id, m, t)
		tree.Label = nil
		tree.RemapById(nil)
		checkLabelId(tree.Label, tree.Id, m, t)
	}
}

func checkLabelId(label []string, id []int, m *bimap.Map, t *testing.T) {
	if len(label) != len(id) {
		t.Errorf("Label has %d labels; Id has %d ids", len(label), len(id))
	}
	for i, l := range label {
		a := id[i]
		b := m.FindByString(l)
		if a != b {
			t.Errorf("expected %d; got %d", b, a)
		}
		if a == bimap.NoInt {
			t.Errorf("unknown word in Label: %s", l)
		}
	}
}

var fillSpanCases = []struct {
	input string
	span  []Span
}{
	{"(())", nil},
	{"((A B))", []Span{{0, 1}, {0, 1}}},
	{"((A (B C) (D (E F) (G H))))", []Span{{0, 3}, {0, 1}, {0, 1}, {1, 3}, {1, 2}, {1, 2}, {2, 3}, {2, 3}}},
}

func TestParseTreeFillSpan(t *testing.T) {
	for _, c := range fillSpanCases {
		tree := FromString(c.input)
		tree.FillSpan()
		numSpans := len(tree.Span)
		numNodes := tree.Topology.NumNodes()
		if numSpans != numNodes {
			t.Errorf("Span has %d spans; Topology has %d nodes", numSpans, numNodes)
		}
		for i, sp := range tree.Span {
			if sp != c.span[i] {
				t.Errorf("expected %v; got %v for node %d %q", c.span[i], sp, i, tree.StringUnder(NodeId(i)))
			}
		}
	}
}

var fillHeadCases = []struct {
	input string
	head  []int
	leaf  []NodeId
}{
	{"(())", nil, nil},
	{"((A B))", []int{0, -1}, []NodeId{1, 1}},
	{"((A (B (C D) (E F)) (G H)))", []int{1, 1, 0, -1, 0, -1, 0, -1}, []NodeId{7, 5, 3, 3, 5, 5, 7, 7}},
}

func TestParseTreeFillHead(t *testing.T) {
	finder := &heads.TableHeadFinder{nil, heads.HEAD_FINAL}
	for _, c := range fillHeadCases {
		tree := FromString(c.input)
		tree.FillHead(finder)
		numHeads := len(tree.Head)
		numNodes := tree.Topology.NumNodes()
		if numHeads != numNodes {
			t.Errorf("Head has %d heads; Topology has %d nodes", numHeads, numNodes)
		}
		for i, child := range tree.Head {
			if child != c.head[i] {
				t.Errorf("expected child %d; got %d as head for %q", c.head[i], child, tree.StringUnder(NodeId(i)))
			}
		}
	}
}

func TestParseTreeFillHeadLeaf(t *testing.T) {
	finder := &heads.TableHeadFinder{nil, heads.HEAD_FINAL}
	for _, c := range fillHeadCases {
		tree := FromString(c.input)
		tree.FillHead(finder)
		tree.FillHeadLeaf()
		numHl := len(tree.HeadLeaf)
		numNodes := tree.Topology.NumNodes()
		if numHl != numNodes {
			t.Errorf("HeadLeaf has %d leaves; Topology has %d nodes", numHl, numNodes)
		}
		for i, leaf := range tree.HeadLeaf {
			if leaf != c.leaf[i] {
				t.Errorf("expected leaf %d; got %d as head for %q", c.leaf[i], leaf, tree.StringUnder(NodeId(i)))
			}
		}
	}
}

var fillYieldPOSCases = []struct {
	input string
	yield []NodeId
	pos   []NodeId
}{
	{"(())", nil, nil},
	{"((A B))", []NodeId{1}, []NodeId{0}},
	{"((A (B (C D) (E F)) (G H)))", []NodeId{3, 5, 7}, []NodeId{2, 4, 6}},
}

func TestParseTreeFillYield(t *testing.T) {
	for _, c := range fillYieldPOSCases {
		tree := FromString(c.input)
		tree.FillYield()
		if !reflect.DeepEqual(tree.Yield, c.yield) {
			t.Errorf("expected %v; got %v for tree %q", c.yield, tree.Yield, c.input)
		}
	}
}

func TestParseTreeFillPOS(t *testing.T) {
	for _, c := range fillYieldPOSCases {
		tree := FromString(c.input)
		tree.FillPOS()
		if !reflect.DeepEqual(tree.POS, c.pos) {
			t.Errorf("expected %v; got %v for tree %q", c.pos, tree.POS, c.input)
		}
	}
}

func TestParseTreeFill(t *testing.T) {
	flags := []int{0, FILL_LABEL_ID, FILL_SPAN, FILL_HEAD, FILL_HEAD_LEAF, FILL_YIELD, FILL_POS, FILL_UP_LINK, FILL_EVERYTHING}
	finder := &heads.TableHeadFinder{nil, heads.HEAD_FINAL}
	m := bimap.New()

	tree := FromString("((A (B C) (D E)))")
	tree.RemapByLabel(m)
	tree.FillSpan()
	tree.FillHead(finder)
	tree.FillHeadLeaf()
	tree.FillYield()
	tree.FillPOS()
	tree.Topology.FillUpLink()

	for _, flag := range flags {
		tree1 := FromString("((A (B C) (D E)))")
		tree1.Fill(flag, m, finder)
		if (flag == FILL_EVERYTHING || flag&FILL_LABEL_ID != 0) && !reflect.DeepEqual(tree.Id, tree1.Id) {
			t.Errorf("expected %v; got %v when FILL_LABEL_ID", tree.Id, tree1.Id)
		}
		if (flag == FILL_EVERYTHING || flag&FILL_SPAN != 0) && !reflect.DeepEqual(tree.Span, tree1.Span) {
			t.Errorf("expected %v; got %v when FILL_SPAN", tree.Span, tree1.Span)
		}
		if (flag == FILL_EVERYTHING || flag&FILL_HEAD != 0) && !reflect.DeepEqual(tree.Head, tree1.Head) {
			t.Errorf("expected %v; got %v when FILL_HEAD", tree.Head, tree1.Head)
		}
		if (flag == FILL_EVERYTHING || flag&FILL_HEAD_LEAF != 0) && !reflect.DeepEqual(tree.HeadLeaf, tree1.HeadLeaf) {
			t.Errorf("expected %v; got %v when FILL_HEAD_LEAF", tree.HeadLeaf, tree1.HeadLeaf)
		}
		if (flag == FILL_EVERYTHING || flag&FILL_YIELD != 0) && !reflect.DeepEqual(tree.Yield, tree1.Yield) {
			t.Errorf("expected %v; got %v when FILL_YIELD", tree.Yield, tree1.Yield)
		}
		if (flag == FILL_EVERYTHING || flag&FILL_POS != 0) && !reflect.DeepEqual(tree.POS, tree1.POS) {
			t.Errorf("expected %v; got %v when FILL_POS", tree.POS, tree1.POS)
		}
		if (flag == FILL_EVERYTHING || flag&FILL_UP_LINK != 0) && !reflect.DeepEqual(tree.Topology.UpLink, tree1.Topology.UpLink) {
			t.Errorf("expected %v; got %v when FILL_UP_LINK", tree.Topology.UpLink, tree1.Topology.UpLink)
		}
	}

	label := tree.Label
	tree.Label = nil
	tree.Fill(FILL_LABEL_ID, nil, nil)
	if !reflect.DeepEqual(label, tree.Label) {
		t.Errorf("expected %v; got %v after filling Label from Id", label, tree.Label)
	}
}

var treeTopsortCases = []struct {
	input, output *ParseTree
	remove        []bool
}{
	{FromString("(())"), FromString("(())"), nil},
	{FromString("((A (B C) (D (E F) (G H))))"), FromString("((A (D (G H))))"), []bool{false, true, false, false, true, false, false, false}},
	{FromString("((A (B C) (D (E F) (G H))))"), FromString("((A (B C) (D (E F) (G H))))"), []bool{false, false, false, false, false, false, false, false}},
}

func TestParseTreeTopsort(t *testing.T) {
	m := bimap.New()
	finder := &heads.TableHeadFinder{nil, heads.HEAD_FINAL}
	for _, c := range treeTopsortCases {
		input, output := c.input, c.output
		input.Topology.Disconnect(c.remove)
		input.RemapByLabel(m)
		input.FillSpan()
		input.FillHead(finder)
		input.FillHeadLeaf()
		input.Topsort()

		output.RemapByLabel(m)
		output.FillSpan()
		output.FillHead(finder)
		output.FillHeadLeaf()

		if !equiv(input, output) {
			t.Errorf("expected %q; got %v", output, input)
		}
	}
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
	{&ParseTree{Topology: fromParents(0, []NodeId{NoNodeId}), Label: []string{"A"}}, false},
	{FromString("((A B))"), true},
	{FromString("((A (B (C D))))"), false},
	{FromString("((A (B C) (D E)))"), false},
}

func TestIsPreTerminal(t *testing.T) {
	for _, c := range isPreTerminalCases {
		a := c.input.Topology.PreTerminal(c.input.Topology.Root)
		b := c.output
		if a != b {
			t.Errorf("expected %v; got %v for PreTerminal(%q)\n",
				b, a, c.input)
		}
	}
}
