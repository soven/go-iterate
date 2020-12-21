package iter

import (
	"github.com/pkg/errors"
)

// Int16Handler is an object handling an item type of int16.
type Int16Handler interface {
	// Handle should do something with item of int16.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Handle(int16) error
}

// Int16Handle is a shortcut implementation
// of Int16Handler based on a function.
type Int16Handle func(int16) error

// Handle does something with item of int16.
// It is suggested to return EndOfInt16Iterator to stop iteration.
func (h Int16Handle) Handle(item int16) error { return h(item) }

// Int16DoNothing does nothing.
var Int16DoNothing Int16Handler = Int16Handle(func(_ int16) error { return nil })

type doubleInt16Handler struct {
	lhs, rhs Int16Handler
}

func (h doubleInt16Handler) Handle(item int16) error {
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

// Int16HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int16HandlerSeries(handlers ...Int16Handler) Int16Handler {
	var series = Int16DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt16Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingInt16Iterator does iteration with
// handling by previously set handler.
type HandlingInt16Iterator struct {
	preparedInt16Item
	handler Int16Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Int16Handling(items Int16Iterator, handlers ...Int16Handler) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &HandlingInt16Iterator{
		preparedInt16Item{base: items}, Int16HandlerSeries(handlers...)}
}

// Int16Range iterates over items and use handlers to each one.
func Int16Range(items Int16Iterator, handlers ...Int16Handler) error {
	return Int16Discard(Int16Handling(items, handlers...))
}

// Int16RangeIterator is an iterator over items.
type Int16RangeIterator interface {
	// Range should iterate over items.
	Range(...Int16Handler) error
}

type sInt16RangeIterator struct {
	iter Int16Iterator
}

// ToInt16RangeIterator constructs an instance implementing Int16RangeIterator
// based on Int16Iterator.
func ToInt16RangeIterator(iter Int16Iterator) Int16RangeIterator {
	if iter == nil {
		iter = EmptyInt16Iterator
	}
	return sInt16RangeIterator{iter: iter}
}

// MakeInt16RangeIterator constructs an instance implementing Int16RangeIterator
// based on Int16IterMaker.
func MakeInt16RangeIterator(maker Int16IterMaker) Int16RangeIterator {
	if maker == nil {
		maker = MakeNoInt16Iter
	}
	return ToInt16RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sInt16RangeIterator) Range(handlers ...Int16Handler) error {
	return Int16Range(r.iter, handlers...)
}

// Int16EnumHandler is an object handling an item type of int16 and its ordered number.
type Int16EnumHandler interface {
	// Handle should do something with item of int16 and its ordered number.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Handle(int, int16) error
}

// Int16EnumHandle is a shortcut implementation
// of Int16EnumHandler based on a function.
type Int16EnumHandle func(int, int16) error

// Handle does something with item of int16 and its ordered number.
// It is suggested to return EndOfInt16Iterator to stop iteration.
func (h Int16EnumHandle) Handle(n int, item int16) error { return h(n, item) }

// Int16DoEnumNothing does nothing.
var Int16DoEnumNothing = Int16EnumHandle(func(_ int, _ int16) error { return nil })

type doubleInt16EnumHandler struct {
	lhs, rhs Int16EnumHandler
}

func (h doubleInt16EnumHandler) Handle(n int, item int16) error {
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

// Int16EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int16EnumHandlerSeries(handlers ...Int16EnumHandler) Int16EnumHandler {
	var series Int16EnumHandler = Int16DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt16EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingInt16Iterator does iteration with
// handling by previously set handler.
type EnumHandlingInt16Iterator struct {
	preparedInt16Item
	handler Int16EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Int16EnumHandling(items Int16Iterator, handlers ...Int16EnumHandler) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &EnumHandlingInt16Iterator{
		preparedInt16Item{base: items}, Int16EnumHandlerSeries(handlers...), 0}
}

// Int16Enum iterates over items and their ordering numbers and use handlers to each one.
func Int16Enum(items Int16Iterator, handlers ...Int16EnumHandler) error {
	return Int16Discard(Int16EnumHandling(items, handlers...))
}

// Int16EnumIterator is an iterator over items and their ordering numbers.
type Int16EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Int16EnumHandler) error
}

type sInt16EnumIterator struct {
	iter Int16Iterator
}

// ToInt16EnumIterator constructs an instance implementing Int16EnumIterator
// based on Int16Iterator.
func ToInt16EnumIterator(iter Int16Iterator) Int16EnumIterator {
	if iter == nil {
		iter = EmptyInt16Iterator
	}
	return sInt16EnumIterator{iter: iter}
}

// MakeInt16EnumIterator constructs an instance implementing Int16EnumIterator
// based on Int16IterMaker.
func MakeInt16EnumIterator(maker Int16IterMaker) Int16EnumIterator {
	if maker == nil {
		maker = MakeNoInt16Iter
	}
	return ToInt16EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sInt16EnumIterator) Enum(handlers ...Int16EnumHandler) error {
	return Int16Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sInt16EnumIterator) Range(handlers ...Int16Handler) error {
	return Int16Range(r.iter, handlers...)
}
