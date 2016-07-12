package stats

import (
	"bytes"
	"fmt"
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
	m.a[pos]++
	m.mu.Unlock()
}

func (m *Meter) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var buf bytes.Buffer
	t := time.Unix(int64(m.startSec), 0).UTC()
	for i := m.start; i < len(m.a); i++ {
		fmt.Fprintf(&buf, "%s %d\n", t.Format("2006-01-02T15:04:05"), m.a[i])
		t = t.Add(time.Second)
	}
	for i := 0; i < m.start; i++ {
		fmt.Fprintf(&buf, "%s %d\n", t.Format("2006-01-02T15:04:05"), m.a[i])
		t = t.Add(time.Second)
	}
	return buf.String()
}
