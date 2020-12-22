package iter

// SliceIterator is an iterator based on a slice of interface{}.
type SliceIterator struct {
	slice []interface{}
	cur   int
}

// NewSliceIterator returns a new instance of SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewSliceIterator(slice []interface{}) *SliceIterator {
	it := &SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *SliceIterator) Next() interface{} {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (SliceIterator) Err() error { return nil }

// SliceIterator is an iterator based on a slice of interface{}
// and doing iteration in back direction.
type InvertingSliceIterator struct {
	slice []interface{}
	cur   int
}

// NewInvertingSliceIterator returns a new instance of InvertingSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingSlice(Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingSliceIterator(slice []interface{}) *InvertingSliceIterator {
	it := &InvertingSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingSliceIterator) Next() interface{} {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingSliceIterator) Err() error { return nil }

// Unroll unrolls items to slice of interface{}.
func Unroll(items Iterator) Slice {
	var slice Slice
	panicIfIteratorError(Discard(Handling(items, Handle(func(item interface{}) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Slice is a slice of interface{}.
type Slice []interface{}

// MakeIter returns a new instance of Iterator to iterate over it.
// It returns EmptyIterator if the error is not nil.
func (s Slice) MakeIter() Iterator {
	return NewSliceIterator(s)
}

// Slice is a slice of interface{} which can make inverting iterator.
type InvertingSlice []interface{}

// MakeIter returns a new instance of Iterator to iterate over it.
// It returns EmptyIterator if the error is not nil.
func (s InvertingSlice) MakeIter() Iterator {
	return NewInvertingSliceIterator(s)
}

// Invert unrolls items and make inverting iterator based on them.
func Invert(items Iterator) Iterator {
	return InvertingSlice(Unroll(items)).MakeIter()
}
