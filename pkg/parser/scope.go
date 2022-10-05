package parser

type scope struct {
	vars  map[string]*Var
	outer *scope

	returnType *Type
}

func newScope() *scope {
	return &scope{vars: map[string]*Var{}, returnType: ANY_TYPE}
}

func newInnerScope(outer *scope) *scope {
	return &scope{vars: map[string]*Var{}, outer: outer, returnType: outer.returnType}
}

func newInnerScopeWithReturnType(outer *scope, returnType *Type) *scope {
	return &scope{vars: map[string]*Var{}, outer: outer, returnType: returnType}
}

func (s *scope) inLocalScope(name string) bool {
	_, ok := s.vars[name]
	return ok
}

func (s *scope) get(name string) (*Var, bool) {
	if s == nil {
		return nil, false
	}
	if v, ok := s.vars[name]; ok {
		return v, ok
	}
	return s.outer.get(name)
}

func (s *scope) set(name string, v *Var) *Var {
	s.vars[name] = v
	return v
}
