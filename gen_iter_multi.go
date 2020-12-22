package iter

import "github.com/pkg/errors"

type doubleIterator struct {
	lhs, rhs Iterator
	inRHS    bool
}

func (it *doubleIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleIterator) Next() interface{} {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperIterator combines all iterators to one.
func SuperIterator(itemList ...Iterator) Iterator {
	var super = EmptyIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Comparer is a strategy to compare two types.
type Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs interface{}) bool
}

// Compare is a shortcut implementation
// of Comparer based on a function.
type Compare func(lhs, rhs interface{}) bool

// IsLess is true if lhs is less than rhs.
func (c Compare) IsLess(lhs, rhs interface{}) bool { return c(lhs, rhs) }

// AlwaysLess is an implementation of Comparer returning always true.
var AlwaysLess = Compare(func(_, _ interface{}) bool { return true })

type priorityIterator struct {
	lhs, rhs preparedItem
	comparer Comparer
}

func (it *priorityIterator) HasNext() bool {
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

func (it *priorityIterator) Next() interface{} {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfIteratorError(
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

func (it priorityIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorIterator(comparer Comparer, itemList ...Iterator) Iterator {
	if comparer == nil {
		comparer = AlwaysLess
	}

	var prior = EmptyIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityIterator{
			lhs:      preparedItem{base: itemList[i]},
			rhs:      preparedItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
