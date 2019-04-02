package bert

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

func write1(w io.Writer, ui8 uint8) { w.Write([]byte{ui8}) }

func write2(w io.Writer, ui16 uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, ui16)
	w.Write(b)
}

func write4(w io.Writer, ui32 uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, ui32)
	w.Write(b)
}

func writeSmallInt(w io.Writer, n uint8) {
	write1(w, SmallIntTag)
	write1(w, n)
}

func writeInt(w io.Writer, n uint32) {
	write1(w, IntTag)
	write4(w, n)
}

func writeFloat(w io.Writer, f float32) {
	write1(w, FloatTag)

	s := fmt.Sprintf("%.20e", float32(f))
	w.Write([]byte(s))

	pad := make([]byte, 31-len(s))
	w.Write(pad)
}

func writeAtom(w io.Writer, a string) {
	write1(w, AtomTag)
	write2(w, uint16(len(a)))
	w.Write([]byte(a))
}

func writeSmallTuple(w io.Writer, t reflect.Value) (err error) {
	write1(w, SmallTupleTag)
	size := t.Len()
	write1(w, uint8(size))

	for i := 0; i < size; i++ {
		err = writeTag(w, t.Index(i))
		if err != nil {
			break
		}
	}
	return
}

func writeBinary(w io.Writer, a []byte) {
	write1(w, BinTag)
	size := len(a)
	write4(w, uint32(size))
	w.Write(a)
}

func writeBitstring(w io.Writer, a []byte, bits uint8) {
	write1(w, BitTag)
	size := len(a)
	write4(w, uint32(size))
	write1(w, bits)
	w.Write(a)
}

func writeNil(w io.Writer) { write1(w, NilTag) }

func writeString(w io.Writer, s string) {
	write1(w, StringTag)
	write2(w, uint16(len(s)))
	w.Write([]byte(s))
}

func writeList(w io.Writer, l reflect.Value) (err error) {
	write1(w, ListTag)
	size := l.Len()
	write4(w, uint32(size))

	for i := 0; i < size; i++ {
		err = writeTag(w, l.Index(i))
		if err != nil {
			break
		}
	}

	writeNil(w)
	return
}

func writeTag(w io.Writer, val reflect.Value) (err error) {
	switch v := val; v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := v.Int()
		if n >= 0 && n < 256 {
			writeSmallInt(w, uint8(n))
		} else {
			writeInt(w, uint32(n))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = ErrUintType
	case reflect.Float32, reflect.Float64:
		writeFloat(w, float32(v.Float()))
	case reflect.String:
		if v.Type().Name() == "Atom" {
			writeAtom(w, v.String())
		} else {
			writeString(w, v.String())
		}
	case reflect.Slice:
		if b, ok := v.Interface().([]byte); ok {
			writeBinary(w, b)
		} else {
			err = writeSmallTuple(w, v)
		}

	case reflect.Array:
		err = writeList(w, v)
	case reflect.Interface:
		err = writeTag(w, v.Elem())
	case reflect.Struct:
		if b, ok := v.Interface().(Bitstring); ok {
			writeBitstring(w, b.Bytes, b.Bits)
		} else if l, ok := v.Interface().(List); ok {
			err = writeList(w, reflect.ValueOf(l.Items))
		} else {
			err = ErrUnknownType
		}
	default:
		if !reflect.Indirect(val).IsValid() {
			writeNil(w)
		} else {
			err = ErrUnknownType
		}
	}

	return
}

// EncodeTo encodes val and writes it to w, returning any error.
func EncodeTo(w io.Writer, val interface{}) (err error) {
	write1(w, VersionTag)
	err = writeTag(w, reflect.ValueOf(val))
	return
}

// Encode encodes val and returns it or an error.
func Encode(val interface{}) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := EncodeTo(buf, val)
	return buf.Bytes(), err
}

// Marshal is an alias for EncodeTo.
func Marshal(w io.Writer, val interface{}) error {
	return EncodeTo(w, val)
}

// MarshalResponse encodes val into a BURP Response struct and writes it to w,
// returning any error.
func MarshalResponse(w io.Writer, val interface{}) (err error) {
	resp, err := Encode(val)

	write4(w, uint32(len(resp)))
	w.Write(resp)

	return
}
