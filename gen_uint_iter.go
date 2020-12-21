package iter

// UintIterator is an iterator over items type of uint.
type UintIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() uint
	// Err contains first met error while Next.
	Err() error
}

type emptyUintIterator struct{}

func (emptyUintIterator) HasNext() bool     { return false }
func (emptyUintIterator) Next() (next uint) { return 0 }
func (emptyUintIterator) Err() error        { return nil }

// EmptyUintIterator is a zero value for UintIterator.
// It is not contains any item to iterate over it.
var EmptyUintIterator UintIterator = emptyUintIterator{}

// UintIterMaker is a maker of UintIterator.
type UintIterMaker interface {
	// MakeIter should return a new instance of UintIterator to iterate over it.
	MakeIter() UintIterator
}

// MakeUintIter is a shortcut implementation
// of UintIterator based on a function.
type MakeUintIter func() UintIterator

// MakeIter returns a new instance of UintIterator to iterate over it.
func (m MakeUintIter) MakeIter() UintIterator { return m() }

// MakeNoUintIter is a zero value for UintIterMaker.
// It always returns EmptyUintIterator and an empty error.
var MakeNoUintIter UintIterMaker = MakeUintIter(
	func() UintIterator { return EmptyUintIterator })

// UintDiscard just range over all items and do nothing with each of them.
func UintDiscard(items UintIterator) error {
	if items == nil {
		items = EmptyUintIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
