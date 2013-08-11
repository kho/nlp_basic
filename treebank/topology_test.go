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
	root := rooted.Root()
	if root != 0 {
		t.Fatalf("NewRootedTopology created a topology with root %d\n",
			root)
	}

	if parent := rooted.Parent(root); parent != NoNodeId {
		t.Errorf("NewRootedTopology created a topology of root with parent %d\n",
			parent)
	}

	if children := rooted.Children(root); len(children) != 0 {
		t.Errorf("NewRootedTopology created a topology of root with children %v\n",
			children)
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

	oldRoot := t1.Root()
	newRoot := t1.AddNode()
	t1.SetRoot(newRoot)
	t1.AppendChild(newRoot, oldRoot)

	if size1, size2 := t1.NumNodes(), t2.NumNodes(); size1 == size2 {
		t.Errorf("sizes do not differ after modification: %d vs %d\n",
			size1, size2)
	}

	if root1, root2 := t1.Root(), t2.Root(); root1 == root2 {
		t.Errorf("roots do not differ after modification: %d vs %d\n",
			root1, root2)
	}

	if parent1, parent2 := t1.Parent(0), t2.Parent(0); parent1 == parent2 {
		t.Errorf("parents do not differ after modification: %d vs %d\n",
			parent1, parent2)
	}
}

func TestTopologyAddNode(t *testing.T) {
	const numInserts = 100
	trees := []*Topology{NewEmptyTopology(), NewRootedTopology()}
	for _, tree := range trees {
		oldSize := tree.NumNodes()
		oldRoot := tree.Root()
		oldTree := tree.Copy()
		for i := 0; i < numInserts; i++ {
			node := tree.AddNode()
			if node != NodeId(i+oldSize) {
				t.Errorf("expected new id %d; got %d\n", i+oldSize, node)
			}
			if parent := tree.Parent(node); parent != NoNodeId {
				t.Errorf("expected parent %d; got %d\n", NoNodeId, parent)
			}
			if children := tree.Children(node); len(children) != 0 {
				t.Errorf("expected tree children; got %v\n", children)
			}
		}
		topologySanityCheck(tree, t)
		if root := tree.Root(); root != oldRoot {
			t.Errorf("expected root %d; got %d\n", oldRoot, root)
		}
		if numNodes := tree.NumNodes(); numNodes != oldSize+numInserts {
			t.Errorf("expected %d nodes after AddNode(); got %d\n",
				oldSize+numInserts, numNodes)
		}
		for i := 0; i < oldSize; i++ {
			n := NodeId(i)

			if oldParent, parent := oldTree.Parent(n), tree.Parent(n); oldParent != parent {
				t.Errorf("parent of old node %d changed from %d to %d\n",
					n, oldParent, parent)
			}
		}
	}
}

func TestTopologySetRoot(t *testing.T) {
	tree := NewRootedTopology()
	tree.AppendChild(tree.Root(), tree.AddNode())
	newRoot := tree.AddNode()
	save := tree.Copy()
	tree.SetRoot(newRoot)
	topologySanityCheck(tree, t)

	if oldSize, newSize := save.NumNodes(), tree.NumNodes(); oldSize != newSize {
		t.Errorf("num nodes changed from %d to %d\n", oldSize, newSize)
	}

	if curRoot := tree.Root(); curRoot != newRoot {
		t.Errorf("expected new root %d; got %d\n", newRoot, curRoot)
	}

	if rootParent := tree.Parent(newRoot); rootParent != NoNodeId {
		t.Errorf("new root %d has parent %d\n", newRoot, rootParent)
	}

	for i := 0; i < tree.NumNodes(); i++ {
		n := NodeId(i)
		if n == save.Root() || n == tree.Root() {
			continue
		}
		if oldParent, newParent := save.Parent(n), tree.Parent(n); oldParent != newParent {
			t.Errorf("SetRoot() changed %d's parent from %d to %d\n",
				n, oldParent, newParent)
		}
	}
}

func TestTopologyAppendChild(t *testing.T) {
	tree := NewRootedTopology()
	tree.AddNode()
	tree.AddNode()
	tree.AppendChild(2, 1)
	topologySanityCheck(tree, t)
	p0, p1, p2 := tree.Parent(0), tree.Parent(1), tree.Parent(2)
	a0, a1, a2 := NoNodeId, NodeId(2), NoNodeId
	if p0 != a0 {
		t.Errorf("expected parent %d; got %d\n", a0, p0)
	}
	if p1 != a1 {
		t.Errorf("expected parent %d; got %d\n", a1, p1)
	}
	if p2 != a2 {
		t.Errorf("expected parent %d; got %d\n", a2, p2)
	}
	// Appending root
	func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Error("expected panic; got nil")
			}
		}()
		tree.AppendChild(1, 0)
	}()
	// Appending twice
	func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Error("expected panic; got nil")
			}
		}()
		tree.AppendChild(0, 1)
	}()
}

