package evaluator

import "foxygo.at/evy/pkg/parser"

type scope struct {
	values map[string]Value
	outer  *scope
}

func newScope() *scope {
	return &scope{values: map[string]Value{}}
}

func newInnerScope(outer *scope) *scope {
	return &scope{values: map[string]Value{}, outer: outer}
}

func (s *scope) get(name string) (Value, bool) {
	if s == nil || name == "_" {
		return nil, false
	}
	if val, ok := s.values[name]; ok {
		return val, true
	}
	return s.outer.get(name)
}

func (s *scope) set(name string, val Value, t *parser.Type) {
	if name == "_" {
		return
	}
	switch val.Type() {
	case parser.GENERIC_ARRAY:
		val.(*Array).T = t
	case parser.GENERIC_MAP:
		val.(*Map).T = t
	}
	s.values[name] = val
}
