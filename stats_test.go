package stats

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestStatsMarshal(t *testing.T) {
	s := New().SetBufSize(2)
	s.Meter("test").Add(testTime, 1)
	s.Meter("test").Add(testTime.Add(time.Second), 2)
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
	s1 := New().SetBufSize(2)
	s1.Meter("test").Add(testTime, 1)
	s1.Meter("test").Add(testTime.Add(time.Second), 2)

	s2 := New().SetBufSize(2)
	s2.Meter("test").Add(testTime, 3)
	s2.Meter("test").Add(testTime.Add(time.Second), 4)

	s2.Merge(s1)

	actual := string(s2.String())
	expected := `{"meters":{"test":[` + strconv.Itoa(int(testTime.Unix())) + `,4,6]}}`
	if actual != expected {
		t.Fatalf("expect %s got %s", expected, actual)
	}
}
