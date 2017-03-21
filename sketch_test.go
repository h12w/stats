package stats

import "testing"

func TestSketch(t *testing.T) {
	s, err := NewRingSketchWithCap(1000000, 0.001, 0, 24)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 50; i++ {
		key := []byte("b")
		s.Update(key, 0)
		if cnt := int(s.Query(key, 0) + 0.5); cnt != i+1 {
			t.Fatal(cnt, i+1)
		}
	}
}
