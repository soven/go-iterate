package iter

// Int32Iterator is an iterator over items type of int32.
type Int32Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() int32
	// Err contains first met error while Next.
	Err() error
}

type emptyInt32Iterator struct{}

func (emptyInt32Iterator) HasNext() bool      { return false }
func (emptyInt32Iterator) Next() (next int32) { return 0 }
func (emptyInt32Iterator) Err() error         { return nil }

// EmptyInt32Iterator is a zero value for Int32Iterator.
// It is not contains any item to iterate over it.
var EmptyInt32Iterator Int32Iterator = emptyInt32Iterator{}

// Int32IterMaker is a maker of Int32Iterator.
type Int32IterMaker interface {
	// MakeIter should return a new instance of Int32Iterator to iterate over it.
	MakeIter() Int32Iterator
}

// MakeInt32Iter is a shortcut implementation
// of Int32Iterator based on a function.
type MakeInt32Iter func() Int32Iterator

// MakeIter returns a new instance of Int32Iterator to iterate over it.
func (m MakeInt32Iter) MakeIter() Int32Iterator { return m() }

// MakeNoInt32Iter is a zero value for Int32IterMaker.
// It always returns EmptyInt32Iterator and an empty error.
var MakeNoInt32Iter Int32IterMaker = MakeInt32Iter(
	func() Int32Iterator { return EmptyInt32Iterator })

// Int32Discard just range over all items and do nothing with each of them.
func Int32Discard(items Int32Iterator) error {
	if items == nil {
		items = EmptyInt32Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
