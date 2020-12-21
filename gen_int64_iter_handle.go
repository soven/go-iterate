package iter

import (
	"github.com/pkg/errors"
)

// Int64Handler is an object handling an item type of int64.
type Int64Handler interface {
	// Handle should do something with item of int64.
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Handle(int64) error
}

// Int64Handle is a shortcut implementation
// of Int64Handler based on a function.
type Int64Handle func(int64) error

// Handle does something with item of int64.
// It is suggested to return EndOfInt64Iterator to stop iteration.
func (h Int64Handle) Handle(item int64) error { return h(item) }

// Int64DoNothing does nothing.
var Int64DoNothing Int64Handler = Int64Handle(func(_ int64) error { return nil })

type doubleInt64Handler struct {
	lhs, rhs Int64Handler
}

func (h doubleInt64Handler) Handle(item int64) error {
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

// Int64HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int64HandlerSeries(handlers ...Int64Handler) Int64Handler {
	var series = Int64DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt64Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingInt64Iterator does iteration with
// handling by previously set handler.
type HandlingInt64Iterator struct {
	preparedInt64Item
	handler Int64Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Int64Handling(items Int64Iterator, handlers ...Int64Handler) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &HandlingInt64Iterator{
		preparedInt64Item{base: items}, Int64HandlerSeries(handlers...)}
}

// Int64Range iterates over items and use handlers to each one.
func Int64Range(items Int64Iterator, handlers ...Int64Handler) error {
	return Int64Discard(Int64Handling(items, handlers...))
}

// Int64RangeIterator is an iterator over items.
type Int64RangeIterator interface {
	// Range should iterate over items.
	Range(...Int64Handler) error
}

type sInt64RangeIterator struct {
	iter Int64Iterator
}

// ToInt64RangeIterator constructs an instance implementing Int64RangeIterator
// based on Int64Iterator.
func ToInt64RangeIterator(iter Int64Iterator) Int64RangeIterator {
	if iter == nil {
		iter = EmptyInt64Iterator
	}
	return sInt64RangeIterator{iter: iter}
}

// MakeInt64RangeIterator constructs an instance implementing Int64RangeIterator
// based on Int64IterMaker.
func MakeInt64RangeIterator(maker Int64IterMaker) Int64RangeIterator {
	if maker == nil {
		maker = MakeNoInt64Iter
	}
	return ToInt64RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sInt64RangeIterator) Range(handlers ...Int64Handler) error {
	return Int64Range(r.iter, handlers...)
}

// Int64EnumHandler is an object handling an item type of int64 and its ordered number.
type Int64EnumHandler interface {
	// Handle should do something with item of int64 and its ordered number.
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Handle(int, int64) error
}

// Int64EnumHandle is a shortcut implementation
// of Int64EnumHandler based on a function.
type Int64EnumHandle func(int, int64) error

// Handle does something with item of int64 and its ordered number.
// It is suggested to return EndOfInt64Iterator to stop iteration.
func (h Int64EnumHandle) Handle(n int, item int64) error { return h(n, item) }

// Int64DoEnumNothing does nothing.
var Int64DoEnumNothing = Int64EnumHandle(func(_ int, _ int64) error { return nil })

type doubleInt64EnumHandler struct {
	lhs, rhs Int64EnumHandler
}

func (h doubleInt64EnumHandler) Handle(n int, item int64) error {
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

// Int64EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int64EnumHandlerSeries(handlers ...Int64EnumHandler) Int64EnumHandler {
	var series Int64EnumHandler = Int64DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt64EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingInt64Iterator does iteration with
// handling by previously set handler.
type EnumHandlingInt64Iterator struct {
	preparedInt64Item
	handler Int64EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Int64EnumHandling(items Int64Iterator, handlers ...Int64EnumHandler) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &EnumHandlingInt64Iterator{
		preparedInt64Item{base: items}, Int64EnumHandlerSeries(handlers...), 0}
}

// Int64Enum iterates over items and their ordering numbers and use handlers to each one.
func Int64Enum(items Int64Iterator, handlers ...Int64EnumHandler) error {
	return Int64Discard(Int64EnumHandling(items, handlers...))
}

// Int64EnumIterator is an iterator over items and their ordering numbers.
type Int64EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Int64EnumHandler) error
}

type sInt64EnumIterator struct {
	iter Int64Iterator
}

// ToInt64EnumIterator constructs an instance implementing Int64EnumIterator
// based on Int64Iterator.
func ToInt64EnumIterator(iter Int64Iterator) Int64EnumIterator {
	if iter == nil {
		iter = EmptyInt64Iterator
	}
	return sInt64EnumIterator{iter: iter}
}

// MakeInt64EnumIterator constructs an instance implementing Int64EnumIterator
// based on Int64IterMaker.
func MakeInt64EnumIterator(maker Int64IterMaker) Int64EnumIterator {
	if maker == nil {
		maker = MakeNoInt64Iter
	}
	return ToInt64EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sInt64EnumIterator) Enum(handlers ...Int64EnumHandler) error {
	return Int64Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sInt64EnumIterator) Range(handlers ...Int64Handler) error {
	return Int64Range(r.iter, handlers...)
}
