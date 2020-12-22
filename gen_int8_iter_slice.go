package iter

// Int8SliceIterator is an iterator based on a slice of int8.
type Int8SliceIterator struct {
	slice []int8
	cur   int
}

// NewInt8SliceIterator returns a new instance of Int8SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Int8Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewInt8SliceIterator(slice []int8) *Int8SliceIterator {
	it := &Int8SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Int8SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Int8SliceIterator) Next() int8 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Int8SliceIterator) Err() error { return nil }

// Int8SliceIterator is an iterator based on a slice of int8
// and doing iteration in back direction.
type InvertingInt8SliceIterator struct {
	slice []int8
	cur   int
}

// NewInvertingInt8SliceIterator returns a new instance of InvertingInt8SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingInt8Slice(Int8Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingInt8SliceIterator(slice []int8) *InvertingInt8SliceIterator {
	it := &InvertingInt8SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingInt8SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingInt8SliceIterator) Next() int8 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingInt8SliceIterator) Err() error { return nil }

// Int8Unroll unrolls items to slice of int8.
func Int8Unroll(items Int8Iterator) Int8Slice {
	var slice Int8Slice
	panicIfInt8IteratorError(Int8Discard(Int8Handling(items, Int8Handle(func(item int8) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Int8Slice is a slice of int8.
type Int8Slice []int8

// MakeIter returns a new instance of Int8Iterator to iterate over it.
// It returns EmptyInt8Iterator if the error is not nil.
func (s Int8Slice) MakeIter() Int8Iterator {
	return NewInt8SliceIterator(s)
}

// Int8Slice is a slice of int8 which can make inverting iterator.
type InvertingInt8Slice []int8

// MakeIter returns a new instance of Int8Iterator to iterate over it.
// It returns EmptyInt8Iterator if the error is not nil.
func (s InvertingInt8Slice) MakeIter() Int8Iterator {
	return NewInvertingInt8SliceIterator(s)
}

// Int8Invert unrolls items and make inverting iterator based on them.
func Int8Invert(items Int8Iterator) Int8Iterator {
	return InvertingInt8Slice(Int8Unroll(items)).MakeIter()
}
