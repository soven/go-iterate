package iter

import "github.com/pkg/errors"

type doubleRuneIterator struct {
	lhs, rhs RuneIterator
	inRHS    bool
}

func (it *doubleRuneIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleRuneIterator) Next() rune {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleRuneIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperRuneIterator combines all iterators to one.
func SuperRuneIterator(itemList ...RuneIterator) RuneIterator {
	var super = EmptyRuneIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleRuneIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// RuneComparer is a strategy to compare two types.
type RuneComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs rune) bool
}

// RuneCompare is a shortcut implementation
// of RuneComparer based on a function.
type RuneCompare func(lhs, rhs rune) bool

// IsLess is true if lhs is less than rhs.
func (c RuneCompare) IsLess(lhs, rhs rune) bool { return c(lhs, rhs) }

// RuneAlwaysLess is an implementation of RuneComparer returning always true.
var RuneAlwaysLess = RuneCompare(func(_, _ rune) bool { return true })

type priorityRuneIterator struct {
	lhs, rhs preparedRuneItem
	comparer RuneComparer
}

func (it *priorityRuneIterator) HasNext() bool {
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

func (it *priorityRuneIterator) Next() rune {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfRuneIteratorError(
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

func (it priorityRuneIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorRuneIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorRuneIterator(comparer RuneComparer, itemList ...RuneIterator) RuneIterator {
	if comparer == nil {
		comparer = RuneAlwaysLess
	}

	var prior = EmptyRuneIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityRuneIterator{
			lhs:      preparedRuneItem{base: itemList[i]},
			rhs:      preparedRuneItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
