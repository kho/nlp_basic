package treebank

type Node struct {
	Label    string
	Children []Node
}

func Copy(tree Node) Node {
	var ret Node
	ret.Label = tree.Label
	ret.Children = make([]Node, len(tree.Children))
	for i, child := range tree.Children {
		ret.Children[i] = Copy(child)
	}
	return ret
}
