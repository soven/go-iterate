// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

type Int16Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should be invoked after check HasNext.
	Next() int16
	// Err contains first met error while Next.
	Err() error
}

type emptyInt16Iterator struct{}

func (emptyInt16Iterator) HasNext() bool      { return false }
func (emptyInt16Iterator) Next() (next int16) { return 0 }
func (emptyInt16Iterator) Err() error         { return nil }

var EmptyInt16Iterator = emptyInt16Iterator{}

type Int16IterMaker interface {
	MakeIter() (Int16Iterator, error)
}

type MakeInt16Iter func() (Int16Iterator, error)

func (m MakeInt16Iter) MakeIter() (Int16Iterator, error) { return m() }

var MakeNoInt16Iter = MakeInt16Iter(func() (Int16Iterator, error) { return EmptyInt16Iterator, nil })

// Int16Discard just range over all items and do nothing with each of them.
func Int16Discard(items Int16Iterator) error {
	if items == nil {
		return nil
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}