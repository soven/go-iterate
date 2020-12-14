// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

type StringIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should be invoked after check HasNext.
	Next() string
	// Err contains first met error while Next.
	Err() error
}

type emptyStringIterator struct{}

func (emptyStringIterator) HasNext() bool       { return false }
func (emptyStringIterator) Next() (next string) { return "" }
func (emptyStringIterator) Err() error          { return nil }

var EmptyStringIterator = emptyStringIterator{}

type StringIterMaker interface {
	MakeIter() (StringIterator, error)
}

type MakeStringIter func() (StringIterator, error)

func (m MakeStringIter) MakeIter() (StringIterator, error) { return m() }

var MakeNoStringIter = MakeStringIter(func() (StringIterator, error) { return EmptyStringIterator, nil })

// StringDiscard just range over all items and do nothing with each of them.
func StringDiscard(items StringIterator) error {
	if items == nil {
		return nil
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}
