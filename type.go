package bert

const (
	VersionTag     = 131
	SmallIntTag    = 97
	IntTag         = 98
	SmallBignumTag = 110
	LargeBignumTag = 111
	FloatTag       = 99
	AtomTag        = 100
	SmallTupleTag  = 104
	LargeTupleTag  = 105
	NilTag         = 106
	StringTag      = 107
	ListTag        = 108
	BinTag         = 109
	BitTag         = 77
)

type Atom string
type Bitstring struct {
	Bytes []byte
	Bits  uint8
}
type List struct {
	Items []Term
}

const (
	BertAtom  = Atom("bert")
	NilAtom   = Atom("nil")
	TrueAtom  = Atom("true")
	FalseAtom = Atom("false")
)

type Term interface{}

type Request struct {
	Kind      Atom
	Module    Atom
	Function  Atom
	Arguments []Term
}
