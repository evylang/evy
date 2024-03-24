package object

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
)

type BuiltinFunction func(args ...Object) Object

type ObjectType string

const (
	NULL_OBJ = "NULL"

	INTEGER_OBJ = "INTEGER"
	BOOLEAN_OBJ = "BOOLEAN"
	STRING_OBJ  = "STRING"
	ARRAY_OBJ   = "ARRAY"
	MAP_OBJ     = "MAP"
)

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Hashable interface {
	HashKey() HashKey
}

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Indexable interface {
	Index(key Object) (Object, error)
}

type Sliceable interface {
	Slice(start, end Object) (Object, error)
}

// TODO: rename this to num to match evy spec
type Integer struct {
	Value float64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%f", i.Value) }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

func (s *String) Index(key Object) (Object, error) {
	index, ok := key.(*Integer)
	if !ok {
		return nil, errors.New("string index must be an integer")
	}
	i := int(index.Value)
	if err := indexOutOfBounds(i, len(s.Value)); err != nil {
		return nil, fmt.Errorf("string %w", err)
	}
	if i < 0 {
		i = len(s.Value) + i
	}
	return &String{Value: s.Value[i : i+1]}, nil
}

func (s *String) Slice(start, end Object) (Object, error) {
	var startIdx, endIdx int
	// if the starting index _is_ provided, overwrite default with provided value
	if _, ok := start.(*Integer); ok {
		startIdx = int(start.(*Integer).Value)
	}
	// if the ending index _is not_ provided, then use the length of the string
	if _, ok := end.(*Null); ok {
		endIdx = len(s.Value)
	} else {
		endIdx = int(end.(*Integer).Value)
	}
	if err := indexOutOfBounds(startIdx, len(s.Value)); err != nil {
		return nil, fmt.Errorf("string slice %w", err)
	}
	if startIdx < 0 {
		startIdx = len(s.Value) + startIdx
	}
	if endIdx < -len(s.Value) || endIdx > len(s.Value) {
		return nil, fmt.Errorf("string slice end index out of bounds: end: %d len: %d", endIdx, len(s.Value))
	}
	if endIdx < 0 {
		endIdx = len(s.Value) + endIdx
	}
	if startIdx > endIdx {
		return nil, fmt.Errorf("string slice index out of bounds: start: %d end: %d", startIdx, endIdx)
	}
	return &String{Value: s.Value[startIdx:endIdx]}, nil
}

type Array struct {
	Elements []Object
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

func (a *Array) Index(key Object) (Object, error) {
	integer, ok := key.(*Integer)
	if !ok {
		return nil, errors.New("map index must be an integer")
	}
	i := int(integer.Value)
	if err := indexOutOfBounds(i, len(a.Elements)); err != nil {
		return nil, fmt.Errorf("array %w", err)
	}
	if i < 0 {
		i = len(a.Elements) + i
	}
	return a.Elements[i], nil
}

func (a *Array) Slice(start, end Object) (Object, error) {
	var startIdx, endIdx int
	// if the starting index _is_ provided, overwrite default with provided value
	if _, ok := start.(*Integer); ok {
		startIdx = int(start.(*Integer).Value)
	}
	// if the ending index _is not_ provided, then use the length of the array
	if _, ok := end.(*Null); ok {
		endIdx = len(a.Elements)
	} else {
		endIdx = int(end.(*Integer).Value)
	}
	if err := indexOutOfBounds(startIdx, len(a.Elements)); err != nil {
		return nil, fmt.Errorf("array slice %w", err)
	}
	if startIdx < 0 {
		startIdx = len(a.Elements) + startIdx
	}
	if endIdx < -len(a.Elements) || endIdx > len(a.Elements) {
		return nil, fmt.Errorf("array slice end index out of bounds: end: %d len: %d", endIdx, len(a.Elements))
	}
	if endIdx < 0 {
		endIdx = len(a.Elements) + endIdx
	}
	if startIdx > endIdx {
		return nil, fmt.Errorf("array slice index out of bounds: start: %d end: %d", startIdx, endIdx)
	}
	return &Array{Elements: a.Elements[startIdx:endIdx]}, nil
}

type Map map[string]Object

func (m Map) Type() ObjectType { return MAP_OBJ }
func (m Map) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s: %s", k, v.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

func (m Map) Index(key Object) (Object, error) {
	index, ok := key.(*String)
	if !ok {
		return nil, errors.New("map index must be a string")
	}
	k := index.Value
	val, ok := m[k]
	if !ok {
		return nil, fmt.Errorf("no key %s in map", k)
	}
	return val, nil
}

func indexOutOfBounds(i, length int) error {
	if i >= length || i < -length {
		return fmt.Errorf("index out of bounds: %d len: %d", i, length)
	}
	return nil
}
