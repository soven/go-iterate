package iter

// Uint64Iterator is an iterator over items type of uint64.
type Uint64Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() uint64
	// Err contains first met error while Next.
	Err() error
}

type emptyUint64Iterator struct{}

func (emptyUint64Iterator) HasNext() bool       { return false }
func (emptyUint64Iterator) Next() (next uint64) { return 0 }
func (emptyUint64Iterator) Err() error          { return nil }

// EmptyUint64Iterator is a zero value for Uint64Iterator.
// It is not contains any item to iterate over it.
var EmptyUint64Iterator Uint64Iterator = emptyUint64Iterator{}

// Uint64IterMaker is a maker of Uint64Iterator.
type Uint64IterMaker interface {
	// MakeIter should return a new instance of Uint64Iterator to iterate over it.
	MakeIter() Uint64Iterator
}

// MakeUint64Iter is a shortcut implementation
// of Uint64Iterator based on a function.
type MakeUint64Iter func() Uint64Iterator

// MakeIter returns a new instance of Uint64Iterator to iterate over it.
func (m MakeUint64Iter) MakeIter() Uint64Iterator { return m() }

// MakeNoUint64Iter is a zero value for Uint64IterMaker.
// It always returns EmptyUint64Iterator and an empty error.
var MakeNoUint64Iter Uint64IterMaker = MakeUint64Iter(
	func() Uint64Iterator { return EmptyUint64Iterator })

// Uint64Discard just range over all items and do nothing with each of them.
func Uint64Discard(items Uint64Iterator) error {
	if items == nil {
		items = EmptyUint64Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
