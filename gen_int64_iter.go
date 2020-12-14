// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

type Int64Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should be invoked after check HasNext.
	Next() int64
	// Err contains first met error while Next.
	Err() error
}

type emptyInt64Iterator struct{}

func (emptyInt64Iterator) HasNext() bool      { return false }
func (emptyInt64Iterator) Next() (next int64) { return 0 }
func (emptyInt64Iterator) Err() error         { return nil }

var EmptyInt64Iterator = emptyInt64Iterator{}

type Int64IterMaker interface {
	MakeIter() (Int64Iterator, error)
}

type MakeInt64Iter func() (Int64Iterator, error)

func (m MakeInt64Iter) MakeIter() (Int64Iterator, error) { return m() }

var MakeNoInt64Iter = MakeInt64Iter(func() (Int64Iterator, error) { return EmptyInt64Iterator, nil })

// Int64Discard just range over all items and do nothing with each of them.
func Int64Discard(items Int64Iterator) error {
	if items == nil {
		return nil
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}