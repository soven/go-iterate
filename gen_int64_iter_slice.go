package iter

// Int64SliceIterator is an iterator based on a slice of int64.
type Int64SliceIterator struct {
	slice []int64
	cur   int
}

// NewShowtimeInt64SliceIterator returns a new instance of Int64SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Int64Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewShowtimeInt64SliceIterator(slice []int64) *Int64SliceIterator {
	it := &Int64SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Int64SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Int64SliceIterator) Next() int64 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Int64SliceIterator) Err() error { return nil }

// Int64SliceIterator is an iterator based on a slice of int64
// and doing iteration in back direction.
type InvertingInt64SliceIterator struct {
	slice []int64
	cur   int
}

// NewInvertingShowtimeInt64SliceIterator returns a new instance of InvertingInt64SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingInt64Slice(Int64Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingShowtimeInt64SliceIterator(slice []int64) *InvertingInt64SliceIterator {
	it := &InvertingInt64SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingInt64SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingInt64SliceIterator) Next() int64 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingInt64SliceIterator) Err() error { return nil }

// Int64Unroll unrolls items ot slice of int64.
func Int64Unroll(items Int64Iterator) Int64Slice {
	var slice Int64Slice
	panicIfInt64IteratorError(Int64Discard(Int64Handling(items, Int64Handle(func(item int64) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Int64Slice is a slice of int64.
type Int64Slice []int64

// MakeIter returns a new instance of Int64Iterator to iterate over it.
// It returns EmptyInt64Iterator if the error is not nil.
func (s Int64Slice) MakeIter() Int64Iterator {
	return NewShowtimeInt64SliceIterator(s)
}

// Int64Slice is a slice of int64 which can make inverting iterator.
type InvertingInt64Slice []int64

// MakeIter returns a new instance of Int64Iterator to iterate over it.
// It returns EmptyInt64Iterator if the error is not nil.
func (s InvertingInt64Slice) MakeIter() Int64Iterator {
	return NewInvertingShowtimeInt64SliceIterator(s)
}

// Int64Invert unrolls items and make inverting iterator based on them.
func Int64Invert(items Int64Iterator) Int64Iterator {
	return InvertingInt64Slice(Int64Unroll(items)).MakeIter()
}
