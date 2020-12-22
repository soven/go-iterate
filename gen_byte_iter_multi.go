package iter

import "github.com/pkg/errors"

type doubleByteIterator struct {
	lhs, rhs ByteIterator
	inRHS    bool
}

func (it *doubleByteIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleByteIterator) Next() byte {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleByteIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperByteIterator combines all iterators to one.
func SuperByteIterator(itemList ...ByteIterator) ByteIterator {
	var super = EmptyByteIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleByteIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// ByteComparer is a strategy to compare two types.
type ByteComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs byte) bool
}

// ByteCompare is a shortcut implementation
// of ByteComparer based on a function.
type ByteCompare func(lhs, rhs byte) bool

// IsLess is true if lhs is less than rhs.
func (c ByteCompare) IsLess(lhs, rhs byte) bool { return c(lhs, rhs) }

// ByteAlwaysLess is an implementation of ByteComparer returning always true.
var ByteAlwaysLess = ByteCompare(func(_, _ byte) bool { return true })

type priorityByteIterator struct {
	lhs, rhs preparedByteItem
	comparer ByteComparer
}

func (it *priorityByteIterator) HasNext() bool {
	if it.lhs.hasNext && it.rhs.hasNext {
		return true
	}
	if !it.lhs.hasNext && it.lhs.HasNext() {
		next := it.lhs.base.Next()
		it.lhs.hasNext = true
		it.lhs.next = next
	}
	if !it.rhs.hasNext && it.rhs.HasNext() {
		next := it.rhs.base.Next()
		it.rhs.hasNext = true
		it.rhs.next = next
	}

	return it.lhs.hasNext || it.rhs.hasNext
}

func (it *priorityByteIterator) Next() byte {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfByteIteratorError(
			errors.New("no next"), "priority: next")
	}

	if !it.lhs.hasNext {
		// it.rhs.hasNext == true
		return it.rhs.Next()
	}
	if !it.rhs.hasNext {
		// it.lhs.hasNext == true
		return it.lhs.Next()
	}

	// both have next
	lhsNext := it.lhs.Next()
	rhsNext := it.rhs.Next()
	if it.comparer.IsLess(lhsNext, rhsNext) {
		// remember rhsNext
		it.rhs.hasNext = true
		it.rhs.next = rhsNext
		return lhsNext
	}

	// rhsNext is less than or equal to lhsNext.
	// remember lhsNext
	it.lhs.hasNext = true
	it.lhs.next = lhsNext
	return rhsNext
}

func (it priorityByteIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorByteIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorByteIterator(comparer ByteComparer, itemList ...ByteIterator) ByteIterator {
	if comparer == nil {
		comparer = ByteAlwaysLess
	}

	var prior = EmptyByteIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityByteIterator{
			lhs:      preparedByteItem{base: itemList[i]},
			rhs:      preparedByteItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
