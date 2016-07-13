package stats

import (
	"bytes"
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
	return &Meter{a: make([]int, bufSize)}
}

func (m *Meter) Inc(t time.Time) {
	m.Add(t, 1)
}

func (m *Meter) Add(t time.Time, value int) {
	m.mu.Lock()
	sec := int(t.Unix())
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
		m.mu.Unlock()
		return // ignore data older than a circle
	}
	pos := m.start + offset
	if pos >= len(m.a) {
		pos -= len(m.a)
	}
	m.a[pos] += value
	m.mu.Unlock()
}

func (m *Meter) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var buf bytes.Buffer
	buf.WriteString("[")
	buf.WriteString(strconv.Itoa(m.startSec))
	buf.WriteByte(',')
	buf.WriteString(strconv.Itoa(m.a[m.start]))
	for i := m.start + 1; i < len(m.a); i++ {
		buf.WriteByte(',')
		buf.WriteString(strconv.Itoa(m.a[i]))
	}
	for i := 0; i < m.start; i++ {
		buf.WriteByte(',')
		buf.WriteString(strconv.Itoa(m.a[i]))
	}
	buf.WriteString("]")
	return buf.Bytes(), nil
}

func (m *Meter) String() string {
	buf, _ := m.MarshalJSON()
	return string(buf)
}