func TestComponents(t *testing.T) {
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
		t.Log(*tree)
		save := tree.Copy()
		oldToNew := tree.Topsort()
		topologySanityCheck(tree, t)
		numComponents := len(tree.Components())
		if save.Root() != NoNodeId && numComponents != 1 {
			t.Errorf("expected 1 component; got %d\n", numComponents)
		}
		if save.Root() == NoNodeId && numComponents != 0 {
			t.Errorf("expected 0 component; got %d\n", numComponents)
		}
		for i := 0; i < tree.NumNodes(); i++ {
			child, parent := NodeId(i), tree.Parent(NodeId(i))
			if parent == NoNodeId {
				if child != tree.Root() {
					t.Errorf("dangling node %d after Topsort(); tree is %v\n",
						child, *tree)
				}
			} else {
				if child <= parent {
					t.Errorf("not in topological order: child %d, parent %d\n",
						child, parent)
				}
			}
		}
		for i := 0; i < save.NumNodes(); i++ {
			oldNode := NodeId(i)
			oldParent := save.Parent(oldNode)
			newNode := oldToNew[i]
			if newNode != NoNodeId {
				checkNodeRange(newNode, NodeId(tree.NumNodes()), t)
				newParent := tree.Parent(newNode)
				if oldParent == NoNodeId {
					if newParent != NoNodeId {
						t.Errorf("old parent is %d; but new parent is %d\n", oldParent, newParent)
					}
				} else if oldToNew[oldParent] != newParent {
					t.Errorf("expected old %d maps to new %d; got %d\n",
						oldParent, newParent, oldToNew[oldParent])
				}
			}
		}
	}
}

func TestTopologyDisconnect(t *testing.T) {
	disconnectCases := []struct {
		input, output *Topology
		remove        []bool
	}{
		{NewEmptyTopology(), NewEmptyTopology(), nil},
		{NewRootedTopology(), NewRootedTopology(), []bool{false}},
		{NewRootedTopology(), fromParents(NoNodeId, []NodeId{NoNodeId}), []bool{true}},
		{fromParents(0, []NodeId{NoNodeId, 0, 1, 0, 3, 4}),
			fromParents(NoNodeId, []NodeId{NoNodeId, 0, 1, 0, 3, 4}),
			[]bool{true, false, false, false, false, false}},
		{fromParents(0, []NodeId{NoNodeId, 0, 1, 0, 3, 4}),
			fromParents(0, []NodeId{NoNodeId, NoNodeId, NoNodeId, NoNodeId, NoNodeId, NoNodeId}),
			[]bool{false, true, true, true, true, true}},
	}

	for _, c := range disconnectCases {
		c.input.Disconnect(c.remove)
		if !c.input.Equal(c.output) {
			t.Errorf("expected %q; got %q after disconnect with %v\n",
				*c.output, *c.input, c.remove)
		}
	}
}

func topologySanityCheck(tree *Topology, t *testing.T) {
	if tree.NumNodes() == 0 {
		if root := tree.Root(); root != NoNodeId {
			t.Errorf("expected root %d; got %d\n", NoNodeId, root)
		}
		if len(tree.parent) != 0 {
			t.Errorf("expected empty parent; got %v\n", tree.parent)
		}
		if len(tree.children) != 0 {
			t.Errorf("expected empty children; got %v\n", tree.children)
		}
	} else {
		upper := NodeId(tree.NumNodes())
		if tree.Root() != NoNodeId {
			checkNodeRange(tree.Root(), upper, t)
		}
		for child, parent := range tree.parent {
			checkNodeRange(NodeId(child), upper, t)
			if parent != NoNodeId {
				checkNodeRange(parent, upper, t)
			}
		}
		numEdges := 0
		for parent, children := range tree.children {
			numEdges += len(children)
			checkNodeRange(NodeId(parent), upper, t)
			for _, child := range children {
				checkNodeRange(child, upper, t)
				if realParent := tree.parent[child]; realParent != NodeId(parent) {
					t.Errorf("children and parent mismatch: %d is stored as child of %d but parent[%d] = %d\n",
						child, parent, child, realParent)
				}
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
		ret.SetRoot(root)
	}
	return ret
}
