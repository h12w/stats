package stats

import (
	"bytes"
	"encoding/json"
	"strconv"
	"sync"
	"time"
)

type Meter struct {
	a        []int
	start    int
	startSec int
	mu       sync.RWMutex
}

func NewMeter(start time.Time, size int) *Meter {
	return &Meter{
		start:    0,
		startSec: int(start.Unix()),
		a:        make([]int, size),
	}
}

func (m *Meter) Inc(t time.Time, value int) {
	m.mu.Lock()
	m.add(int(t.Unix()), value)
	m.mu.Unlock()
}

func (m *Meter) Get(sec int) int {
	m.mu.RLock()
	v := m.get(sec)
	m.mu.RUnlock()
	return v
}

func (m *Meter) add(sec, value int) {
	if sec < m.startSec {
		return
	}
	// TRUE: sec >= m.startSec

	for sec >= m.startSec+len(m.a) {
		m.a[m.start] = 0
		m.start++
		m.startSec++
		if m.start == len(m.a) {
			m.start = 0
		}
	}
	// TRUE: m.startSec <= sec && sec < m.startSec+len(m.a)
	// TRUE: 0 <= sec-m.startSec && sec-m.startSec < len(m.a)

	pos := m.start + (sec - m.startSec)
	if pos >= len(m.a) {
		pos -= len(m.a)
	}
	m.a[pos] += value
}

func (m *Meter) get(sec int) int {
	if sec < m.startSec || sec >= m.startSec+len(m.a) {
		return 0
	}
	pos := m.start + (sec - m.startSec)
	if pos >= len(m.a) {
		pos -= len(m.a)
	}
	return m.a[pos]
}

func (m *Meter) Merge(o *Meter) {
	m.mu.Lock()
	for i := range o.a {
		sec := o.startSec + i
		m.add(sec, o.get(sec))
	}
	m.mu.Unlock()
}

func (m *Meter) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var buf bytes.Buffer
	buf.WriteString("[")
	if len(m.a) > 0 {
		buf.WriteString(strconv.Itoa(m.startSec))
		sec := m.startSec
		buf.WriteByte(',')
		buf.WriteString(strconv.Itoa(m.get(sec)))
		for i := 1; i < len(m.a); i++ {
			sec := m.startSec + i
			buf.WriteByte(',')
			buf.WriteString(strconv.Itoa(m.get(sec)))
		}
	}
	buf.WriteString("]")
	return buf.Bytes(), nil
}

func (m *Meter) UnmarshalJSON(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var ints []int
	if err := json.Unmarshal(data, &ints); err != nil {
		return err
	}
	if len(ints) == 0 {
		return nil
	}
	m.startSec = ints[0]
	size := len(m.a)
	m.a = ints[1:]
	if len(m.a) < size {
		m.a = append(m.a, make([]int, size-len(m.a))...)
	}
	return nil
}

func (m *Meter) String() string {
	buf, _ := m.MarshalJSON()
	return string(buf)
}
