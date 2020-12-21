package iter

// Int32SliceIterator is an iterator based on a slice of int32.
type Int32SliceIterator struct {
	slice []int32
	cur   int
}

// NewShowtimeInt32SliceIterator returns a new instance of Int32SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Int32Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewShowtimeInt32SliceIterator(slice []int32) *Int32SliceIterator {
	it := &Int32SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Int32SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Int32SliceIterator) Next() int32 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Int32SliceIterator) Err() error { return nil }

// Int32SliceIterator is an iterator based on a slice of int32
// and doing iteration in back direction.
type InvertingInt32SliceIterator struct {
	slice []int32
	cur   int
}

// NewInvertingShowtimeInt32SliceIterator returns a new instance of InvertingInt32SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingInt32Slice(Int32Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingShowtimeInt32SliceIterator(slice []int32) *InvertingInt32SliceIterator {
	it := &InvertingInt32SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingInt32SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingInt32SliceIterator) Next() int32 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingInt32SliceIterator) Err() error { return nil }

// Int32Unroll unrolls items ot slice of int32.
func Int32Unroll(items Int32Iterator) Int32Slice {
	var slice Int32Slice
	panicIfInt32IteratorError(Int32Discard(Int32Handling(items, Int32Handle(func(item int32) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Int32Slice is a slice of int32.
type Int32Slice []int32

// MakeIter returns a new instance of Int32Iterator to iterate over it.
// It returns EmptyInt32Iterator if the error is not nil.
func (s Int32Slice) MakeIter() Int32Iterator {
	return NewShowtimeInt32SliceIterator(s)
}

// Int32Slice is a slice of int32 which can make inverting iterator.
type InvertingInt32Slice []int32

// MakeIter returns a new instance of Int32Iterator to iterate over it.
// It returns EmptyInt32Iterator if the error is not nil.
func (s InvertingInt32Slice) MakeIter() Int32Iterator {
	return NewInvertingShowtimeInt32SliceIterator(s)
}

// Int32Invert unrolls items and make inverting iterator based on them.
func Int32Invert(items Int32Iterator) Int32Iterator {
	return InvertingInt32Slice(Int32Unroll(items)).MakeIter()
}
