// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

type StringHandler interface {
	// It is suggested to return EndOfStringIterator to stop iteration.
	Handle(string) error
}

type StringHandle func(string) error

func (h StringHandle) Handle(item string) error { return h(item) }

var StringDoNothing = StringHandle(func(_ string) error { return nil })

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

func StringHandlerSeries(handlers ...StringHandler) StringHandler {
	var series StringHandler = StringDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleStringHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

type HandlingStringIterator struct {
	preparedStringItem
	handler StringHandler
}

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
	return &HandlingStringIterator{preparedStringItem{base: items}, StringHandlerSeries(handlers...)}
}

func StringRange(items StringIterator, handler ...StringHandler) error {
	// no error wrapping since no additional context for the error; just return it.
	return StringDiscard(StringHandling(items, handler...))
}

type StringEnumHandler interface {
	// It is suggested to return EndOfStringIterator to stop iteration.
	Handle(int, string) error
}

type StringEnumHandle func(int, string) error

func (h StringEnumHandle) Handle(n int, item string) error { return h(n, item) }

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

type EnumHandlingStringIterator struct {
	preparedStringItem
	handler StringEnumHandler
	count   int
}

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

func StringEnumerate(items StringIterator, handler ...StringEnumHandler) error {
	// no error wrapping since no additional context for the error; just return it.
	return StringDiscard(StringEnumHandling(items, handler...))
}