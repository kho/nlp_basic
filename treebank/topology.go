package treebank

import (
	"fmt"
)

// Topology stores the tree structure. A topology consists of N nodes,
// with id from 0 to (N-1) where one of them is the root.
type Topology struct {
	root     NodeId
	parent   []NodeId
	children [][]NodeId
}

// NodeId is the tree node id in a Topology. Normal values are
// non-negative. Negative values are used to represent special
// results.
type NodeId int32

// Special NodeId values
const (
	// NoNodeId represents a non-existent node.
	NoNodeId NodeId = -1 - iota
)

// NewRootedTopology creates a topology with a single root node.
func NewRootedTopology() *Topology {
	return &Topology{
		root:     0,
		parent:   []NodeId{NoNodeId},
		children: [][]NodeId{nil},
	}
}

// NewEmptyTopology creates an empty topology.
func NewEmptyTopology() *Topology {
	return &Topology{root: NoNodeId}
}

// Copy creates a deep copy of the given topology. Node: use this
// instead of simple assignment for copying.
func (t *Topology) Copy() *Topology {
	parent := make([]NodeId, len(t.parent))
	children := make([][]NodeId, len(t.children))
	copy(parent, t.parent)
	for i := range t.children {
		children[i] = make([]NodeId, len(t.children[i]))
		copy(children[i], t.children[i])
	}
	return &Topology{t.root, parent, children}
}

// Equal tests if one topology holds identical contents compared with
// the other.
func (t *Topology) Equal(s *Topology) bool {
	if t.root != s.root || t.NumNodes() != s.NumNodes() {
		return false
	}
	for i := 0; i < t.NumNodes(); i++ {
		n := NodeId(i)
		if t.Parent(n) != s.Parent(n) {
			return false
		}
	}
	return true
}

// NumNodes returns the number of nodes in t.
func (t *Topology) NumNodes() int {
	return len(t.children)
}

// Root returns the root of t; or NoNodeId if t is empty.
func (t *Topology) Root() NodeId {
	return t.root
}

// Parent returns the parent node of a node or NoNodeId when the node
// does not yet have a parent (e.g. dangling nodes or root).
func (t *Topology) Parent(n NodeId) NodeId {
	return t.parent[n]
}

// Children returns the children node ids of a node. The returned
// slice should not be shorten or appended. But the elements can be
// rearranged in-place to change the ordering of nodes.
func (t *Topology) Children(n NodeId) []NodeId {
	return t.children[n]
}

// Leaf tests whether the given node is a leaf.
func (t *Topology) Leaf(n NodeId) bool {
	return len(t.children[n]) == 0
}

// PreTerminal tests whether the given node is a pre-terminal,
// i.e. the POS tag node dominating the leaf.
func (t *Topology) PreTerminal(n NodeId) bool {
	return len(t.children[n]) == 1 && t.Leaf(t.children[n][0])
}

// AddNode adds a node without a parent (i.e. a dangling node outside
// the tree) to the topology and returns the node id of the new node.
func (t *Topology) AddNode() NodeId {
	id := NodeId(t.NumNodes())
	t.parent = append(t.parent, NoNodeId)
	t.children = append(t.children, nil)
	return id
}

// SetRoot changes the root of t to r.
func (t *Topology) SetRoot(r NodeId) {
	t.root = r
	t.parent[r] = NoNodeId
}

// AppendChild appends b as the rightmost child of a. b must not be
// root or already have a parent because this may create cyclicity.
func (t *Topology) AppendChild(a NodeId, b NodeId) {
	if t.parent[b] != NoNodeId {
		panic(fmt.Sprintf("node %d already has a parent: %d", b, t.parent[b]))
	}
	if b == t.root {
		panic(fmt.Sprintf("root %d cannot be set to a child of %d", b, a))
	}
	t.children[a] = append(t.children[a], b)
	t.parent[b] = a
}

// Components returns the connect components inside the topology as a
// map from roots to their nodes.
func (t *Topology) Components() map[NodeId][]NodeId {
	p := make([]NodeId, t.NumNodes())
	for i := range p {
		p[i] = NodeId(i)
	}
	for child, parent := range t.parent {
		if parent == NoNodeId {
			continue
		}
		union(parent, NodeId(child), p)
	}
	m := make(map[NodeId][]NodeId)
	for i, c := range p {
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
		m := n
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
// slice. Nodes not in the tree from the root are removed and their
// mapping is set to NoNodeId in the return value.
func (t *Topology) Topsort() []NodeId {
	traverse := make([]NodeId, 0, t.NumNodes())
	if t.Root() != NoNodeId {
		dfsTraverse(t, t.Root(), &traverse)
	}
	oldToNew := remap(t, traverse)
	return oldToNew
}

func dfsTraverse(t *Topology, n NodeId, ns *[]NodeId) {
	*ns = append(*ns, n)
	for _, c := range t.children[n] {
		dfsTraverse(t, c, ns)
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
	newParent := make([]NodeId, newSize)
	newChildren := make([][]NodeId, newSize)
	for n, o := range newToOld {
		newParent[n] = t.parent[o]
		if newParent[n] != NoNodeId {
			newParent[n] = oldToNew[newParent[n]]
		}
		newChildren[n] = t.children[o]
		for i := range newChildren[n] {
			newChildren[n][i] = oldToNew[newChildren[n][i]]
		}
	}
	t.root = newRoot
	t.parent = newParent
	t.children = newChildren
	return oldToNew
}

// Disconnect removes the nodes marked as true in remove from their
// parents. Returns the number of nodes that is disconnected.
func (t *Topology) Disconnect(remove []bool) int {
	numDisconnected := 0
	for i, r := range remove {
		if !r {
			continue
		}
		node := NodeId(i)
		if node == t.root {
			t.root = NoNodeId
		} else {
			parent := t.parent[node]
			if parent == NoNodeId {
				continue
			}
			t.parent[node] = NoNodeId
			j := 0
			for j < len(t.children[parent]) && t.children[parent][j] != node {
				j++
			}
			copy(t.children[parent][j:], t.children[parent][j+1:])
			t.children[parent] = t.children[parent][:len(t.children[parent])-1]
		}
		numDisconnected++
	}
	return numDisconnected
}
