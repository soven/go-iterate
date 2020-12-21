package iter

import (
	"github.com/pkg/errors"
)

// Uint16Handler is an object handling an item type of uint16.
type Uint16Handler interface {
	// Handle should do something with item of uint16.
	// It is suggested to return EndOfUint16Iterator to stop iteration.
	Handle(uint16) error
}

// Uint16Handle is a shortcut implementation
// of Uint16Handler based on a function.
type Uint16Handle func(uint16) error

// Handle does something with item of uint16.
// It is suggested to return EndOfUint16Iterator to stop iteration.
func (h Uint16Handle) Handle(item uint16) error { return h(item) }

// Uint16DoNothing does nothing.
var Uint16DoNothing Uint16Handler = Uint16Handle(func(_ uint16) error { return nil })

type doubleUint16Handler struct {
	lhs, rhs Uint16Handler
}

func (h doubleUint16Handler) Handle(item uint16) error {
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

// Uint16HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint16HandlerSeries(handlers ...Uint16Handler) Uint16Handler {
	var series = Uint16DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint16Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingUint16Iterator does iteration with
// handling by previously set handler.
type HandlingUint16Iterator struct {
	preparedUint16Item
	handler Uint16Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
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

// Uint16Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Uint16Handling(items Uint16Iterator, handlers ...Uint16Handler) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	return &HandlingUint16Iterator{
		preparedUint16Item{base: items}, Uint16HandlerSeries(handlers...)}
}

// Uint16Range iterates over items and use handlers to each one.
func Uint16Range(items Uint16Iterator, handlers ...Uint16Handler) error {
	return Uint16Discard(Uint16Handling(items, handlers...))
}

// Uint16RangeIterator is an iterator over items.
type Uint16RangeIterator interface {
	// Range should iterate over items.
	Range(...Uint16Handler) error
}

type sUint16RangeIterator struct {
	iter Uint16Iterator
}

// ToUint16RangeIterator constructs an instance implementing Uint16RangeIterator
// based on Uint16Iterator.
func ToUint16RangeIterator(iter Uint16Iterator) Uint16RangeIterator {
	if iter == nil {
		iter = EmptyUint16Iterator
	}
	return sUint16RangeIterator{iter: iter}
}

// MakeUint16RangeIterator constructs an instance implementing Uint16RangeIterator
// based on Uint16IterMaker.
func MakeUint16RangeIterator(maker Uint16IterMaker) Uint16RangeIterator {
	if maker == nil {
		maker = MakeNoUint16Iter
	}
	return ToUint16RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sUint16RangeIterator) Range(handlers ...Uint16Handler) error {
	return Uint16Range(r.iter, handlers...)
}

// Uint16EnumHandler is an object handling an item type of uint16 and its ordered number.
type Uint16EnumHandler interface {
	// Handle should do something with item of uint16 and its ordered number.
	// It is suggested to return EndOfUint16Iterator to stop iteration.
	Handle(int, uint16) error
}

// Uint16EnumHandle is a shortcut implementation
// of Uint16EnumHandler based on a function.
type Uint16EnumHandle func(int, uint16) error

// Handle does something with item of uint16 and its ordered number.
// It is suggested to return EndOfUint16Iterator to stop iteration.
func (h Uint16EnumHandle) Handle(n int, item uint16) error { return h(n, item) }

// Uint16DoEnumNothing does nothing.
var Uint16DoEnumNothing = Uint16EnumHandle(func(_ int, _ uint16) error { return nil })

type doubleUint16EnumHandler struct {
	lhs, rhs Uint16EnumHandler
}

func (h doubleUint16EnumHandler) Handle(n int, item uint16) error {
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

// Uint16EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint16EnumHandlerSeries(handlers ...Uint16EnumHandler) Uint16EnumHandler {
	var series Uint16EnumHandler = Uint16DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint16EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingUint16Iterator does iteration with
// handling by previously set handler.
type EnumHandlingUint16Iterator struct {
	preparedUint16Item
	handler Uint16EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
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

// Uint16EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Uint16EnumHandling(items Uint16Iterator, handlers ...Uint16EnumHandler) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	return &EnumHandlingUint16Iterator{
		preparedUint16Item{base: items}, Uint16EnumHandlerSeries(handlers...), 0}
}

// Uint16Enum iterates over items and their ordering numbers and use handlers to each one.
func Uint16Enum(items Uint16Iterator, handlers ...Uint16EnumHandler) error {
	return Uint16Discard(Uint16EnumHandling(items, handlers...))
}

// Uint16EnumIterator is an iterator over items and their ordering numbers.
type Uint16EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Uint16EnumHandler) error
}

type sUint16EnumIterator struct {
	iter Uint16Iterator
}

// ToUint16EnumIterator constructs an instance implementing Uint16EnumIterator
// based on Uint16Iterator.
func ToUint16EnumIterator(iter Uint16Iterator) Uint16EnumIterator {
	if iter == nil {
		iter = EmptyUint16Iterator
	}
	return sUint16EnumIterator{iter: iter}
}

// MakeUint16EnumIterator constructs an instance implementing Uint16EnumIterator
// based on Uint16IterMaker.
func MakeUint16EnumIterator(maker Uint16IterMaker) Uint16EnumIterator {
	if maker == nil {
		maker = MakeNoUint16Iter
	}
	return ToUint16EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sUint16EnumIterator) Enum(handlers ...Uint16EnumHandler) error {
	return Uint16Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sUint16EnumIterator) Range(handlers ...Uint16Handler) error {
	return Uint16Range(r.iter, handlers...)
}
