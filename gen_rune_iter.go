package iter

// RuneIterator is an iterator over items type of rune.
type RuneIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() rune
	// Err contains first met error while Next.
	Err() error
}

type emptyRuneIterator struct{}

func (emptyRuneIterator) HasNext() bool     { return false }
func (emptyRuneIterator) Next() (next rune) { return 0 }
func (emptyRuneIterator) Err() error        { return nil }

// EmptyRuneIterator is a zero value for RuneIterator.
// It is not contains any item to iterate over it.
var EmptyRuneIterator RuneIterator = emptyRuneIterator{}

// RuneIterMaker is a maker of RuneIterator.
type RuneIterMaker interface {
	// MakeIter should return a new instance of RuneIterator to iterate over it.
	MakeIter() RuneIterator
}

// MakeRuneIter is a shortcut implementation
// of RuneIterator based on a function.
type MakeRuneIter func() RuneIterator

// MakeIter returns a new instance of RuneIterator to iterate over it.
func (m MakeRuneIter) MakeIter() RuneIterator { return m() }

// MakeNoRuneIter is a zero value for RuneIterMaker.
// It always returns EmptyRuneIterator and an empty error.
var MakeNoRuneIter RuneIterMaker = MakeRuneIter(
	func() RuneIterator { return EmptyRuneIterator })

// RuneDiscard just range over all items and do nothing with each of them.
func RuneDiscard(items RuneIterator) error {
	if items == nil {
		items = EmptyRuneIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
