// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import (
	"github.com/pkg/errors"
)

// Uint8Handler is an object handling an item type of uint8.
type Uint8Handler interface {
	// Handle should do something with item of uint8.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Handle(uint8) error
}

// Uint8Handle is a shortcut implementation
// of Uint8Handler based on a function.
type Uint8Handle func(uint8) error

// Handle does something with item of uint8.
// It is suggested to return EndOfUint8Iterator to stop iteration.
func (h Uint8Handle) Handle(item uint8) error { return h(item) }

// Uint8DoNothing does nothing.
var Uint8DoNothing Uint8Handler = Uint8Handle(func(_ uint8) error { return nil })

type doubleUint8Handler struct {
	lhs, rhs Uint8Handler
}

func (h doubleUint8Handler) Handle(item uint8) error {
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

// Uint8HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint8HandlerSeries(handlers ...Uint8Handler) Uint8Handler {
	var series = Uint8DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint8Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingUint8Iterator does iteration with
// handling by previously set handler.
type HandlingUint8Iterator struct {
	preparedUint8Item
	handler Uint8Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
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

// Uint8Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Uint8Handling(items Uint8Iterator, handlers ...Uint8Handler) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &HandlingUint8Iterator{
		preparedUint8Item{base: items}, Uint8HandlerSeries(handlers...)}
}

// Uint8Range iterates over items and use handlers to each one.
func Uint8Range(items Uint8Iterator, handlers ...Uint8Handler) error {
	return Uint8Discard(Uint8Handling(items, handlers...))
}

// Uint8RangeIterator is an iterator over items.
type Uint8RangeIterator interface {
	// Range should iterate over items.
	Range(...Uint8Handler) error
}

type sUint8RangeIterator struct {
	iter Uint8Iterator
}

// ToUint8RangeIterator constructs an instance implementing Uint8RangeIterator
// based on Uint8Iterator.
func ToUint8RangeIterator(iter Uint8Iterator) Uint8RangeIterator {
	if iter == nil {
		iter = EmptyUint8Iterator
	}
	return sUint8RangeIterator{iter: iter}
}

// MakeUint8RangeIterator constructs an instance implementing Uint8RangeIterator
// based on Uint8IterMaker.
func MakeUint8RangeIterator(maker Uint8IterMaker) Uint8RangeIterator {
	if maker == nil {
		maker = MakeNoUint8Iter
	}
	return ToUint8RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sUint8RangeIterator) Range(handlers ...Uint8Handler) error {
	return Uint8Range(r.iter, handlers...)
}

// Uint8EnumHandler is an object handling an item type of uint8 and its ordered number.
type Uint8EnumHandler interface {
	// Handle should do something with item of uint8 and its ordered number.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Handle(int, uint8) error
}

// Uint8EnumHandle is a shortcut implementation
// of Uint8EnumHandler based on a function.
type Uint8EnumHandle func(int, uint8) error

// Handle does something with item of uint8 and its ordered number.
// It is suggested to return EndOfUint8Iterator to stop iteration.
func (h Uint8EnumHandle) Handle(n int, item uint8) error { return h(n, item) }

// Uint8DoEnumNothing does nothing.
var Uint8DoEnumNothing = Uint8EnumHandle(func(_ int, _ uint8) error { return nil })

type doubleUint8EnumHandler struct {
	lhs, rhs Uint8EnumHandler
}

func (h doubleUint8EnumHandler) Handle(n int, item uint8) error {
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

// Uint8EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint8EnumHandlerSeries(handlers ...Uint8EnumHandler) Uint8EnumHandler {
	var series Uint8EnumHandler = Uint8DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint8EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingUint8Iterator does iteration with
// handling by previously set handler.
type EnumHandlingUint8Iterator struct {
	preparedUint8Item
	handler Uint8EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
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

// Uint8EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Uint8EnumHandling(items Uint8Iterator, handlers ...Uint8EnumHandler) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &EnumHandlingUint8Iterator{
		preparedUint8Item{base: items}, Uint8EnumHandlerSeries(handlers...), 0}
}

// Uint8Enum iterates over items and their ordering numbers and use handlers to each one.
func Uint8Enum(items Uint8Iterator, handlers ...Uint8EnumHandler) error {
	return Uint8Discard(Uint8EnumHandling(items, handlers...))
}

// Uint8EnumIterator is an iterator over items and their ordering numbers.
type Uint8EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Uint8EnumHandler) error
}

type sUint8EnumIterator struct {
	iter Uint8Iterator
}

// ToUint8EnumIterator constructs an instance implementing Uint8EnumIterator
// based on Uint8Iterator.
func ToUint8EnumIterator(iter Uint8Iterator) Uint8EnumIterator {
	if iter == nil {
		iter = EmptyUint8Iterator
	}
	return sUint8EnumIterator{iter: iter}
}

// MakeUint8EnumIterator constructs an instance implementing Uint8EnumIterator
// based on Uint8IterMaker.
func MakeUint8EnumIterator(maker Uint8IterMaker) Uint8EnumIterator {
	if maker == nil {
		maker = MakeNoUint8Iter
	}
	return ToUint8EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sUint8EnumIterator) Enum(handlers ...Uint8EnumHandler) error {
	return Uint8Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sUint8EnumIterator) Range(handlers ...Uint8Handler) error {
	return Uint8Range(r.iter, handlers...)
}
