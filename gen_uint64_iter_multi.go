package iter

import "github.com/pkg/errors"

type doubleUint64Iterator struct {
	lhs, rhs Uint64Iterator
	inRHS    bool
}

func (it *doubleUint64Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint64Iterator) Next() uint64 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint64Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUint64Iterator combines all iterators to one.
func SuperUint64Iterator(itemList ...Uint64Iterator) Uint64Iterator {
	var super = EmptyUint64Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint64Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Uint64Comparer is a strategy to compare two types.
type Uint64Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs uint64) bool
}

// Uint64Compare is a shortcut implementation
// of Uint64Comparer based on a function.
type Uint64Compare func(lhs, rhs uint64) bool

// IsLess is true if lhs is less than rhs.
func (c Uint64Compare) IsLess(lhs, rhs uint64) bool { return c(lhs, rhs) }

// Uint64AlwaysLess is an implementation of Uint64Comparer returning always true.
var Uint64AlwaysLess = Uint64Compare(func(_, _ uint64) bool { return true })

type priorityUint64Iterator struct {
	lhs, rhs preparedUint64Item
	comparer Uint64Comparer
}

func (it *priorityUint64Iterator) HasNext() bool {
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

func (it *priorityUint64Iterator) Next() uint64 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUint64IteratorError(
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

func (it priorityUint64Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUint64Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUint64Iterator(comparer Uint64Comparer, itemList ...Uint64Iterator) Uint64Iterator {
	if comparer == nil {
		comparer = Uint64AlwaysLess
	}

	var prior = EmptyUint64Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUint64Iterator{
			lhs:      preparedUint64Item{base: itemList[i]},
			rhs:      preparedUint64Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
