package heads

import (
	"fmt"
)

// HeadFinder finds the head constituent in a CFG rule expressed as
// parent and children.
type HeadFinder interface {
	// FindHead returns the index of the head child. The returned value
	// must be a member of children. If the head cannot be found, it
	// should panic.
	FindHead(parent string, children []string) int
}

// Three possible directions of the head.
const (
	UNKNOWN      = 0
	HEAD_INITIAL = iota
	HEAD_FINAL   = iota
)

type HeadRule struct {
	// Whether the constituent is head-initial or head-final.
	Direction int
	// A mapping from labels to priority levels in the range of
	// [0:len(Priority)). 0 is the highest. Labels not in the table all
	// have the lowest priority (i.e. len(Priority)).
	Priority map[string]int
}

// NewHeadRule creates a HeadRule with Direction being dir, and
// Priority storing label priorities stores in decreasing order of the
// labels in match. match can be empty, in which case dir alone
// decides the head. Direction must be either HEAD_INITIAL or
// HEAD_FINAL.
func NewHeadRule(dir int, match []string) *HeadRule {
	if dir != HEAD_INITIAL && dir != HEAD_FINAL {
		panic("head direction must be either HEAD_INITIAL or HEAD_FINAL")
	}
	priority := make(map[string]int)
	for i, v := range match {
		priority[v] = i
	}
	return &HeadRule{dir, priority}
}

func (rule *HeadRule) LabelPriority(label string) int {
	p, ok := rule.Priority[label]
	if !ok {
		return len(rule.Priority)
	}
	return p
}

// TableHeadFinder finds the head by looking up a table of
// HeadRules. Initial value is an empty finder that panics whenever
// FindHead() is called.
type TableHeadFinder struct {
	// Head rules
	Table map[string]*HeadRule
	// Fallback direction when the parent category is not known (UNKNOWN
	// = panic).
	Fallback int
}

func (finder *TableHeadFinder) FindHead(parent string, children []string) int {
	if len(children) == 0 {
		panic("trying to find the head of a leaf: " + parent)
	}
	rule, ok := finder.Table[parent]
	if !ok {
		switch finder.Fallback {
		case HEAD_INITIAL:
			return 0
		case HEAD_FINAL:
			return len(children) - 1
		default:
			panic("unknown category: " + parent)
		}
	}
	switch rule.Direction {
	case HEAD_INITIAL:
		i := 0
		p := rule.LabelPriority(children[i])
		for j := 1; j < len(children); j++ {
			if pp := rule.LabelPriority(children[j]); pp < p {
				i = j
				p = pp
			}
		}
		return i
	case HEAD_FINAL:
		i := len(children) - 1
		p := rule.LabelPriority(children[i])
		for j := i - 1; j >= 0; j-- {
			if pp := rule.LabelPriority(children[j]); pp < p {
				i = j
				p = pp
			}
		}
		return i
	default:
		panic(fmt.Sprintf("invalid rule.Direction: %d", rule.Direction))
	}
}

// EnglishHeadFinder is a head-finder for English Penn Treebank
// trees. It overrides certain NP head rules. See [2] in
// http://www.cs.columbia.edu/~mcollins/papers/heads for details.
type EnglishHeadFinder TableHeadFinder

