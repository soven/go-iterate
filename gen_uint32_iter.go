package iter

// Uint32Iterator is an iterator over items type of uint32.
type Uint32Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() uint32
	// Err contains first met error while Next.
	Err() error
}

type emptyUint32Iterator struct{}

func (emptyUint32Iterator) HasNext() bool       { return false }
func (emptyUint32Iterator) Next() (next uint32) { return 0 }
func (emptyUint32Iterator) Err() error          { return nil }

// EmptyUint32Iterator is a zero value for Uint32Iterator.
// It is not contains any item to iterate over it.
var EmptyUint32Iterator Uint32Iterator = emptyUint32Iterator{}

// Uint32IterMaker is a maker of Uint32Iterator.
type Uint32IterMaker interface {
	// MakeIter should return a new instance of Uint32Iterator to iterate over it.
	MakeIter() Uint32Iterator
}

// MakeUint32Iter is a shortcut implementation
// of Uint32Iterator based on a function.
type MakeUint32Iter func() Uint32Iterator

// MakeIter returns a new instance of Uint32Iterator to iterate over it.
func (m MakeUint32Iter) MakeIter() Uint32Iterator { return m() }

// MakeNoUint32Iter is a zero value for Uint32IterMaker.
// It always returns EmptyUint32Iterator and an empty error.
var MakeNoUint32Iter Uint32IterMaker = MakeUint32Iter(
	func() Uint32Iterator { return EmptyUint32Iterator })

// Uint32Discard just range over all items and do nothing with each of them.
func Uint32Discard(items Uint32Iterator) error {
	if items == nil {
		items = EmptyUint32Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
