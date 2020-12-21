package iter

// Uint16SliceIterator is an iterator based on a slice of uint16.
type Uint16SliceIterator struct {
	slice []uint16
	cur   int
}

// NewShowtimeUint16SliceIterator returns a new instance of Uint16SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Uint16Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewShowtimeUint16SliceIterator(slice []uint16) *Uint16SliceIterator {
	it := &Uint16SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Uint16SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Uint16SliceIterator) Next() uint16 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Uint16SliceIterator) Err() error { return nil }

// Uint16SliceIterator is an iterator based on a slice of uint16
// and doing iteration in back direction.
type InvertingUint16SliceIterator struct {
	slice []uint16
	cur   int
}

// NewInvertingShowtimeUint16SliceIterator returns a new instance of InvertingUint16SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingUint16Slice(Uint16Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingShowtimeUint16SliceIterator(slice []uint16) *InvertingUint16SliceIterator {
	it := &InvertingUint16SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingUint16SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingUint16SliceIterator) Next() uint16 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingUint16SliceIterator) Err() error { return nil }

// Uint16Unroll unrolls items ot slice of uint16.
func Uint16Unroll(items Uint16Iterator) Uint16Slice {
	var slice Uint16Slice
	panicIfUint16IteratorError(Uint16Discard(Uint16Handling(items, Uint16Handle(func(item uint16) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Uint16Slice is a slice of uint16.
type Uint16Slice []uint16

// MakeIter returns a new instance of Uint16Iterator to iterate over it.
// It returns EmptyUint16Iterator if the error is not nil.
func (s Uint16Slice) MakeIter() Uint16Iterator {
	return NewShowtimeUint16SliceIterator(s)
}

// Uint16Slice is a slice of uint16 which can make inverting iterator.
type InvertingUint16Slice []uint16

// MakeIter returns a new instance of Uint16Iterator to iterate over it.
// It returns EmptyUint16Iterator if the error is not nil.
func (s InvertingUint16Slice) MakeIter() Uint16Iterator {
	return NewInvertingShowtimeUint16SliceIterator(s)
}

// Uint16Invert unrolls items and make inverting iterator based on them.
func Uint16Invert(items Uint16Iterator) Uint16Iterator {
	return InvertingUint16Slice(Uint16Unroll(items)).MakeIter()
}
