package bimap

import (
	"testing"
)

func TestMap(t *testing.T) {
	strs := []string{"a", "b", "c"}
	m := New()
	if size := m.Size(); size != 0 {
		t.Errorf("expected empty map; got size %d\n", size)
	}
	for i, s := range strs {
		if id := m.Add(s); id != i {
			t.Errorf("expected %d; got %d\n", i, id)
		}
	}
	for i, s := range strs {
		if id := m.FindString(s); id != i {
			t.Errorf("expected %d; got %d\n", i, id)
		}
		if ss := m.FindInt(i); ss != s {
			t.Errorf("expected %q; got %q\n", s, ss)
		}
	}
	if s := m.FindInt(-1); s != "" {
		t.Errorf("expected empty; got %q\n", s)
	}
	if s := m.FindInt(m.Size()); s != "" {
		t.Errorf("expected empty; got %q\n", s)
	}
	if i := m.FindString("abc"); i != NoInt {
		t.Errorf("expected NoInt; got %d\n", i)
	}

	var intArray [3]int
	intSlice := intArray[:0]
	m.AppendString(strs, &intSlice)
	if intArray != [...]int{0, 1, 2} {
		t.Errorf("expected [0, 1, 2]; got %v\n", intArray)
	}

	if n := copy(intArray[:], m.TranslateString(strs)); n != 3 {
		t.Errorf("expected %d elements copied; got %d\n", 3, n)
	}
	if intArray != [...]int{0, 1, 2} {
		t.Errorf("expected [0, 1, 2]; got %v\n", intArray)
	}

	var strArray [3]string
	strSlice := strArray[:0]
	m.AppendInt([]int{0, 1, 2}, &strSlice)
	if strArray != [...]string{"a", "b", "c"} {
		t.Errorf("expected [a, b, c]; got %v\n", strArray)
	}

	if n := copy(strArray[:], m.TranslateInt([]int{0, 1, 2})); n != 3 {
		t.Errorf("expected %d element copied; got %d\n", 3, n)
	}
	if strArray != [...]string{"a", "b", "c"} {
		t.Errorf("expected [a, b, c]; got %v\n", strArray)
	}

	func() {
		defer func() {
			err := recover()
			if err == nil {
				t.Errorf("expected error; got nil\n")
			}
		}()
		m.Add("")
	}()
}
