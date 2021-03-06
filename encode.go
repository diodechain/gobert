package bert

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
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

func writeNumber(w io.Writer, n big.Int) {
	if n.IsInt64() {
		x := n.Int64()
		if x >= 0 && x < 256 {
			writeSmallInt(w, uint8(x))
			return
		}
		if x >= -2147483648 && x <= 2147483647 {
			writeInt(w, uint32(x))
			return
		}
	}

	write1(w, SmallBignumTag)
	bytes := n.Bytes()
	// converting big endian to small endian
	// http://erlang.org/doc/apps/erts/erl_ext_dist.html#small_big_ext
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	write1(w, uint8(len(bytes)))
	if n.Sign() > -1 {
		write1(w, 0)
	} else {
		write1(w, 1)
	}
	w.Write(bytes)
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
	size := (int(bits) + 7) / 8
	write4(w, uint32(size))
	write1(w, bits%8)

	for len(a) < size {
		a = append([]byte{0}, a...)
	}
	w.Write(a[:size])
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
	val = reflect.Indirect(val)
	switch v := val; v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := v.Int()
		writeNumber(w, *big.NewInt(n))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := v.Uint()
		var bn big.Int
		bn.SetUint64(n)
		writeNumber(w, bn)
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
			if b.Bits%8 != 0 {
				writeBitstring(w, b.Bytes, b.Bits)
			} else {
				writeBinary(w, b.Bytes[0:b.Bits/8])
			}
		} else if l, ok := v.Interface().(List); ok {
			err = writeList(w, reflect.ValueOf(l.Items))
		} else if bn, ok := v.Interface().(big.Int); ok {
			writeNumber(w, bn)
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
