package treebank

import (
	"testing"
)

func TestNewRootedTopology(t *testing.T) {
	rooted := NewRootedTopology()
	if numNodes := rooted.NumNodes(); numNodes != 1 {
		t.Fatalf("NewRootedTopology created a topology with %d nodes\n",
			numNodes)
	}
	root := rooted.Root
	if root != 0 {
		t.Fatalf("NewRootedTopology created a topology with root %d\n",
			root)
	}

	if children := rooted.Children[root]; len(children) != 0 {
		t.Errorf("NewRootedTopology created a topology of root with children %v\n", children)
	}
}

func TestNewEmptyTopology(t *testing.T) {
	empty := NewEmptyTopology()
	numNodes := empty.NumNodes()
	if numNodes != 0 {
		t.Fatalf("NewEmptyTopology created a topology with %d nodes\n",
			numNodes)
	}
}

func TestTopologyCopy(t *testing.T) {
	t1 := NewRootedTopology()
	t2 := t1.Copy()

	if !t1.Equal(t2) {
		t.Errorf("topologies are not equal after copy: %v vs %v\n",
			*t1, *t2)
	}

	oldRoot := t1.Root
	newRoot := t1.AddNode()
	t1.Root = newRoot
	t1.AppendChild(newRoot, oldRoot)

	if size1, size2 := t1.NumNodes(), t2.NumNodes(); size1 == size2 {
		t.Errorf("sizes do not differ after modification: %d vs %d\n",
			size1, size2)
	}

	if root1, root2 := t1.Root, t2.Root; root1 == root2 {
		t.Errorf("roots do not differ after modification: %d vs %d\n",
			root1, root2)
	}
}

func TestTopologyAddNode(t *testing.T) {
	const numInserts = 100
	trees := []*Topology{NewEmptyTopology(), NewRootedTopology()}
	for _, tree := range trees {
		oldSize := tree.NumNodes()
		oldRoot := tree.Root
		for i := 0; i < numInserts; i++ {
			node := tree.AddNode()
			if node != NodeId(i+oldSize) {
				t.Errorf("expected new id %d; got %d\n", i+oldSize, node)
			}
			if children := tree.Children[node]; len(children) != 0 {
				t.Errorf("expected tree children; got %v\n", children)
			}
		}
		topologySanityCheck(tree, t)
		if root := tree.Root; root != oldRoot {
			t.Errorf("expected root %d; got %d\n", oldRoot, root)
		}
		if numNodes := tree.NumNodes(); numNodes != oldSize+numInserts {
			t.Errorf("expected %d nodes after AddNode(); got %d\n",
				oldSize+numInserts, numNodes)
		}
	}
}

func TestTopologyFillUpLink(t *testing.T) {
	tree := NewEmptyTopology()
	tree.FillUpLink()
	if len(tree.UpLink) != 0 {
		t.Errorf("non-empty UpLink out of empty topology %v", tree.UpLink)
	}
	a, b, c, d, e := tree.AddNode(), tree.AddNode(), tree.AddNode(), tree.AddNode(), tree.AddNode()
	tree.AppendChild(a, b)
	tree.AppendChild(a, c)
	tree.AppendChild(e, d)
	tree.UpLink = []UpLink{{}, {}, {}}
	tree.FillUpLink()
	if len(tree.UpLink) != tree.NumNodes() {
		t.Errorf("Topology has %d nodes; UpLink has %d elements", tree.NumNodes(), len(tree.UpLink))
	}
	answer := []UpLink{{NoNodeId, 0}, {a, 0}, {a, 1}, {e, 0}, {NoNodeId, 0}}
	for i, uplink := range tree.UpLink {
		if uplink != answer[i] {
			t.Errorf("expected %v; got %v", answer[i], uplink)
		}
	}
}

func TestTopologyComponents(t *testing.T) {
	if c := NewEmptyTopology().Components(); len(c) != 0 {
		t.Errorf("expected empty components; got %v\n", c)
	}
	if c := NewRootedTopology().Components(); len(c) != 1 || len(c[0]) != 1 || c[0][0] != 0 {
		t.Errorf("expected single element component; got %v\n", c)
	}
	tree := NewRootedTopology()
	tree.AddNode()
	tree.AddNode()
	for c, ns := range tree.Components() {
		if len(ns) != 1 {
			t.Errorf("expected single element for component %d; got %v\n", c, ns)
		}
		if ns[0] != c {
			t.Errorf("expected [%d]; got %v\n", c, ns)
		}
	}
	tree.AppendChild(1, 2)
	c := tree.Components()
	if len(c) != 2 {
		t.Errorf("expected 2 components; got %d\n", len(c))
	}
	if len(c[0]) != 1 || len(c[1]) != 2 {
		t.Errorf("expected components with size 1 and 2; got %v and %v\n", c[0], c[1])
	}
}

