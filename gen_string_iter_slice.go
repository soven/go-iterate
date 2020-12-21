package iter

// StringSliceIterator is an iterator based on a slice of string.
type StringSliceIterator struct {
	slice []string
	cur   int
}

// NewShowtimeStringSliceIterator returns a new instance of StringSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use StringUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewShowtimeStringSliceIterator(slice []string) *StringSliceIterator {
	it := &StringSliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it StringSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *StringSliceIterator) Next() string {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (StringSliceIterator) Err() error { return nil }

// StringSliceIterator is an iterator based on a slice of string
// and doing iteration in back direction.
type InvertingStringSliceIterator struct {
	slice []string
	cur   int
}

// NewInvertingShowtimeStringSliceIterator returns a new instance of InvertingStringSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingStringSlice(StringUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingShowtimeStringSliceIterator(slice []string) *InvertingStringSliceIterator {
	it := &InvertingStringSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingStringSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingStringSliceIterator) Next() string {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingStringSliceIterator) Err() error { return nil }

// StringUnroll unrolls items ot slice of string.
func StringUnroll(items StringIterator) StringSlice {
	var slice StringSlice
	panicIfStringIteratorError(StringDiscard(StringHandling(items, StringHandle(func(item string) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// StringSlice is a slice of string.
type StringSlice []string

// MakeIter returns a new instance of StringIterator to iterate over it.
// It returns EmptyStringIterator if the error is not nil.
func (s StringSlice) MakeIter() StringIterator {
	return NewShowtimeStringSliceIterator(s)
}

// StringSlice is a slice of string which can make inverting iterator.
type InvertingStringSlice []string

// MakeIter returns a new instance of StringIterator to iterate over it.
// It returns EmptyStringIterator if the error is not nil.
func (s InvertingStringSlice) MakeIter() StringIterator {
	return NewInvertingShowtimeStringSliceIterator(s)
}

// StringInvert unrolls items and make inverting iterator based on them.
func StringInvert(items StringIterator) StringIterator {
	return InvertingStringSlice(StringUnroll(items)).MakeIter()
}
