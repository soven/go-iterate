package resembled

// PrefixIterator is an iterator over items type of Type.
type PrefixIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() Type
	// Err contains first met error while Next.
	Err() error
}

type emptyPrefixIterator struct{}

func (emptyPrefixIterator) HasNext() bool     { return false }
func (emptyPrefixIterator) Next() (next Type) { return Zero }
func (emptyPrefixIterator) Err() error        { return nil }

// EmptyPrefixIterator is a zero value for PrefixIterator.
// It is not contains any item to iterate over it.
var EmptyPrefixIterator PrefixIterator = emptyPrefixIterator{}

// PrefixIterMaker is a maker of PrefixIterator.
type PrefixIterMaker interface {
	// MakeIter should return a new instance of PrefixIterator to iterate over it.
	// It is suggested to return EmptyPrefixIterator if the error is not nil.
	MakeIter() (PrefixIterator, error)
}

// MakePrefixIter is a shortcut implementation
// of PrefixIterator based on a function.
type MakePrefixIter func() (PrefixIterator, error)

// MakeIter returns a new instance of PrefixIterator to iterate over it.
// It returns EmptyPrefixIterator if the error is not nil.
func (m MakePrefixIter) MakeIter() (PrefixIterator, error) { return m() }

// MakeNoPrefixIter is a zero value for PrefixIterMaker.
// It always returns EmptyPrefixIterator and an empty error.
var MakeNoPrefixIter PrefixIterMaker = MakePrefixIter(
	func() (PrefixIterator, error) { return EmptyPrefixIterator, nil })

// PrefixDiscard just range over all items and do nothing with each of them.
func PrefixDiscard(items PrefixIterator) error {
	if items == nil {
		items = EmptyPrefixIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
