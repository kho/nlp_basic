package treebank

import (
	"errors"
	"io"
	"strings"
)

// Parsing errors
var (
	ParseError        = errors.New("something went wrong")
	NoCloseParen      = errors.New("expect )")
	NoOpenParen       = errors.New("expect (")
	NoCategory        = errors.New("expect category")
	NoWordOrOpenParen = errors.New("expect word or (")
	ResidualInput     = errors.New("residual input")
)

// ParseString parses a single string to extract one tree and discards
// the rest of the string.
func ParseString(input string) (*ParseTree, error) {
	return NewParser(strings.NewReader(input)).Next()
}

// FromString converts a single string to extract one tree. Panics if
// there is any error.
func FromString(input string) *ParseTree {
	p := NewParser(strings.NewReader(input))
	tree, err := p.Next()
	if err != nil {
		panic(err)
	}
	_, err = p.Next()
	if err != io.EOF {
		panic(ResidualInput)
	}
	return tree
}

// ParseAll extracts all the trees from the remaining input until the
// end of input or first parse error. A nil pointer is stored
// everytime a NoParse is encountered.
func ParseAll(input io.ByteScanner) (trees []*ParseTree, err error) {
	p := NewParser(input)
	tree, err := p.Next()
	for err == nil {
		trees = append(trees, tree)
		tree, err = p.Next()
	}
	if err == io.EOF {
		err = nil
	}
	return
}

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

// Next extracts the next parse tree from input. When succeeds, it
// returns the tree and nil error. When it encounters an error when
// reading the first token, it returns the IO error from the scanner;
// otherwise it returns one of the above parser errors.
func (p *Parser) Next() (*ParseTree, error) {
	tree := &ParseTree{NewEmptyTopology(), make([]string, 0, 16)}
	_, err := p.parseS(tree)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

// parseS is the entry point of the following recursive descent parser
// (note the grammar is stricter than ordinary sexp because of the
// constraints in Treebank trees):
//
//   S -> '(' Tree ')'
//   Tree -> '(' { ')' | Node ')' }
//   Node -> Label { Label | Children }
//   Children -> '(' Node ')' { eps | Children }
//
// All parse* methods has a similar interface: they take the tree
// being constructed and returns the id of its created node(s) or any
// error.
//
// It returns the node id of the top-level node and sets it as root
// when the next tree is not no parse; or NoNodeId when the next tree
// is no parse. If it fails to read the first token, the IO error from
// the input is returned; otherwise the returned error is one of the
// parse errors.
func (p *Parser) parseS(tree *ParseTree) (NodeId, error) {
	// (
	_, kind, err := p.nextToken()
	if err != nil {
		return NoNodeId, err
	}
	if kind != kOpen {
		return NoNodeId, NoOpenParen
	}

	root, err := p.parseTree(tree)
	if err != nil {
		return NoNodeId, err
	}

	// )
	_, kind, err = p.nextToken()
	if err != nil || kind != kClose {
		return NoNodeId, NoCloseParen
	}

	// Success; set the root when the tree is not empty.
	if root != NoNodeId {
		tree.Topology.SetRoot(root)
	}
	return root, nil
}

// parseTree parses a tree using the following rule,
//   Tree -> '(' { ')' | Node ')' }
// In case 1, the tree is empty and NoNodeId is returned as the
// created node. In case 2, a node is actually created and
// returned. Any IO error is translated to parse errors and returned.
func (p *Parser) parseTree(tree *ParseTree) (NodeId, error) {
	_, kind, err := p.nextToken()
	if err != nil || kind != kOpen {
		return NoNodeId, NoOpenParen
	}

	_, kind, err = p.peekToken()
	if err == nil && kind == kClose {
		p.nextToken()
		return NoNodeId, nil
	}

	node, err := p.parseNode(tree)
	if err != nil {
		return NoNodeId, err
	}

	_, kind, err = p.nextToken()
	if err != nil || kind != kClose {
		return NoNodeId, NoCloseParen
	}

	return node, nil
}

// parseNode parses a tree node using the following rule,
//   Node -> Label { Label | Children }
// When succeeds, it adds the parsed node to the tree and returns the
// node id. Otherwise it returns one of the parse errors.
func (p *Parser) parseNode(tree *ParseTree) (NodeId, error) {
	// First Label --- Category
	token, kind, err := p.nextToken()
	if err != nil || kind != kWord {
		return NoNodeId, NoCategory
	}

	// Create the node
	node := tree.Topology.AddNode()
	tree.Label = append(tree.Label, string(token))

	// ( or word
	_, kind, err = p.peekToken()
	if err != nil || kind == kClose {
		return NoNodeId, NoWordOrOpenParen
	}

	switch kind {
	case kWord:
		// This is a pre-terminal
		token, _, _ := p.nextToken()
		child := tree.Topology.AddNode()
		tree.Label = append(tree.Label, string(token))
		tree.Topology.AppendChild(node, child)
	case kOpen:
		// This is a non-terminal
		children, err := p.parseChildren(tree)
		if err != nil {
			return NoNodeId, err
		}
		// Directly use the returned children to avoid reallocation
		tree.Topology.children[node] = children
		for _, child := range children {
			tree.Topology.parent[child] = node
		}
	default:
		err = ParseError
		return NoNodeId, err
	}

	return node, nil
}

// parseChildren parses a list of children using the following rule,
//   Children -> '(' Node ')' { eps | Children }
// It returns a slice of children node ids or any error. The caller
// owns the returend slice.
func (p *Parser) parseChildren(tree *ParseTree) ([]NodeId, error) {
	children := make([]NodeId, 0, 4)
	// First node
	_, kind, err := p.nextToken()
	if err != nil || kind != kOpen {
		return nil, NoOpenParen
	}

	child, err := p.parseNode(tree)
	if err != nil {
		return nil, err
	}
	children = append(children, child)

	_, kind, err = p.nextToken()
	if err != nil || kind != kClose {
		return nil, NoCloseParen
	}

	// The rest
	_, kind, err = p.peekToken()
	for err == nil && kind == kOpen {
		p.nextToken()

		child, err = p.parseNode(tree)
		if err != nil {
			return nil, err
		}
		children = append(children, child)

		_, kind, err = p.nextToken()
		if err != nil || kind != kClose {
			return nil, NoCloseParen
		}

		_, kind, err = p.peekToken()
	}

	return children, nil
}

// kind is the kind of token found by the parser. It only takes the
// following 3 constant values.
type kind int

// Three kinds of tokens.
const (
	kOpen  kind = 0
	kClose      = iota
	kWord       = iota
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
	p.token = p.token[:1]
	p.token[0] = c
	if c == '(' {
		kind = kOpen
	} else if c == ')' {
		kind = kClose
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
		kind = kWord
	}
	token = p.token
	return
}
