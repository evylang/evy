package evaluator

type ranger interface {
	next(scope *scope, loopVarName string) bool
}

type stepRange struct {
	cur  float64
	stop float64
	step float64
}

type arrayRange struct {
	cur   int
	array *arrayVal
}

type mapRange struct {
	cur    int // index of Map.Order slice of keys
	mapVal *mapVal
	order  []string // copy of order in case map entry gets deleted during iteration
}

type stringRange struct {
	cur   int
	str   *stringVal
	runes []rune
}

func (s *stepRange) next(scope *scope, loopVarName string) bool {
	if s.step > 0 && s.cur >= s.stop {
		return false
	}
	if s.step < 0 && s.cur <= s.stop {
		return false
	}
	scope.update(loopVarName, &numVal{V: s.cur})
	s.cur += s.step
	return true
}

func (a *arrayRange) next(scope *scope, loopVarName string) bool {
	elements := *a.array.Elements
	if a.cur >= len(elements) {
		return false
	}
	scope.update(loopVarName, elements[a.cur])
	a.cur++
	return true
}

func (m *mapRange) next(scope *scope, loopVarName string) bool {
	for m.cur < len(m.order) {
		key := m.order[m.cur]
		m.cur++
		if _, ok := m.mapVal.Pairs[key]; ok { // ensure value hasn't been deleted
			scope.update(loopVarName, &stringVal{V: key})
			return true
		}
	}
	return false
}

func (s *stringRange) next(scope *scope, loopVarName string) bool {
	if s.runes == nil {
		s.runes = s.str.runes()
	}
	if s.cur >= len(s.runes) {
		return false
	}
	scope.update(loopVarName, &stringVal{V: string(s.runes[s.cur])})

	s.cur++
	return true
}
