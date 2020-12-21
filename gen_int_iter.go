package iter

// IntIterator is an iterator over items type of int.
type IntIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() int
	// Err contains first met error while Next.
	Err() error
}

type emptyIntIterator struct{}

func (emptyIntIterator) HasNext() bool    { return false }
func (emptyIntIterator) Next() (next int) { return 0 }
func (emptyIntIterator) Err() error       { return nil }

// EmptyIntIterator is a zero value for IntIterator.
// It is not contains any item to iterate over it.
var EmptyIntIterator IntIterator = emptyIntIterator{}

// IntIterMaker is a maker of IntIterator.
type IntIterMaker interface {
	// MakeIter should return a new instance of IntIterator to iterate over it.
	MakeIter() IntIterator
}

// MakeIntIter is a shortcut implementation
// of IntIterator based on a function.
type MakeIntIter func() IntIterator

// MakeIter returns a new instance of IntIterator to iterate over it.
func (m MakeIntIter) MakeIter() IntIterator { return m() }

// MakeNoIntIter is a zero value for IntIterMaker.
// It always returns EmptyIntIterator and an empty error.
var MakeNoIntIter IntIterMaker = MakeIntIter(
	func() IntIterator { return EmptyIntIterator })

// IntDiscard just range over all items and do nothing with each of them.
func IntDiscard(items IntIterator) error {
	if items == nil {
		items = EmptyIntIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
