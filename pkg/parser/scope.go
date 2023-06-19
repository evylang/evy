package parser

type scope struct {
	vars       map[string]*Var
	outer      *scope
	block      Node
	returnType *Type
}

func newScope(outer *scope, node Node) *scope {
	if outer == nil {
		return newScopeWithReturnType(nil, node, nil)
	}
	return newScopeWithReturnType(outer, node, outer.returnType)
}

func newScopeWithReturnType(outer *scope, node Node, returnType *Type) *scope {
	return &scope{
		vars:       map[string]*Var{},
		block:      node,
		outer:      outer,
		returnType: returnType,
	}
}

func (s *scope) inLocalScope(name string) bool {
	_, ok := s.vars[name]
	return ok
}

func (s *scope) get(name string) (*Var, bool) {
	if s == nil || name == "_" {
		return nil, false
	}
	if v, ok := s.vars[name]; ok {
		return v, ok
	}
	return s.outer.get(name)
}

func (s *scope) set(name string, v *Var) {
	if name != "_" {
		s.vars[name] = v
	}
}
