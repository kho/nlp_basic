package treebank

import (
	"bytes"
)

// Node is a parse tree node. Its parts are references so itself is
// quite small and should be passed by value unless being modified.
type Node struct {
	Label    string
	Children []Node
}

func (node Node) String() string {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	buf.WriteByte('(')
	dfsString(node, buf)
	buf.WriteByte(')')
	return buf.String()
}

// Copy deep-copies a Node.
func Copy(tree Node) Node {
	var ret Node
	ret.Label = tree.Label
	ret.Children = make([]Node, len(tree.Children))
	for i, child := range tree.Children {
		ret.Children[i] = Copy(child)
	}
	return ret
}

// IsPreTerminal tests if the node is a pre-terminal
func IsPreTerminal(node Node) bool {
	return len(node.Children) == 1 && len(node.Children[0].Children) == 0
}

// StripAnnotation strips off rich treebank annotation (e.g. NP-1,
// NP-SUBJ, etc) and returns the tree itself.
func (node *Node) StripAnnotation() *Node {
	stripAnnotation(node)
	return node
}

// RemoveNone removes -NONE- and its unary ancestors.
func (node *Node) RemoveNone() *Node {
	removeNone(node)
	return node
}

func dfsString(node Node, buf *bytes.Buffer) {
	if len(node.Children) > 0 {
		buf.WriteByte('(')
		buf.WriteString(node.Label)
		for i := range node.Children {
			buf.WriteByte(' ')
			dfsString(node.Children[i], buf)
		}
		buf.WriteByte(')')
	} else {
		buf.WriteString(node.Label)
	}
}

func stripAnnotation(node *Node) {
	if len(node.Children) == 0 {
		// Only strip if starting with * (i.e. *pro*, *T*, *PRO*, etc.)
		if len(node.Label) > 0 && node.Label[0] == '*' {
			node.Label = stripLabelAnnotation(node.Label)
		}
	} else {
		// Do not strip if this is -NONE-
		if len(node.Label) > 0 && node.Label[0] != '-' {
			node.Label = stripLabelAnnotation(node.Label)
		}
		for i := range node.Children {
			stripAnnotation(&node.Children[i])
		}
	}
}

func stripLabelAnnotation(label string) string {
	i := 0
	for i < len(label) && label[i] != '-' && label[i] != '=' {
		i++
	}
	return label[:i]
}

func removeNone(node *Node) {
	old_num_children := len(node.Children)
	for i := range node.Children {
		if node.Children[i].Label != "-NONE-" {
			removeNone(&node.Children[i])
		}
	}
	var children []Node
	for i := range node.Children {
		if node.Children[i].Label != "-NONE-" {
			children = append(children, node.Children[i])
		}
	}
	node.Children = children
	if len(children) == 0 && old_num_children > 0 {
		node.Label = "-NONE-"
	}
}