func TestTopologyTopsort(t *testing.T) {
	topsortCases := []*Topology{
		NewEmptyTopology(), NewRootedTopology(),
		fromParents(1, []NodeId{1, NoNodeId, 1}),
		fromParents(1, []NodeId{1, NoNodeId, 1, NoNodeId, 3, 3}),
		fromParents(NoNodeId, []NodeId{1, NoNodeId, 1, NoNodeId, 3, 3}),
	}
	for _, tree := range topsortCases {
		save := tree.Copy()
		oldToNew := tree.Topsort()
		topologySanityCheck(tree, t)
		numComponents := len(tree.Components())
		if save.Root != NoNodeId && numComponents != 1 {
			t.Errorf("expected 1 component; got %d\n", numComponents)
		}
		if save.Root == NoNodeId && numComponents != 0 {
			t.Errorf("expected 0 component; got %d\n", numComponents)
		}
		for i := 0; i < save.NumNodes(); i++ {
			newNode := oldToNew[i]
			if newNode != NoNodeId {
				checkNodeRange(newNode, NodeId(tree.NumNodes()), t)
			}
		}
	}
}

func TestTopologyTopsortCycle(t *testing.T) {
	tree := NewEmptyTopology()
	a, b := tree.AddNode(), tree.AddNode()
	tree.Root = a
	tree.AppendChild(a, b)
	tree.AppendChild(b, a)
	func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Errorf("expected error; got nil")
			}
		}()
		tree.Topsort()
	}()
}

func TestTopologyDisconnect(t *testing.T) {
	disconnectCases := []struct {
		input, output *Topology
		remove        []bool
	}{
		{NewEmptyTopology(), NewEmptyTopology(), nil},
		{NewRootedTopology(), NewRootedTopology(), []bool{false}},
		{NewRootedTopology(), fromParents(NoNodeId, []NodeId{NoNodeId}), []bool{true}},
		{
			fromParents(0, []NodeId{NoNodeId, 0, 1, 0, 3, 4}),
			fromParents(NoNodeId, []NodeId{NoNodeId, 0, 1, 0, 3, 4}),
			[]bool{true, false, false, false, false, false},
		},
		{
			fromParents(0, []NodeId{NoNodeId, 0, 1, 0, 3, 4}),
			fromParents(0, []NodeId{NoNodeId, NoNodeId, NoNodeId, NoNodeId, NoNodeId, NoNodeId}),
			[]bool{false, true, true, true, true, true},
		},
		{
			fromParents(0, []NodeId{NoNodeId, 0, 1, 2, 3, 1, 5, 0, 7, 8, 7, 10}),
			fromParents(0, []NodeId{NoNodeId, NoNodeId, 1, 2, 3, 1, 5, 0, NoNodeId, 8, 7, 10}),
			[]bool{false, true, false, false, false, false, false, false, true, false, false, false},
		},
	}

	for _, c := range disconnectCases {
		c.input.Disconnect(c.remove)
		if !c.input.Equal(c.output) {
			t.Errorf("expected %v; got %v after disconnect with %v\n",
				*c.output, *c.input, c.remove)
		}
	}
}

func topologySanityCheck(tree *Topology, t *testing.T) {
	if tree.NumNodes() == 0 {
		if root := tree.Root; root != NoNodeId {
			t.Errorf("expected root %d; got %d\n", NoNodeId, root)
		}
		if len(tree.Children) != 0 {
			t.Errorf("expected empty children; got %v\n", tree.Children)
		}
	} else {
		upper := NodeId(tree.NumNodes())
		if tree.Root != NoNodeId {
			checkNodeRange(tree.Root, upper, t)
		}
		numEdges := 0
		for parent, children := range tree.Children {
			numEdges += len(children)
			checkNodeRange(NodeId(parent), upper, t)
			for _, child := range children {
				checkNodeRange(child, upper, t)
			}
		}
		numComponents := len(tree.Components())
		if numEdges+numComponents != tree.NumNodes() {
			t.Errorf("got %d edges, %d components but %d nodes; there are cycles\n",
				numEdges, numComponents, tree.NumNodes())
		}
	}
}

func checkNodeRange(n NodeId, upper NodeId, t *testing.T) {
	if !(0 <= n && n < upper) {
		t.Errorf("expected node within range [0, %d); got %d\n", upper, n)
	}
}

func fromParents(root NodeId, parent []NodeId) *Topology {
	ret := NewEmptyTopology()
	for _ = range parent {
		ret.AddNode()
	}
	for i, p := range parent {
		if p != NoNodeId {
			ret.AppendChild(p, NodeId(i))
		}
	}
	if root != NoNodeId {
		ret.Root = root
	}
	return ret
}
