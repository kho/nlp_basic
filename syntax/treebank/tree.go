package treebank

import (
	"bytes"
	"github.com/kho/nlp_basic/bimap"
	"github.com/kho/nlp_basic/syntax/heads"
)

// ParseTree is a tree topology with rich annotations of nodes stored
// as slices addressed by NodeIds from the topology. Certain
// annotations may not be available, in which case a nil slice is
// stored instead. Once a slice is assigned to a ParseTree, the tree
// owns the underlying memory.
type ParseTree struct {
	// Topology is the required field which should never be nil.
	Topology *Topology
	// The following may be nil or not up-to-date.
	Map      *bimap.Map // Map between label and Id
	Label    []string   // Node label as string
	Id       []int      // Node label as int
	Span     []Span     // A [left, right) input span of the node
	Head     []int      // The position of the head child of a give node; leaf's head child is some undefined value.
	HeadLeaf []NodeId   // The head leaf of a give node; leaf's head is itself
}

type Span struct{ Left, Right int }

// Group of constants that decides what to fill in ParseTree.Fill().
const (
	// When Label is available, always fills Id; otherwise use Id to
	// fill Label.
	FILL_LABEL_ID = 1 << iota
	FILL_SPAN
	FILL_HEAD
	FILL_HEAD_LEAF
	FILL_UP_LINK
	FILL_EVERYTHING = FILL_LABEL_ID | FILL_SPAN | FILL_HEAD | FILL_HEAD_LEAF | FILL_UP_LINK
)

// Fill fills annotations specified by the flags. m may be nil as long
// as the corresponding Remap call is valid. h may be nil as long as
// FILL_HEAD or FILL_HEAD_LEAF are not marked.
func (tree *ParseTree) Fill(flags int, m *bimap.Map, finder heads.HeadFinder) {
	if flags&FILL_LABEL_ID != 0 {
		if tree.Topology.NumNodes() == len(tree.Label) {
			tree.RemapByLabel(m)
		} else {
			tree.RemapById(m)
		}
	}
	if flags&FILL_SPAN != 0 {
		tree.FillSpan()
	}
	if flags&(FILL_HEAD|FILL_HEAD_LEAF) != 0 {
		tree.FillHead(finder)
	}
	if flags&FILL_HEAD_LEAF != 0 {
		tree.FillHeadLeaf()
	}
	if flags&FILL_UP_LINK != 0 {
		tree.Topology.FillUpLink()
	}
}

// RemapByLabel remaps Id by Label using the given mapping. If m is
// nil, the mapping that is already stored is used.
func (tree *ParseTree) RemapByLabel(m *bimap.Map) {
	if len(tree.Label) != tree.Topology.NumNodes() {
		panic("Label and Topology do not match in size")
	}
	if m == nil {
		if tree.Map == nil {
			panic("RemapByLabel without specifying a map")
		}
	} else {
		tree.Map = m
	}
	tree.Id = tree.Id[:0]
	tree.Map.AppendByString(tree.Label, &tree.Id)
}

// RemapById remaps Label by Id using the given mapping. If m is nil,
// the mapping that is already stored is used.
func (tree *ParseTree) RemapById(m *bimap.Map) {
	if len(tree.Id) != tree.Topology.NumNodes() {
		panic("Id and Topology do not match in size")
	}
	if m == nil {
		if tree.Map == nil {
			panic("RemapById without specifying a map")
		}
	} else {
		tree.Map = m
	}
	tree.Label = tree.Label[:0]
	tree.Map.AppendByInt(tree.Id, &tree.Label)
}

// FillSpan fills the Span slice.
func (tree *ParseTree) FillSpan() {
	numNodes := tree.Topology.NumNodes()
	if cap(tree.Span) >= numNodes {
		tree.Span = tree.Span[:numNodes]
	} else {
		tree.Span = make([]Span, numNodes)
	}
	if tree.Topology.Root != NoNodeId {
		dfsFillSpan(tree, tree.Topology.Root, 0)
	}
}

func dfsFillSpan(tree *ParseTree, node NodeId, left int) int {
	tree.Span[node].Left = left
	var right int
	if tree.Topology.Leaf(node) {
		right = left + 1
	} else {
		right = left
		for _, child := range tree.Topology.Children[node] {
			right = dfsFillSpan(tree, child, right)
		}
	}
	tree.Span[node].Right = right
	return right
}

// FillHead fills the Head slice with the given head finder. A valid
// Label slice must present.
func (tree *ParseTree) FillHead(finder heads.HeadFinder) {
	numNodes := tree.Topology.NumNodes()
	if len(tree.Label) != numNodes {
		panic("Label and Topology do not match in size")
	}
	children := make([]string, 0, 16)
	if cap(tree.Head) >= numNodes {
		tree.Head = tree.Head[:numNodes]
	} else {
		tree.Head = make([]int, numNodes)
	}
	for i := range tree.Head {
		node := NodeId(i)
		if tree.Topology.Leaf(node) {
			tree.Head[i] = -1
		} else {
			children = children[:0]
			for _, child := range tree.Topology.Children[node] {
				children = append(children, tree.Label[child])
			}
			tree.Head[i] = finder.FindHead(tree.Label[node], children)
		}
	}
}

