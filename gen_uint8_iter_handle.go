// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

type Uint8Handler interface {
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Handle(uint8) error
}

type Uint8Handle func(uint8) error

func (h Uint8Handle) Handle(item uint8) error { return h(item) }

var Uint8DoNothing = Uint8Handle(func(_ uint8) error { return nil })

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

func Uint8HandlerSeries(handlers ...Uint8Handler) Uint8Handler {
	var series Uint8Handler = Uint8DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint8Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

type HandlingUint8Iterator struct {
	preparedUint8Item
	handler Uint8Handler
}

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
	return &HandlingUint8Iterator{preparedUint8Item{base: items}, Uint8HandlerSeries(handlers...)}
}

func Uint8Range(items Uint8Iterator, handler ...Uint8Handler) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint8Discard(Uint8Handling(items, handler...))
}

type Uint8EnumHandler interface {
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Handle(int, uint8) error
}

type Uint8EnumHandle func(int, uint8) error

func (h Uint8EnumHandle) Handle(n int, item uint8) error { return h(n, item) }

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

type EnumHandlingUint8Iterator struct {
	preparedUint8Item
	handler Uint8EnumHandler
	count   int
}

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

func Uint8Enumerate(items Uint8Iterator, handler ...Uint8EnumHandler) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint8Discard(Uint8EnumHandling(items, handler...))
}
