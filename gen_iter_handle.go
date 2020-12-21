package iter

import (
	"github.com/pkg/errors"
)

// Handler is an object handling an item type of interface{}.
type Handler interface {
	// Handle should do something with item of interface{}.
	// It is suggested to return EndOfIterator to stop iteration.
	Handle(interface{}) error
}

// Handle is a shortcut implementation
// of Handler based on a function.
type Handle func(interface{}) error

// Handle does something with item of interface{}.
// It is suggested to return EndOfIterator to stop iteration.
func (h Handle) Handle(item interface{}) error { return h(item) }

// DoNothing does nothing.
var DoNothing Handler = Handle(func(_ interface{}) error { return nil })

type doubleHandler struct {
	lhs, rhs Handler
}

func (h doubleHandler) Handle(item interface{}) error {
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

// HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func HandlerSeries(handlers ...Handler) Handler {
	var series = DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingIterator does iteration with
// handling by previously set handler.
type HandlingIterator struct {
	preparedItem
	handler Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfIterator(err) {
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

// Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Handling(items Iterator, handlers ...Handler) Iterator {
	if items == nil {
		return EmptyIterator
	}
	return &HandlingIterator{
		preparedItem{base: items}, HandlerSeries(handlers...)}
}

// Range iterates over items and use handlers to each one.
func Range(items Iterator, handlers ...Handler) error {
	return Discard(Handling(items, handlers...))
}

// RangeIterator is an iterator over items.
type RangeIterator interface {
	// Range should iterate over items.
	Range(...Handler) error
}

type sRangeIterator struct {
	iter Iterator
}

// ToRangeIterator constructs an instance implementing RangeIterator
// based on Iterator.
func ToRangeIterator(iter Iterator) RangeIterator {
	if iter == nil {
		iter = EmptyIterator
	}
	return sRangeIterator{iter: iter}
}

// MakeRangeIterator constructs an instance implementing RangeIterator
// based on IterMaker.
func MakeRangeIterator(maker IterMaker) RangeIterator {
	if maker == nil {
		maker = MakeNoIter
	}
	return ToRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sRangeIterator) Range(handlers ...Handler) error {
	return Range(r.iter, handlers...)
}

// EnumHandler is an object handling an item type of interface{} and its ordered number.
type EnumHandler interface {
	// Handle should do something with item of interface{} and its ordered number.
	// It is suggested to return EndOfIterator to stop iteration.
	Handle(int, interface{}) error
}

// EnumHandle is a shortcut implementation
// of EnumHandler based on a function.
type EnumHandle func(int, interface{}) error

// Handle does something with item of interface{} and its ordered number.
// It is suggested to return EndOfIterator to stop iteration.
func (h EnumHandle) Handle(n int, item interface{}) error { return h(n, item) }

// DoEnumNothing does nothing.
var DoEnumNothing = EnumHandle(func(_ int, _ interface{}) error { return nil })

type doubleEnumHandler struct {
	lhs, rhs EnumHandler
}

func (h doubleEnumHandler) Handle(n int, item interface{}) error {
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

// EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func EnumHandlerSeries(handlers ...EnumHandler) EnumHandler {
	var series EnumHandler = DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingIterator does iteration with
// handling by previously set handler.
type EnumHandlingIterator struct {
	preparedItem
	handler EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfIterator(err) {
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

// EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func EnumHandling(items Iterator, handlers ...EnumHandler) Iterator {
	if items == nil {
		return EmptyIterator
	}
	return &EnumHandlingIterator{
		preparedItem{base: items}, EnumHandlerSeries(handlers...), 0}
}

// Enum iterates over items and their ordering numbers and use handlers to each one.
func Enum(items Iterator, handlers ...EnumHandler) error {
	return Discard(EnumHandling(items, handlers...))
}

// EnumIterator is an iterator over items and their ordering numbers.
type EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...EnumHandler) error
}

type sEnumIterator struct {
	iter Iterator
}

// ToEnumIterator constructs an instance implementing EnumIterator
// based on Iterator.
func ToEnumIterator(iter Iterator) EnumIterator {
	if iter == nil {
		iter = EmptyIterator
	}
	return sEnumIterator{iter: iter}
}

// MakeEnumIterator constructs an instance implementing EnumIterator
// based on IterMaker.
func MakeEnumIterator(maker IterMaker) EnumIterator {
	if maker == nil {
		maker = MakeNoIter
	}
	return ToEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sEnumIterator) Enum(handlers ...EnumHandler) error {
	return Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sEnumIterator) Range(handlers ...Handler) error {
	return Range(r.iter, handlers...)
}
