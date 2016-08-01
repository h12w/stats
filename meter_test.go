package stats

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"
)

var (
//testTimeStr = "2006-01-02T03:04:05Z"
//testTime, _ = time.Parse(testTimeStr, testTimeStr)
)

func TestMeterInc(t *testing.T) {
	now := time.Now()
	m := NewMeter(now, 3)

	start := now
	{
		m.Inc(now, 1)
		m.Inc(now.Add(-time.Second), 1)
		m.Inc(now.Add(time.Second), 2)
		m.Inc(now.Add(2*time.Second), 3)
		actual := getMeterValues(m, start, 3*time.Second)
		expected := []int{1, 2, 3}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("expect %v got %v", actual, expected)
		}
	}

	now = now.Add(3 * time.Second)
	{
		m.Inc(now, 2)
		start = start.Add(time.Second)
		actual := getMeterValues(m, start, 3*time.Second)
		expected := []int{2, 3, 2}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("expect %v got %v", actual, expected)
		}
	}

	now = now.Add(10 * time.Second)
	{
		m.Inc(now, 1)
		start = now.Add(-2 * time.Second)
		actual := getMeterValues(m, start, 3*time.Second)
		expected := []int{0, 0, 1}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("expect %v got %v", expected, actual)
		}
	}
}
func getMeterValues(m *Meter, start time.Time, du time.Duration) (values []int) {
	for i := int(start.Unix()); i < int(start.Add(du).Unix()); i++ {
		values = append(values, m.Get(i))
	}
	return values
}

func TestMeterJSON(t *testing.T) {
	now := time.Now()
	m := NewMeter(now, 3)
	m.Inc(now, 1)
	m.Inc(now.Add(time.Second), 2)
	m.Inc(now.Add(2*time.Second), 3)
	m.Inc(now.Add(3*time.Second), 4)

	jsonBuf, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	{
		actual := string(jsonBuf)
		expected := "[" + strconv.Itoa(int(now.Add(time.Second).Unix())) + `,2,3,4]`
		if actual != expected {
			t.Fatalf("expect \n%s\n but got\n%s", expected, actual)
		}
	}
	{
		m := NewMeter(now, 3)
		if err := json.Unmarshal(jsonBuf, &m); err != nil {
			t.Fatal(err)
		}
		jsonBuf2, err := json.Marshal(&m)
		if err != nil {
			t.Fatal(err)
		}
		if string(jsonBuf2) != string(jsonBuf) {
			t.Fatalf("expect %s but got %s", string(jsonBuf), string(jsonBuf2))
		}
	}
}

func TestMeterMerge(t *testing.T) {
	now := time.Now()

	m1 := NewMeter(now, 2)
	m1.Inc(now, 1)
	m1.Inc(now.Add(time.Second), 2)

	{
		m2 := NewMeter(now, 2)
		m2.Inc(now.Add(time.Second), 3)
		m2.Inc(now.Add(2*time.Second), 4)

		m2.Merge(m1)
		expectedM2 := `[` + unixStr(now.Add(time.Second)) + `,5,4]`
		if expectedM2 != m2.String() {
			t.Fatalf("expect %s but got %s", expectedM2, m2.String())
		}
	}
	{
		m2 := NewMeter(now, 2)
		m2.Inc(now.Add(-time.Second), 3)
		m2.Inc(now, 4)

		m2.Merge(m1)
		expectedM2 := `[` + unixStr(now) + `,5,2]`
		if expectedM2 != m2.String() {
			t.Fatalf("expect %s but got %s", expectedM2, m2.String())
		}
	}
}

func BenchmarkMeter(b *testing.B) {
	m := NewMeter(time.Now(), 600)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Inc(time.Now(), 1)
	}
}

func unixStr(t time.Time) string {
	return strconv.Itoa(int(t.Unix()))
}
