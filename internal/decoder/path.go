package decoder

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-json/internal/errors"
)

// PathString represents JSON Path
//
// JSON Path rule
// .     : child operator
// ..    : recursive descent
// [num] : object/element of array by number
// [*]   : all objects/elements for array.
type PathString string

func (s PathString) Build() (Path, error) {
	buf := []rune(s)
	length := len(buf)
	cursor := 0
	builder := &pathBuilder{}
	start := 0
	for cursor < length {
		c := buf[cursor]
		switch c {
		case '.':
			if start < cursor {
				builder.child(string(buf[start:cursor]))
			}
			c, err := builder.parsePathDot(buf, cursor)
			if err != nil {
				return nil, err
			}
			cursor = c
			start = cursor
		case '[':
			c, err := builder.parsePathIndex(buf, cursor)
			if err != nil {
				return nil, err
			}
			cursor = c
			start = cursor
		}
		cursor++
	}
	if start < cursor {
		builder.child(string(buf[start:cursor]))
	}
	return builder.Build(), nil
}

func (b *pathBuilder) parsePathRecursive(buf []rune, cursor int) (int, error) {
	length := len(buf)
	cursor += 2 // skip .. characters
	start := cursor
	for ; cursor < length; cursor++ {
		c := buf[cursor]
		switch c {
		case '$', '*', ']':
			return 0, fmt.Errorf("specified '%c' after '..' character", c)
		case '.', '[':
			goto end
		}
	}
end:
	if start == cursor {
		return 0, fmt.Errorf("not found recursive selector")
	}
	b.recursive(string(buf[start:cursor]))
	return cursor, nil
}

func (b *pathBuilder) parsePathDot(buf []rune, cursor int) (int, error) {
	length := len(buf)
	if cursor+1 < length && buf[cursor+1] == '.' {
		c, err := b.parsePathRecursive(buf, cursor)
		if err != nil {
			return 0, err
		}
		return c, nil
	}
	cursor++ // skip . character
	start := cursor
	for ; cursor < length; cursor++ {
		c := buf[cursor]
		switch c {
		case '$', '*', ']':
			return 0, fmt.Errorf("specified '%c' after '.' character", c)
		case '.', '[':
			goto end
		}
	}
end:
	if start == cursor {
		return 0, fmt.Errorf("not found child selector")
	}
	b.child(string(buf[start:cursor]))
	return cursor, nil
}

func (b *pathBuilder) parsePathIndex(buf []rune, cursor int) (int, error) {
	length := len(buf)
	cursor++ // skip '[' character
	if length <= cursor {
		return 0, fmt.Errorf("unexpected end of JSON Path")
	}
	c := buf[cursor]
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '*':
		start := cursor
		cursor++
		for ; cursor < length; cursor++ {
			c := buf[cursor]
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				continue
			}
			break
		}
		if buf[cursor] != ']' {
			return 0, fmt.Errorf("invalid character %s at %d", string(buf[cursor]), cursor)
		}
		numOrAll := string(buf[start:cursor])
		if numOrAll == "*" {
			b.indexAll()
			return cursor + 1, nil
		}
		num, err := strconv.ParseInt(numOrAll, 10, 64)
		if err != nil {
			return 0, err
		}
		b.index(int(num))
		return cursor + 1, nil
	}
	return 0, fmt.Errorf("invalid character %s at %d", c, cursor)
}

type pathBuilder struct {
	root Path
	node Path
}

func (b *pathBuilder) indexAll() *pathBuilder {
	node := newIndexAllNode()
	if b.root == nil {
		b.root = node
		b.node = node
	} else {
		b.node = b.node.chain(node)
	}
	return b
}

func (b *pathBuilder) recursive(selector string) *pathBuilder {
	node := newRecursiveNode(selector)
	if b.root == nil {
		b.root = node
		b.node = node
	} else {
		b.node = b.node.chain(node)
	}
	return b
}

func (b *pathBuilder) child(name string) *pathBuilder {
	node := newSelectorNode(name)
	if b.root == nil {
		b.root = node
		b.node = node
	} else {
		b.node = b.node.chain(node)
	}
	return b
}

func (b *pathBuilder) index(idx int) *pathBuilder {
	node := newIndexNode(idx)
	if b.root == nil {
		b.root = node
		b.node = node
	} else {
		b.node = b.node.chain(node)
	}
	return b
}

func (b *pathBuilder) Build() Path {
	return b.root
}

type Path interface {
	fmt.Stringer
	chain(Path) Path
	Index(int) (Path, bool, error)
	Field(string) (Path, bool, error)
	target() bool
	allRead() bool
	single() bool
}

type basePathNode struct {
	child Path
}

func (n *basePathNode) allRead() bool {
	return true
}

func (n *basePathNode) chain(node Path) Path {
	n.child = node
	return node
}

func (n *basePathNode) target() bool {
	return n.child == nil
}

func (n *basePathNode) single() bool {
	return true
}

type selectorNode struct {
	*basePathNode
	selector string
}

func newSelectorNode(selector string) *selectorNode {
	return &selectorNode{
		basePathNode: &basePathNode{},
		selector:     strings.ToLower(selector),
	}
}

func (n *selectorNode) Index(idx int) (Path, bool, error) {
	return nil, false, &errors.PathError{}
}

func (n *selectorNode) Field(fieldName string) (Path, bool, error) {
	if n.selector == fieldName {
		return n.child, true, nil
	}
	return nil, false, nil
}

func (n *selectorNode) String() string {
	s := fmt.Sprintf(".%s", n.selector)
	if n.child != nil {
		s += n.child.String()
	}
	return s
}

type indexNode struct {
	*basePathNode
	selector int
}

func newIndexNode(selector int) *indexNode {
	return &indexNode{
		basePathNode: &basePathNode{},
		selector:     selector,
	}
}

func (n *indexNode) Index(idx int) (Path, bool, error) {
	if n.selector == idx {
		return n.child, true, nil
	}
	return nil, false, nil
}

func (n *indexNode) Field(fieldName string) (Path, bool, error) {
	return nil, false, &errors.PathError{}
}

func (n *indexNode) String() string {
	s := fmt.Sprintf("[%d]", n.selector)
	if n.child != nil {
		s += n.child.String()
	}
	return s
}

type indexAllNode struct {
	*basePathNode
}

func newIndexAllNode() *indexAllNode {
	return &indexAllNode{
		basePathNode: &basePathNode{},
	}
}

func (n *indexAllNode) Index(idx int) (Path, bool, error) {
	return n.child, true, nil
}

func (n *indexAllNode) Field(fieldName string) (Path, bool, error) {
	return nil, false, &errors.PathError{}
}

func (n *indexAllNode) String() string {
	s := "[*]"
	if n.child != nil {
		s += n.child.String()
	}
	return s
}

type recursiveNode struct {
	*basePathNode
	selector string
}

func newRecursiveNode(selector string) *recursiveNode {
	return &recursiveNode{
		basePathNode: &basePathNode{},
		selector:     selector,
	}
}

func (n *recursiveNode) Field(fieldName string) (Path, bool, error) {
	if n.selector == fieldName {
		return n.child, true, nil
	}
	return nil, false, nil
}

func (n *recursiveNode) Index(_ int) (Path, bool, error) {
	return n, true, nil
}

func (n *recursiveNode) String() string {
	s := fmt.Sprintf("..%s", n.selector)
	if n.child != nil {
		s += n.child.String()
	}
	return s
}
