package abi

type Scope struct {
	values map[string]Value
	Outer  *Scope
}

func NewScope() *Scope {
	return &Scope{values: map[string]Value{}}
}

func NewInnerScope(outer *Scope) *Scope {
	return &Scope{values: map[string]Value{}, Outer: outer}
}

func (s *Scope) Get(name string) (Value, bool) {
	if s == nil || name == "_" {
		return nil, false
	}
	if val, ok := s.values[name]; ok {
		return val, true
	}
	return s.Outer.Get(name)
}

func (s *Scope) Set(name string, val Value) {
	if name == "_" {
		return
	}
	s.values[name] = val
}