func NewEnglishHeadFinder() *EnglishHeadFinder {
	return (*EnglishHeadFinder)(&TableHeadFinder{
		map[string]*HeadRule{
			"ADJP":   NewHeadRule(HEAD_FINAL, []string{"NNS", "QP", "NN", "$", "ADVP", "JJ", "VBN", "VBG", "ADJP", "JJR", "NP", "JJS", "DT", "FW", "RBR", "RBS", "SBAR", "RB"}),
			"ADVP":   NewHeadRule(HEAD_INITIAL, []string{"RB", "RBR", "RBS", "FW", "ADVP", "TO", "CD", "JJR", "JJ", "IN", "NP", "JJS", "NN"}),
			"CONJP":  NewHeadRule(HEAD_INITIAL, []string{"CC", "RB", "IN"}),
			"FRAG":   NewHeadRule(HEAD_INITIAL, nil),
			"INTJ":   NewHeadRule(HEAD_FINAL, nil),
			"LST":    NewHeadRule(HEAD_INITIAL, []string{"LS", ":"}),
			"NAC":    NewHeadRule(HEAD_FINAL, []string{"NN", "NNS", "NNP", "NNPS", "NP", "NAC", "EX", "$", "CD", "QP", "PRP", "VBG", "JJ", "JJS", "JJR", "ADJP", "FW"}),
			"PP":     NewHeadRule(HEAD_INITIAL, []string{"IN", "TO", "VBG", "VBN", "RP", "FW"}),
			"PRN":    NewHeadRule(HEAD_FINAL, nil),
			"PRT":    NewHeadRule(HEAD_INITIAL, []string{"RP"}),
			"QP":     NewHeadRule(HEAD_FINAL, []string{"$", "IN", "NNS", "NN", "JJ", "RB", "DT", "CD", "NCD", "QP", "JJR", "JJS"}),
			"RRC":    NewHeadRule(HEAD_INITIAL, []string{"VP", "NP", "ADVP", "ADJP", "PP"}),
			"S":      NewHeadRule(HEAD_FINAL, []string{"TO", "IN", "VP", "S", "SBAR", "ADJP", "UCP", "NP"}),
			"SBAR":   NewHeadRule(HEAD_FINAL, []string{"WHNP", "WHPP", "WHADVP", "WHADJP", "IN", "DT", "S", "SQ", "SINV", "SBAR", "FRAG"}),
			"SBARQ":  NewHeadRule(HEAD_FINAL, []string{"SQ", "S", "SINV", "SBARQ", "FRAG"}),
			"SINV":   NewHeadRule(HEAD_FINAL, []string{"VBZ", "VBD", "VBP", "VB", "MD", "VP", "S", "SINV", "ADJP", "NP"}),
			"SQ":     NewHeadRule(HEAD_FINAL, []string{"VBZ", "VBD", "VBP", "VB", "MD", "VP", "SQ"}),
			"UCP":    NewHeadRule(HEAD_INITIAL, nil),
			"VP":     NewHeadRule(HEAD_FINAL, []string{"TO", "VBD", "VBN", "MD", "VBZ", "VB", "VBG", "VBP", "VP", "ADJP", "NN", "NNS", "NP"}),
			"WHADJP": NewHeadRule(HEAD_FINAL, []string{"CC", "WRB", "JJ", "ADJP"}),
			"WHADVP": NewHeadRule(HEAD_INITIAL, []string{"CC", "WRB"}),
			"WHNP":   NewHeadRule(HEAD_FINAL, []string{"WDT", "WP", "WP$", "WHADJP", "WHPP", "WHNP"}),
			"WHPP":   NewHeadRule(HEAD_INITIAL, []string{"IN", "TO", "FW"}),
		},
		UNKNOWN,
	})
}

// Quoting http://www.cs.columbia.edu/~mcollins/papers/heads:
//
//   Ignore the row for NPs -- I use a special set of rules for
//   this. For these I initially remove ADJPs, QPs, and also NPs which
//   dominate a possesive (tagged POS, e.g. (NP (NP the man 's)
//   telescope ) becomes (NP the man 's telescope)).
//
// This needs to be done outside the head finder.
// TODO: implement this transformation.
func (finder *EnglishHeadFinder) FindHead(parent string, children []string) int {
	if parent == "NP" {
		if len(children) == 0 {
			panic("trying to find the head of a leaf: " + parent)
		}
		// If the last word is tagged POS, return (last-word);
		if children[len(children)-1] == "POS" {
			return len(children) - 1
		}
		// Else search from right to left for the first child which is
		// an NN, NNP, NNPS, NNS, NX, POS, or JJR
		for i := len(children) - 1; i >= 0; i-- {
			child := children[i]
			if child == "NN" || child == "NNP" || child == "NNPS" || child == "NNS" || child == "NX" || child == "POS" || child == "JJR" {
				return i
			}
		}
		// Else search from left to right for first child which is an NP
		for i, child := range children {
			if child == "NP" {
				return i
			}
		}
		// Else search from right to left for the first child which is a
		// $, ADJP or PRN
		for i := len(children) - 1; i >= 0; i-- {
			child := children[i]
			if child == "$" || child == "ADJP" || child == "PRN" {
				return i
			}
		}
		// Else search from right to left for the first child which is a CD
		for i := len(children) - 1; i >= 0; i-- {
			child := children[i]
			if child == "CD" {
				return i
			}
		}
		// Else search from right to left for the first child which is a JJ, JJS, RB or QP
		for i := len(children) - 1; i >= 0; i-- {
			child := children[i]
			if child == "JJ" || child == "JJS" || child == "RB" || child == "QP" {
				return i
			}
		}
		// Else return the last word
		return len(children) - 1
	}
	return (*TableHeadFinder)(finder).FindHead(parent, children)
}

