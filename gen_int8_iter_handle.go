package iter

import (
	"github.com/pkg/errors"
)

// Int8Handler is an object handling an item type of int8.
type Int8Handler interface {
	// Handle should do something with item of int8.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Handle(int8) error
}

// Int8Handle is a shortcut implementation
// of Int8Handler based on a function.
type Int8Handle func(int8) error

// Handle does something with item of int8.
// It is suggested to return EndOfInt8Iterator to stop iteration.
func (h Int8Handle) Handle(item int8) error { return h(item) }

// Int8DoNothing does nothing.
var Int8DoNothing Int8Handler = Int8Handle(func(_ int8) error { return nil })

type doubleInt8Handler struct {
	lhs, rhs Int8Handler
}

func (h doubleInt8Handler) Handle(item int8) error {
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

// Int8HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int8HandlerSeries(handlers ...Int8Handler) Int8Handler {
	var series = Int8DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt8Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingInt8Iterator does iteration with
// handling by previously set handler.
type HandlingInt8Iterator struct {
	preparedInt8Item
	handler Int8Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Int8Handling(items Int8Iterator, handlers ...Int8Handler) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &HandlingInt8Iterator{
		preparedInt8Item{base: items}, Int8HandlerSeries(handlers...)}
}

// Int8Range iterates over items and use handlers to each one.
func Int8Range(items Int8Iterator, handlers ...Int8Handler) error {
	return Int8Discard(Int8Handling(items, handlers...))
}

// Int8RangeIterator is an iterator over items.
type Int8RangeIterator interface {
	// Range should iterate over items.
	Range(...Int8Handler) error
}

type sInt8RangeIterator struct {
	iter Int8Iterator
}

// ToInt8RangeIterator constructs an instance implementing Int8RangeIterator
// based on Int8Iterator.
func ToInt8RangeIterator(iter Int8Iterator) Int8RangeIterator {
	if iter == nil {
		iter = EmptyInt8Iterator
	}
	return sInt8RangeIterator{iter: iter}
}

// MakeInt8RangeIterator constructs an instance implementing Int8RangeIterator
// based on Int8IterMaker.
func MakeInt8RangeIterator(maker Int8IterMaker) Int8RangeIterator {
	if maker == nil {
		maker = MakeNoInt8Iter
	}
	return ToInt8RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sInt8RangeIterator) Range(handlers ...Int8Handler) error {
	return Int8Range(r.iter, handlers...)
}

// Int8EnumHandler is an object handling an item type of int8 and its ordered number.
type Int8EnumHandler interface {
	// Handle should do something with item of int8 and its ordered number.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Handle(int, int8) error
}

// Int8EnumHandle is a shortcut implementation
// of Int8EnumHandler based on a function.
type Int8EnumHandle func(int, int8) error

// Handle does something with item of int8 and its ordered number.
// It is suggested to return EndOfInt8Iterator to stop iteration.
func (h Int8EnumHandle) Handle(n int, item int8) error { return h(n, item) }

// Int8DoEnumNothing does nothing.
var Int8DoEnumNothing = Int8EnumHandle(func(_ int, _ int8) error { return nil })

type doubleInt8EnumHandler struct {
	lhs, rhs Int8EnumHandler
}

func (h doubleInt8EnumHandler) Handle(n int, item int8) error {
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

// Int8EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int8EnumHandlerSeries(handlers ...Int8EnumHandler) Int8EnumHandler {
	var series Int8EnumHandler = Int8DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt8EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingInt8Iterator does iteration with
// handling by previously set handler.
type EnumHandlingInt8Iterator struct {
	preparedInt8Item
	handler Int8EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Int8EnumHandling(items Int8Iterator, handlers ...Int8EnumHandler) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &EnumHandlingInt8Iterator{
		preparedInt8Item{base: items}, Int8EnumHandlerSeries(handlers...), 0}
}

// Int8Enum iterates over items and their ordering numbers and use handlers to each one.
func Int8Enum(items Int8Iterator, handlers ...Int8EnumHandler) error {
	return Int8Discard(Int8EnumHandling(items, handlers...))
}

// Int8EnumIterator is an iterator over items and their ordering numbers.
type Int8EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Int8EnumHandler) error
}

type sInt8EnumIterator struct {
	iter Int8Iterator
}

// ToInt8EnumIterator constructs an instance implementing Int8EnumIterator
// based on Int8Iterator.
func ToInt8EnumIterator(iter Int8Iterator) Int8EnumIterator {
	if iter == nil {
		iter = EmptyInt8Iterator
	}
	return sInt8EnumIterator{iter: iter}
}

// MakeInt8EnumIterator constructs an instance implementing Int8EnumIterator
// based on Int8IterMaker.
func MakeInt8EnumIterator(maker Int8IterMaker) Int8EnumIterator {
	if maker == nil {
		maker = MakeNoInt8Iter
	}
	return ToInt8EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sInt8EnumIterator) Enum(handlers ...Int8EnumHandler) error {
	return Int8Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sInt8EnumIterator) Range(handlers ...Int8Handler) error {
	return Int8Range(r.iter, handlers...)
}
