package evaluator

import "fmt"

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

func (s *scope) set(name string, val value) {
	if name == "_" {
		return
	}
	s.values[name] = val
}

func (s *scope) update(name string, val value) {
	if name == "_" {
		return
	}
	if _, ok := s.values[name]; ok {
		s.values[name] = val
		return
	}
	if s.outer == nil {
		panic(fmt.Errorf("%w: update of unknown variable %q", ErrAssignmentTarget, name))
	}
	s.outer.update(name, val)
}
