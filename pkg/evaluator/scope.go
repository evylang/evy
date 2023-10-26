package evaluator

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
