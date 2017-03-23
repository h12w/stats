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
	"io"
	"math"

	"github.com/dgryski/go-farm"
	"github.com/dgryski/go-pcgr"
)

/*
Sketch is a Count-Min-Log Sketch 16-bit registers
*/
type Sketch struct {
	w   int32
	d   int32
	exp float64

	store []uint16
}

/*
NewSketch returns a new Count-Min-Log Sketch with 16-bit registers
*/
func newSketch(w int32, d int32, exp float64) (*Sketch, error) {
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

	return newSketch(int32(m/w), int32(w), 1.00026)
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
	w := int(cml.w)
	sk := make([]*uint16, cml.d, cml.d)
	c := uint16(math.MaxUint16)

	hsum := farm.Hash64(e)
	h1 := uint32(hsum & 0xffffffff)
	h2 := uint32((hsum >> 32) & 0xffffffff)

	for i := range sk {
		saltedHash := int((h1 + uint32(i)*h2))
		if sk[i] = &cml.store[i*w+(saltedHash%w)]; *sk[i] < c {
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
	w := int(cml.w)
	d := int(cml.d)
	c := uint16(math.MaxUint16)

	hsum := farm.Hash64(e)
	h1 := uint32(hsum & 0xffffffff)
	h2 := uint32((hsum >> 32) & 0xffffffff)

	for i := 0; i < d; i++ {
		saltedHash := int((h1 + uint32(i)*h2))
		if sk := cml.store[i*w+(saltedHash%w)]; sk < c {
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

func (s *Sketch) WriteTo(w io.Writer) (int64, error) {
	var err error
	var nn int
	n := 0
	nn, err = writeInt32(w, s.w)
	if err != nil {
		return int64(n), err
	}
	n += nn
	nn, err = writeInt32(w, s.d)
	if err != nil {
		return int64(n), err
	}
	n += nn
	nn, err = writeFloat64(w, s.exp)
	if err != nil {
		return int64(n), err
	}
	n += nn
	nn, err = writeUint64Slice(w, s.store)
	if err != nil {
		return int64(n), err
	}
	n += nn
	return int64(n), nil
}

func (s *Sketch) ReadFrom(r io.Reader) (int64, error) {
	var err error
	var nn int
	n := 0
	nn, err = readInt32(r, &s.w)
	if err != nil {
		return int64(n), err
	}
	n += nn
	nn, err = readInt32(r, &s.d)
	if err != nil {
		return int64(n), err
	}
	n += nn
	nn, err = readFloat64(r, &s.exp)
	if err != nil {
		return int64(n), err
	}
	n += nn
	nn, err = readUint64Slice(r, &s.store)
	if err != nil {
		return int64(n), err
	}
	n += nn
	return int64(n), nil
}

func writeUint64Slice(w io.Writer, s []uint16) (int, error) {
	var err error
	var nn int
	n := 0
	nn, err = writeInt64(w, int64(len(s)))
	if err != nil {
		return n, err
	}
	n += nn
	for _, v := range s {
		nn, err = writeUint16(w, v)
		if err != nil {
			return n, err
		}
		n += nn
	}
	return n, nil
}

func readUint64Slice(r io.Reader, s *[]uint16) (int, error) {
	var err error
	var nn int
	n := 0
	var size int64
	nn, err = readInt64(r, &size)
	if err != nil {
		return n, err
	}
	n += nn
	*s = make([]uint16, int(size))
	for i := range *s {
		nn, err = readUint16(r, &(*s)[i])
		if err != nil {
			return n, err
		}
		n += nn
	}
	return n, nil
}

func writeInt64(w io.Writer, i int64) (int, error) {
	return w.Write([]byte{byte(i >> 56), byte(i >> 48), byte(i >> 40), byte(i >> 32), byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
}

func readInt64(r io.Reader, i *int64) (int, error) {
	var b [8]byte
	n, err := io.ReadFull(r, b[:])
	if err != nil {
		return n, err
	}
	*i = int64(b[0])<<56 | int64(b[1])<<48 | int64(b[2])<<40 | int64(b[3])<<32 |
		int64(b[4])<<24 | int64(b[5])<<16 | int64(b[6])<<8 | int64(b[7])
	return n, nil
}

func writeInt32(w io.Writer, i int32) (int, error) {
	return w.Write([]byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
}

func readInt32(r io.Reader, i *int32) (int, error) {
	var b [4]byte
	n, err := io.ReadFull(r, b[:])
	if err != nil {
		return n, err
	}
	*i = int32(b[0])<<24 | int32(b[1])<<16 | int32(b[2])<<8 | int32(b[3])
	return n, nil
}

func writeFloat64(w io.Writer, i float64) (int, error) {
	return w.Write([]byte{byte(uint64(i) >> 56), byte(uint64(i) >> 48), byte(uint64(i) >> 40), byte(uint64(i) >> 32), byte(uint64(i) >> 24), byte(uint64(i) >> 16), byte(uint64(i) >> 8), byte(uint64(i))})
}

func readFloat64(r io.Reader, i *float64) (int, error) {
	var b [8]byte
	n, err := io.ReadFull(r, b[:])
	if err != nil {
		return n, err
	}
	*i = float64(uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7]))
	return n, nil
}

func writeUint16(w io.Writer, i uint16) (int, error) {
	return w.Write([]byte{byte(i >> 8), byte(i)})
}

func readUint16(r io.Reader, i *uint16) (int, error) {
	var b [2]byte
	n, err := io.ReadFull(r, b[:])
	if err != nil {
		return n, err
	}
	*i = uint16(b[0])<<8 | uint16(b[1])
	return n, nil
}
