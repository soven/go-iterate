package iter

import "github.com/pkg/errors"

// EndOfUint16Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfUint16Iterator = errors.New("end of uint16 iterator")

func isEndOfUint16Iterator(err error) bool {
	return errors.Is(err, EndOfUint16Iterator)
}

func wrapIfNotEndOfUint16Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfUint16Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfUint16IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedUint16Item struct {
	base    Uint16Iterator
	hasNext bool
	next    uint16
	err     error
}

func (it *preparedUint16Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedUint16Item) Next() uint16 {
	if !it.hasNext {
		panicIfUint16IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedUint16Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfUint16Iterator) {
		return it.err
	}
	return it.base.Err()
}
