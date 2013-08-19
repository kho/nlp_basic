package treebank

// Topology stores the tree structure. A topology consists of N nodes,
// with id from 0 to (N-1) forming a forest. The tree under Root is
// the tree that is represented by the Topology. Or when Root is
// NoNodeId, the empty tree is represented (possibly by a non-empty
// Topology). The user should not modify any field other than Root.
type Topology struct {
	Root     NodeId
	Children [][]NodeId
	// UpLink is the link to the parent of a node. This is
	// optional. When it is not present, it is set to nil. However, when
	// it is present, it is not updated when the Topology is modified
	// (e.g. via AddNode(), AppendChild(), Disconnect()). On the other
	// hand, all Topology methods do not read from this. It is thus
	// recommended to FillUpLink() only after the Topology is finalized.
	UpLink []UpLink
}

// NodeId is the tree node id in a Topology. Normal values are
// non-negative. Negative values are used to represent special
// results.
type NodeId int32

// UpLink is the link to the parent of a node.
type UpLink struct {
	// Parent is the parent node id. A root (not only the Root of a
	// Topology) has NoNodeId as its Parent.
	Parent NodeId
	// NthChild is the position of a node in its parent's children
	// slice. This may hold arbitrary value when Parent is NoNodeId.
	NthChild int
}

// Special NodeId values
const (
	// NoNodeId represents a non-existent node.
	NoNodeId NodeId = -1
)

// NewRootedTopology creates a topology with a single root node.
func NewRootedTopology() *Topology {
	return &Topology{Root: 0, Children: [][]NodeId{nil}}
}

// NewEmptyTopology creates an empty topology.
func NewEmptyTopology() *Topology {
	return &Topology{Root: NoNodeId}
}

// Copy creates a deep copy of the given topology. Use this instead of
// simple assignment for copying. UpLink is *not* copied, following
// the convention that no Topology method reads it.
func (t *Topology) Copy() *Topology {
	var children [][]NodeId
	if len(t.Children) != 0 {
		children = make([][]NodeId, len(t.Children))
		for i := range t.Children {
			// Only create a slice for internal nodes
			if len(t.Children[i]) != 0 {
				children[i] = make([]NodeId, len(t.Children[i]))
				copy(children[i], t.Children[i])
			}
		}
	}
	return &Topology{t.Root, children, nil}
}

// Equal tests if one topology holds identical contents compared with
// the other. The contents in UpLink are ignored.
func (t *Topology) Equal(s *Topology) bool {
	if t.Root != s.Root || t.NumNodes() != s.NumNodes() {
		return false
	}
	numNodes := t.NumNodes()
	for i := 0; i < numNodes; i++ {
		if len(t.Children[i]) != len(s.Children[i]) {
			return false
		}
		for j, v := range t.Children[i] {
			if v != s.Children[i][j] {
				return false
			}
		}
	}
	return true
}

// NumNodes returns the total number of nodes in t.
func (t *Topology) NumNodes() int {
	return len(t.Children)
}

// FillUpLink fills the UpLink information.
func (t *Topology) FillUpLink() {
	numNodes := t.NumNodes()
	if cap(t.UpLink) < numNodes {
		t.UpLink = make([]UpLink, numNodes)
	} else {
		t.UpLink = t.UpLink[:numNodes]
	}
	for i := 0; i < numNodes; i++ {
		t.UpLink[i].Parent = NoNodeId
	}
	for parent, children := range t.Children {
		for nth, child := range children {
			t.UpLink[child] = UpLink{NodeId(parent), nth}
		}
	}
}

// Leaf tests whether the given node is a leaf in its own tree.
func (t *Topology) Leaf(n NodeId) bool {
	return len(t.Children[n]) == 0
}

// PreTerminal tests whether the given node is a pre-terminal in its
// own tree, i.e. the POS tag node dominating the leaf.
func (t *Topology) PreTerminal(n NodeId) bool {
	return len(t.Children[n]) == 1 && t.Leaf(t.Children[n][0])
}

