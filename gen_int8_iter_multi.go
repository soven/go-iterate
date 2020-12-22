package iter

import "github.com/pkg/errors"

type doubleInt8Iterator struct {
	lhs, rhs Int8Iterator
	inRHS    bool
}

func (it *doubleInt8Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleInt8Iterator) Next() int8 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleInt8Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperInt8Iterator combines all iterators to one.
func SuperInt8Iterator(itemList ...Int8Iterator) Int8Iterator {
	var super = EmptyInt8Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleInt8Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Int8Comparer is a strategy to compare two types.
type Int8Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs int8) bool
}

// Int8Compare is a shortcut implementation
// of Int8Comparer based on a function.
type Int8Compare func(lhs, rhs int8) bool

// IsLess is true if lhs is less than rhs.
func (c Int8Compare) IsLess(lhs, rhs int8) bool { return c(lhs, rhs) }

// Int8AlwaysLess is an implementation of Int8Comparer returning always true.
var Int8AlwaysLess = Int8Compare(func(_, _ int8) bool { return true })

type priorityInt8Iterator struct {
	lhs, rhs preparedInt8Item
	comparer Int8Comparer
}

func (it *priorityInt8Iterator) HasNext() bool {
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

func (it *priorityInt8Iterator) Next() int8 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfInt8IteratorError(
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

func (it priorityInt8Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorInt8Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorInt8Iterator(comparer Int8Comparer, itemList ...Int8Iterator) Int8Iterator {
	if comparer == nil {
		comparer = Int8AlwaysLess
	}

	var prior = EmptyInt8Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityInt8Iterator{
			lhs:      preparedInt8Item{base: itemList[i]},
			rhs:      preparedInt8Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
