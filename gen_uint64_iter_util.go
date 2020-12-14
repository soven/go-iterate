// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

var EndOfUint64Iterator = errors.New("end of uint64 iterator")

func isEndOfUint64Iterator(err error) bool {
	return errors.Is(err, EndOfUint64Iterator)
}

func wrapIfNotEndOfUint64Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfUint64Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfUint64IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedUint64Item struct {
	base    Uint64Iterator
	hasNext bool
	next    uint64
	err     error
}

func (it *preparedUint64Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedUint64Item) Next() uint64 {
	if !it.hasNext {
		panicIfUint64IteratorError(errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedUint64Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfUint64Iterator) {
		return it.err
	}
	return it.base.Err()
}
