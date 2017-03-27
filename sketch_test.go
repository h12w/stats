package stats

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"h12.me/gspec/util"
)

func TestSketch(t *testing.T) {
	testSketch(t, NewCMLRingSketcher(24, 1000000, 0))
	testSketch(t, NewMapRingSketcher(2, 1000000, 0))
}

func testSketch(t *testing.T, s *RingSketcher) {
	key := []byte("b")
	for i := 0; i < 50; i++ {
		s.Inc(0, key)
		if cnt := int(s.Get(0, key) + 0.5); cnt != i+1 {
			t.Fatal(cnt, i+1)
		}
	}
	{
		f, err := os.Create("ring_sketch.bin")
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
		f, err := os.Open("ring_sketch.bin")
		if err != nil {
			t.Fatal(err)
		}
		buf := bufio.NewReader(f)
		if _, err := s.ReadFrom(buf); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	if cnt := int(s.Get(0, key) + 0.5); cnt != 50 {
		t.Fatal(cnt, 50)
	}

}

func TestMapSize(t *testing.T) {
	if testing.Short() {
		return
	}
	fmt.Println(util.RandString(61))
	before := util.MemAlloc()
	m := make(map[string]uint8)
	for i := 0; i < 1000*1000; i++ {
		m[util.RandString(61)] = uint8(rand.Intn(255))
	}
	fmt.Println(len(m))
	fmt.Println((float64(util.MemAlloc()-before) / 1024 / 1024))
	f, err := os.Create("test.bin")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	start := time.Now()
	if err := gob.NewEncoder(f).Encode(m); err != nil {
		log.Fatal(err)
	}
	fmt.Println(time.Since(start))
}
