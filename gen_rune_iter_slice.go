package iter

// RuneSliceIterator is an iterator based on a slice of rune.
type RuneSliceIterator struct {
	slice []rune
	cur   int
}

// NewRuneSliceIterator returns a new instance of RuneSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use RuneUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewRuneSliceIterator(slice []rune) *RuneSliceIterator {
	it := &RuneSliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it RuneSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *RuneSliceIterator) Next() rune {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (RuneSliceIterator) Err() error { return nil }

// RuneSliceIterator is an iterator based on a slice of rune
// and doing iteration in back direction.
type InvertingRuneSliceIterator struct {
	slice []rune
	cur   int
}

// NewInvertingRuneSliceIterator returns a new instance of InvertingRuneSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingRuneSlice(RuneUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingRuneSliceIterator(slice []rune) *InvertingRuneSliceIterator {
	it := &InvertingRuneSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingRuneSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingRuneSliceIterator) Next() rune {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingRuneSliceIterator) Err() error { return nil }

// RuneUnroll unrolls items to slice of rune.
func RuneUnroll(items RuneIterator) RuneSlice {
	var slice RuneSlice
	panicIfRuneIteratorError(RuneDiscard(RuneHandling(items, RuneHandle(func(item rune) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// RuneSlice is a slice of rune.
type RuneSlice []rune

// MakeIter returns a new instance of RuneIterator to iterate over it.
// It returns EmptyRuneIterator if the error is not nil.
func (s RuneSlice) MakeIter() RuneIterator {
	return NewRuneSliceIterator(s)
}

// RuneSlice is a slice of rune which can make inverting iterator.
type InvertingRuneSlice []rune

// MakeIter returns a new instance of RuneIterator to iterate over it.
// It returns EmptyRuneIterator if the error is not nil.
func (s InvertingRuneSlice) MakeIter() RuneIterator {
	return NewInvertingRuneSliceIterator(s)
}

// RuneInvert unrolls items and make inverting iterator based on them.
func RuneInvert(items RuneIterator) RuneIterator {
	return InvertingRuneSlice(RuneUnroll(items)).MakeIter()
}
