package stats

import (
	"io"

	"math"

	"h12.me/stats/binary"
	"h12.me/stats/internal/cml"
)

func NewCMLRingSketcher(ringSize int, elemCap int, startOffset int64) *RingSketcher {
	a := make([]Sketcher, ringSize)
	for i := range a {
		var err error
		a[i], err = cml.New(elemCap, 0.001)
		if err != nil {
			panic(err)
		}
	}
	return NewRingSketcher(startOffset, a)
}

func NewMapRingSketcher(ringSize int, elemCap int, startOffset int64) *RingSketcher {
	if elemCap >= math.MaxInt32 {
		panic("elemCap should be within int32")
	}
	a := make([]Sketcher, ringSize)
	for i := range a {
		var err error
		a[i] = NewUint8MapSketcher(int32(elemCap))
		if err != nil {
			panic(err)
		}
	}
	return NewRingSketcher(startOffset, a)
}

type Uint8MapSketcher struct {
	m        map[string]uint8
	capLimit int32
}

func NewUint8MapSketcher(capLimit int32) *Uint8MapSketcher {
	return &Uint8MapSketcher{
		m:        make(map[string]uint8),
		capLimit: capLimit,
	}
}

func (s *Uint8MapSketcher) Get(key []byte) float64 {
	return float64(s.m[string(key)])
}

func (s *Uint8MapSketcher) Inc(key []byte) {
	if len(s.m) >= int(s.capLimit) {
		return
	}

	k := string(key)
	if s.m[k] == 255 {
		return
	}

	s.m[k]++
}

func (s *Uint8MapSketcher) Reset() {
	s.m = make(map[string]uint8)
}

func (s *Uint8MapSketcher) WriteTo(w io.Writer) (int64, error) {
	var nn int
	var err error
	n := 0
	nn, err = binary.WriteInt32(w, int32(s.capLimit))
	n += nn
	if err != nil {
		return int64(n), err
	}
	nn, err = binary.WriteInt32(w, int32(len(s.m)))
	n += nn
	if err != nil {
		return int64(n), err
	}
	for k, v := range s.m {
		nn, err = binary.WriteString(w, k)
		n += nn
		if err != nil {
			return int64(n), err
		}
		nn, err = binary.WriteUint8(w, v)
		n += nn
		if err != nil {
			return int64(n), err
		}
	}
	return int64(n), nil
}

func (s *Uint8MapSketcher) ReadFrom(r io.Reader) (int64, error) {
	var nn int
	var err error
	n := 0
	nn, err = binary.ReadInt32(r, &s.capLimit)
	n += nn
	if err != nil {
		return int64(n), err
	}
	var size int32
	nn, err = binary.ReadInt32(r, &size)
	n += nn
	if err != nil {
		return int64(n), err
	}
	s.m = make(map[string]uint8)
	for i := 0; i < int(size); i++ {
		var k string
		var v uint8
		nn, err = binary.ReadString(r, &k)
		n += nn
		if err != nil {
			return int64(n), err
		}
		nn, err = binary.ReadUint8(r, &v)
		n += nn
		if err != nil {
			return int64(n), err
		}
		s.m[k] = v
	}
	return int64(n), nil
}
