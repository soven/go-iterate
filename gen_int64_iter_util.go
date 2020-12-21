package iter

import "github.com/pkg/errors"

// EndOfInt64Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfInt64Iterator = errors.New("end of int64 iterator")

func isEndOfInt64Iterator(err error) bool {
	return errors.Is(err, EndOfInt64Iterator)
}

func wrapIfNotEndOfInt64Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfInt64Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfInt64IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedInt64Item struct {
	base    Int64Iterator
	hasNext bool
	next    int64
	err     error
}

func (it *preparedInt64Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedInt64Item) Next() int64 {
	if !it.hasNext {
		panicIfInt64IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedInt64Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfInt64Iterator) {
		return it.err
	}
	return it.base.Err()
}
