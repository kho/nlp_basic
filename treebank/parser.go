package treebank

import (
	"errors"
	"io"
)

var (
	ParseError        = errors.New("something went wrong")
	NoParse           = errors.New("no parse")
	NoCloseParen      = errors.New("expect )")
	NoOpenParen       = errors.New("expect (")
	NoCategory        = errors.New("expect category")
	NoWordOrOpenParen = errors.New("expect word or (")
)

// Parse extracts the next parse tree from input. When succeeds, it
// returns the node and nil error. For a special tree form (()), it
// returns NoParse. When it encounters an error when reading the first
// token, it returns the IO error from the scanner; otherwise it
// returns one of the above errors.
func Parse(input io.ByteScanner) (node Node, err error) {
	tok := newTokenizer(input)

	// (
	_, kind, err := tok.Next()
	if err != nil {
		return
	}
	if kind != OPEN {
		err = NoOpenParen
		return
	}

	// ...
	node, err = parseNode(tok)
	if err != nil {
		return
	}

	// )
	_, kind, err = tok.Next()
	if err != nil || kind != CLOSE {
		err = NoCloseParen
	}

	return
}

// tokenizer tokenizes bytes from input.
type tokenizer struct {
	input io.ByteScanner
	peek  bool
	// these are only valid when peek is true
	token string
	kind  kind
	err   error
}

// kind is the kind of token found by the tokenizer. It only takes the
// following 3 constant values.
type kind int

const (
	OPEN  kind = 0
	CLOSE      = iota
	WORD       = iota
)

func newTokenizer(input io.ByteScanner) *tokenizer {
	return &tokenizer{input: input}
}

// Peek peeks at the next token. See Next() for its return values.
func (tok *tokenizer) Peek() (string, kind, error) {
	if !tok.peek {
		tok.token, tok.kind, tok.err = tok.Next()
		tok.peek = true
	}
	return tok.token, tok.kind, tok.err
}

// Next returns the next token as a string and its kind. Or when the
// read fails, it returns the IO error.
func (tok *tokenizer) Next() (token string, kind kind, err error) {
	if tok.peek {
		token, kind, err = tok.token, tok.kind, tok.err
		tok.peek = false
		return
	}
	// Skip spaces
	c, err := tok.input.ReadByte()
	for err == nil && (c == ' ' || c == '\t' || c == '\n') {
		c, err = tok.input.ReadByte()
	}
	if err != nil {
		return
	}
	// Find out token's type
	if c == '(' {
		token = "("
		kind = OPEN
	} else if c == ')' {
		token = ")"
		kind = CLOSE
	} else {
		tBuf := make([]byte, 1, 8)
		tBuf[0] = c
		// Continue until a white-space or parentheses
		c, err = tok.input.ReadByte()
		for err == nil && c != ' ' && c != '\t' && c != '\n' && c != '(' && c != ')' {
			tBuf = append(tBuf, c)
			c, err = tok.input.ReadByte()
		}
		if err == nil {
			tok.input.UnreadByte()
		} else {
			// We have successfully read something; postpone this error.
			err = nil
		}
		kind = WORD
		token = string(tBuf)
	}
	return
}

// parseNode tries to parse a tree node from tok. When succeeds, it
// returns the tree node. When the next expr is no parse (()), it
// returns NoParse. Otherwise it returns other errors defined above.
func parseNode(tok *tokenizer) (node Node, err error) {
	// (
	_, kind, err := tok.Next()
	if err != nil || kind != OPEN {
		err = NoOpenParen
		return
	}
	_, kind, err = tok.Peek()
	if err != nil {
		return
	}
	if kind == CLOSE {
		err = NoParse
		return
	}
	// Category
	token, kind, err := tok.Next()
	if err != nil || kind != WORD {
		err = NoCategory
		return
	}
	node.Label = token

	// ( or word
	token, kind, err = tok.Peek()
	if err != nil || kind == CLOSE {
		err = NoWordOrOpenParen
		return
	}

	switch kind {
	case WORD:
		node.Children = append(node.Children, Node{token, nil})
		tok.Next()
	case OPEN:
		node.Children, err = parseChildren(tok)
		if err != nil {
			return
		}
	default:
		err = ParseError
		return
	}

	_, kind, err = tok.Next()
	if err != nil || kind != CLOSE {
		err = NoCloseParen
		return
	}

	return
}

// parseChildren parses at least one node form and appends these to
// children until it encounters the closing parenthesis or an error.
func parseChildren(tok *tokenizer) (children []Node, err error) {
	// (...)
	child, err := parseNode(tok)
	if err != nil {
		return
	}
	children = append(children, child)

	_, kind, err := tok.Peek()
	for err == nil && kind == OPEN {
		child, err = parseNode(tok)
		if err != nil {
			return
		}
		children = append(children, child)
		_, kind, err = tok.Peek()
	}
	err = nil
	return
}
