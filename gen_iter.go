package iter

// Iterator is an iterator over items type of interface{}.
type Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() interface{}
	// Err contains first met error while Next.
	Err() error
}

type emptyIterator struct{}

func (emptyIterator) HasNext() bool            { return false }
func (emptyIterator) Next() (next interface{}) { return nil }
func (emptyIterator) Err() error               { return nil }

// EmptyIterator is a zero value for Iterator.
// It is not contains any item to iterate over it.
var EmptyIterator Iterator = emptyIterator{}

// IterMaker is a maker of Iterator.
type IterMaker interface {
	// MakeIter should return a new instance of Iterator to iterate over it.
	MakeIter() Iterator
}

// MakeIter is a shortcut implementation
// of Iterator based on a function.
type MakeIter func() Iterator

// MakeIter returns a new instance of Iterator to iterate over it.
func (m MakeIter) MakeIter() Iterator { return m() }

// MakeNoIter is a zero value for IterMaker.
// It always returns EmptyIterator and an empty error.
var MakeNoIter IterMaker = MakeIter(
	func() Iterator { return EmptyIterator })

// Discard just range over all items and do nothing with each of them.
func Discard(items Iterator) error {
	if items == nil {
		items = EmptyIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
