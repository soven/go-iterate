package iter

// Uint8SliceIterator is an iterator based on a slice of uint8.
type Uint8SliceIterator struct {
	slice []uint8
	cur   int
}

// NewUint8SliceIterator returns a new instance of Uint8SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Uint8Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewUint8SliceIterator(slice []uint8) *Uint8SliceIterator {
	it := &Uint8SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Uint8SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Uint8SliceIterator) Next() uint8 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Uint8SliceIterator) Err() error { return nil }

// Uint8SliceIterator is an iterator based on a slice of uint8
// and doing iteration in back direction.
type InvertingUint8SliceIterator struct {
	slice []uint8
	cur   int
}

// NewInvertingUint8SliceIterator returns a new instance of InvertingUint8SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingUint8Slice(Uint8Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingUint8SliceIterator(slice []uint8) *InvertingUint8SliceIterator {
	it := &InvertingUint8SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingUint8SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingUint8SliceIterator) Next() uint8 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingUint8SliceIterator) Err() error { return nil }

// Uint8Unroll unrolls items to slice of uint8.
func Uint8Unroll(items Uint8Iterator) Uint8Slice {
	var slice Uint8Slice
	panicIfUint8IteratorError(Uint8Discard(Uint8Handling(items, Uint8Handle(func(item uint8) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Uint8Slice is a slice of uint8.
type Uint8Slice []uint8

// MakeIter returns a new instance of Uint8Iterator to iterate over it.
// It returns EmptyUint8Iterator if the error is not nil.
func (s Uint8Slice) MakeIter() Uint8Iterator {
	return NewUint8SliceIterator(s)
}

// Uint8Slice is a slice of uint8 which can make inverting iterator.
type InvertingUint8Slice []uint8

// MakeIter returns a new instance of Uint8Iterator to iterate over it.
// It returns EmptyUint8Iterator if the error is not nil.
func (s InvertingUint8Slice) MakeIter() Uint8Iterator {
	return NewInvertingUint8SliceIterator(s)
}

// Uint8Invert unrolls items and make inverting iterator based on them.
func Uint8Invert(items Uint8Iterator) Uint8Iterator {
	return InvertingUint8Slice(Uint8Unroll(items)).MakeIter()
}
