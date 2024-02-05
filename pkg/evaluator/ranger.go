package evaluator

import "evylang.dev/evy/pkg/abi"

type ranger interface {
	next() bool
}

type stepRange struct {
	loopVar *abi.NumVal
	cur     float64
	stop    float64
	step    float64
}

type arrayRange struct {
	loopVar abi.Value
	cur     int
	array   *abi.ArrayVal
}

type mapRange struct {
	loopVar  abi.Value
	cur      int // index of Map.Order slice of keys
	mapValal *abi.MapVal
	order    []string // copy of order in case map entry gets deleted during iteration
}

type stringRange struct {
	loopVar *abi.StringVal
	cur     int
	str     *abi.StringVal
	runes   []rune
}

func (s *stepRange) next() bool {
	if s.step > 0 && s.cur >= s.stop {
		return false
	}
	if s.step < 0 && s.cur <= s.stop {
		return false
	}
	if s.loopVar != nil {
		s.loopVar.V = s.cur
	}
	s.cur += s.step
	return true
}

func (a *arrayRange) next() bool {
	elements := *a.array.Elements
	if a.cur >= len(elements) {
		return false
	}
	if a.loopVar != nil {
		a.loopVar.Set(elements[a.cur])
	}

	a.cur++
	return true
}

func (m *mapRange) next() bool {
	for m.cur < len(m.order) {
		key := m.order[m.cur]
		m.cur++
		if _, ok := m.mapValal.Pairs[key]; ok { // ensure value hasn't been deleted
			if m.loopVar != nil {
				m.loopVar.(*abi.StringVal).V = key
			}
			return true
		}
	}
	return false
}

func (s *stringRange) next() bool {
	if s.runes == nil {
		s.runes = s.str.Runes()
	}
	if s.cur >= len(s.runes) {
		return false
	}
	if s.loopVar != nil {
		s.loopVar.V = string(s.runes[s.cur])
	}
	s.cur++
	return true
}
