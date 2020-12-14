// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

type IntSliceIterator struct {
	slice []int
	cur   int
}

// NewShowtimeIntSliceIterator returns a new instance of IntSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use IntUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewShowtimeIntSliceIterator(slice []int) *IntSliceIterator {
	it := &IntSliceIterator{slice: slice}
	return it
}

func (it IntSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

func (it *IntSliceIterator) Next() int {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

func (IntSliceIterator) Err() error { return nil }

type InvertingIntSliceIterator struct {
	slice []int
	cur   int
}

// NewInvertingShowtimeIntSliceIterator returns a new instance of InvertingIntSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingIntSlice(IntUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingShowtimeIntSliceIterator(slice []int) *InvertingIntSliceIterator {
	it := &InvertingIntSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

func (it InvertingIntSliceIterator) HasNext() bool {
	return it.cur >= 0
}

func (it *InvertingIntSliceIterator) Next() int {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

func (InvertingIntSliceIterator) Err() error { return nil }

func IntUnroll(items IntIterator) IntSlice {
	var slice IntSlice
	panicIfIntIteratorError(IntRange(items, IntHandle(func(item int) error {
		slice = append(slice, item)
		return nil
	})), "unroll: discard")

	return slice
}

type IntSlice []int

func (s IntSlice) MakeIter() (IntIterator, error) {
	return NewShowtimeIntSliceIterator(s), nil
}

func MakeIntSliceIter(slice []int) IntIterator {
	items, err := IntSlice(slice).MakeIter()
	panicIfIntIteratorError(err, "make slice iter")
	return items
}

type InvertingIntSlice []int

func (s InvertingIntSlice) MakeIter() (IntIterator, error) {
	return NewInvertingShowtimeIntSliceIterator(s), nil
}

func MakeInvertingIntSliceIter(slice []int) IntIterator {
	items, err := InvertingIntSlice(slice).MakeIter()
	panicIfIntIteratorError(err, "make inverting slice iter")
	return items
}

func IntInvert(items IntIterator) IntIterator {
	return MakeInvertingIntSliceIter(IntUnroll(items))
}
