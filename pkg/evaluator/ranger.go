package evaluator

type ranger interface {
	next() bool
}

type stepRange struct {
	loopVar *Num
	cur     float64
	stop    float64
	step    float64
}

type arrayRange struct {
	loopVar Value
	cur     int
	array   *Array
}

type mapRange struct {
	loopVar Value
	cur     int // index of Map.Order slice of keys
	mapVal  *Map
	order   []string // copy of order in case map entry gets deleted during iteration
}

type stringRange struct {
	loopVar *String
	cur     int
	str     *String
	runes   []rune
}

func (s *stepRange) next() bool {
	if s.step > 0 && s.cur >= s.stop {
		return false
	}
	if s.step < 0 && s.cur <= s.stop {
		return false
	}
	s.loopVar.Val = s.cur
	s.cur += s.step
	return true
}

func (a *arrayRange) next() bool {
	elements := *a.array.Elements
	if a.cur >= len(elements) {
		return false
	}
	a.loopVar.Set(elements[a.cur])
	a.cur++
	return true
}

func (m *mapRange) next() bool {
	for m.cur < len(m.order) {
		key := m.order[m.cur]
		m.cur++
		if _, ok := m.mapVal.Pairs[key]; ok { // ensure value hasn't been deleted
			m.loopVar.(*String).Val = key
			return true
		}
	}
	return false
}

func (s *stringRange) next() bool {
	if s.runes == nil {
		s.runes = s.str.runes()
	}
	if s.cur >= len(s.runes) {
		return false
	}
	s.loopVar.Val = string(s.runes[s.cur])
	s.cur++
	return true
}
