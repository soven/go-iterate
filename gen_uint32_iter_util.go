// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

var EndOfUint32Iterator = errors.New("end of uint32 iterator")

func isEndOfUint32Iterator(err error) bool {
	return errors.Is(err, EndOfUint32Iterator)
}

func wrapIfNotEndOfUint32Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfUint32Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfUint32IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedUint32Item struct {
	base    Uint32Iterator
	hasNext bool
	next    uint32
	err     error
}

func (it *preparedUint32Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedUint32Item) Next() uint32 {
	if !it.hasNext {
		panicIfUint32IteratorError(errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedUint32Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfUint32Iterator) {
		return it.err
	}
	return it.base.Err()
}
