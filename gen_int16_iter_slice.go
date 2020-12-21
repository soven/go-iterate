package iter

// Int16SliceIterator is an iterator based on a slice of int16.
type Int16SliceIterator struct {
	slice []int16
	cur   int
}

// NewShowtimeInt16SliceIterator returns a new instance of Int16SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Int16Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewShowtimeInt16SliceIterator(slice []int16) *Int16SliceIterator {
	it := &Int16SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Int16SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Int16SliceIterator) Next() int16 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Int16SliceIterator) Err() error { return nil }

// Int16SliceIterator is an iterator based on a slice of int16
// and doing iteration in back direction.
type InvertingInt16SliceIterator struct {
	slice []int16
	cur   int
}

// NewInvertingShowtimeInt16SliceIterator returns a new instance of InvertingInt16SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingInt16Slice(Int16Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingShowtimeInt16SliceIterator(slice []int16) *InvertingInt16SliceIterator {
	it := &InvertingInt16SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingInt16SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingInt16SliceIterator) Next() int16 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingInt16SliceIterator) Err() error { return nil }

// Int16Unroll unrolls items ot slice of int16.
func Int16Unroll(items Int16Iterator) Int16Slice {
	var slice Int16Slice
	panicIfInt16IteratorError(Int16Discard(Int16Handling(items, Int16Handle(func(item int16) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Int16Slice is a slice of int16.
type Int16Slice []int16

// MakeIter returns a new instance of Int16Iterator to iterate over it.
// It returns EmptyInt16Iterator if the error is not nil.
func (s Int16Slice) MakeIter() Int16Iterator {
	return NewShowtimeInt16SliceIterator(s)
}

// Int16Slice is a slice of int16 which can make inverting iterator.
type InvertingInt16Slice []int16

// MakeIter returns a new instance of Int16Iterator to iterate over it.
// It returns EmptyInt16Iterator if the error is not nil.
func (s InvertingInt16Slice) MakeIter() Int16Iterator {
	return NewInvertingShowtimeInt16SliceIterator(s)
}

// Int16Invert unrolls items and make inverting iterator based on them.
func Int16Invert(items Int16Iterator) Int16Iterator {
	return InvertingInt16Slice(Int16Unroll(items)).MakeIter()
}
