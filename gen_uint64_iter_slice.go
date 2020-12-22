package iter

// Uint64SliceIterator is an iterator based on a slice of uint64.
type Uint64SliceIterator struct {
	slice []uint64
	cur   int
}

// NewUint64SliceIterator returns a new instance of Uint64SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Uint64Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewUint64SliceIterator(slice []uint64) *Uint64SliceIterator {
	it := &Uint64SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Uint64SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Uint64SliceIterator) Next() uint64 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Uint64SliceIterator) Err() error { return nil }

// Uint64SliceIterator is an iterator based on a slice of uint64
// and doing iteration in back direction.
type InvertingUint64SliceIterator struct {
	slice []uint64
	cur   int
}

// NewInvertingUint64SliceIterator returns a new instance of InvertingUint64SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingUint64Slice(Uint64Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingUint64SliceIterator(slice []uint64) *InvertingUint64SliceIterator {
	it := &InvertingUint64SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingUint64SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingUint64SliceIterator) Next() uint64 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingUint64SliceIterator) Err() error { return nil }

// Uint64Unroll unrolls items to slice of uint64.
func Uint64Unroll(items Uint64Iterator) Uint64Slice {
	var slice Uint64Slice
	panicIfUint64IteratorError(Uint64Discard(Uint64Handling(items, Uint64Handle(func(item uint64) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Uint64Slice is a slice of uint64.
type Uint64Slice []uint64

// MakeIter returns a new instance of Uint64Iterator to iterate over it.
// It returns EmptyUint64Iterator if the error is not nil.
func (s Uint64Slice) MakeIter() Uint64Iterator {
	return NewUint64SliceIterator(s)
}

// Uint64Slice is a slice of uint64 which can make inverting iterator.
type InvertingUint64Slice []uint64

// MakeIter returns a new instance of Uint64Iterator to iterate over it.
// It returns EmptyUint64Iterator if the error is not nil.
func (s InvertingUint64Slice) MakeIter() Uint64Iterator {
	return NewInvertingUint64SliceIterator(s)
}

// Uint64Invert unrolls items and make inverting iterator based on them.
func Uint64Invert(items Uint64Iterator) Uint64Iterator {
	return InvertingUint64Slice(Uint64Unroll(items)).MakeIter()
}
