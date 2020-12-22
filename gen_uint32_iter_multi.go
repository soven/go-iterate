package iter

import "github.com/pkg/errors"

type doubleUint32Iterator struct {
	lhs, rhs Uint32Iterator
	inRHS    bool
}

func (it *doubleUint32Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint32Iterator) Next() uint32 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint32Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUint32Iterator combines all iterators to one.
func SuperUint32Iterator(itemList ...Uint32Iterator) Uint32Iterator {
	var super = EmptyUint32Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint32Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Uint32Comparer is a strategy to compare two types.
type Uint32Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs uint32) bool
}

// Uint32Compare is a shortcut implementation
// of Uint32Comparer based on a function.
type Uint32Compare func(lhs, rhs uint32) bool

// IsLess is true if lhs is less than rhs.
func (c Uint32Compare) IsLess(lhs, rhs uint32) bool { return c(lhs, rhs) }

// Uint32AlwaysLess is an implementation of Uint32Comparer returning always true.
var Uint32AlwaysLess = Uint32Compare(func(_, _ uint32) bool { return true })

type priorityUint32Iterator struct {
	lhs, rhs preparedUint32Item
	comparer Uint32Comparer
}

func (it *priorityUint32Iterator) HasNext() bool {
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

func (it *priorityUint32Iterator) Next() uint32 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUint32IteratorError(
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

func (it priorityUint32Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUint32Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUint32Iterator(comparer Uint32Comparer, itemList ...Uint32Iterator) Uint32Iterator {
	if comparer == nil {
		comparer = Uint32AlwaysLess
	}

	var prior = EmptyUint32Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUint32Iterator{
			lhs:      preparedUint32Item{base: itemList[i]},
			rhs:      preparedUint32Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
