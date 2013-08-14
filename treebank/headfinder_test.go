package treebank

import (
	"testing"
)

func TestHeadRule(t *testing.T) {
	rules := []*HeadRule{
		NewHeadRule(HEAD_INITIAL, nil),
		NewHeadRule(HEAD_INITIAL, []string{"a"}),
		NewHeadRule(HEAD_FINAL, []string{"a", "b"}),
	}

	labels := []string{"a", "b", "c"}
	priorities := [][]int{{0, 0, 0}, {0, 1, 1}, {0, 1, 2}}

	for i, rule := range rules {
		for j, label := range labels {
			if p := rule.LabelPriority(label); p != priorities[i][j] {
				t.Errorf("expected %d; got %d\n", priorities[i][j], p)
			}
		}
	}

	// panic when direction is UNKNOWN
	func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Error("expected error; got nil")
			}
		}()
		NewHeadRule(UNKNOWN, nil)
	}()
}

func TestTableHeadFinder(t *testing.T) {
	tables := []*TableHeadFinder{
		&TableHeadFinder{nil, HEAD_INITIAL},
		&TableHeadFinder{nil, HEAD_FINAL},
		&TableHeadFinder{
			map[string]*HeadRule{
				"a": NewHeadRule(HEAD_INITIAL, []string{"a", "b", "c"}),
				"b": NewHeadRule(HEAD_FINAL, []string{"a", "b", "c"}),
			},
			UNKNOWN,
		},
	}

	inputs := []struct {
		parent   string
		children []string
	}{
		{"a", []string{"b"}}, {"a", []string{"b", "a"}}, {"a", []string{"c", "a", "b", "a"}},
		{"b", []string{"b"}}, {"b", []string{"a", "b"}}, {"b", []string{"a", "b", "a", "c"}},
	}
	outputs := [][]int{
		{0, 0, 0, 0, 0, 0},
		{0, 1, 3, 0, 1, 3},
		{0, 1, 1, 0, 0, 2},
	}

	for i, table := range tables {
		for j, input := range inputs {
			if head := table.FindHead(input.parent, input.children); head != outputs[i][j] {
				t.Errorf("expected %d; got %d as head of %q -> %q\n", outputs[i][j], head, input.parent, input.children)
			}
		}
	}

	// panic when Fallback is unknown
	func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Error("expected error; got nil")
			}
		}()
		(&TableHeadFinder{nil, UNKNOWN}).FindHead("a", []string{"a"})
	}()

	// panic when finding the head of a leaf
	func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Error("expected error; got nil")
			}
		}()
		(&TableHeadFinder{nil, HEAD_INITIAL}).FindHead("a", nil)
	}()
}

func TestEnglishHeadFinderNP(t *testing.T) {
	finder := NewEnglishHeadFinder()
	inputs := []struct {
		children []string
		head     int
	}{
		{[]string{"NN", "POS"}, 1},
		{[]string{"NP", "NN"}, 1},
		{[]string{"NP", "NNP"}, 1},
		{[]string{"NP", "NNPS"}, 1},
		{[]string{"NP", "NNS"}, 1},
		{[]string{"NP", "NX"}, 1},
		{[]string{"NP", "JJR"}, 1},
		{[]string{"NP", "$"}, 0},
		{[]string{"$", "CD"}, 0},
		{[]string{"ADJP", "CD"}, 0},
		{[]string{"PRN", "CD"}, 0},
		{[]string{"CD", "JJ"}, 0},
		{[]string{"JJ", "X"}, 0},
		{[]string{"JJS", "X"}, 0},
		{[]string{"RB", "X"}, 0},
		{[]string{"QP", "X"}, 0},
		{[]string{"X"}, 0},
	}
	for _, input := range inputs {
		if head := finder.FindHead("NP", input.children); head != input.head {
			t.Errorf("expected %d; got %d as head of %q -> %q\n", input.head, head, "NP", input.children)
		}
	}
}

func TestChineseHeadFinderDP(t *testing.T) {
	finder := NewChineseHeadFinder()
	inputs := []struct {
		children []string
		head     int
	}{
		{[]string{"x", "DP", "y"}, 1},
		{[]string{"x", "DT", "y"}, 1},
		{[]string{"x", "OD", "y"}, 1},
		{[]string{"DP", "M"}, 1},
		{[]string{"DT", "M"}, 1},
		{[]string{"OD", "M"}, 1},
	}
	for _, input := range inputs {
		if head := finder.FindHead("DP", input.children); head != input.head {
			t.Errorf("expected %d; got %d as head of %q -> %q\n", input.head, head, "DP", input.children)
		}
	}
}
