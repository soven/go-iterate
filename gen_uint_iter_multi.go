package iter

import "github.com/pkg/errors"

type doubleUintIterator struct {
	lhs, rhs UintIterator
	inRHS    bool
}

func (it *doubleUintIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUintIterator) Next() uint {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUintIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUintIterator combines all iterators to one.
func SuperUintIterator(itemList ...UintIterator) UintIterator {
	var super = EmptyUintIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUintIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// UintComparer is a strategy to compare two types.
type UintComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs uint) bool
}

// UintCompare is a shortcut implementation
// of UintComparer based on a function.
type UintCompare func(lhs, rhs uint) bool

// IsLess is true if lhs is less than rhs.
func (c UintCompare) IsLess(lhs, rhs uint) bool { return c(lhs, rhs) }

// UintAlwaysLess is an implementation of UintComparer returning always true.
var UintAlwaysLess = UintCompare(func(_, _ uint) bool { return true })

type priorityUintIterator struct {
	lhs, rhs preparedUintItem
	comparer UintComparer
}

func (it *priorityUintIterator) HasNext() bool {
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

func (it *priorityUintIterator) Next() uint {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUintIteratorError(
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

func (it priorityUintIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUintIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUintIterator(comparer UintComparer, itemList ...UintIterator) UintIterator {
	if comparer == nil {
		comparer = UintAlwaysLess
	}

	var prior = EmptyUintIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUintIterator{
			lhs:      preparedUintItem{base: itemList[i]},
			rhs:      preparedUintItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
