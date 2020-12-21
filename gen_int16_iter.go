package iter

// Int16Iterator is an iterator over items type of int16.
type Int16Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() int16
	// Err contains first met error while Next.
	Err() error
}

type emptyInt16Iterator struct{}

func (emptyInt16Iterator) HasNext() bool      { return false }
func (emptyInt16Iterator) Next() (next int16) { return 0 }
func (emptyInt16Iterator) Err() error         { return nil }

// EmptyInt16Iterator is a zero value for Int16Iterator.
// It is not contains any item to iterate over it.
var EmptyInt16Iterator Int16Iterator = emptyInt16Iterator{}

// Int16IterMaker is a maker of Int16Iterator.
type Int16IterMaker interface {
	// MakeIter should return a new instance of Int16Iterator to iterate over it.
	MakeIter() Int16Iterator
}

// MakeInt16Iter is a shortcut implementation
// of Int16Iterator based on a function.
type MakeInt16Iter func() Int16Iterator

// MakeIter returns a new instance of Int16Iterator to iterate over it.
func (m MakeInt16Iter) MakeIter() Int16Iterator { return m() }

// MakeNoInt16Iter is a zero value for Int16IterMaker.
// It always returns EmptyInt16Iterator and an empty error.
var MakeNoInt16Iter Int16IterMaker = MakeInt16Iter(
	func() Int16Iterator { return EmptyInt16Iterator })

// Int16Discard just range over all items and do nothing with each of them.
func Int16Discard(items Int16Iterator) error {
	if items == nil {
		items = EmptyInt16Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
