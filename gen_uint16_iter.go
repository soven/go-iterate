package iter

// Uint16Iterator is an iterator over items type of uint16.
type Uint16Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() uint16
	// Err contains first met error while Next.
	Err() error
}

type emptyUint16Iterator struct{}

func (emptyUint16Iterator) HasNext() bool       { return false }
func (emptyUint16Iterator) Next() (next uint16) { return 0 }
func (emptyUint16Iterator) Err() error          { return nil }

// EmptyUint16Iterator is a zero value for Uint16Iterator.
// It is not contains any item to iterate over it.
var EmptyUint16Iterator Uint16Iterator = emptyUint16Iterator{}

// Uint16IterMaker is a maker of Uint16Iterator.
type Uint16IterMaker interface {
	// MakeIter should return a new instance of Uint16Iterator to iterate over it.
	MakeIter() Uint16Iterator
}

// MakeUint16Iter is a shortcut implementation
// of Uint16Iterator based on a function.
type MakeUint16Iter func() Uint16Iterator

// MakeIter returns a new instance of Uint16Iterator to iterate over it.
func (m MakeUint16Iter) MakeIter() Uint16Iterator { return m() }

// MakeNoUint16Iter is a zero value for Uint16IterMaker.
// It always returns EmptyUint16Iterator and an empty error.
var MakeNoUint16Iter Uint16IterMaker = MakeUint16Iter(
	func() Uint16Iterator { return EmptyUint16Iterator })

// Uint16Discard just range over all items and do nothing with each of them.
func Uint16Discard(items Uint16Iterator) error {
	if items == nil {
		items = EmptyUint16Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
