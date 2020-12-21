package iter

import "github.com/pkg/errors"

// EndOfInt16Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfInt16Iterator = errors.New("end of int16 iterator")

func isEndOfInt16Iterator(err error) bool {
	return errors.Is(err, EndOfInt16Iterator)
}

func wrapIfNotEndOfInt16Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfInt16Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfInt16IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedInt16Item struct {
	base    Int16Iterator
	hasNext bool
	next    int16
	err     error
}

func (it *preparedInt16Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedInt16Item) Next() int16 {
	if !it.hasNext {
		panicIfInt16IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedInt16Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfInt16Iterator) {
		return it.err
	}
	return it.base.Err()
}
