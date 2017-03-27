package stats

import (
	"io"

	"h12.me/stats/binary"
)

type (
	RingSketcher struct {
		start       int64
		offset      int64
		a           []Sketcher
		newSketcher func() Sketcher
	}
	Sketcher interface {
		Get([]byte) float64
		Inc([]byte)
		Reset()
		WriteTo(io.Writer) (int64, error)
		ReadFrom(io.Reader) (int64, error)
	}
)

func NewRingSketcher(offset int64, a []Sketcher, newSketcher func() Sketcher) *RingSketcher {
	if len(a) < 1 {
		panic("must have at least 1 sketcher")
	}
	return &RingSketcher{
		start:       0,
		offset:      offset,
		a:           a,
		newSketcher: newSketcher,
	}
}

func (m *RingSketcher) Offset() int64 {
	return m.offset
}

func (m *RingSketcher) Get(offset int64, key []byte) float64 {
	if offset < m.offset || offset >= m.offset+int64(len(m.a)) {
		return 0
	}
	pos := int(m.start + offset - m.offset)
	if pos >= len(m.a) {
		pos -= len(m.a)
	}
	return m.a[pos].Get(key)
}

func (m *RingSketcher) Inc(offset int64, key []byte) {
	if offset < m.offset {
		return
	}
	// TRUE: offset >= m.offset

	for offset >= m.offset+int64(len(m.a)) {
		m.a[m.start].Reset()
		m.start++
		m.offset++
		if int(m.start) == len(m.a) {
			m.start = 0
		}
	}
	// TRUE: m.offset <= offset && offset < m.offset+len(m.a)

	delta := offset - m.offset
	// TRUE: 0 <= delta && delta < len(m.a)

	pos := int(m.start + delta)
	if pos >= len(m.a) {
		pos -= len(m.a)
	}
	m.a[pos].Inc(key)
}

func (s *RingSketcher) WriteTo(w io.Writer) (int64, error) {
	var err error
	var nn int
	n := 0
	nn, err = binary.WriteInt64(w, s.start)
	if err != nil {
		return int64(n), err
	}
	n += nn
	nn, err = binary.WriteInt64(w, s.offset)
	if err != nil {
		return int64(n), err
	}
	n += nn
	nn, err = binary.WriteInt64(w, int64(len(s.a)))
	if err != nil {
		return int64(n), err
	}
	n += nn
	for i := range s.a {
		nn, err := s.a[i].WriteTo(w)
		if err != nil {
			return int64(n), err
		}
		n += int(nn)
	}
	return int64(n), nil
}

func (s *RingSketcher) ReadFrom(r io.Reader) (int64, error) {
	var err error
	var nn int
	n := 0
	nn, err = binary.ReadInt64(r, &s.start)
	if err != nil {
		return int64(n), err
	}
	n += nn
	nn, err = binary.ReadInt64(r, &s.offset)
	if err != nil {
		return int64(n), err
	}
	n += nn
	var size int64
	nn, err = binary.ReadInt64(r, &size)
	if err != nil {
		return int64(n), err
	}
	n += nn
	s.a = make([]Sketcher, int(size))
	for i := range s.a {
		s.a[i] = s.newSketcher()
		nn, err := s.a[i].ReadFrom(r)
		if err != nil {
			return int64(n), err
		}
		n += int(nn)
	}
	return int64(n), nil
}
