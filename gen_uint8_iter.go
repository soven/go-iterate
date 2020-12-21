package iter

// Uint8Iterator is an iterator over items type of uint8.
type Uint8Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() uint8
	// Err contains first met error while Next.
	Err() error
}

type emptyUint8Iterator struct{}

func (emptyUint8Iterator) HasNext() bool      { return false }
func (emptyUint8Iterator) Next() (next uint8) { return 0 }
func (emptyUint8Iterator) Err() error         { return nil }

// EmptyUint8Iterator is a zero value for Uint8Iterator.
// It is not contains any item to iterate over it.
var EmptyUint8Iterator Uint8Iterator = emptyUint8Iterator{}

// Uint8IterMaker is a maker of Uint8Iterator.
type Uint8IterMaker interface {
	// MakeIter should return a new instance of Uint8Iterator to iterate over it.
	MakeIter() Uint8Iterator
}

// MakeUint8Iter is a shortcut implementation
// of Uint8Iterator based on a function.
type MakeUint8Iter func() Uint8Iterator

// MakeIter returns a new instance of Uint8Iterator to iterate over it.
func (m MakeUint8Iter) MakeIter() Uint8Iterator { return m() }

// MakeNoUint8Iter is a zero value for Uint8IterMaker.
// It always returns EmptyUint8Iterator and an empty error.
var MakeNoUint8Iter Uint8IterMaker = MakeUint8Iter(
	func() Uint8Iterator { return EmptyUint8Iterator })

// Uint8Discard just range over all items and do nothing with each of them.
func Uint8Discard(items Uint8Iterator) error {
	if items == nil {
		items = EmptyUint8Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
