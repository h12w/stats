package binary

import (
	"bytes"
	"reflect"
	"testing"
)

func TestSparse(t *testing.T) {
	for _, testcase := range [][]uint16{
		[]uint16{},
		[]uint16{0},
		[]uint16{0xABCD},
		[]uint16{0, 1},
		[]uint16{1, 0},
		[]uint16{1, 1},
		[]uint16{1, 0, 1},
		[]uint16{0, 1, 0},
		[]uint16{1, 0, 1, 0},
		[]uint16{0, 1, 0, 1},
		[]uint16{0, 1, 1, 0},
		[]uint16{1, 1, 0, 1, 1},
	} {
		w := new(bytes.Buffer)
		n, err := WriteUint16SliceSparse(w, testcase)
		if err != nil {
			t.Fatal(err)
		}
		bs := w.Bytes()
		if n != len(bs) {
			t.Fatal("size mismatch", n, len(bs))
		}
		var s []uint16
		n, err = ReadUint16SliceSparse(bytes.NewReader(bs), &s)
		if n != len(bs) {
			t.Fatal("size mismatch", n, len(bs))
		}
		if !reflect.DeepEqual(testcase, s) {
			t.Fatal("slice mismatch", testcase)
		}
	}
}
