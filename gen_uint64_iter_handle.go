package iter

import (
	"github.com/pkg/errors"
)

// Uint64Handler is an object handling an item type of uint64.
type Uint64Handler interface {
	// Handle should do something with item of uint64.
	// It is suggested to return EndOfUint64Iterator to stop iteration.
	Handle(uint64) error
}

// Uint64Handle is a shortcut implementation
// of Uint64Handler based on a function.
type Uint64Handle func(uint64) error

// Handle does something with item of uint64.
// It is suggested to return EndOfUint64Iterator to stop iteration.
func (h Uint64Handle) Handle(item uint64) error { return h(item) }

// Uint64DoNothing does nothing.
var Uint64DoNothing Uint64Handler = Uint64Handle(func(_ uint64) error { return nil })

type doubleUint64Handler struct {
	lhs, rhs Uint64Handler
}

func (h doubleUint64Handler) Handle(item uint64) error {
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

// Uint64HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint64HandlerSeries(handlers ...Uint64Handler) Uint64Handler {
	var series = Uint64DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint64Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingUint64Iterator does iteration with
// handling by previously set handler.
type HandlingUint64Iterator struct {
	preparedUint64Item
	handler Uint64Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
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

// Uint64Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Uint64Handling(items Uint64Iterator, handlers ...Uint64Handler) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &HandlingUint64Iterator{
		preparedUint64Item{base: items}, Uint64HandlerSeries(handlers...)}
}

// Uint64Range iterates over items and use handlers to each one.
func Uint64Range(items Uint64Iterator, handlers ...Uint64Handler) error {
	return Uint64Discard(Uint64Handling(items, handlers...))
}

// Uint64RangeIterator is an iterator over items.
type Uint64RangeIterator interface {
	// Range should iterate over items.
	Range(...Uint64Handler) error
}

type sUint64RangeIterator struct {
	iter Uint64Iterator
}

// ToUint64RangeIterator constructs an instance implementing Uint64RangeIterator
// based on Uint64Iterator.
func ToUint64RangeIterator(iter Uint64Iterator) Uint64RangeIterator {
	if iter == nil {
		iter = EmptyUint64Iterator
	}
	return sUint64RangeIterator{iter: iter}
}

// MakeUint64RangeIterator constructs an instance implementing Uint64RangeIterator
// based on Uint64IterMaker.
func MakeUint64RangeIterator(maker Uint64IterMaker) Uint64RangeIterator {
	if maker == nil {
		maker = MakeNoUint64Iter
	}
	return ToUint64RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sUint64RangeIterator) Range(handlers ...Uint64Handler) error {
	return Uint64Range(r.iter, handlers...)
}

// Uint64EnumHandler is an object handling an item type of uint64 and its ordered number.
type Uint64EnumHandler interface {
	// Handle should do something with item of uint64 and its ordered number.
	// It is suggested to return EndOfUint64Iterator to stop iteration.
	Handle(int, uint64) error
}

// Uint64EnumHandle is a shortcut implementation
// of Uint64EnumHandler based on a function.
type Uint64EnumHandle func(int, uint64) error

// Handle does something with item of uint64 and its ordered number.
// It is suggested to return EndOfUint64Iterator to stop iteration.
func (h Uint64EnumHandle) Handle(n int, item uint64) error { return h(n, item) }

// Uint64DoEnumNothing does nothing.
var Uint64DoEnumNothing = Uint64EnumHandle(func(_ int, _ uint64) error { return nil })

type doubleUint64EnumHandler struct {
	lhs, rhs Uint64EnumHandler
}

func (h doubleUint64EnumHandler) Handle(n int, item uint64) error {
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

// Uint64EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint64EnumHandlerSeries(handlers ...Uint64EnumHandler) Uint64EnumHandler {
	var series Uint64EnumHandler = Uint64DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint64EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingUint64Iterator does iteration with
// handling by previously set handler.
type EnumHandlingUint64Iterator struct {
	preparedUint64Item
	handler Uint64EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
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

// Uint64EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Uint64EnumHandling(items Uint64Iterator, handlers ...Uint64EnumHandler) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &EnumHandlingUint64Iterator{
		preparedUint64Item{base: items}, Uint64EnumHandlerSeries(handlers...), 0}
}

// Uint64Enum iterates over items and their ordering numbers and use handlers to each one.
func Uint64Enum(items Uint64Iterator, handlers ...Uint64EnumHandler) error {
	return Uint64Discard(Uint64EnumHandling(items, handlers...))
}

// Uint64EnumIterator is an iterator over items and their ordering numbers.
type Uint64EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Uint64EnumHandler) error
}

type sUint64EnumIterator struct {
	iter Uint64Iterator
}

// ToUint64EnumIterator constructs an instance implementing Uint64EnumIterator
// based on Uint64Iterator.
func ToUint64EnumIterator(iter Uint64Iterator) Uint64EnumIterator {
	if iter == nil {
		iter = EmptyUint64Iterator
	}
	return sUint64EnumIterator{iter: iter}
}

// MakeUint64EnumIterator constructs an instance implementing Uint64EnumIterator
// based on Uint64IterMaker.
func MakeUint64EnumIterator(maker Uint64IterMaker) Uint64EnumIterator {
	if maker == nil {
		maker = MakeNoUint64Iter
	}
	return ToUint64EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sUint64EnumIterator) Enum(handlers ...Uint64EnumHandler) error {
	return Uint64Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sUint64EnumIterator) Range(handlers ...Uint64Handler) error {
	return Uint64Range(r.iter, handlers...)
}
