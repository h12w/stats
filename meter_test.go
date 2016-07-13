package stats

import (
	"strconv"
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

	start := testTime
	{
		m.Inc(now)
		m.Inc(now.Add(-time.Second))

		m.Inc(now.Add(time.Second))
		m.Inc(now.Add(time.Second))

		m.Inc(now.Add(2 * time.Second))
		m.Inc(now.Add(2 * time.Second))
		m.Inc(now.Add(2 * time.Second))
		expected := "[" + strconv.Itoa(int(start.Unix())) + ",1,2,3]"
		if expected != m.String() {
			t.Fatalf("expect %s got %s", expected, m.String())
		}
	}

	now = now.Add(3 * time.Second)
	{
		m.Inc(now)
		m.Inc(now)
		start = start.Add(time.Second)

		expected := "[" + strconv.Itoa(int(start.Unix())) + ",2,3,2]"
		if expected != m.String() {
			t.Fatalf("expect %s got %s", expected, m.String())
		}
	}

	now = now.Add(3 * time.Second)
	{
		m.Inc(now)
		start = start.Add(3 * time.Second)

		expected := "[" + strconv.Itoa(int(start.Unix())) + ",0,0,1]"
		if expected != m.String() {
			t.Fatalf("expect %s got %s", expected, m.String())
		}
	}
}

func TestMeterJSON(t *testing.T) {
	m := NewMeter(3)
	now := testTime
	m.Add(now, 1)
	m.Add(now.Add(time.Second), 2)
	m.Add(now.Add(2*time.Second), 3)
	m.Add(now.Add(3*time.Second), 4)

	jsonBuf, err := m.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	actual := string(jsonBuf)
	expected := "[" + strconv.Itoa(int(testTime.Add(time.Second).Unix())) + `,2,3,4]`
	if actual != expected {
		t.Fatalf("expect \n%s\n but got\n%s", expected, actual)
	}
}

func BenchmarkMeter(b *testing.B) {
	m := NewMeter(600)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Inc(time.Now())
	}
}
