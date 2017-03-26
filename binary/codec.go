package binary

import (
	"fmt"
	"io"
	"math"
)

func WriteUint16SliceSparse(w io.Writer, s []uint16) (int, error) {
	var err error
	var nn int
	n := 0
	nn, err = WriteInt64(w, int64(len(s))) // start sparse array
	if err != nil {
		return n, err
	}
	n += nn

	i := 0
	for i < len(s) {
		for ; ; i++ {
			if i >= len(s) {
				goto Return
			}
			if s[i] != 0 {
				break
			}
		}

		start := i + 1
		nn, err := WriteInt64(w, int64(start)) // start nonzero sequence
		if err != nil {
			return n, err
		}
		n += nn

		for ; ; i++ {
			if i >= len(s) {
				goto Return
			}
			if s[i] == 0 {
				break
			}
			nn, err := WriteUint16(w, s[i])
			if err != nil {
				return n, err
			}
			n += nn
		}

		nn, err = WriteUint16(w, 0) // end nonzero sequence
		if err != nil {
			return n, err
		}
		n += nn
	}

Return:
	nn, err = WriteInt64(w, 0) // end sparse array
	if err != nil {
		return n, err
	}
	n += nn
	return n, nil
}

func ReadUint16SliceSparse(r io.Reader, s *[]uint16) (int, error) {
	var err error
	var nn int
	n := 0
	var size int64
	nn, err = ReadInt64(r, &size)
	n += nn
	if err != nil {
		return n, err
	}
	*s = make([]uint16, int(size))
	for {
		var start int64
		nn, err := ReadInt64(r, &start)
		n += nn
		if err == io.EOF {
			goto Return
		} else if err != nil {
			return n, err
		}
		if start == 0 {
			break
		}
		start--
		for i := start; ; i++ {
			var v uint16
			nn, err := ReadUint16(r, &v)
			n += nn
			if err == io.EOF {
				goto Return
			} else if err != nil {
				return n, err
			}
			if v == 0 {
				break
			}
			if int(i) >= len(*s) {
				return n, fmt.Errorf("out of range %d in %d", i, len(*s))
			}
			(*s)[i] = v
		}
	}
Return:
	return n, nil
}

func WriteString(w io.Writer, s string) (int, error) {
	var err error
	var nn int
	n := 0
	nn, err = WriteInt64(w, int64(len(s)))
	n += nn
	if err != nil {
		return n, err
	}
	nn, err = w.Write([]byte(s))
	n += nn
	if err != nil {
		return n, err
	}
	return n, nil
}

func ReadString(r io.Reader, s *string) (int, error) {
	var err error
	var nn int
	n := 0
	var size int64
	nn, err = ReadInt64(r, &size)
	n += nn
	if err != nil {
		return n, err
	}
	buf := make([]byte, int(size))
	nn, err = io.ReadFull(r, buf)
	n += nn
	if err != nil {
		return n, err
	}
	*s = string(buf)
	return n, nil
}

func WriteUint16Slice(w io.Writer, s []uint16) (int, error) {
	var err error
	var nn int
	n := 0
	nn, err = WriteInt64(w, int64(len(s)))
	if err != nil {
		return n, err
	}
	n += nn
	for _, v := range s {
		nn, err = WriteUint16(w, v)
		if err != nil {
			return n, err
		}
		n += nn
	}
	return n, nil
}

func ReadUint16Slice(r io.Reader, s *[]uint16) (int, error) {
	var err error
	var nn int
	n := 0
	var size int64
	nn, err = ReadInt64(r, &size)
	if err != nil {
		return n, err
	}
	n += nn
	*s = make([]uint16, int(size))
	for i := range *s {
		nn, err = ReadUint16(r, &(*s)[i])
		if err != nil {
			return n, err
		}
		n += nn
	}
	return n, nil
}

func WriteUint64(w io.Writer, i uint64) (int, error) {
	return w.Write([]byte{byte(i >> 56), byte(i >> 48), byte(i >> 40), byte(i >> 32),
		byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
}

func ReadUint64(r io.Reader, i *uint64) (int, error) {
	var b [8]byte
	n, err := io.ReadFull(r, b[:])
	if err != nil {
		return n, err
	}
	*i = uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
	return n, nil
}

func WriteInt64(w io.Writer, i int64) (int, error) {
	return w.Write([]byte{byte(i >> 56), byte(i >> 48), byte(i >> 40), byte(i >> 32),
		byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
}

func ReadInt64(r io.Reader, i *int64) (int, error) {
	var b [8]byte
	n, err := io.ReadFull(r, b[:])
	if err != nil {
		return n, err
	}
	*i = int64(b[0])<<56 | int64(b[1])<<48 | int64(b[2])<<40 | int64(b[3])<<32 |
		int64(b[4])<<24 | int64(b[5])<<16 | int64(b[6])<<8 | int64(b[7])
	return n, nil
}

func WriteInt32(w io.Writer, i int32) (int, error) {
	return w.Write([]byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
}

func ReadInt32(r io.Reader, i *int32) (int, error) {
	var b [4]byte
	n, err := io.ReadFull(r, b[:])
	if err != nil {
		return n, err
	}
	*i = int32(b[0])<<24 | int32(b[1])<<16 | int32(b[2])<<8 | int32(b[3])
	return n, nil
}

func WriteFloat64(w io.Writer, f float64) (int, error) {
	return WriteUint64(w, math.Float64bits(f))
}

func ReadFloat64(r io.Reader, f *float64) (int, error) {
	var i uint64
	n, err := ReadUint64(r, &i)
	if err != nil {
		return n, err
	}
	*f = math.Float64frombits(i)
	return n, nil
}

func WriteUint16(w io.Writer, i uint16) (int, error) {
	return w.Write([]byte{byte(i >> 8), byte(i)})
}

func ReadUint16(r io.Reader, i *uint16) (int, error) {
	var b [2]byte
	n, err := io.ReadFull(r, b[:])
	if err != nil {
		return n, err
	}
	*i = uint16(b[0])<<8 | uint16(b[1])
	return n, nil
}

func WriteUint8(w io.Writer, i uint8) (int, error) {
	return w.Write([]uint8{i})
}

func ReadUint8(r io.Reader, i *uint8) (int, error) {
	var b [1]byte
	n, err := io.ReadFull(r, b[:])
	if err != nil {
		return n, err
	}
	*i = b[0]
	return n, nil
}
