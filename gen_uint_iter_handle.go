package iter

import (
	"github.com/pkg/errors"
)

// UintHandler is an object handling an item type of uint.
type UintHandler interface {
	// Handle should do something with item of uint.
	// It is suggested to return EndOfUintIterator to stop iteration.
	Handle(uint) error
}

// UintHandle is a shortcut implementation
// of UintHandler based on a function.
type UintHandle func(uint) error

// Handle does something with item of uint.
// It is suggested to return EndOfUintIterator to stop iteration.
func (h UintHandle) Handle(item uint) error { return h(item) }

// UintDoNothing does nothing.
var UintDoNothing UintHandler = UintHandle(func(_ uint) error { return nil })

type doubleUintHandler struct {
	lhs, rhs UintHandler
}

func (h doubleUintHandler) Handle(item uint) error {
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

// UintHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func UintHandlerSeries(handlers ...UintHandler) UintHandler {
	var series = UintDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUintHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingUintIterator does iteration with
// handling by previously set handler.
type HandlingUintIterator struct {
	preparedUintItem
	handler UintHandler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUintItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfUintIterator(err) {
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

// UintHandling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func UintHandling(items UintIterator, handlers ...UintHandler) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	return &HandlingUintIterator{
		preparedUintItem{base: items}, UintHandlerSeries(handlers...)}
}

// UintRange iterates over items and use handlers to each one.
func UintRange(items UintIterator, handlers ...UintHandler) error {
	return UintDiscard(UintHandling(items, handlers...))
}

// UintRangeIterator is an iterator over items.
type UintRangeIterator interface {
	// Range should iterate over items.
	Range(...UintHandler) error
}

type sUintRangeIterator struct {
	iter UintIterator
}

// ToUintRangeIterator constructs an instance implementing UintRangeIterator
// based on UintIterator.
func ToUintRangeIterator(iter UintIterator) UintRangeIterator {
	if iter == nil {
		iter = EmptyUintIterator
	}
	return sUintRangeIterator{iter: iter}
}

// MakeUintRangeIterator constructs an instance implementing UintRangeIterator
// based on UintIterMaker.
func MakeUintRangeIterator(maker UintIterMaker) UintRangeIterator {
	if maker == nil {
		maker = MakeNoUintIter
	}
	return ToUintRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sUintRangeIterator) Range(handlers ...UintHandler) error {
	return UintRange(r.iter, handlers...)
}

// UintEnumHandler is an object handling an item type of uint and its ordered number.
type UintEnumHandler interface {
	// Handle should do something with item of uint and its ordered number.
	// It is suggested to return EndOfUintIterator to stop iteration.
	Handle(int, uint) error
}

// UintEnumHandle is a shortcut implementation
// of UintEnumHandler based on a function.
type UintEnumHandle func(int, uint) error

// Handle does something with item of uint and its ordered number.
// It is suggested to return EndOfUintIterator to stop iteration.
func (h UintEnumHandle) Handle(n int, item uint) error { return h(n, item) }

// UintDoEnumNothing does nothing.
var UintDoEnumNothing = UintEnumHandle(func(_ int, _ uint) error { return nil })

type doubleUintEnumHandler struct {
	lhs, rhs UintEnumHandler
}

func (h doubleUintEnumHandler) Handle(n int, item uint) error {
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

// UintEnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func UintEnumHandlerSeries(handlers ...UintEnumHandler) UintEnumHandler {
	var series UintEnumHandler = UintDoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUintEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingUintIterator does iteration with
// handling by previously set handler.
type EnumHandlingUintIterator struct {
	preparedUintItem
	handler UintEnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUintItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfUintIterator(err) {
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

// UintEnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func UintEnumHandling(items UintIterator, handlers ...UintEnumHandler) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	return &EnumHandlingUintIterator{
		preparedUintItem{base: items}, UintEnumHandlerSeries(handlers...), 0}
}

// UintEnum iterates over items and their ordering numbers and use handlers to each one.
func UintEnum(items UintIterator, handlers ...UintEnumHandler) error {
	return UintDiscard(UintEnumHandling(items, handlers...))
}

// UintEnumIterator is an iterator over items and their ordering numbers.
type UintEnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...UintEnumHandler) error
}

type sUintEnumIterator struct {
	iter UintIterator
}

// ToUintEnumIterator constructs an instance implementing UintEnumIterator
// based on UintIterator.
func ToUintEnumIterator(iter UintIterator) UintEnumIterator {
	if iter == nil {
		iter = EmptyUintIterator
	}
	return sUintEnumIterator{iter: iter}
}

// MakeUintEnumIterator constructs an instance implementing UintEnumIterator
// based on UintIterMaker.
func MakeUintEnumIterator(maker UintIterMaker) UintEnumIterator {
	if maker == nil {
		maker = MakeNoUintIter
	}
	return ToUintEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sUintEnumIterator) Enum(handlers ...UintEnumHandler) error {
	return UintEnum(r.iter, handlers...)
}

// Range iterates over items.
func (r sUintEnumIterator) Range(handlers ...UintHandler) error {
	return UintRange(r.iter, handlers...)
}
