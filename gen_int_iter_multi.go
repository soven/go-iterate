package iter

import "github.com/pkg/errors"

type doubleIntIterator struct {
	lhs, rhs IntIterator
	inRHS    bool
}

func (it *doubleIntIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleIntIterator) Next() int {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleIntIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperIntIterator combines all iterators to one.
func SuperIntIterator(itemList ...IntIterator) IntIterator {
	var super = EmptyIntIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleIntIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// IntComparer is a strategy to compare two types.
type IntComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs int) bool
}

// IntCompare is a shortcut implementation
// of IntComparer based on a function.
type IntCompare func(lhs, rhs int) bool

// IsLess is true if lhs is less than rhs.
func (c IntCompare) IsLess(lhs, rhs int) bool { return c(lhs, rhs) }

// IntAlwaysLess is an implementation of IntComparer returning always true.
var IntAlwaysLess = IntCompare(func(_, _ int) bool { return true })

type priorityIntIterator struct {
	lhs, rhs preparedIntItem
	comparer IntComparer
}

func (it *priorityIntIterator) HasNext() bool {
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

func (it *priorityIntIterator) Next() int {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfIntIteratorError(
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

func (it priorityIntIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorIntIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorIntIterator(comparer IntComparer, itemList ...IntIterator) IntIterator {
	if comparer == nil {
		comparer = IntAlwaysLess
	}

	var prior = EmptyIntIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityIntIterator{
			lhs:      preparedIntItem{base: itemList[i]},
			rhs:      preparedIntItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
