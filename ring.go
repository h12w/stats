package stats

type (
	RingSketcher struct {
		a      []Sketcher
		start  int
		offset int
	}
	Sketcher interface {
		Get([]byte) float64
		Inc([]byte)
		Reset()
	}
)

func NewRingSketcher(offset int, a []Sketcher) *RingSketcher {
	return &RingSketcher{
		a:      a,
		start:  0,
		offset: offset,
	}
}

func (m *RingSketcher) Offset() int {
	return m.offset
}

func (m *RingSketcher) Get(offset int, key []byte) float64 {
	if offset < m.offset || offset >= m.offset+len(m.a) {
		return 0
	}
	pos := m.start + (offset - m.offset)
	if pos >= len(m.a) {
		pos -= len(m.a)
	}
	return m.a[pos].Get(key)
}

func (m *RingSketcher) Inc(offset int, key []byte) {
	if offset < m.offset {
		return
	}
	// TRUE: offset >= m.offset

	for offset >= m.offset+len(m.a) {
		m.a[m.start].Reset()
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
	m.a[pos].Inc(key)
}
