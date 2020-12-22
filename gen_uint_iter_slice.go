package iter

// UintSliceIterator is an iterator based on a slice of uint.
type UintSliceIterator struct {
	slice []uint
	cur   int
}

// NewUintSliceIterator returns a new instance of UintSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use UintUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewUintSliceIterator(slice []uint) *UintSliceIterator {
	it := &UintSliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it UintSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *UintSliceIterator) Next() uint {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (UintSliceIterator) Err() error { return nil }

// UintSliceIterator is an iterator based on a slice of uint
// and doing iteration in back direction.
type InvertingUintSliceIterator struct {
	slice []uint
	cur   int
}

// NewInvertingUintSliceIterator returns a new instance of InvertingUintSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingUintSlice(UintUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingUintSliceIterator(slice []uint) *InvertingUintSliceIterator {
	it := &InvertingUintSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingUintSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingUintSliceIterator) Next() uint {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingUintSliceIterator) Err() error { return nil }

// UintUnroll unrolls items to slice of uint.
func UintUnroll(items UintIterator) UintSlice {
	var slice UintSlice
	panicIfUintIteratorError(UintDiscard(UintHandling(items, UintHandle(func(item uint) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// UintSlice is a slice of uint.
type UintSlice []uint

// MakeIter returns a new instance of UintIterator to iterate over it.
// It returns EmptyUintIterator if the error is not nil.
func (s UintSlice) MakeIter() UintIterator {
	return NewUintSliceIterator(s)
}

// UintSlice is a slice of uint which can make inverting iterator.
type InvertingUintSlice []uint

// MakeIter returns a new instance of UintIterator to iterate over it.
// It returns EmptyUintIterator if the error is not nil.
func (s InvertingUintSlice) MakeIter() UintIterator {
	return NewInvertingUintSliceIterator(s)
}

// UintInvert unrolls items and make inverting iterator based on them.
func UintInvert(items UintIterator) UintIterator {
	return InvertingUintSlice(UintUnroll(items)).MakeIter()
}