// ChineseHeadFinder is a head-finder for Chinese Treebank trees. It
// overrides DP head rules and has a fallback choice. See Table 8 in
// http://www.aclweb.org/anthology-new/D/D08/D08-1059.pdf for details.
type ChineseHeadFinder TableHeadFinder

func NewChineseHeadFinder() *ChineseHeadFinder {
	return (*ChineseHeadFinder)(&TableHeadFinder{
		map[string]*HeadRule{
			"ADJP": NewHeadRule(HEAD_FINAL, []string{"ADJP", "JJ", "AD"}),
			"ADVP": NewHeadRule(HEAD_FINAL, []string{"ADVP", "AD", "CS", "JJ", "NP", "PP", "P", "VA", "VV"}),
			"CLP":  NewHeadRule(HEAD_FINAL, []string{"CLP", "M", "NN", "NP"}),
			"CP":   NewHeadRule(HEAD_FINAL, []string{"CP", "IP", "VP"}),
			"DNP":  NewHeadRule(HEAD_FINAL, []string{"DEG", "DNP", "DEC", "QP"}),
			"DP":   NewHeadRule(HEAD_INITIAL, []string{"DP", "DT", "OD"}),
			"DVP":  NewHeadRule(HEAD_FINAL, []string{"DEV", "AD", "VP"}),
			"FRAG": NewHeadRule(HEAD_FINAL, []string{"VV", "NR", "NN", "NT"}),
			"IP":   NewHeadRule(HEAD_FINAL, []string{"VP", "IP", "NP"}),
			"LCP":  NewHeadRule(HEAD_FINAL, []string{"LCP", "LC"}),
			"LST":  NewHeadRule(HEAD_FINAL, []string{"CD", "NP", "QP"}),
			"NP":   NewHeadRule(HEAD_FINAL, []string{"NP", "NN", "IP", "NR", "NT"}),
			"NN":   NewHeadRule(HEAD_FINAL, []string{"NP", "NN", "IP", "NR", "NT"}),
			"PP":   NewHeadRule(HEAD_INITIAL, []string{"P", "PP"}),
			"PRN":  NewHeadRule(HEAD_INITIAL, []string{"PU"}),
			"QP":   NewHeadRule(HEAD_FINAL, []string{"QP", "CLP", "CD"}),
			"UCP":  NewHeadRule(HEAD_INITIAL, []string{"IP", "NP", "VP"}),
			"VCD":  NewHeadRule(HEAD_INITIAL, []string{"VV", "VA", "VE"}),
			"VP":   NewHeadRule(HEAD_INITIAL, []string{"VE", "VC", "VV", "VNV", "VPT", "VRD", "VSB", "VCD", "VP"}),
			"VPT":  NewHeadRule(HEAD_INITIAL, []string{"VA", "VV"}),
			"VRD":  NewHeadRule(HEAD_INITIAL, []string{"VVI", "VA"}),
			"VSB":  NewHeadRule(HEAD_FINAL, []string{"VV", "VE"}),
		},
		HEAD_FINAL,
	})
}

func (finder *ChineseHeadFinder) FindHead(parent string, children []string) int {
	if parent == "DP" {
		for i := len(children) - 1; i >= 0; i-- {
			if children[i] == "M" {
				return i
			}
		}
	}
	return (*TableHeadFinder)(finder).FindHead(parent, children)
}
