package stats

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"
)

func TestSMarshal(t *testing.T) {
	s := New(2)
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
