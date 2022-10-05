package evaluator

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
	if s == nil {
		return nil, false
	}
	if val, ok := s.values[name]; ok {
		return val, true
	}
	return s.outer.get(name)
}

func (s *scope) getScope(name string) (*scope, bool) {
	if s == nil {
		return nil, false
	}
	if _, ok := s.values[name]; ok {
		return s, true
	}
	return s.outer.getScope(name)
}

func (s *scope) set(name string, val Value) {
	s.values[name] = val
}
