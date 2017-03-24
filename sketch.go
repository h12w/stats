package stats

import "h12.me/stats/internal/cml"

func NewRingCMLSketcher(ringSize int, elemCap int, startOffset int64) *RingSketcher {
	a := make([]Sketcher, ringSize)
	for i := range a {
		var err error
		a[i], err = cml.New(elemCap, 0.001)
		if err != nil {
			panic(err)
		}
	}
	return NewRingSketcher(startOffset, a)
}
