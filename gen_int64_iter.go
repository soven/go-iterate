package iter

// Int64Iterator is an iterator over items type of int64.
type Int64Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() int64
	// Err contains first met error while Next.
	Err() error
}

type emptyInt64Iterator struct{}

func (emptyInt64Iterator) HasNext() bool      { return false }
func (emptyInt64Iterator) Next() (next int64) { return 0 }
func (emptyInt64Iterator) Err() error         { return nil }

// EmptyInt64Iterator is a zero value for Int64Iterator.
// It is not contains any item to iterate over it.
var EmptyInt64Iterator Int64Iterator = emptyInt64Iterator{}

// Int64IterMaker is a maker of Int64Iterator.
type Int64IterMaker interface {
	// MakeIter should return a new instance of Int64Iterator to iterate over it.
	MakeIter() Int64Iterator
}

// MakeInt64Iter is a shortcut implementation
// of Int64Iterator based on a function.
type MakeInt64Iter func() Int64Iterator

// MakeIter returns a new instance of Int64Iterator to iterate over it.
func (m MakeInt64Iter) MakeIter() Int64Iterator { return m() }

// MakeNoInt64Iter is a zero value for Int64IterMaker.
// It always returns EmptyInt64Iterator and an empty error.
var MakeNoInt64Iter Int64IterMaker = MakeInt64Iter(
	func() Int64Iterator { return EmptyInt64Iterator })

// Int64Discard just range over all items and do nothing with each of them.
func Int64Discard(items Int64Iterator) error {
	if items == nil {
		items = EmptyInt64Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
