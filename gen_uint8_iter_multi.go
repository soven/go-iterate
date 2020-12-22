package iter

import "github.com/pkg/errors"

type doubleUint8Iterator struct {
	lhs, rhs Uint8Iterator
	inRHS    bool
}

func (it *doubleUint8Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint8Iterator) Next() uint8 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint8Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUint8Iterator combines all iterators to one.
func SuperUint8Iterator(itemList ...Uint8Iterator) Uint8Iterator {
	var super = EmptyUint8Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint8Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Uint8Comparer is a strategy to compare two types.
type Uint8Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs uint8) bool
}

// Uint8Compare is a shortcut implementation
// of Uint8Comparer based on a function.
type Uint8Compare func(lhs, rhs uint8) bool

// IsLess is true if lhs is less than rhs.
func (c Uint8Compare) IsLess(lhs, rhs uint8) bool { return c(lhs, rhs) }

// Uint8AlwaysLess is an implementation of Uint8Comparer returning always true.
var Uint8AlwaysLess = Uint8Compare(func(_, _ uint8) bool { return true })

type priorityUint8Iterator struct {
	lhs, rhs preparedUint8Item
	comparer Uint8Comparer
}

func (it *priorityUint8Iterator) HasNext() bool {
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

func (it *priorityUint8Iterator) Next() uint8 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUint8IteratorError(
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

func (it priorityUint8Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUint8Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUint8Iterator(comparer Uint8Comparer, itemList ...Uint8Iterator) Uint8Iterator {
	if comparer == nil {
		comparer = Uint8AlwaysLess
	}

	var prior = EmptyUint8Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUint8Iterator{
			lhs:      preparedUint8Item{base: itemList[i]},
			rhs:      preparedUint8Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
