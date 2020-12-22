package iter

import "github.com/pkg/errors"

type doubleUint16Iterator struct {
	lhs, rhs Uint16Iterator
	inRHS    bool
}

func (it *doubleUint16Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint16Iterator) Next() uint16 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint16Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUint16Iterator combines all iterators to one.
func SuperUint16Iterator(itemList ...Uint16Iterator) Uint16Iterator {
	var super = EmptyUint16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint16Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Uint16Comparer is a strategy to compare two types.
type Uint16Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs uint16) bool
}

// Uint16Compare is a shortcut implementation
// of Uint16Comparer based on a function.
type Uint16Compare func(lhs, rhs uint16) bool

// IsLess is true if lhs is less than rhs.
func (c Uint16Compare) IsLess(lhs, rhs uint16) bool { return c(lhs, rhs) }

// Uint16AlwaysLess is an implementation of Uint16Comparer returning always true.
var Uint16AlwaysLess = Uint16Compare(func(_, _ uint16) bool { return true })

type priorityUint16Iterator struct {
	lhs, rhs preparedUint16Item
	comparer Uint16Comparer
}

func (it *priorityUint16Iterator) HasNext() bool {
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

func (it *priorityUint16Iterator) Next() uint16 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUint16IteratorError(
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

func (it priorityUint16Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUint16Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUint16Iterator(comparer Uint16Comparer, itemList ...Uint16Iterator) Uint16Iterator {
	if comparer == nil {
		comparer = Uint16AlwaysLess
	}

	var prior = EmptyUint16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUint16Iterator{
			lhs:      preparedUint16Item{base: itemList[i]},
			rhs:      preparedUint16Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
