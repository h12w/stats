package stats

import "sync"

// S is the container for all statistics
type S struct {
	Meters  map[string]*Meter `json:"meters"`
	bufSize int               `json:"-"`
	mu      sync.RWMutex      `json:"-"`
}

// New creates a new S
func New(bufSize int) *S {
	return &S{
		Meters:  make(map[string]*Meter),
		bufSize: bufSize,
	}
}

// Meter gets or creates a meter by name
func (s *S) Meter(name string) *Meter {
	s.mu.RLock()
	m, ok := s.Meters[name]
	s.mu.RUnlock()
	if !ok {
		m = NewMeter(s.bufSize)
		s.mu.Lock()
		s.Meters[name] = m
		s.mu.Unlock()
	}
	return m
}