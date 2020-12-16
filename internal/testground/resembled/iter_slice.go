package resembled

// PrefixSliceIterator is an iterator based on a slice of Type.
type PrefixSliceIterator struct {
	slice []Type
	cur   int
}

// NewShowtimePrefixSliceIterator returns a new instance of PrefixSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use PrefixUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewShowtimePrefixSliceIterator(slice []Type) *PrefixSliceIterator {
	it := &PrefixSliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it PrefixSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *PrefixSliceIterator) Next() Type {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (PrefixSliceIterator) Err() error { return nil }

// PrefixSliceIterator is an iterator based on a slice of Type
// and doing iteration in back direction.
type InvertingPrefixSliceIterator struct {
	slice []Type
	cur   int
}

// NewInvertingShowtimePrefixSliceIterator returns a new instance of InvertingPrefixSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingPrefixSlice(PrefixUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingShowtimePrefixSliceIterator(slice []Type) *InvertingPrefixSliceIterator {
	it := &InvertingPrefixSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingPrefixSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingPrefixSliceIterator) Next() Type {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingPrefixSliceIterator) Err() error { return nil }

// PrefixUnroll unrolls items ot slice of Type.
func PrefixUnroll(items PrefixIterator) PrefixSlice {
	var slice PrefixSlice
	panicIfPrefixIteratorError(PrefixDiscard(PrefixHandling(items, PrefixHandle(func(item Type) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// PrefixSlice is a slice of Type.
type PrefixSlice []Type

// MakeIter returns a new instance of PrefixIterator to iterate over it.
// It returns EmptyPrefixIterator if the error is not nil.
func (s PrefixSlice) MakeIter() (PrefixIterator, error) {
	return NewShowtimePrefixSliceIterator(s), nil
}

// MakePrefixSliceIter it is shortcut for PrefixSlice.MakeIter since
// it does not return an error.
func MakePrefixSliceIter(slice []Type) PrefixIterator {
	items, err := PrefixSlice(slice).MakeIter()
	panicIfPrefixIteratorError(err, "make slice iter")
	return items
}

// PrefixSlice is a slice of Type which can make inverting iterator.
type InvertingPrefixSlice []Type

// MakeIter returns a new instance of PrefixIterator to iterate over it.
// It returns EmptyPrefixIterator if the error is not nil.
func (s InvertingPrefixSlice) MakeIter() (PrefixIterator, error) {
	return NewInvertingShowtimePrefixSliceIterator(s), nil
}

// MakePrefixSliceIter it is shortcut for InvertingPrefixSlice.MakeIter since
// it does not return an error.
func MakeInvertingPrefixSliceIter(slice []Type) PrefixIterator {
	items, err := InvertingPrefixSlice(slice).MakeIter()
	panicIfPrefixIteratorError(err, "make inverting slice iter")
	return items
}

// PrefixInvert unrolls items and make inverting iterator based on them.
func PrefixInvert(items PrefixIterator) PrefixIterator {
	return MakeInvertingPrefixSliceIter(PrefixUnroll(items))
}
