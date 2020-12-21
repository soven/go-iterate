package iter

import "github.com/pkg/errors"

// EndOfByteIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfByteIterator = errors.New("end of byte iterator")

func isEndOfByteIterator(err error) bool {
	return errors.Is(err, EndOfByteIterator)
}

func wrapIfNotEndOfByteIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfByteIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfByteIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedByteItem struct {
	base    ByteIterator
	hasNext bool
	next    byte
	err     error
}

func (it *preparedByteItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedByteItem) Next() byte {
	if !it.hasNext {
		panicIfByteIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedByteItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfByteIterator) {
		return it.err
	}
	return it.base.Err()
}
