package iter

import "github.com/pkg/errors"

type doubleInt64Iterator struct {
	lhs, rhs Int64Iterator
	inRHS    bool
}

func (it *doubleInt64Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleInt64Iterator) Next() int64 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleInt64Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperInt64Iterator combines all iterators to one.
func SuperInt64Iterator(itemList ...Int64Iterator) Int64Iterator {
	var super = EmptyInt64Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleInt64Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Int64Comparer is a strategy to compare two types.
type Int64Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs int64) bool
}

// Int64Compare is a shortcut implementation
// of Int64Comparer based on a function.
type Int64Compare func(lhs, rhs int64) bool

// IsLess is true if lhs is less than rhs.
func (c Int64Compare) IsLess(lhs, rhs int64) bool { return c(lhs, rhs) }

// Int64AlwaysLess is an implementation of Int64Comparer returning always true.
var Int64AlwaysLess = Int64Compare(func(_, _ int64) bool { return true })

type priorityInt64Iterator struct {
	lhs, rhs preparedInt64Item
	comparer Int64Comparer
}

func (it *priorityInt64Iterator) HasNext() bool {
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

func (it *priorityInt64Iterator) Next() int64 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfInt64IteratorError(
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

func (it priorityInt64Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorInt64Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorInt64Iterator(comparer Int64Comparer, itemList ...Int64Iterator) Int64Iterator {
	if comparer == nil {
		comparer = Int64AlwaysLess
	}

	var prior = EmptyInt64Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityInt64Iterator{
			lhs:      preparedInt64Item{base: itemList[i]},
			rhs:      preparedInt64Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
