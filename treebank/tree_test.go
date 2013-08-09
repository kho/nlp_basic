package treebank

import (
	"testing"
)

var copyCases = []Node{
	{"a", nil},
	{"a", []Node{{"b", nil}, {"c", nil}}},
	{"a", []Node{{"b", []Node{{"c", nil}}}}},
}

var modifyFuncs = []func(*Node){clearLabels, addChild}

func TestCopy(t *testing.T) {
	for _, tree := range copyCases {
		for _, f := range modifyFuncs {
			treeCopy := Copy(tree)
			if !equiv(tree, treeCopy) {
				t.Errorf("expected %v; got %v after copy\n", tree, treeCopy)
			}
			f(&treeCopy)
			if equiv(tree, treeCopy) {
				t.Errorf("modification is propagated between copies: %v\n", tree)
			}
		}
	}
}

func BenchmarkCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := range copyCases {
			_ = Copy(copyCases[j])
		}
	}
}

func clearLabels(node *Node) {
	node.Label = ""
	for i := range node.Children {
		clearLabels(&node.Children[i])
	}
}

func addChild(node *Node) {
	for i := range node.Children {
		addChild(&node.Children[i])
	}
	node.Children = append(node.Children, Node{"x", nil})
}
