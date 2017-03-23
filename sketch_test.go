package stats

import "testing"

func TestSketch(t *testing.T) {
	s := NewRingCMLSketcher(24, 1000000, 0)
	for i := 0; i < 50; i++ {
		key := []byte("b")
		s.Inc(0, key)
		if cnt := int(s.Get(0, key) + 0.5); cnt != i+1 {
			t.Fatal(cnt, i+1)
		}
	}
}
