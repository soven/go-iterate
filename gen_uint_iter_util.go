package iter

import "github.com/pkg/errors"

// EndOfUintIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfUintIterator = errors.New("end of uint iterator")

func isEndOfUintIterator(err error) bool {
	return errors.Is(err, EndOfUintIterator)
}

func wrapIfNotEndOfUintIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfUintIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfUintIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedUintItem struct {
	base    UintIterator
	hasNext bool
	next    uint
	err     error
}

func (it *preparedUintItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedUintItem) Next() uint {
	if !it.hasNext {
		panicIfUintIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedUintItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfUintIterator) {
		return it.err
	}
	return it.base.Err()
}
