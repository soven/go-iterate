package iter

import (
	"github.com/pkg/errors"
)

// RuneHandler is an object handling an item type of rune.
type RuneHandler interface {
	// Handle should do something with item of rune.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Handle(rune) error
}

// RuneHandle is a shortcut implementation
// of RuneHandler based on a function.
type RuneHandle func(rune) error

// Handle does something with item of rune.
// It is suggested to return EndOfRuneIterator to stop iteration.
func (h RuneHandle) Handle(item rune) error { return h(item) }

// RuneDoNothing does nothing.
var RuneDoNothing RuneHandler = RuneHandle(func(_ rune) error { return nil })

type doubleRuneHandler struct {
	lhs, rhs RuneHandler
}

func (h doubleRuneHandler) Handle(item rune) error {
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

// RuneHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func RuneHandlerSeries(handlers ...RuneHandler) RuneHandler {
	var series = RuneDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleRuneHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingRuneIterator does iteration with
// handling by previously set handler.
type HandlingRuneIterator struct {
	preparedRuneItem
	handler RuneHandler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneHandling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func RuneHandling(items RuneIterator, handlers ...RuneHandler) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &HandlingRuneIterator{
		preparedRuneItem{base: items}, RuneHandlerSeries(handlers...)}
}

// RuneRange iterates over items and use handlers to each one.
func RuneRange(items RuneIterator, handlers ...RuneHandler) error {
	return RuneDiscard(RuneHandling(items, handlers...))
}

// RuneRangeIterator is an iterator over items.
type RuneRangeIterator interface {
	// Range should iterate over items.
	Range(...RuneHandler) error
}

type sRuneRangeIterator struct {
	iter RuneIterator
}

// ToRuneRangeIterator constructs an instance implementing RuneRangeIterator
// based on RuneIterator.
func ToRuneRangeIterator(iter RuneIterator) RuneRangeIterator {
	if iter == nil {
		iter = EmptyRuneIterator
	}
	return sRuneRangeIterator{iter: iter}
}

// MakeRuneRangeIterator constructs an instance implementing RuneRangeIterator
// based on RuneIterMaker.
func MakeRuneRangeIterator(maker RuneIterMaker) RuneRangeIterator {
	if maker == nil {
		maker = MakeNoRuneIter
	}
	return ToRuneRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sRuneRangeIterator) Range(handlers ...RuneHandler) error {
	return RuneRange(r.iter, handlers...)
}

// RuneEnumHandler is an object handling an item type of rune and its ordered number.
type RuneEnumHandler interface {
	// Handle should do something with item of rune and its ordered number.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Handle(int, rune) error
}

// RuneEnumHandle is a shortcut implementation
// of RuneEnumHandler based on a function.
type RuneEnumHandle func(int, rune) error

// Handle does something with item of rune and its ordered number.
// It is suggested to return EndOfRuneIterator to stop iteration.
func (h RuneEnumHandle) Handle(n int, item rune) error { return h(n, item) }

// RuneDoEnumNothing does nothing.
var RuneDoEnumNothing = RuneEnumHandle(func(_ int, _ rune) error { return nil })

type doubleRuneEnumHandler struct {
	lhs, rhs RuneEnumHandler
}

func (h doubleRuneEnumHandler) Handle(n int, item rune) error {
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

// RuneEnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func RuneEnumHandlerSeries(handlers ...RuneEnumHandler) RuneEnumHandler {
	var series RuneEnumHandler = RuneDoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleRuneEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingRuneIterator does iteration with
// handling by previously set handler.
type EnumHandlingRuneIterator struct {
	preparedRuneItem
	handler RuneEnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneEnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func RuneEnumHandling(items RuneIterator, handlers ...RuneEnumHandler) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &EnumHandlingRuneIterator{
		preparedRuneItem{base: items}, RuneEnumHandlerSeries(handlers...), 0}
}

// RuneEnum iterates over items and their ordering numbers and use handlers to each one.
func RuneEnum(items RuneIterator, handlers ...RuneEnumHandler) error {
	return RuneDiscard(RuneEnumHandling(items, handlers...))
}

// RuneEnumIterator is an iterator over items and their ordering numbers.
type RuneEnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...RuneEnumHandler) error
}

type sRuneEnumIterator struct {
	iter RuneIterator
}

// ToRuneEnumIterator constructs an instance implementing RuneEnumIterator
// based on RuneIterator.
func ToRuneEnumIterator(iter RuneIterator) RuneEnumIterator {
	if iter == nil {
		iter = EmptyRuneIterator
	}
	return sRuneEnumIterator{iter: iter}
}

// MakeRuneEnumIterator constructs an instance implementing RuneEnumIterator
// based on RuneIterMaker.
func MakeRuneEnumIterator(maker RuneIterMaker) RuneEnumIterator {
	if maker == nil {
		maker = MakeNoRuneIter
	}
	return ToRuneEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sRuneEnumIterator) Enum(handlers ...RuneEnumHandler) error {
	return RuneEnum(r.iter, handlers...)
}

// Range iterates over items.
func (r sRuneEnumIterator) Range(handlers ...RuneHandler) error {
	return RuneRange(r.iter, handlers...)
}
