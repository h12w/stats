package stats

import (
	"encoding/json"
	"sync"
)

// S is the container for all statistics
type S struct {
	Meters         map[Key]*Meter `json:"meters"`
	defaultBufSize int            `json:"-"`
	mu             sync.RWMutex   `json:"-"`
}

// New creates a new S
func New() *S {
	return &S{
		Meters:         make(map[Key]*Meter),
		defaultBufSize: 60,
	}
}

// Meter gets or creates a meter by name
func (s *S) Meter(name string, tags Tags) *Meter {
	return s.meter(NewKey(name, tags))
}

func (s *S) meter(key Key) *Meter {
	s.mu.RLock()
	defaultBufSize := s.defaultBufSize
	m, ok := s.Meters[key]
	s.mu.RUnlock()
	if !ok {
		m = NewMeter(defaultBufSize)
		s.mu.Lock()
		s.Meters[key] = m
		s.mu.Unlock()
	}
	return m
}

func (s *S) Merge(o *S) {
	o.mu.RLock()
	for key, meter := range o.Meters {
		s.meter(key).Merge(meter)
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
