package iter

import "github.com/pkg/errors"

// EndOfIntIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfIntIterator = errors.New("end of int iterator")

func isEndOfIntIterator(err error) bool {
	return errors.Is(err, EndOfIntIterator)
}

func wrapIfNotEndOfIntIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfIntIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfIntIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedIntItem struct {
	base    IntIterator
	hasNext bool
	next    int
	err     error
}

func (it *preparedIntItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedIntItem) Next() int {
	if !it.hasNext {
		panicIfIntIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedIntItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfIntIterator) {
		return it.err
	}
	return it.base.Err()
}
