package iter

import "github.com/pkg/errors"

// EndOfRuneIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfRuneIterator = errors.New("end of rune iterator")

func isEndOfRuneIterator(err error) bool {
	return errors.Is(err, EndOfRuneIterator)
}

func wrapIfNotEndOfRuneIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfRuneIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfRuneIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedRuneItem struct {
	base    RuneIterator
	hasNext bool
	next    rune
	err     error
}

func (it *preparedRuneItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedRuneItem) Next() rune {
	if !it.hasNext {
		panicIfRuneIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedRuneItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfRuneIterator) {
		return it.err
	}
	return it.base.Err()
}
