package iter

import "github.com/pkg/errors"

// EndOfUint8Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfUint8Iterator = errors.New("end of uint8 iterator")

func isEndOfUint8Iterator(err error) bool {
	return errors.Is(err, EndOfUint8Iterator)
}

func wrapIfNotEndOfUint8Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfUint8Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfUint8IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedUint8Item struct {
	base    Uint8Iterator
	hasNext bool
	next    uint8
	err     error
}

func (it *preparedUint8Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedUint8Item) Next() uint8 {
	if !it.hasNext {
		panicIfUint8IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedUint8Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfUint8Iterator) {
		return it.err
	}
	return it.base.Err()
}
