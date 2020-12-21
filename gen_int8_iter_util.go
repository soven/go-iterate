package iter

import "github.com/pkg/errors"

// EndOfInt8Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfInt8Iterator = errors.New("end of int8 iterator")

func isEndOfInt8Iterator(err error) bool {
	return errors.Is(err, EndOfInt8Iterator)
}

func wrapIfNotEndOfInt8Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfInt8Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfInt8IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedInt8Item struct {
	base    Int8Iterator
	hasNext bool
	next    int8
	err     error
}

func (it *preparedInt8Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedInt8Item) Next() int8 {
	if !it.hasNext {
		panicIfInt8IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedInt8Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfInt8Iterator) {
		return it.err
	}
	return it.base.Err()
}
