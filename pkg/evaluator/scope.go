package evaluator

func NewEnclosedScope(outer *Scope) *Scope {
	env := NewScope()
	env.outer = outer
	return env
}

func NewScope() *Scope {
	s := make(map[string]Object)
	return &Scope{store: s, outer: nil}
}

type Scope struct {
	store map[string]Object
	outer *Scope
}

func (s *Scope) Get(name string) (Object, bool) {
	obj, ok := s.store[name]
	if !ok && s.outer != nil {
		obj, ok = s.outer.Get(name)
	}
	return obj, ok
}

func (s *Scope) Set(name string, val Object) Object {
	s.store[name] = val
	return val
}
