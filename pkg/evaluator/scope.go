package evaluator

type Scope struct {
	store map[string]Value
	outer *Scope
}

func NewScope() *Scope {
	return &Scope{store: map[string]Value{}}
}

func NewEnclosedScope(outer *Scope) *Scope {
	return &Scope{store: map[string]Value{}, outer: outer}
}

func (s *Scope) Get(name string) (Value, bool) {
	if s == nil {
		return nil, false
	}
	if val, ok := s.store[name]; ok {
		return val, ok
	}
	return s.outer.Get(name)
}

func (s *Scope) Set(name string, val Value) Value {
	s.store[name] = val
	return val
}
