package iter

// Uint32SliceIterator is an iterator based on a slice of uint32.
type Uint32SliceIterator struct {
	slice []uint32
	cur   int
}

// NewUint32SliceIterator returns a new instance of Uint32SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Uint32Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewUint32SliceIterator(slice []uint32) *Uint32SliceIterator {
	it := &Uint32SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Uint32SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Uint32SliceIterator) Next() uint32 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Uint32SliceIterator) Err() error { return nil }

// Uint32SliceIterator is an iterator based on a slice of uint32
// and doing iteration in back direction.
type InvertingUint32SliceIterator struct {
	slice []uint32
	cur   int
}

// NewInvertingUint32SliceIterator returns a new instance of InvertingUint32SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingUint32Slice(Uint32Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingUint32SliceIterator(slice []uint32) *InvertingUint32SliceIterator {
	it := &InvertingUint32SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingUint32SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingUint32SliceIterator) Next() uint32 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingUint32SliceIterator) Err() error { return nil }

// Uint32Unroll unrolls items to slice of uint32.
func Uint32Unroll(items Uint32Iterator) Uint32Slice {
	var slice Uint32Slice
	panicIfUint32IteratorError(Uint32Discard(Uint32Handling(items, Uint32Handle(func(item uint32) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Uint32Slice is a slice of uint32.
type Uint32Slice []uint32

// MakeIter returns a new instance of Uint32Iterator to iterate over it.
// It returns EmptyUint32Iterator if the error is not nil.
func (s Uint32Slice) MakeIter() Uint32Iterator {
	return NewUint32SliceIterator(s)
}

// Uint32Slice is a slice of uint32 which can make inverting iterator.
type InvertingUint32Slice []uint32

// MakeIter returns a new instance of Uint32Iterator to iterate over it.
// It returns EmptyUint32Iterator if the error is not nil.
func (s InvertingUint32Slice) MakeIter() Uint32Iterator {
	return NewInvertingUint32SliceIterator(s)
}

// Uint32Invert unrolls items and make inverting iterator based on them.
func Uint32Invert(items Uint32Iterator) Uint32Iterator {
	return InvertingUint32Slice(Uint32Unroll(items)).MakeIter()
}
