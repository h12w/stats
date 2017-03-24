package stats

import "testing"
import "os"
import "bufio"
import "h12.me/stats/internal/cml"

func TestSketch(t *testing.T) {
	s := NewRingCMLSketcher(24, 1000000, 0)
	key := []byte("b")
	for i := 0; i < 50; i++ {
		s.Inc(0, key)
		if cnt := int(s.Get(0, key) + 0.5); cnt != i+1 {
			t.Fatal(cnt, i+1)
		}
	}
	{
		f, err := os.Create("ring_cml_sketch.bin")
		if err != nil {
			t.Fatal(err)
		}
		buf := bufio.NewWriter(f)
		if _, err := s.WriteTo(buf); err != nil {
			t.Fatal(err)
		}
		buf.Flush()
		f.Close()
	}
	{
		f, err := os.Open("ring_cml_sketch.bin")
		if err != nil {
			t.Fatal(err)
		}
		buf := bufio.NewReader(f)
		if _, err := s.ReadFrom(buf, func() Sketcher { return &cml.Sketch{} }); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	if cnt := int(s.Get(0, key) + 0.5); cnt != 50 {
		t.Fatal(cnt, 50)
	}

}
