// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

type StringSliceIterator struct {
	slice []string
	cur   int
}

// NewShowtimeStringSliceIterator returns a new instance of StringSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use StringUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewShowtimeStringSliceIterator(slice []string) *StringSliceIterator {
	it := &StringSliceIterator{slice: slice}
	return it
}

func (it StringSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

func (it *StringSliceIterator) Next() string {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

func (StringSliceIterator) Err() error { return nil }

type InvertingStringSliceIterator struct {
	slice []string
	cur   int
}

// NewInvertingShowtimeStringSliceIterator returns a new instance of InvertingStringSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingStringSlice(StringUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingShowtimeStringSliceIterator(slice []string) *InvertingStringSliceIterator {
	it := &InvertingStringSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

func (it InvertingStringSliceIterator) HasNext() bool {
	return it.cur >= 0
}

func (it *InvertingStringSliceIterator) Next() string {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

func (InvertingStringSliceIterator) Err() error { return nil }

func StringUnroll(items StringIterator) StringSlice {
	var slice StringSlice
	panicIfStringIteratorError(StringRange(items, StringHandle(func(item string) error {
		slice = append(slice, item)
		return nil
	})), "unroll: discard")

	return slice
}

type StringSlice []string

func (s StringSlice) MakeIter() (StringIterator, error) {
	return NewShowtimeStringSliceIterator(s), nil
}

func MakeStringSliceIter(slice []string) StringIterator {
	items, err := StringSlice(slice).MakeIter()
	panicIfStringIteratorError(err, "make slice iter")
	return items
}

type InvertingStringSlice []string

func (s InvertingStringSlice) MakeIter() (StringIterator, error) {
	return NewInvertingShowtimeStringSliceIterator(s), nil
}

func MakeInvertingStringSliceIter(slice []string) StringIterator {
	items, err := InvertingStringSlice(slice).MakeIter()
	panicIfStringIteratorError(err, "make inverting slice iter")
	return items
}

func StringInvert(items StringIterator) StringIterator {
	return MakeInvertingStringSliceIter(StringUnroll(items))
}
