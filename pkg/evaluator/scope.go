package evaluator

import "foxygo.at/evy/pkg/parser"

type scope struct {
	values map[string]value
	outer  *scope
}

func newScope() *scope {
	return &scope{values: map[string]value{}}
}

func newInnerScope(outer *scope) *scope {
	return &scope{values: map[string]value{}, outer: outer}
}

func (s *scope) get(name string) (value, bool) {
	if s == nil || name == "_" {
		return nil, false
	}
	if val, ok := s.values[name]; ok {
		return val, true
	}
	return s.outer.get(name)
}

func (s *scope) set(name string, val value, t *parser.Type) {
	if name == "_" {
		return
	}
	switch val.Type() {
	case parser.UNTYPED_ARRAY:
		val.(*arrayVal).T = t
	case parser.UNTYPED_MAP:
		val.(*mapVal).T = t
	}
	s.values[name] = val
}
