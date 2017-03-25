package stats

import (
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"h12.me/gspec/util"
)

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
