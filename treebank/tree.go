package treebank

// Node is a parse tree node. Its parts are references so itself is
// quite small and should be passed by value unless being modified.
type Node struct {
	Label    string
	Children []Node
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
