package parser

type scope struct {
	vars  map[string]*Var
	outer *scope
}

func newScope() *scope {
	return &scope{vars: map[string]*Var{}}
}

func newInnerScope(outer *scope) *scope {
	return &scope{vars: map[string]*Var{}, outer: outer}
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
