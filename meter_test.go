package stats

import (
	"testing"
	"time"
)

var (
	testTimeStr = "2006-01-02T03:04:05Z"
	testTime, _ = time.Parse(testTimeStr, testTimeStr)
)

func TestMeter(t *testing.T) {
	now := testTime
	m := NewMeter(3)

	{
		m.Inc(now)
		m.Inc(now.Add(-time.Second))

		m.Inc(now.Add(time.Second))
		m.Inc(now.Add(time.Second))

		m.Inc(now.Add(2 * time.Second))
		m.Inc(now.Add(2 * time.Second))
		m.Inc(now.Add(2 * time.Second))
		expected := `2006-01-02T03:04:05 1
2006-01-02T03:04:06 2
2006-01-02T03:04:07 3
`
		if expected != m.String() {
			t.Fatalf("expect %s got %s", expected, m.String())
		}
	}

	now = now.Add(3 * time.Second)
	{
		m.Inc(now)
		m.Inc(now)

		expected := `2006-01-02T03:04:06 2
2006-01-02T03:04:07 3
2006-01-02T03:04:08 2
`
		if expected != m.String() {
			t.Fatalf("expect %s got %s", expected, m.String())
		}
	}

	now = now.Add(3 * time.Second)
	{
		m.Inc(now)

		expected := `2006-01-02T03:04:09 0
2006-01-02T03:04:10 0
2006-01-02T03:04:11 1
`
		if expected != m.String() {
			t.Fatalf("expect %s got %s", expected, m.String())
		}
	}
}

func BenchmarkMeter(b *testing.B) {
	m := NewMeter(600)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Inc(time.Now())
	}
}
