package iter

import "github.com/pkg/errors"

// EndOfStringIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfStringIterator = errors.New("end of string iterator")

func isEndOfStringIterator(err error) bool {
	return errors.Is(err, EndOfStringIterator)
}

func wrapIfNotEndOfStringIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfStringIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfStringIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedStringItem struct {
	base    StringIterator
	hasNext bool
	next    string
	err     error
}

func (it *preparedStringItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedStringItem) Next() string {
	if !it.hasNext {
		panicIfStringIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = ""
	return next
}

func (it preparedStringItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfStringIterator) {
		return it.err
	}
	return it.base.Err()
}
