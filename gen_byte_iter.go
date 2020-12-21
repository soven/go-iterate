package iter

// ByteIterator is an iterator over items type of byte.
type ByteIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() byte
	// Err contains first met error while Next.
	Err() error
}

type emptyByteIterator struct{}

func (emptyByteIterator) HasNext() bool     { return false }
func (emptyByteIterator) Next() (next byte) { return 0 }
func (emptyByteIterator) Err() error        { return nil }

// EmptyByteIterator is a zero value for ByteIterator.
// It is not contains any item to iterate over it.
var EmptyByteIterator ByteIterator = emptyByteIterator{}

// ByteIterMaker is a maker of ByteIterator.
type ByteIterMaker interface {
	// MakeIter should return a new instance of ByteIterator to iterate over it.
	MakeIter() ByteIterator
}

// MakeByteIter is a shortcut implementation
// of ByteIterator based on a function.
type MakeByteIter func() ByteIterator

// MakeIter returns a new instance of ByteIterator to iterate over it.
func (m MakeByteIter) MakeIter() ByteIterator { return m() }

// MakeNoByteIter is a zero value for ByteIterMaker.
// It always returns EmptyByteIterator and an empty error.
var MakeNoByteIter ByteIterMaker = MakeByteIter(
	func() ByteIterator { return EmptyByteIterator })

// ByteDiscard just range over all items and do nothing with each of them.
func ByteDiscard(items ByteIterator) error {
	if items == nil {
		items = EmptyByteIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
