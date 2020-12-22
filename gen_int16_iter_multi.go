package iter

import "github.com/pkg/errors"

type doubleInt16Iterator struct {
	lhs, rhs Int16Iterator
	inRHS    bool
}

func (it *doubleInt16Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleInt16Iterator) Next() int16 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleInt16Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperInt16Iterator combines all iterators to one.
func SuperInt16Iterator(itemList ...Int16Iterator) Int16Iterator {
	var super = EmptyInt16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleInt16Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Int16Comparer is a strategy to compare two types.
type Int16Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs int16) bool
}

// Int16Compare is a shortcut implementation
// of Int16Comparer based on a function.
type Int16Compare func(lhs, rhs int16) bool

// IsLess is true if lhs is less than rhs.
func (c Int16Compare) IsLess(lhs, rhs int16) bool { return c(lhs, rhs) }

// Int16AlwaysLess is an implementation of Int16Comparer returning always true.
var Int16AlwaysLess = Int16Compare(func(_, _ int16) bool { return true })

type priorityInt16Iterator struct {
	lhs, rhs preparedInt16Item
	comparer Int16Comparer
}

func (it *priorityInt16Iterator) HasNext() bool {
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

func (it *priorityInt16Iterator) Next() int16 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfInt16IteratorError(
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

func (it priorityInt16Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorInt16Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorInt16Iterator(comparer Int16Comparer, itemList ...Int16Iterator) Int16Iterator {
	if comparer == nil {
		comparer = Int16AlwaysLess
	}

	var prior = EmptyInt16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityInt16Iterator{
			lhs:      preparedInt16Item{base: itemList[i]},
			rhs:      preparedInt16Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
