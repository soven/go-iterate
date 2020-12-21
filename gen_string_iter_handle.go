package iter

import (
	"github.com/pkg/errors"
)

// StringHandler is an object handling an item type of string.
type StringHandler interface {
	// Handle should do something with item of string.
	// It is suggested to return EndOfStringIterator to stop iteration.
	Handle(string) error
}

// StringHandle is a shortcut implementation
// of StringHandler based on a function.
type StringHandle func(string) error

// Handle does something with item of string.
// It is suggested to return EndOfStringIterator to stop iteration.
func (h StringHandle) Handle(item string) error { return h(item) }

// StringDoNothing does nothing.
var StringDoNothing StringHandler = StringHandle(func(_ string) error { return nil })

type doubleStringHandler struct {
	lhs, rhs StringHandler
}

func (h doubleStringHandler) Handle(item string) error {
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

// StringHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func StringHandlerSeries(handlers ...StringHandler) StringHandler {
	var series = StringDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleStringHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingStringIterator does iteration with
// handling by previously set handler.
type HandlingStringIterator struct {
	preparedStringItem
	handler StringHandler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedStringItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfStringIterator(err) {
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

// StringHandling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func StringHandling(items StringIterator, handlers ...StringHandler) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	return &HandlingStringIterator{
		preparedStringItem{base: items}, StringHandlerSeries(handlers...)}
}

// StringRange iterates over items and use handlers to each one.
func StringRange(items StringIterator, handlers ...StringHandler) error {
	return StringDiscard(StringHandling(items, handlers...))
}

// StringRangeIterator is an iterator over items.
type StringRangeIterator interface {
	// Range should iterate over items.
	Range(...StringHandler) error
}

type sStringRangeIterator struct {
	iter StringIterator
}

// ToStringRangeIterator constructs an instance implementing StringRangeIterator
// based on StringIterator.
func ToStringRangeIterator(iter StringIterator) StringRangeIterator {
	if iter == nil {
		iter = EmptyStringIterator
	}
	return sStringRangeIterator{iter: iter}
}

// MakeStringRangeIterator constructs an instance implementing StringRangeIterator
// based on StringIterMaker.
func MakeStringRangeIterator(maker StringIterMaker) StringRangeIterator {
	if maker == nil {
		maker = MakeNoStringIter
	}
	return ToStringRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sStringRangeIterator) Range(handlers ...StringHandler) error {
	return StringRange(r.iter, handlers...)
}

// StringEnumHandler is an object handling an item type of string and its ordered number.
type StringEnumHandler interface {
	// Handle should do something with item of string and its ordered number.
	// It is suggested to return EndOfStringIterator to stop iteration.
	Handle(int, string) error
}

// StringEnumHandle is a shortcut implementation
// of StringEnumHandler based on a function.
type StringEnumHandle func(int, string) error

// Handle does something with item of string and its ordered number.
// It is suggested to return EndOfStringIterator to stop iteration.
func (h StringEnumHandle) Handle(n int, item string) error { return h(n, item) }

// StringDoEnumNothing does nothing.
var StringDoEnumNothing = StringEnumHandle(func(_ int, _ string) error { return nil })

type doubleStringEnumHandler struct {
	lhs, rhs StringEnumHandler
}

func (h doubleStringEnumHandler) Handle(n int, item string) error {
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

// StringEnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func StringEnumHandlerSeries(handlers ...StringEnumHandler) StringEnumHandler {
	var series StringEnumHandler = StringDoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleStringEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingStringIterator does iteration with
// handling by previously set handler.
type EnumHandlingStringIterator struct {
	preparedStringItem
	handler StringEnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedStringItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfStringIterator(err) {
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

// StringEnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func StringEnumHandling(items StringIterator, handlers ...StringEnumHandler) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	return &EnumHandlingStringIterator{
		preparedStringItem{base: items}, StringEnumHandlerSeries(handlers...), 0}
}

// StringEnum iterates over items and their ordering numbers and use handlers to each one.
func StringEnum(items StringIterator, handlers ...StringEnumHandler) error {
	return StringDiscard(StringEnumHandling(items, handlers...))
}

// StringEnumIterator is an iterator over items and their ordering numbers.
type StringEnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...StringEnumHandler) error
}

type sStringEnumIterator struct {
	iter StringIterator
}

// ToStringEnumIterator constructs an instance implementing StringEnumIterator
// based on StringIterator.
func ToStringEnumIterator(iter StringIterator) StringEnumIterator {
	if iter == nil {
		iter = EmptyStringIterator
	}
	return sStringEnumIterator{iter: iter}
}

// MakeStringEnumIterator constructs an instance implementing StringEnumIterator
// based on StringIterMaker.
func MakeStringEnumIterator(maker StringIterMaker) StringEnumIterator {
	if maker == nil {
		maker = MakeNoStringIter
	}
	return ToStringEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sStringEnumIterator) Enum(handlers ...StringEnumHandler) error {
	return StringEnum(r.iter, handlers...)
}

// Range iterates over items.
func (r sStringEnumIterator) Range(handlers ...StringHandler) error {
	return StringRange(r.iter, handlers...)
}
