package stats

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestStatsMarshal(t *testing.T) {
	testTime := time.Now()
	s := New().SetBufSize(2)
	s.Meter("test", nil).Inc(testTime, 1)
	s.Meter("test", nil).Inc(testTime.Add(time.Second), 2)
	jsonText := `{"meters":{"test":[` + strconv.Itoa(int(testTime.Unix())) + `,1,2]}}`

	{
		expected := jsonText
		jsonBuf, err := json.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(jsonBuf)
		if actual != expected {
			t.Fatalf("expect %s got %s", expected, actual)
		}
	}
	{
		expected := s
		actual := New().SetBufSize(2)
		if err := json.Unmarshal([]byte(jsonText), actual); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("expect\n%+v\ngot\n%+v", expected, actual)
		}
	}
}

func TestStatsMerge(t *testing.T) {
	testTime := time.Now()
	s1 := New().SetBufSize(2)
	s1.Meter("test", nil).Inc(testTime, 1)
	s1.Meter("test", nil).Inc(testTime.Add(time.Second), 2)

	s2 := New().SetBufSize(2)
	s2.Meter("test", nil).Inc(testTime, 3)
	s2.Meter("test", nil).Inc(testTime.Add(time.Second), 4)

	s2.Merge(s1, testTime)

	actual := string(s2.String())
	expected := `{"meters":{"test":[` + strconv.Itoa(int(testTime.Unix())) + `,4,6]}}`
	if actual != expected {
		t.Fatalf("expect %s got %s", expected, actual)
	}
}

func TestStatsMergeWithTags(t *testing.T) {
	testTime := time.Now()
	s1 := New().SetBufSize(2)
	s1.Meter("test", nil).Inc(testTime, 1)
	s1.Meter("test", nil).Inc(testTime.Add(time.Second), 2)

	s2 := New().SetBufSize(2)

	s2.MergeWithTags(s1, testTime, Tags{"host": "a"})

	actual := string(s2.String())
	expected := `{"meters":{"test host=a":[` + strconv.Itoa(int(testTime.Unix())) + `,1,2]}}`
	if actual != expected {
		t.Fatalf("expect %s got %s", expected, actual)
	}
}
