package resembled

import "github.com/pkg/errors"

// PrefixHandler is an object handling an item type of Type.
type PrefixHandler interface {
	// Handle should do something with item of Type.
	// It is suggested to return EndOfPrefixIterator to stop iteration.
	Handle(Type) error
}

// PrefixHandle is a shortcut implementation
// of PrefixHandler based on a function.
type PrefixHandle func(Type) error

// Handle does something with item of Type.
// It is suggested to return EndOfPrefixIterator to stop iteration.
func (h PrefixHandle) Handle(item Type) error { return h(item) }

// PrefixDoNothing does nothing.
var PrefixDoNothing PrefixHandler = PrefixHandle(func(_ Type) error { return nil })

type doublePrefixHandler struct {
	lhs, rhs PrefixHandler
}

func (h doublePrefixHandler) Handle(item Type) error {
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

// PrefixHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func PrefixHandlerSeries(handlers ...PrefixHandler) PrefixHandler {
	var series = PrefixDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doublePrefixHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingPrefixIterator does iteration with
// handling by previously set handler.
type HandlingPrefixIterator struct {
	preparedPrefixItem
	handler PrefixHandler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingPrefixIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedPrefixItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfPrefixIterator(err) {
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

// PrefixHandling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func PrefixHandling(items PrefixIterator, handlers ...PrefixHandler) PrefixIterator {
	if items == nil {
		return EmptyPrefixIterator
	}
	return &HandlingPrefixIterator{
		preparedPrefixItem{base: items}, PrefixHandlerSeries(handlers...)}
}

// PrefixRange iterates over items and use handlers to each one.
func PrefixRange(items PrefixIterator, handlers ...PrefixHandler) error {
	return PrefixDiscard(PrefixHandling(items, handlers...))
}

// PrefixRangeIterator is an iterator over items.
type PrefixRangeIterator interface {
	// Range should iterate over items.
	Range(...PrefixHandler) error
}

type sPrefixRangeIterator struct {
	iter PrefixIterator
}

// ToPrefixRangeIterator constructs an instance implementing PrefixRangeIterator
// based on PrefixIterator.
func ToPrefixRangeIterator(iter PrefixIterator) PrefixRangeIterator {
	if iter == nil {
		iter = EmptyPrefixIterator
	}
	return sPrefixRangeIterator{iter: iter}
}

// MakePrefixRangeIterator constructs an instance implementing PrefixRangeIterator
// based on PrefixIterMaker.
func MakePrefixRangeIterator(maker PrefixIterMaker) PrefixRangeIterator {
	if maker == nil {
		maker = MakeNoPrefixIter
	}
	return ToPrefixRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sPrefixRangeIterator) Range(handlers ...PrefixHandler) error {
	return PrefixRange(r.iter, handlers...)
}

// PrefixEnumHandler is an object handling an item type of Type and its ordered number.
type PrefixEnumHandler interface {
	// Handle should do something with item of Type and its ordered number.
	// It is suggested to return EndOfPrefixIterator to stop iteration.
	Handle(int, Type) error
}

// PrefixEnumHandle is a shortcut implementation
// of PrefixEnumHandler based on a function.
type PrefixEnumHandle func(int, Type) error

// Handle does something with item of Type and its ordered number.
// It is suggested to return EndOfPrefixIterator to stop iteration.
func (h PrefixEnumHandle) Handle(n int, item Type) error { return h(n, item) }

// PrefixDoEnumNothing does nothing.
var PrefixDoEnumNothing = PrefixEnumHandle(func(_ int, _ Type) error { return nil })

type doublePrefixEnumHandler struct {
	lhs, rhs PrefixEnumHandler
}

func (h doublePrefixEnumHandler) Handle(n int, item Type) error {
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

// PrefixEnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func PrefixEnumHandlerSeries(handlers ...PrefixEnumHandler) PrefixEnumHandler {
	var series PrefixEnumHandler = PrefixDoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doublePrefixEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingPrefixIterator does iteration with
// handling by previously set handler.
type EnumHandlingPrefixIterator struct {
	preparedPrefixItem
	handler PrefixEnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingPrefixIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedPrefixItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfPrefixIterator(err) {
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

// PrefixEnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func PrefixEnumHandling(items PrefixIterator, handlers ...PrefixEnumHandler) PrefixIterator {
	if items == nil {
		return EmptyPrefixIterator
	}
	return &EnumHandlingPrefixIterator{
		preparedPrefixItem{base: items}, PrefixEnumHandlerSeries(handlers...), 0}
}

// PrefixEnum iterates over items and their ordering numbers and use handlers to each one.
func PrefixEnum(items PrefixIterator, handlers ...PrefixEnumHandler) error {
	return PrefixDiscard(PrefixEnumHandling(items, handlers...))
}

// PrefixEnumIterator is an iterator over items and their ordering numbers.
type PrefixEnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...PrefixEnumHandler) error
}

type sPrefixEnumIterator struct {
	iter PrefixIterator
}

// ToPrefixEnumIterator constructs an instance implementing PrefixEnumIterator
// based on PrefixIterator.
func ToPrefixEnumIterator(iter PrefixIterator) PrefixEnumIterator {
	if iter == nil {
		iter = EmptyPrefixIterator
	}
	return sPrefixEnumIterator{iter: iter}
}

// MakePrefixEnumIterator constructs an instance implementing PrefixEnumIterator
// based on PrefixIterMaker.
func MakePrefixEnumIterator(maker PrefixIterMaker) PrefixEnumIterator {
	if maker == nil {
		maker = MakeNoPrefixIter
	}
	return ToPrefixEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sPrefixEnumIterator) Enum(handlers ...PrefixEnumHandler) error {
	return PrefixEnum(r.iter, handlers...)
}

// Range iterates over items.
func (r sPrefixEnumIterator) Range(handlers ...PrefixHandler) error {
	return PrefixRange(r.iter, handlers...)
}
