package stats

import (
	"encoding/json"
	"sync"
	"time"
)

var DefaultBufferSize = 60

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
		defaultBufSize: DefaultBufferSize,
	}
}

// Meter gets or creates a meter by name
func (s *S) Meter(name string, tags Tags) *Meter {
	return s.meter(NewKey(name, tags), time.Now())
}

func (s *S) meter(key Key, start time.Time) *Meter {
	s.mu.Lock()
	defer s.mu.Unlock()
	defaultBufSize := s.defaultBufSize
	m, ok := s.Meters[key]
	if !ok {
		m = NewMeter(start, defaultBufSize)
		s.Meters[key] = m
	}
	return m
}

func (s *S) Merge(o *S, start time.Time) {
	o.mu.RLock() // lock o during reading
	for key, meter := range o.Meters {
		s.meter(key, start).Merge(meter)
	}
	o.mu.RUnlock()
}

func (s *S) MergeWithTags(o *S, start time.Time, tags Tags) error {
	o.mu.RLock() // lock o during reading
	for key, meter := range o.Meters {
		name, keyTags, err := key.Decode()
		if err != nil {
			return err
		}
		for k, v := range tags {
			keyTags[k] = v
		}
		s.meter(NewKey(name, keyTags), start).Merge(meter)
	}
	o.mu.RUnlock()
	return nil
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
