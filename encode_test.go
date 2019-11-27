package bert

import (
	"bytes"
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	// Small Integer
	assertEncode(t, 1, []byte{131, 97, 1})
	assertEncode(t, 42, []byte{131, 97, 42})

	// Integer
	assertEncode(t, 257, []byte{131, 98, 0, 0, 1, 1})
	assertEncode(t, 1025, []byte{131, 98, 0, 0, 4, 1})
	assertEncode(t, -1, []byte{131, 98, 255, 255, 255, 255})
	assertEncode(t, -8, []byte{131, 98, 255, 255, 255, 248})
	assertEncode(t, 5000, []byte{131, 98, 0, 0, 19, 136})
	assertEncode(t, -5000, []byte{131, 98, 255, 255, 236, 120})

	// Float
	assertEncode(t, 0.5, []byte{131, 99, 53, 46, 48, 48, 48, 48, 48, 48,
		48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 101,
		45, 48, 49, 0, 0, 0, 0, 0,
	})
	assertEncode(t, 3.14159, []byte{131, 99, 51, 46, 49, 52, 49, 53, 57,
		48, 49, 49, 56, 52, 48, 56, 50, 48, 51, 49, 50, 53, 48, 48,
		101, 43, 48, 48, 0, 0, 0, 0, 0,
	})
	assertEncode(t, -3.14159, []byte{131, 99, 45, 51, 46, 49, 52, 49, 53,
		57, 48, 49, 49, 56, 52, 48, 56, 50, 48, 51, 49, 50, 53, 48,
		48, 101, 43, 48, 48, 0, 0, 0, 0,
	})

	// Atom
	assertEncode(t, Atom("foo"),
		[]byte{131, 100, 0, 3, 102, 111, 111})

	// Small Tuple
	assertEncode(t, []Term{Atom("foo")},
		[]byte{131, 104, 1, 100, 0, 3, 102, 111, 111})
	assertEncode(t, []Term{Atom("foo"), Atom("bar")},
		[]byte{131, 104, 2,
			100, 0, 3, 102, 111, 111,
			100, 0, 3, 98, 97, 114,
		})
	assertEncode(t, []Term{Atom("coord"), 23, 42},
		[]byte{131, 104, 3,
			100, 0, 5, 99, 111, 111, 114, 100,
			97, 23,
			97, 42,
		})
	tuple := make([][]byte, 2)
	tuple[0] = []byte("0")
	tuple[1] = []byte("1")
	assertEncode(t, tuple, []byte{131, 104, 2, 109, 0, 0, 0, 1, 48, 109, 0, 0, 0, 1, 49})

	// Nil
	assertEncode(t, nil, []byte{131, 106})

	// String
	assertEncode(t, "foo", []byte{131, 107, 0, 3, 102, 111, 111})

	// Binary
	assertEncode(t, []byte{1, 2, 3, 4},
		[]byte{131, 109, 0, 0, 0, 4, 1, 2, 3, 4})

	// Bitstring
	assertEncode(t, Bitstring{[]byte{128}, 1}, []byte{131, 77, 0, 0, 0, 1, 1, 128})

	// List
	assertEncode(t, [1]Term{1},
		[]byte{131, 108, 0, 0, 0, 1, 97, 1, 106})
	assertEncode(t, [3]Term{1, 2, 3},
		[]byte{131, 108, 0, 0, 0, 3,
			97, 1, 97, 2, 97, 3,
			106,
		})
	assertEncode(t, [2]Term{Atom("a"), Atom("b")},
		[]byte{131, 108, 0, 0, 0, 2,
			100, 0, 1, 97, 100, 0, 1, 98,
			106,
		})
	// dynamic list
	list := List{}
	list.Items = append(list.Items, 1)
	list.Items = append(list.Items, 2)
	list.Items = append(list.Items, 3)
	assertEncode(t, list, []byte{131, 108, 0, 0, 0, 3, 97, 1, 97, 2, 97, 3, 106})
	assertEncode(t, [2]Term{uint(1), uint(2)}, []byte{131, 108, 0, 0, 0, 2, 97, 1, 97, 2, 106})
	assertEncode(t, uint(1), []byte{131, 97, 1})

	// larger than 32-bit
	big := 100000000000
	assertEncode(t, uint(big), []byte{131, 110, 5, 0, 0, 232, 118, 72, 23})
	assertEncode(t, big, []byte{131, 110, 5, 0, 0, 232, 118, 72, 23})
	assertEncode(t, -big, []byte{131, 110, 5, 1, 0, 232, 118, 72, 23})
}

func TestMarshal(t *testing.T) {
	var buf bytes.Buffer
	Marshal(&buf, 42)
	assertEqual(t, []byte{131, 97, 42}, buf.Bytes())
}

func TestMarshalResponse(t *testing.T) {
	var buf bytes.Buffer
	MarshalResponse(&buf, []Term{Atom("reply"), 42})
	assertEqual(t, []byte{0, 0, 0, 13,
		131, 104, 2,
		100, 0, 5, 114, 101, 112, 108,
		121, 97, 42,
	},
		buf.Bytes())
}

func assertEncode(t *testing.T, actual interface{}, expected []byte) {
	val, err := Encode(actual)
	if err != nil {
		t.Errorf("Encode(%v) returned error '%v'", actual, err)
	} else if !reflect.DeepEqual(val, expected) {
		t.Errorf("Decode(%v) = %v, expected %v", actual, val, expected)
	}
}

func assertNotEncode(t *testing.T, actual interface{}, errorMessage string) {
	_, err := Encode(actual)
	if err != nil {
		if err.Error() != errorMessage {
			t.Errorf("Encode(%v) should return error %s but not '%v'", actual, errorMessage, err)
		}
	} else {
		t.Errorf("Encode(%v) expected error %s", actual, errorMessage)
	}
}
