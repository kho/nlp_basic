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

// Parser parses treebank trees from a io.ByteScanner.
type Parser struct {
	input io.ByteScanner
	// tokenizer information
	peek  bool
	token []byte
	kind  kind
	err   error
}

// NewParser creates a new parser that reads from input.
func NewParser(input io.ByteScanner) *Parser {
	return &Parser{input: input, token: make([]byte, 256)}
}

// Next extracts the next parse tree from input. When succeeds,
// it returns the node and nil error. For a special tree form (()), it
// returns NoParse. When it encounters an error when reading the first
// token, it returns the IO error from the scanner; otherwise it
// returns one of the above parser errors.
func (parser *Parser) Next() (node Node, err error) {
	// (
	_, kind, err := parser.nextToken()
	if err != nil {
		return
	}
	if kind != OPEN {
		err = NoOpenParen
		return
	}

	node, err = parser.parseNode()
	// If the error is NoParse, we still need to consume the closing )
	if err != nil && err != NoParse {
		return
	}

	// )
	_, kind, err2 := parser.nextToken()
	if err2 != nil || kind != CLOSE {
		err = NoCloseParen
	}

	return
}

// ParseAll extracts all the trees from the remaining input until the
// end of input or first parse error. A nil pointer is stored
// everytime a NoParse is encountered.
func ParseAll(input io.ByteScanner) (trees []*Node, err error) {
	p := NewParser(input)
	node, err := p.Next()
	for err == nil || err == NoParse {
		if err == NoParse {
			trees = append(trees, nil)
		} else {
			trees = append(trees, &Node{node.Label, node.Children})
		}
		node, err = p.Next()
	}
	if err == io.EOF {
		err = nil
	}
	return
}

// kind is the kind of token found by the parser. It only takes the
// following 3 constant values.
type kind int

const (
	OPEN  kind = 0
	CLOSE      = iota
	WORD       = iota
)

// peekToken peeks at the next token. See nextToken() for its return values.
func (p *Parser) peekToken() ([]byte, kind, error) {
	if !p.peek {
		p.token, p.kind, p.err = p.nextToken()
		p.peek = true
	}
	return p.token, p.kind, p.err
}

// nextToken returns the next token as a byte buffer and its kind. Or
// when the read fails, it returns the IO error.
func (p *Parser) nextToken() (token []byte, kind kind, err error) {
	if p.peek {
		token, kind, err = p.token, p.kind, p.err
		p.peek = false
		return
	}
	// Skip spaces
	c, err := p.input.ReadByte()
	for err == nil && (c == ' ' || c == '\t' || c == '\n') {
		c, err = p.input.ReadByte()
	}
	if err != nil {
		return
	}
	// Find out token's type
	p.token = p.token[:0]
	p.token = append(p.token, c)
	if c == '(' {
		kind = OPEN
	} else if c == ')' {
		kind = CLOSE
	} else {
		// Continue until a white-space or parentheses
		c, err = p.input.ReadByte()
		for err == nil && c != ' ' && c != '\t' && c != '\n' && c != '(' && c != ')' {
			p.token = append(p.token, c)
			c, err = p.input.ReadByte()
		}
		if err == nil {
			p.input.UnreadByte()
		} else {
			// We have successfully read something; postpone this error.
			err = nil
		}
		kind = WORD
	}
	token = p.token
	return
}

// parseNode tries to parse a tree node from tok. When succeeds, it
// returns the tree node. When the next expr is no parse (()), it
// returns NoParse. Otherwise it returns other errors defined above.
func (p *Parser) parseNode() (node Node, err error) {
	// (
	_, kind, err := p.nextToken()
	if err != nil || kind != OPEN {
		err = NoOpenParen
		return
	}
	_, kind, err = p.peekToken()
	if err != nil {
		return
	}
	if kind == CLOSE {
		p.nextToken()
		err = NoParse
		return
	}
	// Category
	token, kind, err := p.nextToken()
	if err != nil || kind != WORD {
		err = NoCategory
		return
	}
	node.Label = string(token)

	// ( or word
	token, kind, err = p.peekToken()
	if err != nil || kind == CLOSE {
		err = NoWordOrOpenParen
		return
	}

	switch kind {
	case WORD:
		node.Children = append(node.Children, Node{string(token), nil})
		p.nextToken() // consume the peeked token
	case OPEN:
		node.Children, err = p.parseChildren()
		if err != nil {
			return
		}
	default:
		err = ParseError
		return
	}

	_, kind, err = p.nextToken()
	if err != nil || kind != CLOSE {
		err = NoCloseParen
		return
	}

	return
}

// parseChildren parses at least one node form and appends these to
// children until it encounters the closing parenthesis or an error.
func (p *Parser) parseChildren() (children []Node, err error) {
	children = make([]Node, 0, 4)
	// (...)
	child, err := p.parseNode()
	if err != nil {
		return
	}
	children = append(children, child)

	_, kind, err := p.peekToken()
	for err == nil && kind == OPEN {
		child, err = p.parseNode()
		if err != nil {
			return
		}
		children = append(children, child)
		_, kind, err = p.peekToken()
	}
	err = nil
	return
}
