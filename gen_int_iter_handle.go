package iter

import (
	"github.com/pkg/errors"
)

// IntHandler is an object handling an item type of int.
type IntHandler interface {
	// Handle should do something with item of int.
	// It is suggested to return EndOfIntIterator to stop iteration.
	Handle(int) error
}

// IntHandle is a shortcut implementation
// of IntHandler based on a function.
type IntHandle func(int) error

// Handle does something with item of int.
// It is suggested to return EndOfIntIterator to stop iteration.
func (h IntHandle) Handle(item int) error { return h(item) }

// IntDoNothing does nothing.
var IntDoNothing IntHandler = IntHandle(func(_ int) error { return nil })

type doubleIntHandler struct {
	lhs, rhs IntHandler
}

func (h doubleIntHandler) Handle(item int) error {
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

// IntHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func IntHandlerSeries(handlers ...IntHandler) IntHandler {
	var series = IntDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleIntHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingIntIterator does iteration with
// handling by previously set handler.
type HandlingIntIterator struct {
	preparedIntItem
	handler IntHandler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedIntItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfIntIterator(err) {
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

// IntHandling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func IntHandling(items IntIterator, handlers ...IntHandler) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	return &HandlingIntIterator{
		preparedIntItem{base: items}, IntHandlerSeries(handlers...)}
}

// IntRange iterates over items and use handlers to each one.
func IntRange(items IntIterator, handlers ...IntHandler) error {
	return IntDiscard(IntHandling(items, handlers...))
}

// IntRangeIterator is an iterator over items.
type IntRangeIterator interface {
	// Range should iterate over items.
	Range(...IntHandler) error
}

type sIntRangeIterator struct {
	iter IntIterator
}

// ToIntRangeIterator constructs an instance implementing IntRangeIterator
// based on IntIterator.
func ToIntRangeIterator(iter IntIterator) IntRangeIterator {
	if iter == nil {
		iter = EmptyIntIterator
	}
	return sIntRangeIterator{iter: iter}
}

// MakeIntRangeIterator constructs an instance implementing IntRangeIterator
// based on IntIterMaker.
func MakeIntRangeIterator(maker IntIterMaker) IntRangeIterator {
	if maker == nil {
		maker = MakeNoIntIter
	}
	return ToIntRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sIntRangeIterator) Range(handlers ...IntHandler) error {
	return IntRange(r.iter, handlers...)
}

// IntEnumHandler is an object handling an item type of int and its ordered number.
type IntEnumHandler interface {
	// Handle should do something with item of int and its ordered number.
	// It is suggested to return EndOfIntIterator to stop iteration.
	Handle(int, int) error
}

// IntEnumHandle is a shortcut implementation
// of IntEnumHandler based on a function.
type IntEnumHandle func(int, int) error

// Handle does something with item of int and its ordered number.
// It is suggested to return EndOfIntIterator to stop iteration.
func (h IntEnumHandle) Handle(n int, item int) error { return h(n, item) }

// IntDoEnumNothing does nothing.
var IntDoEnumNothing = IntEnumHandle(func(_ int, _ int) error { return nil })

type doubleIntEnumHandler struct {
	lhs, rhs IntEnumHandler
}

func (h doubleIntEnumHandler) Handle(n int, item int) error {
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

// IntEnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func IntEnumHandlerSeries(handlers ...IntEnumHandler) IntEnumHandler {
	var series IntEnumHandler = IntDoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleIntEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingIntIterator does iteration with
// handling by previously set handler.
type EnumHandlingIntIterator struct {
	preparedIntItem
	handler IntEnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedIntItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfIntIterator(err) {
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

// IntEnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func IntEnumHandling(items IntIterator, handlers ...IntEnumHandler) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	return &EnumHandlingIntIterator{
		preparedIntItem{base: items}, IntEnumHandlerSeries(handlers...), 0}
}

// IntEnum iterates over items and their ordering numbers and use handlers to each one.
func IntEnum(items IntIterator, handlers ...IntEnumHandler) error {
	return IntDiscard(IntEnumHandling(items, handlers...))
}

// IntEnumIterator is an iterator over items and their ordering numbers.
type IntEnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...IntEnumHandler) error
}

type sIntEnumIterator struct {
	iter IntIterator
}

// ToIntEnumIterator constructs an instance implementing IntEnumIterator
// based on IntIterator.
func ToIntEnumIterator(iter IntIterator) IntEnumIterator {
	if iter == nil {
		iter = EmptyIntIterator
	}
	return sIntEnumIterator{iter: iter}
}

// MakeIntEnumIterator constructs an instance implementing IntEnumIterator
// based on IntIterMaker.
func MakeIntEnumIterator(maker IntIterMaker) IntEnumIterator {
	if maker == nil {
		maker = MakeNoIntIter
	}
	return ToIntEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sIntEnumIterator) Enum(handlers ...IntEnumHandler) error {
	return IntEnum(r.iter, handlers...)
}

// Range iterates over items.
func (r sIntEnumIterator) Range(handlers ...IntHandler) error {
	return IntRange(r.iter, handlers...)
}
