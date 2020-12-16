// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

// IntSliceIterator is an iterator based on a slice of int.
type IntSliceIterator struct {
	slice []int
	cur   int
}

// NewShowtimeIntSliceIterator returns a new instance of IntSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use IntUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewShowtimeIntSliceIterator(slice []int) *IntSliceIterator {
	it := &IntSliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it IntSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *IntSliceIterator) Next() int {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (IntSliceIterator) Err() error { return nil }

// IntSliceIterator is an iterator based on a slice of int
// and doing iteration in back direction.
type InvertingIntSliceIterator struct {
	slice []int
	cur   int
}

// NewInvertingShowtimeIntSliceIterator returns a new instance of InvertingIntSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingIntSlice(IntUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingShowtimeIntSliceIterator(slice []int) *InvertingIntSliceIterator {
	it := &InvertingIntSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingIntSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingIntSliceIterator) Next() int {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingIntSliceIterator) Err() error { return nil }

// IntUnroll unrolls items ot slice of int.
func IntUnroll(items IntIterator) IntSlice {
	var slice IntSlice
	panicIfIntIteratorError(IntDiscard(IntHandling(items, IntHandle(func(item int) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// IntSlice is a slice of int.
type IntSlice []int

// MakeIter returns a new instance of IntIterator to iterate over it.
// It returns EmptyIntIterator if the error is not nil.
func (s IntSlice) MakeIter() (IntIterator, error) {
	return NewShowtimeIntSliceIterator(s), nil
}

// MakeIntSliceIter it is shortcut for IntSlice.MakeIter since
// it does not return an error.
func MakeIntSliceIter(slice []int) IntIterator {
	items, err := IntSlice(slice).MakeIter()
	panicIfIntIteratorError(err, "make slice iter")
	return items
}

// IntSlice is a slice of int which can make inverting iterator.
type InvertingIntSlice []int

// MakeIter returns a new instance of IntIterator to iterate over it.
// It returns EmptyIntIterator if the error is not nil.
func (s InvertingIntSlice) MakeIter() (IntIterator, error) {
	return NewInvertingShowtimeIntSliceIterator(s), nil
}

// MakeIntSliceIter it is shortcut for InvertingIntSlice.MakeIter since
// it does not return an error.
func MakeInvertingIntSliceIter(slice []int) IntIterator {
	items, err := InvertingIntSlice(slice).MakeIter()
	panicIfIntIteratorError(err, "make inverting slice iter")
	return items
}

// IntInvert unrolls items and make inverting iterator based on them.
func IntInvert(items IntIterator) IntIterator {
	return MakeInvertingIntSliceIter(IntUnroll(items))
}
