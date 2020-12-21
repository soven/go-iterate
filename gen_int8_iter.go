package iter

// Int8Iterator is an iterator over items type of int8.
type Int8Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() int8
	// Err contains first met error while Next.
	Err() error
}

type emptyInt8Iterator struct{}

func (emptyInt8Iterator) HasNext() bool     { return false }
func (emptyInt8Iterator) Next() (next int8) { return 0 }
func (emptyInt8Iterator) Err() error        { return nil }

// EmptyInt8Iterator is a zero value for Int8Iterator.
// It is not contains any item to iterate over it.
var EmptyInt8Iterator Int8Iterator = emptyInt8Iterator{}

// Int8IterMaker is a maker of Int8Iterator.
type Int8IterMaker interface {
	// MakeIter should return a new instance of Int8Iterator to iterate over it.
	MakeIter() Int8Iterator
}

// MakeInt8Iter is a shortcut implementation
// of Int8Iterator based on a function.
type MakeInt8Iter func() Int8Iterator

// MakeIter returns a new instance of Int8Iterator to iterate over it.
func (m MakeInt8Iter) MakeIter() Int8Iterator { return m() }

// MakeNoInt8Iter is a zero value for Int8IterMaker.
// It always returns EmptyInt8Iterator and an empty error.
var MakeNoInt8Iter Int8IterMaker = MakeInt8Iter(
	func() Int8Iterator { return EmptyInt8Iterator })

// Int8Discard just range over all items and do nothing with each of them.
func Int8Discard(items Int8Iterator) error {
	if items == nil {
		items = EmptyInt8Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
