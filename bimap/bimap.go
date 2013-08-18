// Package bimap implements bi-directional mapping between strings and
// integral ids.

package bimap

// Speical constants that may be returned from certain methods that
// access a Map.
const (
	NoInt int = -1
)

// Map is a bi-directional mapping between strings and integers. The
// integers forms a dense range starting a 0.
type Map struct {
	strToInt map[string]int
	intToStr []string
}

// New creates an empty Map
func New() *Map {
	return &Map{make(map[string]int), make([]string, 0, 1024)}
}

// FromSlice creates an empty Map from the given slice
func FromSlice(s []string) *Map {
	m := New()
	for _, i := range s {
		m.Add(i)
	}
	return m
}

// Add adds the given string into the map and returns its id. The
// string being added should not be empty. This is not thread safe.
func (m *Map) Add(s string) int {
	if len(s) == 0 {
		panic("trying to add an empty string")
	}
	i, ok := m.strToInt[s]
	if !ok {
		i = len(m.intToStr)
		m.strToInt[s] = i
		m.intToStr = append(m.intToStr, s)
	}
	return i
}

// FindByString finds the id or returns NoInt if the string is not in
// the map.
func (m *Map) FindByString(s string) int {
	i, ok := m.strToInt[s]
	if ok {
		return i
	}
	return NoInt
}

// FindByInt finds the string corresponding to the given integral
// id. Returns the string if the id is in the map; or an empty string
// if it is not.
func (m *Map) FindByInt(i int) string {
	if 0 <= i && i < len(m.intToStr) {
		return m.intToStr[i]
	}
	return ""
}

// AppendByString translates a slice of string and appends the result
// to the given slice. New words are also added to the Map.
func (m *Map) AppendByString(strs []string, ints *[]int) {
	for _, s := range strs {
		*ints = append(*ints, m.Add(s))
	}
}

// TranslateByString translates a slice of string into a slice of
// integers. New words are also added to the Map.
func (m *Map) TranslateByString(strs []string) []int {
	ints := make([]int, len(strs))
	for i, s := range strs {
		ints[i] = m.Add(s)
	}
	return ints
}

// AppendByInt translates a slice of integers and appends the result
// to the given slice.
func (m *Map) AppendByInt(ints []int, strs *[]string) {
	for _, id := range ints {
		*strs = append(*strs, m.FindByInt(id))
	}
}

// TranslateByInt translates a slice of integers into a slice of
// strings.
func (m *Map) TranslateByInt(ints []int) []string {
	strs := make([]string, len(ints))
	for i, id := range ints {
		strs[i] = m.FindByInt(id)
	}
	return strs
}

// Size returns the size of the map, which is also the next id to be
// assigned.
func (m *Map) Size() int {
	return len(m.intToStr)
}
