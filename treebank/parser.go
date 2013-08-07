package treebank

import (
	"errors"
)

type Node struct {
	Label    string
	Children []Node
}

var EndOfInput = errors.New("end of input")
var (
	ParseError        = errors.New("something went wrong")
	NoCloseParen      = errors.New("expect )")
	NoOpenParen       = errors.New("expect (")
	NoCategory        = errors.New("expect category")
	NoWordOrOpenParen = errors.New("expect word or (")
)

type Tokenizer struct {
	input string
	pos   int
	peek  *peek
}

type Part struct {
	Start, End int
}

type Kind int

const (
	OPEN  Kind = 0
	CLOSE      = iota
	WORD       = iota
)

type peek struct {
	p Part
	k Kind
	e error
}

func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{input, 0, nil}
}

func (tok *Tokenizer) Pos() int {
	return tok.pos
}

func (tok *Tokenizer) Token(p Part) string {
	return tok.input[p.Start:p.End]
}

func (tok *Tokenizer) Peek() (Part, Kind, error) {
	if tok.peek == nil {
		p, k, e := tok.Next()
		tok.peek = &peek{p, k, e}
	}
	return tok.peek.p, tok.peek.k, tok.peek.e
}

func (tok *Tokenizer) Next() (p Part, k Kind, e error) {
	if tok.peek != nil {
		p, k, e = tok.peek.p, tok.peek.k, tok.peek.e
		tok.peek = nil
		return
	}
	pos := tok.pos
	input := tok.input
	// Skip spaces
	for pos < len(input) && (input[pos] == ' ' || input[pos] == '\t' || input[pos] == '\n') {
		pos++
	}
	if pos == len(input) {
		e = EndOfInput
		return
	}
	// Find out token's type
	if input[pos] == '(' {
		p = Part{pos, pos + 1}
		k = OPEN
		pos++
	} else if input[pos] == ')' {
		p = Part{pos, pos + 1}
		k = CLOSE
		pos++
	} else {
		k = WORD
		p.Start = pos
		pos++
		// Continue until meet a white-space or parentheses
		for pos < len(input) && input[pos] != ' ' && input[pos] != '\t' && input[pos] != '\n' && input[pos] != '(' && input[pos] != ')' {
			pos++
		}
		p.End = pos
	}
	// Store next read position
	tok.pos = pos
	return
}

func Parse(tok *Tokenizer) (ret Node, err error) {
	_, k, err := tok.Next()
	if err != nil || k != OPEN {
		err = NoOpenParen
		return
	}

	ret, err = parseNode(tok)

	if err != nil {
		return
	}

	_, k, err = tok.Next()
	if err != nil || k != CLOSE {
		err = NoCloseParen
	}

	return
}

func parseNode(tok *Tokenizer) (ret Node, err error) {
	// (
	_, k, err := tok.Next()
	if err != nil || k != OPEN {
		err = NoOpenParen
		return
	}
	// Category
	p, k, err := tok.Next()
	if err != nil || k != WORD {
		err = NoCategory
		return
	}
	ret.Label = tok.Token(p)

	// ( or word
	p, k, err = tok.Peek()
	if err != nil || k == CLOSE {
		err = NoWordOrOpenParen
		return
	}

	switch k {
	case WORD:
		ret.Children = append(ret.Children, Node{tok.Token(p), nil})
		tok.Next()
	case OPEN:
		ret.Children, err = parseChildren(tok)
		if err != nil {
			return
		}
	default:
		err = ParseError
		return
	}

	_, k, err = tok.Next()
	if err != nil || k != CLOSE {
		err = NoCloseParen
		return
	}

	return
}

func parseChildren(tok *Tokenizer) (children []Node, err error) {
	// (...)
	child, err := parseNode(tok)
	if err != nil {
		return
	}
	children = append(children, child)

	_, k, err := tok.Peek()
	for err == nil && k == OPEN {
		child, err = parseNode(tok)
		if err != nil {
			return
		}
		children = append(children, child)
		_, k, err = tok.Peek()
	}
	err = nil
	return
}
