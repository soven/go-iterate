package iter

import (
	"github.com/pkg/errors"
)

// Uint32Handler is an object handling an item type of uint32.
type Uint32Handler interface {
	// Handle should do something with item of uint32.
	// It is suggested to return EndOfUint32Iterator to stop iteration.
	Handle(uint32) error
}

// Uint32Handle is a shortcut implementation
// of Uint32Handler based on a function.
type Uint32Handle func(uint32) error

// Handle does something with item of uint32.
// It is suggested to return EndOfUint32Iterator to stop iteration.
func (h Uint32Handle) Handle(item uint32) error { return h(item) }

// Uint32DoNothing does nothing.
var Uint32DoNothing Uint32Handler = Uint32Handle(func(_ uint32) error { return nil })

type doubleUint32Handler struct {
	lhs, rhs Uint32Handler
}

func (h doubleUint32Handler) Handle(item uint32) error {
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

// Uint32HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint32HandlerSeries(handlers ...Uint32Handler) Uint32Handler {
	var series = Uint32DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint32Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingUint32Iterator does iteration with
// handling by previously set handler.
type HandlingUint32Iterator struct {
	preparedUint32Item
	handler Uint32Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingUint32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint32Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfUint32Iterator(err) {
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

// Uint32Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Uint32Handling(items Uint32Iterator, handlers ...Uint32Handler) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	return &HandlingUint32Iterator{
		preparedUint32Item{base: items}, Uint32HandlerSeries(handlers...)}
}

// Uint32Range iterates over items and use handlers to each one.
func Uint32Range(items Uint32Iterator, handlers ...Uint32Handler) error {
	return Uint32Discard(Uint32Handling(items, handlers...))
}

// Uint32RangeIterator is an iterator over items.
type Uint32RangeIterator interface {
	// Range should iterate over items.
	Range(...Uint32Handler) error
}

type sUint32RangeIterator struct {
	iter Uint32Iterator
}

// ToUint32RangeIterator constructs an instance implementing Uint32RangeIterator
// based on Uint32Iterator.
func ToUint32RangeIterator(iter Uint32Iterator) Uint32RangeIterator {
	if iter == nil {
		iter = EmptyUint32Iterator
	}
	return sUint32RangeIterator{iter: iter}
}

// MakeUint32RangeIterator constructs an instance implementing Uint32RangeIterator
// based on Uint32IterMaker.
func MakeUint32RangeIterator(maker Uint32IterMaker) Uint32RangeIterator {
	if maker == nil {
		maker = MakeNoUint32Iter
	}
	return ToUint32RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sUint32RangeIterator) Range(handlers ...Uint32Handler) error {
	return Uint32Range(r.iter, handlers...)
}

// Uint32EnumHandler is an object handling an item type of uint32 and its ordered number.
type Uint32EnumHandler interface {
	// Handle should do something with item of uint32 and its ordered number.
	// It is suggested to return EndOfUint32Iterator to stop iteration.
	Handle(int, uint32) error
}

// Uint32EnumHandle is a shortcut implementation
// of Uint32EnumHandler based on a function.
type Uint32EnumHandle func(int, uint32) error

// Handle does something with item of uint32 and its ordered number.
// It is suggested to return EndOfUint32Iterator to stop iteration.
func (h Uint32EnumHandle) Handle(n int, item uint32) error { return h(n, item) }

// Uint32DoEnumNothing does nothing.
var Uint32DoEnumNothing = Uint32EnumHandle(func(_ int, _ uint32) error { return nil })

type doubleUint32EnumHandler struct {
	lhs, rhs Uint32EnumHandler
}

func (h doubleUint32EnumHandler) Handle(n int, item uint32) error {
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

// Uint32EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint32EnumHandlerSeries(handlers ...Uint32EnumHandler) Uint32EnumHandler {
	var series Uint32EnumHandler = Uint32DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint32EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingUint32Iterator does iteration with
// handling by previously set handler.
type EnumHandlingUint32Iterator struct {
	preparedUint32Item
	handler Uint32EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingUint32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint32Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfUint32Iterator(err) {
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

// Uint32EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Uint32EnumHandling(items Uint32Iterator, handlers ...Uint32EnumHandler) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	return &EnumHandlingUint32Iterator{
		preparedUint32Item{base: items}, Uint32EnumHandlerSeries(handlers...), 0}
}

// Uint32Enum iterates over items and their ordering numbers and use handlers to each one.
func Uint32Enum(items Uint32Iterator, handlers ...Uint32EnumHandler) error {
	return Uint32Discard(Uint32EnumHandling(items, handlers...))
}

// Uint32EnumIterator is an iterator over items and their ordering numbers.
type Uint32EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Uint32EnumHandler) error
}

type sUint32EnumIterator struct {
	iter Uint32Iterator
}

// ToUint32EnumIterator constructs an instance implementing Uint32EnumIterator
// based on Uint32Iterator.
func ToUint32EnumIterator(iter Uint32Iterator) Uint32EnumIterator {
	if iter == nil {
		iter = EmptyUint32Iterator
	}
	return sUint32EnumIterator{iter: iter}
}

// MakeUint32EnumIterator constructs an instance implementing Uint32EnumIterator
// based on Uint32IterMaker.
func MakeUint32EnumIterator(maker Uint32IterMaker) Uint32EnumIterator {
	if maker == nil {
		maker = MakeNoUint32Iter
	}
	return ToUint32EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sUint32EnumIterator) Enum(handlers ...Uint32EnumHandler) error {
	return Uint32Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sUint32EnumIterator) Range(handlers ...Uint32Handler) error {
	return Uint32Range(r.iter, handlers...)
}
