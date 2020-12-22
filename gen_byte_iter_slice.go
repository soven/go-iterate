package iter

// ByteSliceIterator is an iterator based on a slice of byte.
type ByteSliceIterator struct {
	slice []byte
	cur   int
}

// NewByteSliceIterator returns a new instance of ByteSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use ByteUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewByteSliceIterator(slice []byte) *ByteSliceIterator {
	it := &ByteSliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it ByteSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *ByteSliceIterator) Next() byte {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (ByteSliceIterator) Err() error { return nil }

// ByteSliceIterator is an iterator based on a slice of byte
// and doing iteration in back direction.
type InvertingByteSliceIterator struct {
	slice []byte
	cur   int
}

// NewInvertingByteSliceIterator returns a new instance of InvertingByteSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingByteSlice(ByteUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingByteSliceIterator(slice []byte) *InvertingByteSliceIterator {
	it := &InvertingByteSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingByteSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingByteSliceIterator) Next() byte {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingByteSliceIterator) Err() error { return nil }

// ByteUnroll unrolls items to slice of byte.
func ByteUnroll(items ByteIterator) ByteSlice {
	var slice ByteSlice
	panicIfByteIteratorError(ByteDiscard(ByteHandling(items, ByteHandle(func(item byte) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// ByteSlice is a slice of byte.
type ByteSlice []byte

// MakeIter returns a new instance of ByteIterator to iterate over it.
// It returns EmptyByteIterator if the error is not nil.
func (s ByteSlice) MakeIter() ByteIterator {
	return NewByteSliceIterator(s)
}

// ByteSlice is a slice of byte which can make inverting iterator.
type InvertingByteSlice []byte

// MakeIter returns a new instance of ByteIterator to iterate over it.
// It returns EmptyByteIterator if the error is not nil.
func (s InvertingByteSlice) MakeIter() ByteIterator {
	return NewInvertingByteSliceIterator(s)
}

// ByteInvert unrolls items and make inverting iterator based on them.
func ByteInvert(items ByteIterator) ByteIterator {
	return InvertingByteSlice(ByteUnroll(items)).MakeIter()
}
