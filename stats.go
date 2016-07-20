package stats

import (
	"encoding/json"
	"sync"
)

// S is the container for all statistics
type S struct {
	Meters         map[string]*Meter `json:"meters"`
	defaultBufSize int               `json:"-"`
	mu             sync.RWMutex      `json:"-"`
}

// New creates a new S
func New() *S {
	return &S{
		Meters:         make(map[string]*Meter),
		defaultBufSize: 60,
	}
}

// Meter gets or creates a meter by name
func (s *S) Meter(name string) *Meter {
	s.mu.RLock()
	defaultBufSize := s.defaultBufSize
	m, ok := s.Meters[name]
	s.mu.RUnlock()
	if !ok {
		m = NewMeter(defaultBufSize)
		s.mu.Lock()
		s.Meters[name] = m
		s.mu.Unlock()
	}
	return m
}

func (s *S) Merge(o *S) {
	o.mu.RLock()
	for name, meter := range o.Meters {
		s.Meter(name).Merge(meter)
	}
	o.mu.RUnlock()
}

func (s *S) SetBufSize(defaultBufSize int) *S {
	s.mu.Lock()
	s.defaultBufSize = defaultBufSize
	s.mu.Unlock()
	return s
}

func (s *S) String() string {
	buf, _ := json.Marshal(s)
	return string(buf)
}
