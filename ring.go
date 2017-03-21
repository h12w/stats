package stats

type Ring16 struct {
	a      []uint16
	start  int
	offset int
}

func NewRing16(offset int, size int) Ring16 {
	return Ring16{
		start:  0,
		offset: offset,
		a:      make([]uint16, size),
	}
}

func (m *Ring16) Offset() int {
	return m.offset
}

func (m *Ring16) Get(offset int) uint16 {
	if offset < m.offset || offset >= m.offset+len(m.a) {
		return 0
	}
	pos := m.start + (offset - m.offset)
	if pos >= len(m.a) {
		pos -= len(m.a)
	}
	return m.a[pos]
}

func (m *Ring16) Set(offset int, value uint16) {
	if offset < m.offset {
		return
	}
	// TRUE: offset >= m.offset

	for offset >= m.offset+len(m.a) {
		m.a[m.start] = 0
		m.start++
		m.offset++
		if m.start == len(m.a) {
			m.start = 0
		}
	}
	// TRUE: m.offset <= offset && offset < m.offset+len(m.a)

	delta := offset - m.offset
	// TRUE: 0 <= delta && delta < len(m.a)

	pos := m.start + delta
	if pos >= len(m.a) {
		pos -= len(m.a)
	}
	m.a[pos] = value
}
