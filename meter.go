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

func NewMeter(bufSize int) *Meter {
	return &Meter{
		a: make([]int, bufSize),
	}
}

func (m *Meter) Inc(t time.Time) {
	m.Add(t, 1)
}

func (m *Meter) Add(t time.Time, value int) {
	m.mu.Lock()
	m.add(int(t.Unix()), value)
	m.mu.Unlock()
}

func (m *Meter) add(sec, value int) {
	if m.startSec == 0 {
		m.startSec = sec
	}
	for sec-m.startSec >= len(m.a) {
		m.a[m.start] = 0
		m.start++
		if m.start == len(m.a) {
			m.start = 0
		}
		m.startSec++
	}
	offset := sec - m.startSec
	if offset < 0 {
		return // ignore data older than a circle
	}
	pos := m.start + offset
	if pos >= len(m.a) {
		pos -= len(m.a)
	}
	m.a[pos] += value
}

func (m *Meter) get(sec int) int {
	if sec < m.startSec || sec >= m.startSec+len(m.a) {
		return 0
	}
	pos := m.start + sec - m.startSec
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
	var a []int
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	if len(a) == 0 {
		return nil
	}
	m.startSec = a[0]
	copy(m.a, a[1:])
	return nil
}

func (m *Meter) String() string {
	buf, _ := m.MarshalJSON()
	return string(buf)
}
