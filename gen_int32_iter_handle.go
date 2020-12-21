package iter

import (
	"github.com/pkg/errors"
)

// Int32Handler is an object handling an item type of int32.
type Int32Handler interface {
	// Handle should do something with item of int32.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Handle(int32) error
}

// Int32Handle is a shortcut implementation
// of Int32Handler based on a function.
type Int32Handle func(int32) error

// Handle does something with item of int32.
// It is suggested to return EndOfInt32Iterator to stop iteration.
func (h Int32Handle) Handle(item int32) error { return h(item) }

// Int32DoNothing does nothing.
var Int32DoNothing Int32Handler = Int32Handle(func(_ int32) error { return nil })

type doubleInt32Handler struct {
	lhs, rhs Int32Handler
}

func (h doubleInt32Handler) Handle(item int32) error {
	err := h.lhs.Handle(item)
	if err != nil {
		return errors.Wrap(err, "handle lhs")
	}
	err = h.rhs.Handle(item)
	if err != nil {
		return errors.Wrap(err, "handle rhs")
	}
	return nil
}

// Int32HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int32HandlerSeries(handlers ...Int32Handler) Int32Handler {
	var series = Int32DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt32Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingInt32Iterator does iteration with
// handling by previously set handler.
type HandlingInt32Iterator struct {
	preparedInt32Item
	handler Int32Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
				err = errors.Wrap(err, "handling iterator: check")
			}
			it.err = err
			return false
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int32Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Int32Handling(items Int32Iterator, handlers ...Int32Handler) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &HandlingInt32Iterator{
		preparedInt32Item{base: items}, Int32HandlerSeries(handlers...)}
}

// Int32Range iterates over items and use handlers to each one.
func Int32Range(items Int32Iterator, handlers ...Int32Handler) error {
	return Int32Discard(Int32Handling(items, handlers...))
}

// Int32RangeIterator is an iterator over items.
type Int32RangeIterator interface {
	// Range should iterate over items.
	Range(...Int32Handler) error
}

type sInt32RangeIterator struct {
	iter Int32Iterator
}

// ToInt32RangeIterator constructs an instance implementing Int32RangeIterator
// based on Int32Iterator.
func ToInt32RangeIterator(iter Int32Iterator) Int32RangeIterator {
	if iter == nil {
		iter = EmptyInt32Iterator
	}
	return sInt32RangeIterator{iter: iter}
}

// MakeInt32RangeIterator constructs an instance implementing Int32RangeIterator
// based on Int32IterMaker.
func MakeInt32RangeIterator(maker Int32IterMaker) Int32RangeIterator {
	if maker == nil {
		maker = MakeNoInt32Iter
	}
	return ToInt32RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sInt32RangeIterator) Range(handlers ...Int32Handler) error {
	return Int32Range(r.iter, handlers...)
}

// Int32EnumHandler is an object handling an item type of int32 and its ordered number.
type Int32EnumHandler interface {
	// Handle should do something with item of int32 and its ordered number.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Handle(int, int32) error
}

// Int32EnumHandle is a shortcut implementation
// of Int32EnumHandler based on a function.
type Int32EnumHandle func(int, int32) error

// Handle does something with item of int32 and its ordered number.
// It is suggested to return EndOfInt32Iterator to stop iteration.
func (h Int32EnumHandle) Handle(n int, item int32) error { return h(n, item) }

// Int32DoEnumNothing does nothing.
var Int32DoEnumNothing = Int32EnumHandle(func(_ int, _ int32) error { return nil })

type doubleInt32EnumHandler struct {
	lhs, rhs Int32EnumHandler
}

func (h doubleInt32EnumHandler) Handle(n int, item int32) error {
	err := h.lhs.Handle(n, item)
	if err != nil {
		return errors.Wrap(err, "handle lhs")
	}
	err = h.rhs.Handle(n, item)
	if err != nil {
		return errors.Wrap(err, "handle rhs")
	}
	return nil
}

// Int32EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int32EnumHandlerSeries(handlers ...Int32EnumHandler) Int32EnumHandler {
	var series Int32EnumHandler = Int32DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt32EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingInt32Iterator does iteration with
// handling by previously set handler.
type EnumHandlingInt32Iterator struct {
	preparedInt32Item
	handler Int32EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
				err = errors.Wrap(err, "enum handling iterator: check")
			}
			it.err = err
			return false
		}
		it.count++

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int32EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Int32EnumHandling(items Int32Iterator, handlers ...Int32EnumHandler) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &EnumHandlingInt32Iterator{
		preparedInt32Item{base: items}, Int32EnumHandlerSeries(handlers...), 0}
}

// Int32Enum iterates over items and their ordering numbers and use handlers to each one.
func Int32Enum(items Int32Iterator, handlers ...Int32EnumHandler) error {
	return Int32Discard(Int32EnumHandling(items, handlers...))
}

// Int32EnumIterator is an iterator over items and their ordering numbers.
type Int32EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Int32EnumHandler) error
}

type sInt32EnumIterator struct {
	iter Int32Iterator
}

// ToInt32EnumIterator constructs an instance implementing Int32EnumIterator
// based on Int32Iterator.
func ToInt32EnumIterator(iter Int32Iterator) Int32EnumIterator {
	if iter == nil {
		iter = EmptyInt32Iterator
	}
	return sInt32EnumIterator{iter: iter}
}

// MakeInt32EnumIterator constructs an instance implementing Int32EnumIterator
// based on Int32IterMaker.
func MakeInt32EnumIterator(maker Int32IterMaker) Int32EnumIterator {
	if maker == nil {
		maker = MakeNoInt32Iter
	}
	return ToInt32EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sInt32EnumIterator) Enum(handlers ...Int32EnumHandler) error {
	return Int32Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sInt32EnumIterator) Range(handlers ...Int32Handler) error {
	return Int32Range(r.iter, handlers...)
}
