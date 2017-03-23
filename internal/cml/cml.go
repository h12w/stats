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

package cml

import (
	"errors"
	"math"

	"github.com/dgryski/go-farm"
	"github.com/dgryski/go-pcgr"
)

/*
Sketch is a Count-Min-Log Sketch 16-bit registers
*/
type Sketch struct {
	w   int
	d   int
	exp float64

	store []uint16
}

/*
NewSketch returns a new Count-Min-Log Sketch with 16-bit registers
*/
func newSketch(w int, d int, exp float64) (*Sketch, error) {
	store := make([]uint16, d*w)
	return &Sketch{
		w:     w,
		d:     d,
		exp:   exp,
		store: store,
	}, nil
}

/*
New returns a new Count-Min-Log Sketch with 16-bit registers optimized for a given max capacity and expected error rate
*/
func New(capacity int, e float64) (*Sketch, error) {
	if !(e >= 0.001 && e < 1.0) {
		return nil, errors.New("e needs to be >= 0.001 and < 1.0")
	}
	if capacity < 1000000 {
		capacity = 1000000
	}

	m := math.Ceil((float64(capacity) * math.Log(e)) / math.Log(1.0/(math.Pow(2.0, math.Log(2.0)))))
	w := math.Ceil(math.Log(2.0) * m / float64(capacity))

	return newSketch(int(m/w), int(w), 1.00026)
}

func (s *Sketch) Reset() {
	for i := range s.store {
		s.store[i] = 0
	}
}

func (cml *Sketch) increaseDecision(c uint16) bool {
	return randFloat() < 1/math.Pow(cml.exp, float64(c))
}

/*
Update increases the count of `s` by one, return true if added and the current count of `s`
*/
func (cml *Sketch) Inc(e []byte) {
	sk := make([]*uint16, cml.d, cml.d)
	c := uint16(math.MaxUint16)

	hsum := farm.Hash64(e)
	h1 := uint32(hsum & 0xffffffff)
	h2 := uint32((hsum >> 32) & 0xffffffff)

	for i := range sk {
		saltedHash := int((h1 + uint32(i)*h2))
		if sk[i] = &cml.store[i*cml.w+(saltedHash%cml.w)]; *sk[i] < c {
			c = *sk[i]
		}
	}

	if cml.increaseDecision(c) {
		for _, k := range sk {
			if *k == c {
				*k = c + 1
			}
		}
	}
}

func (cml *Sketch) pointValue(c uint16) float64 {
	if c == 0 {
		return 0
	}
	return math.Pow(cml.exp, float64(c-1))
}

func (cml *Sketch) value(c uint16) float64 {
	if c <= 1 {
		return cml.pointValue(c)
	}
	v := cml.pointValue(c + 1)
	return (1 - v) / (1 - cml.exp)
}

/*
Query returns the count of `e`
*/
func (cml *Sketch) Get(e []byte) float64 {
	c := uint16(math.MaxUint16)

	hsum := farm.Hash64(e)
	h1 := uint32(hsum & 0xffffffff)
	h2 := uint32((hsum >> 32) & 0xffffffff)

	for i := 0; i < cml.d; i++ {
		saltedHash := int((h1 + uint32(i)*h2))
		if sk := cml.store[i*cml.w+(saltedHash%cml.w)]; sk < c {
			c = sk
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
