package stats

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"
)

func TestStatsMarshal(t *testing.T) {
	s := New().SetBufSize(2)
	s.Meter("test").Add(testTime, 1)
	s.Meter("test").Add(testTime.Add(time.Second), 2)
	jsonBuf, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(jsonBuf)
	expected := `{"meters":{"test":[` + strconv.Itoa(int(testTime.Unix())) + `,1,2]}}`
	if actual != expected {
		t.Fatalf("expect %s got %s", expected, actual)
	}
}

func TestStatsMerge(t *testing.T) {
	s1 := New().SetBufSize(2)
	s1.Meter("test").Add(testTime, 1)
	s1.Meter("test").Add(testTime.Add(time.Second), 2)

	s2 := New().SetBufSize(2)
	s2.Meter("test").Add(testTime, 3)
	s2.Meter("test").Add(testTime.Add(time.Second), 4)

	s2.Merge(s1)

	jsonBuf, err := json.Marshal(s2)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(jsonBuf)
	expected := `{"meters":{"test":[` + strconv.Itoa(int(testTime.Unix())) + `,4,6]}}`
	if actual != expected {
		t.Fatalf("expect %s got %s", expected, actual)
	}
}
