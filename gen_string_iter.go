package iter

// StringIterator is an iterator over items type of string.
type StringIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() string
	// Err contains first met error while Next.
	Err() error
}

type emptyStringIterator struct{}

func (emptyStringIterator) HasNext() bool       { return false }
func (emptyStringIterator) Next() (next string) { return "" }
func (emptyStringIterator) Err() error          { return nil }

// EmptyStringIterator is a zero value for StringIterator.
// It is not contains any item to iterate over it.
var EmptyStringIterator StringIterator = emptyStringIterator{}

// StringIterMaker is a maker of StringIterator.
type StringIterMaker interface {
	// MakeIter should return a new instance of StringIterator to iterate over it.
	MakeIter() StringIterator
}

// MakeStringIter is a shortcut implementation
// of StringIterator based on a function.
type MakeStringIter func() StringIterator

// MakeIter returns a new instance of StringIterator to iterate over it.
func (m MakeStringIter) MakeIter() StringIterator { return m() }

// MakeNoStringIter is a zero value for StringIterMaker.
// It always returns EmptyStringIterator and an empty error.
var MakeNoStringIter StringIterMaker = MakeStringIter(
	func() StringIterator { return EmptyStringIterator })

// StringDiscard just range over all items and do nothing with each of them.
func StringDiscard(items StringIterator) error {
	if items == nil {
		items = EmptyStringIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
