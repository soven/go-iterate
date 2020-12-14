// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

type Int16Handler interface {
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Handle(int16) error
}

type Int16Handle func(int16) error

func (h Int16Handle) Handle(item int16) error { return h(item) }

var Int16DoNothing = Int16Handle(func(_ int16) error { return nil })

type doubleInt16Handler struct {
	lhs, rhs Int16Handler
}

func (h doubleInt16Handler) Handle(item int16) error {
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

func Int16HandlerSeries(handlers ...Int16Handler) Int16Handler {
	var series Int16Handler = Int16DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt16Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

type HandlingInt16Iterator struct {
	preparedInt16Item
	handler Int16Handler
}

func (it *HandlingInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Int16Handling(items Int16Iterator, handlers ...Int16Handler) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &HandlingInt16Iterator{preparedInt16Item{base: items}, Int16HandlerSeries(handlers...)}
}

func Int16Range(items Int16Iterator, handler ...Int16Handler) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int16Discard(Int16Handling(items, handler...))
}

type Int16EnumHandler interface {
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Handle(int, int16) error
}

type Int16EnumHandle func(int, int16) error

func (h Int16EnumHandle) Handle(n int, item int16) error { return h(n, item) }

var Int16DoEnumNothing = Int16EnumHandle(func(_ int, _ int16) error { return nil })

type doubleInt16EnumHandler struct {
	lhs, rhs Int16EnumHandler
}

func (h doubleInt16EnumHandler) Handle(n int, item int16) error {
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

func Int16EnumHandlerSeries(handlers ...Int16EnumHandler) Int16EnumHandler {
	var series Int16EnumHandler = Int16DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt16EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

type EnumHandlingInt16Iterator struct {
	preparedInt16Item
	handler Int16EnumHandler
	count   int
}

func (it *EnumHandlingInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Int16EnumHandling(items Int16Iterator, handlers ...Int16EnumHandler) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &EnumHandlingInt16Iterator{
		preparedInt16Item{base: items}, Int16EnumHandlerSeries(handlers...), 0}
}

func Int16Enumerate(items Int16Iterator, handler ...Int16EnumHandler) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int16Discard(Int16EnumHandling(items, handler...))
}