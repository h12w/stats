package stats

import (
	"reflect"
	"testing"
)

func TestKey(t *testing.T) {
	for _, testcase := range []struct {
		name string
		tags Tags
		key  Key
	}{
		{
			name: "a",
			tags: nil,
			key:  "a",
		},
		{
			name: "a",
			tags: Tags{
				"x": "1",
			},
			key: "a x=1",
		},
	} {
		if key := NewKey(testcase.name, testcase.tags); key != testcase.key {
			t.Fatalf("expect '%s' got '%s'", testcase.key, key)
		}
		name, tags, err := testcase.key.Decode()
		if err != nil {
			t.Fatal(err)
		}
		if name != testcase.name {
			t.Fatalf("expect '%s' got '%s'", testcase.name, name)
		}
		if !reflect.DeepEqual(tags, testcase.tags) {
			t.Fatalf("expect %v got %v", testcase.tags, tags)
		}
	}
}
