package resembled

import "github.com/pkg/errors"

// EndOfPrefixIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfPrefixIterator = errors.New("end of Type iterator")

func isEndOfPrefixIterator(err error) bool {
	return errors.Is(err, EndOfPrefixIterator)
}

func wrapIfNotEndOfPrefixIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfPrefixIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfPrefixIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedPrefixItem struct {
	base    PrefixIterator
	hasNext bool
	next    Type
	err     error
}

func (it *preparedPrefixItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedPrefixItem) Next() Type {
	if !it.hasNext {
		panicIfPrefixIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = Zero
	return next
}

func (it preparedPrefixItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfPrefixIterator) {
		return it.err
	}
	return it.base.Err()
}