// FillHeadLeaf fills the HeadLeaf slice. A valid Head slice must
// present.
func (tree *ParseTree) FillHeadLeaf() {
	numNodes := tree.Topology.NumNodes()
	if len(tree.Head) != numNodes {
		panic("Head and Topology do not match in size")
	}
	var hl []NodeId
	if cap(tree.HeadLeaf) >= numNodes {
		hl = tree.HeadLeaf[:numNodes]
	} else {
		hl = make([]NodeId, numNodes)
	}
	for i, h := range tree.Head {
		if h >= 0 {
			hl[i] = tree.Topology.Children[i][h]
		} else {
			hl[i] = NodeId(i)
		}
	}
	for i, h := range hl {
		hh := hl[h]
		for hh != hl[hh] {
			hh = hl[hh]
		}
		for hl[i] != hh {
			j := hl[i]
			hl[i] = hh
			i = int(j)
		}
	}
	tree.HeadLeaf = hl
}

// String writes out the tree in standard Treebank format. Label must
// be valid; or if Map and Id are available, Label will be constructed
// and used.
func (tree *ParseTree) String() string {
	if len(tree.Label) != tree.Topology.NumNodes() {
		if tree.Map != nil && len(tree.Id) == tree.Topology.NumNodes() {
			tree.RemapById(nil)
		} else {
			panic("Cannot get valid Label")
		}
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	buf.WriteByte('(')
	if tree.Topology.Root == NoNodeId {
		buf.WriteString("()")
	} else {
		dfsString(tree, tree.Topology.Root, buf)
	}
	buf.WriteByte(')')
	return buf.String()
}

// StringUnder writes out the tree under the given node in sexp.
func (tree *ParseTree) StringUnder(node NodeId) string {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	if node != NoNodeId {
		dfsString(tree, node, buf)
	}
	return buf.String()
}

// dfsString traverses a non-empty tree starting at node and writes
// the string representation to buf.
func dfsString(tree *ParseTree, node NodeId, buf *bytes.Buffer) {
	if tree.Topology.Leaf(node) {
		buf.WriteString(tree.Label[node])
	} else {
		buf.WriteByte('(')
		buf.WriteString(tree.Label[node])
		for _, child := range tree.Topology.Children[node] {
			buf.WriteByte(' ')
			dfsString(tree, child, buf)
		}
		buf.WriteByte(')')
	}
}

// TopSort topologically sorts the tree and re-organizes the optional
// properties into a top-down order. Invalid properties are cleared to
// nil. The mapping from old NodeId to new ones is returned.
func (tree *ParseTree) Topsort() []NodeId {
	oldNumNodes := tree.Topology.NumNodes()
	oldToNew := tree.Topology.Topsort()
	numNodes := tree.Topology.NumNodes()

	var (
		newLabel    []string
		newId       []int
		newSpan     []Span
		newHead     []int
		newHeadLeaf []NodeId
	)

	mapLabel := len(tree.Label) == oldNumNodes
	mapId := len(tree.Id) == oldNumNodes
	mapSpan := len(tree.Span) == oldNumNodes
	mapHead := len(tree.Head) == oldNumNodes
	mapHeadLeaf := len(tree.HeadLeaf) == oldNumNodes

	if mapLabel {
		newLabel = make([]string, numNodes)
	}
	if mapId {
		newId = make([]int, numNodes)
	}
	if mapSpan {
		newSpan = make([]Span, numNodes)
	}
	if mapHead {
		newHead = make([]int, numNodes)
	}
	if mapHeadLeaf {
		newHeadLeaf = make([]NodeId, numNodes)
	}

	for o, n := range oldToNew {
		if n == NoNodeId {
			continue
		}
		if mapLabel {
			newLabel[n] = tree.Label[o]
		}
		if mapId {
			newId[n] = tree.Id[o]
		}
		if mapSpan {
			newSpan[n] = tree.Span[o]
		}
		if mapHead {
			newHead[n] = tree.Head[o]
		}
		if mapHeadLeaf {
			newHeadLeaf[n] = oldToNew[tree.HeadLeaf[o]]
		}
	}

	tree.Id = newId
	tree.Label = newLabel
	tree.Span = newSpan
	tree.Head = newHead
	tree.HeadLeaf = newHeadLeaf

	return oldToNew
}

// StripAnnotation strips off rich treebank annotation (e.g. NP-1,
// NP-SUBJ, etc) and returns the tree itself.
func (tree *ParseTree) StripAnnotation() *ParseTree {
	for i, label := range tree.Label {
		node := NodeId(i)
		if tree.Topology.Leaf(node) {
			// Only strip if starting with * (i.e. *pro*, *T*, *PRO*, etc.)
			if len(label) > 0 && label[0] == '*' {
				tree.Label[i] = stripLabelAnnotation(label)
			}
		} else {
			// Do not strip if this is -NONE-
			if len(label) > 0 && label[0] != '-' {
				tree.Label[i] = stripLabelAnnotation(label)
			}
		}
	}
	return tree
}

func stripLabelAnnotation(label string) string {
	i := 0
	for i < len(label) && label[i] != '-' && label[i] != '=' {
		i++
	}
	return label[:i]
}

// RemoveNone removes -NONE- and its unary ancestors.
func (tree *ParseTree) RemoveNone() *ParseTree {
	tree.Topsort()
	numNodes := tree.Topology.NumNodes()
	invisible := make([]bool, numNodes)
	// Mark in bottom-up order
	for i := numNodes; i > 0; i-- {
		node := NodeId(i - 1)
		label := tree.Label[node]
		if label == "-NONE-" {
			invisible[node] = true
		} else if len(tree.Topology.Children[node]) > 0 {
			invisible[node] = true
			for _, child := range tree.Topology.Children[node] {
				if !invisible[child] {
					invisible[node] = false
					break
				}
			}
			if invisible[node] {
				for _, child := range tree.Topology.Children[node] {
					invisible[child] = false
				}
			}
		}
	}
	tree.Topology.Disconnect(invisible)
	tree.Topsort()
	return tree
}