// AddNode adds a node without a parent (i.e. forming a singleton
// tree) to the topology and returns the node id of the new node. The
// newly added node does not have Parent information.
func (t *Topology) AddNode() NodeId {
	id := NodeId(t.NumNodes())
	t.Children = append(t.Children, nil)
	return id
}

// AppendChild appends child as the rightmost child of parent. The
// user must ensure that child does not already have a parent because
// this creates cyclicity. However, child may be Root, in which case
// the Topology still represents the subtree under child.
func (t *Topology) AppendChild(parent NodeId, child NodeId) {
	t.Children[parent] = append(t.Children[parent], child)
}

// Components returns the connect components inside the topology as a
// map from roots to their nodes. This does not modify the Topology.
func (t *Topology) Components() map[NodeId][]NodeId {
	p := make([]NodeId, t.NumNodes())
	for i := range p {
		p[i] = NodeId(i)
	}
	for parent, children := range t.Children {
		for _, child := range children {
			union(NodeId(parent), NodeId(child), p)
		}
	}
	m := make(map[NodeId][]NodeId)
	for i := range p {
		c := find(NodeId(i), p)
		m[c] = append(m[c], NodeId(i))
	}
	return m
}

func find(n NodeId, p []NodeId) NodeId {
	r := n
	for p[r] != r {
		r = p[r]
	}
	for n != r {
		m := p[n]
		p[n] = r
		n = m
	}
	return r
}

func union(parent NodeId, child NodeId, p []NodeId) {
	a := find(parent, p)
	b := find(child, p)
	p[b] = a
}

// Topsort toplogically sorts the topology in top-down order and
// returns the mapping from old node ids to new node ids as a
// slice. Nodes not in the tree under Root are removed and their
// mapping is set to NoNodeId in the return value. Panics if there is
// cycle.
func (t *Topology) Topsort() []NodeId {
	traverse := make([]NodeId, 0, t.NumNodes())
	visited := make([]bool, t.NumNodes())
	if t.Root != NoNodeId {
		dfsTraverse(t, t.Root, &traverse, visited)
	}
	oldToNew := remap(t, traverse)
	return oldToNew
}

func dfsTraverse(t *Topology, n NodeId, ns *[]NodeId, visited []bool) {
	if visited[n] {
		panic("cycle in Topology")
	}
	visited[n] = true
	*ns = append(*ns, n)
	for _, c := range t.Children[n] {
		dfsTraverse(t, c, ns, visited)
	}
}

func remap(t *Topology, newToOld []NodeId) []NodeId {
	// Build the mapping to new node ids.
	oldToNew := make([]NodeId, t.NumNodes())
	for i := range oldToNew {
		oldToNew[i] = NoNodeId
	}
	for n, o := range newToOld {
		oldToNew[o] = NodeId(n)
	}
	// Build new topology; reuse the children slices.
	newRoot := NodeId(0)
	newSize := len(newToOld)
	if newSize == 0 {
		// New tree is empty
		newRoot = NoNodeId
	}
	newChildren := make([][]NodeId, newSize)
	for n, o := range newToOld {
		newChildren[n] = t.Children[o]
		for i := range newChildren[n] {
			newChildren[n][i] = oldToNew[newChildren[n][i]]
		}
	}
	t.Root = newRoot
	t.UpLink = nil
	t.Children = newChildren
	return oldToNew
}

// Disconnect disconnects nodes marked as true in remove from their
// parents.
func (t *Topology) Disconnect(remove []bool) {
	if t.Root != NoNodeId && remove[t.Root] {
		t.Root = NoNodeId
	}
	for parent, children := range t.Children {
		w := 0 // write position
		for _, child := range children {
			if !remove[child] {
				children[w] = child
				w++
			}
		}
		t.Children[parent] = children[:w]
	}
}
