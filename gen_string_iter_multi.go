package iter

import "github.com/pkg/errors"

type doubleStringIterator struct {
	lhs, rhs StringIterator
	inRHS    bool
}

func (it *doubleStringIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleStringIterator) Next() string {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleStringIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperStringIterator combines all iterators to one.
func SuperStringIterator(itemList ...StringIterator) StringIterator {
	var super = EmptyStringIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleStringIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// StringComparer is a strategy to compare two types.
type StringComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs string) bool
}

// StringCompare is a shortcut implementation
// of StringComparer based on a function.
type StringCompare func(lhs, rhs string) bool

// IsLess is true if lhs is less than rhs.
func (c StringCompare) IsLess(lhs, rhs string) bool { return c(lhs, rhs) }

// StringAlwaysLess is an implementation of StringComparer returning always true.
var StringAlwaysLess = StringCompare(func(_, _ string) bool { return true })

type priorityStringIterator struct {
	lhs, rhs preparedStringItem
	comparer StringComparer
}

func (it *priorityStringIterator) HasNext() bool {
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

func (it *priorityStringIterator) Next() string {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfStringIteratorError(
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

func (it priorityStringIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorStringIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorStringIterator(comparer StringComparer, itemList ...StringIterator) StringIterator {
	if comparer == nil {
		comparer = StringAlwaysLess
	}

	var prior = EmptyStringIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityStringIterator{
			lhs:      preparedStringItem{base: itemList[i]},
			rhs:      preparedStringItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
