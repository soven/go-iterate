// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

type Handler interface {
	// It is suggested to return EndOfIterator to stop iteration.
	Handle(interface{}) error
}

type Handle func(interface{}) error

func (h Handle) Handle(item interface{}) error { return h(item) }

var DoNothing = Handle(func(_ interface{}) error { return nil })

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

func HandlerSeries(handlers ...Handler) Handler {
	var series Handler = DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

type HandlingIterator struct {
	preparedItem
	handler Handler
}

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
	return &HandlingIterator{preparedItem{base: items}, HandlerSeries(handlers...)}
}

func Range(items Iterator, handler ...Handler) error {
	// no error wrapping since no additional context for the error; just return it.
	return Discard(Handling(items, handler...))
}

type EnumHandler interface {
	// It is suggested to return EndOfIterator to stop iteration.
	Handle(int, interface{}) error
}

type EnumHandle func(int, interface{}) error

func (h EnumHandle) Handle(n int, item interface{}) error { return h(n, item) }

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

type EnumHandlingIterator struct {
	preparedItem
	handler EnumHandler
	count   int
}

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

func Enumerate(items Iterator, handler ...EnumHandler) error {
	// no error wrapping since no additional context for the error; just return it.
	return Discard(EnumHandling(items, handler...))
}
