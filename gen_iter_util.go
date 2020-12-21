package iter

import "github.com/pkg/errors"

// EndOfIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfIterator = errors.New("end of interface{} iterator")

func isEndOfIterator(err error) bool {
	return errors.Is(err, EndOfIterator)
}

func wrapIfNotEndOfIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedItem struct {
	base    Iterator
	hasNext bool
	next    interface{}
	err     error
}

func (it *preparedItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedItem) Next() interface{} {
	if !it.hasNext {
		panicIfIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = nil
	return next
}

func (it preparedItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfIterator) {
		return it.err
	}
	return it.base.Err()
}
