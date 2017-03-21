/*
This is a fork from github.com/seiflotfy/count-min-log

The MIT License (MIT)

Copyright (c) 2015 Seif Lotfy <seif.lotfy@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package stats

import (
	"errors"
	"math"

	"github.com/dgryski/go-farm"
	"github.com/dgryski/go-pcgr"
)

/*
RingSketch is a Count-Min-Log Sketch with Ring16 registers
*/
type RingSketch struct {
	w   uint
	d   uint
	exp float64

	store [][]Ring16
}

/*
NewRingSketch returns a new Count-Min-Log Sketch with 16-bit registers
*/
func NewRingSketch(w uint, d uint, exp float64, offset, ringSize int) (*RingSketch, error) {
	store := make([][]Ring16, d, d)
	for i := range store {
		store[i] = make([]Ring16, w, w)
		for j := range store[i] {
			store[i][j] = NewRing16(offset, ringSize)
		}
	}
	return &RingSketch{
		w:     w,
		d:     d,
		exp:   exp,
		store: store,
	}, nil
}

/*
NewForCapacity16 returns a new Count-Min-Log Sketch with 16-bit registers optimized for a given max capacity and expected error rate
*/
func NewRingSketchWithCap(capacity uint64, e float64, offset, ringSize int) (*RingSketch, error) {
	if !(e >= 0.001 && e < 1.0) {
		return nil, errors.New("e needs to be >= 0.001 and < 1.0")
	}
	if capacity < 1000000 {
		capacity = 1000000
	}

	m := math.Ceil((float64(capacity) * math.Log(e)) / math.Log(1.0/(math.Pow(2.0, math.Log(2.0)))))
	w := math.Ceil(math.Log(2.0) * m / float64(capacity))

	return NewRingSketch(uint(m/w), uint(w), 1.00026, offset, ringSize)
}

func (cml *RingSketch) Size() int {
	return int(cml.d)*int(cml.w)*2 + 24
}

func (cml *RingSketch) increaseDecision(c uint16) bool {
	return randFloat() < 1/math.Pow(cml.exp, float64(c))
}

/*
Update increases the count of `s` by one, return true if added and the current count of `s`
*/
func (cml *RingSketch) Update(e []byte, offset int) bool {
	sk := make([]*Ring16, cml.d, cml.d)
	c := uint16(math.MaxUint16)

	hsum := farm.Hash64(e)
	h1 := uint32(hsum & 0xffffffff)
	h2 := uint32((hsum >> 32) & 0xffffffff)

	for i := range sk {
		saltedHash := uint((h1 + uint32(i)*h2))
		sk[i] = &cml.store[i][(saltedHash % cml.w)]
		if v := sk[i].Get(offset); v < c {
			c = v
		}
	}

	if cml.increaseDecision(c) {
		for _, k := range sk {
			if k.Get(offset) == c {
				k.Set(offset, c+1)
			}
		}
	}
	return true
}

func (cml *RingSketch) pointValue(c uint16) float64 {
	if c == 0 {
		return 0
	}
	return math.Pow(cml.exp, float64(c-1))
}

func (cml *RingSketch) value(c uint16) float64 {
	if c <= 1 {
		return cml.pointValue(c)
	}
	v := cml.pointValue(c + 1)
	return (1 - v) / (1 - cml.exp)
}

/*
Query returns the count of `e`
*/
func (cml *RingSketch) Query(e []byte, offset int) float64 {
	c := uint16(math.MaxUint16)

	hsum := farm.Hash64(e)
	h1 := uint32(hsum & 0xffffffff)
	h2 := uint32((hsum >> 32) & 0xffffffff)

	for i := range cml.store {
		saltedHash := uint((h1 + uint32(i)*h2))
		sk := cml.store[i][(saltedHash % cml.w)]
		if v := sk.Get(offset); v < c {
			c = v
		}
	}
	return cml.value(c)
}

var rnd = pcgr.Rand{
	State: 0x0ddc0ffeebadf00d,
	Inc:   0xcafebabe,
}

func randFloat() float64 {
	return float64(rnd.Next()%10e5) / 10e5
}
