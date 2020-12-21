package iter

import (
	"github.com/pkg/errors"
)

// ByteHandler is an object handling an item type of byte.
type ByteHandler interface {
	// Handle should do something with item of byte.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Handle(byte) error
}

// ByteHandle is a shortcut implementation
// of ByteHandler based on a function.
type ByteHandle func(byte) error

// Handle does something with item of byte.
// It is suggested to return EndOfByteIterator to stop iteration.
func (h ByteHandle) Handle(item byte) error { return h(item) }

// ByteDoNothing does nothing.
var ByteDoNothing ByteHandler = ByteHandle(func(_ byte) error { return nil })

type doubleByteHandler struct {
	lhs, rhs ByteHandler
}

func (h doubleByteHandler) Handle(item byte) error {
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

// ByteHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func ByteHandlerSeries(handlers ...ByteHandler) ByteHandler {
	var series = ByteDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleByteHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingByteIterator does iteration with
// handling by previously set handler.
type HandlingByteIterator struct {
	preparedByteItem
	handler ByteHandler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedByteItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteHandling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func ByteHandling(items ByteIterator, handlers ...ByteHandler) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &HandlingByteIterator{
		preparedByteItem{base: items}, ByteHandlerSeries(handlers...)}
}

// ByteRange iterates over items and use handlers to each one.
func ByteRange(items ByteIterator, handlers ...ByteHandler) error {
	return ByteDiscard(ByteHandling(items, handlers...))
}

// ByteRangeIterator is an iterator over items.
type ByteRangeIterator interface {
	// Range should iterate over items.
	Range(...ByteHandler) error
}

type sByteRangeIterator struct {
	iter ByteIterator
}

// ToByteRangeIterator constructs an instance implementing ByteRangeIterator
// based on ByteIterator.
func ToByteRangeIterator(iter ByteIterator) ByteRangeIterator {
	if iter == nil {
		iter = EmptyByteIterator
	}
	return sByteRangeIterator{iter: iter}
}

// MakeByteRangeIterator constructs an instance implementing ByteRangeIterator
// based on ByteIterMaker.
func MakeByteRangeIterator(maker ByteIterMaker) ByteRangeIterator {
	if maker == nil {
		maker = MakeNoByteIter
	}
	return ToByteRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sByteRangeIterator) Range(handlers ...ByteHandler) error {
	return ByteRange(r.iter, handlers...)
}

// ByteEnumHandler is an object handling an item type of byte and its ordered number.
type ByteEnumHandler interface {
	// Handle should do something with item of byte and its ordered number.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Handle(int, byte) error
}

// ByteEnumHandle is a shortcut implementation
// of ByteEnumHandler based on a function.
type ByteEnumHandle func(int, byte) error

// Handle does something with item of byte and its ordered number.
// It is suggested to return EndOfByteIterator to stop iteration.
func (h ByteEnumHandle) Handle(n int, item byte) error { return h(n, item) }

// ByteDoEnumNothing does nothing.
var ByteDoEnumNothing = ByteEnumHandle(func(_ int, _ byte) error { return nil })

type doubleByteEnumHandler struct {
	lhs, rhs ByteEnumHandler
}

func (h doubleByteEnumHandler) Handle(n int, item byte) error {
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

// ByteEnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func ByteEnumHandlerSeries(handlers ...ByteEnumHandler) ByteEnumHandler {
	var series ByteEnumHandler = ByteDoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleByteEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingByteIterator does iteration with
// handling by previously set handler.
type EnumHandlingByteIterator struct {
	preparedByteItem
	handler ByteEnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedByteItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteEnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func ByteEnumHandling(items ByteIterator, handlers ...ByteEnumHandler) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &EnumHandlingByteIterator{
		preparedByteItem{base: items}, ByteEnumHandlerSeries(handlers...), 0}
}

// ByteEnum iterates over items and their ordering numbers and use handlers to each one.
func ByteEnum(items ByteIterator, handlers ...ByteEnumHandler) error {
	return ByteDiscard(ByteEnumHandling(items, handlers...))
}

// ByteEnumIterator is an iterator over items and their ordering numbers.
type ByteEnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...ByteEnumHandler) error
}

type sByteEnumIterator struct {
	iter ByteIterator
}

// ToByteEnumIterator constructs an instance implementing ByteEnumIterator
// based on ByteIterator.
func ToByteEnumIterator(iter ByteIterator) ByteEnumIterator {
	if iter == nil {
		iter = EmptyByteIterator
	}
	return sByteEnumIterator{iter: iter}
}

// MakeByteEnumIterator constructs an instance implementing ByteEnumIterator
// based on ByteIterMaker.
func MakeByteEnumIterator(maker ByteIterMaker) ByteEnumIterator {
	if maker == nil {
		maker = MakeNoByteIter
	}
	return ToByteEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sByteEnumIterator) Enum(handlers ...ByteEnumHandler) error {
	return ByteEnum(r.iter, handlers...)
}

// Range iterates over items.
func (r sByteEnumIterator) Range(handlers ...ByteHandler) error {
	return ByteRange(r.iter, handlers...)
}
