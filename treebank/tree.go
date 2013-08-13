package treebank

import (
	"bytes"
)

// LabelTree is a parse tree with its labels as strings
type LabelTree struct {
	Topology *Topology
	Label    []string
}

func (tree *LabelTree) NumNodes() int {
	return len(tree.Label)
}

// String writes out the tree in standard Treebank format
func (tree *LabelTree) String() string {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	buf.WriteByte('(')
	if tree.Topology.Root() == NoNodeId {
		buf.WriteString("()")
	} else {
		dfsString(tree, tree.Topology.Root(), buf)
	}
	buf.WriteByte(')')
	return buf.String()
}

// StringUnder writes out the tree under the given node in sexp.
func (tree *LabelTree) StringUnder(node NodeId) string {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	if node != NoNodeId {
		dfsString(tree, node, buf)
	}
	return buf.String()
}

// dfsString traverses a non-empty tree starting at node and writes
// the string representation to buf.
func dfsString(tree *LabelTree, node NodeId, buf *bytes.Buffer) {
	if tree.Topology.Leaf(node) {
		buf.WriteString(tree.Label[node])
	} else {
		buf.WriteByte('(')
		buf.WriteString(tree.Label[node])
		for _, child := range tree.Topology.Children(node) {
			buf.WriteByte(' ')
			dfsString(tree, child, buf)
		}
		buf.WriteByte(')')
	}
}

// TopSort topologically sorts the tree and re-organizes the labels
// into a top-down order.
func (tree *LabelTree) Topsort() *LabelTree {
	oldToNew := tree.Topology.Topsort()
	newLabel := make([]string, tree.Topology.NumNodes())
	for oldId, label := range tree.Label {
		newId := oldToNew[oldId]
		if newId != NoNodeId {
			newLabel[newId] = label
		}
	}
	tree.Label = newLabel
	return tree
}

// StripAnnotation strips off rich treebank annotation (e.g. NP-1,
// NP-SUBJ, etc) and returns the tree itself.
func (tree *LabelTree) StripAnnotation() *LabelTree {
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
func (tree *LabelTree) RemoveNone() *LabelTree {
	tree.Topsort()
	invisible := make([]bool, tree.NumNodes())
	// Mark in bottom-up order
	for i := tree.NumNodes(); i > 0; i-- {
		node := NodeId(i - 1)
		label := tree.Label[node]
		if label == "-NONE-" {
			invisible[node] = true
		} else if len(tree.Topology.Children(node)) > 0 {
			invisible[node] = true
			for _, child := range tree.Topology.Children(node) {
				if !invisible[child] {
					invisible[node] = false
					break
				}
			}
			if invisible[node] {
				for _, child := range tree.Topology.Children(node) {
					invisible[child] = false
				}
			}
		}
	}
	tree.Topology.Disconnect(invisible)
	tree.Topsort()
	return tree
}
