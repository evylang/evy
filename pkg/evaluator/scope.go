package evaluator

type scope struct {
	values map[string]Value
	outer  *scope
}

func newScope() *scope {
	return &scope{values: map[string]Value{}}
}

func (s *scope) get(name string) (Value, bool) {
	if s == nil {
		return nil, false
	}
	if val, ok := s.values[name]; ok {
		return val, ok
	}
	return s.outer.get(name)
}

func (s *scope) set(name string, val Value) Value {
	s.values[name] = val
	return val
}
