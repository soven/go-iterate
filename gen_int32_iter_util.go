package iter

import "github.com/pkg/errors"

// EndOfInt32Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfInt32Iterator = errors.New("end of int32 iterator")

func isEndOfInt32Iterator(err error) bool {
	return errors.Is(err, EndOfInt32Iterator)
}

func wrapIfNotEndOfInt32Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfInt32Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfInt32IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedInt32Item struct {
	base    Int32Iterator
	hasNext bool
	next    int32
	err     error
}

func (it *preparedInt32Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedInt32Item) Next() int32 {
	if !it.hasNext {
		panicIfInt32IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedInt32Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfInt32Iterator) {
		return it.err
	}
	return it.base.Err()
}
