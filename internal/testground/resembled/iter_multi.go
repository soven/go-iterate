package resembled

import "github.com/pkg/errors"

type doublePrefixIterator struct {
	lhs, rhs PrefixIterator
	inRHS    bool
}

func (it *doublePrefixIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doublePrefixIterator) Next() Type {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doublePrefixIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperPrefixIterator combines all iterators to one.
func SuperPrefixIterator(itemList ...PrefixIterator) PrefixIterator {
	var super = EmptyPrefixIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doublePrefixIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// PrefixEnumComparer is a strategy to compare two types.
type PrefixComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs Type) bool
}

// PrefixCompare is a shortcut implementation
// of PrefixEnumComparer based on a function.
type PrefixCompare func(lhs, rhs Type) bool

// IsLess is true if lhs is less than rhs.
func (c PrefixCompare) IsLess(lhs, rhs Type) bool { return c(lhs, rhs) }

// EnumPrefixAlwaysLess is an implementation of PrefixEnumComparer returning always true.
var PrefixAlwaysLess PrefixComparer = PrefixCompare(func(_, _ Type) bool { return true })

type priorityPrefixIterator struct {
	lhs, rhs preparedPrefixItem
	comparer PrefixComparer
}

func (it *priorityPrefixIterator) HasNext() bool {
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

func (it *priorityPrefixIterator) Next() Type {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfPrefixIteratorError(
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

func (it priorityPrefixIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorPrefixIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorPrefixIterator(comparer PrefixComparer, itemList ...PrefixIterator) PrefixIterator {
	if comparer == nil {
		comparer = PrefixAlwaysLess
	}

	var prior = EmptyPrefixIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityPrefixIterator{
			lhs:      preparedPrefixItem{base: itemList[i]},
			rhs:      preparedPrefixItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// PrefixEnumComparer is a strategy to compare two types and their order numbers.
type PrefixEnumComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(nLHS int, lhs Type, nRHS int, rhs Type) bool
}

// PrefixEnumCompare is a shortcut implementation
// of PrefixEnumComparer based on a function.
type PrefixEnumCompare func(nLHS int, lhs Type, nRHS int, rhs Type) bool

// IsLess is true if lhs is less than rhs.
func (c PrefixEnumCompare) IsLess(nLHS int, lhs Type, nRHS int, rhs Type) bool {
	return c(nLHS, lhs, nRHS, rhs)
}

// EnumPrefixAlwaysLess is an implementation of PrefixEnumComparer returning always true.
var EnumPrefixAlwaysLess PrefixEnumComparer = PrefixEnumCompare(
	func(_ int, _ Type, _ int, _ Type) bool { return true })

type priorityPrefixEnumIterator struct {
	lhs, rhs           preparedPrefixItem
	countLHS, countRHS int
	comparer           PrefixEnumComparer
}

func (it *priorityPrefixEnumIterator) HasNext() bool {
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

func (it *priorityPrefixEnumIterator) Next() Type {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfPrefixIteratorError(
			errors.New("no next"), "priority enum: next")
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
	if it.comparer.IsLess(it.countLHS, lhsNext, it.countRHS, rhsNext) {
		// remember rhsNext
		it.rhs.hasNext = true
		it.rhs.next = rhsNext
		it.countLHS++
		return lhsNext
	}

	// rhsNext is less than or equal to lhsNext.
	// remember lhsNext
	it.lhs.hasNext = true
	it.lhs.next = lhsNext
	it.countRHS++
	return rhsNext
}

func (it priorityPrefixEnumIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorPrefixEnumIterator compare one by one items and their ordering numbers fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorPrefixEnumIterator(comparer PrefixEnumComparer, itemList ...PrefixIterator) PrefixIterator {
	if comparer == nil {
		comparer = EnumPrefixAlwaysLess
	}

	var prior = EmptyPrefixIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityPrefixEnumIterator{
			lhs:      preparedPrefixItem{base: itemList[i]},
			rhs:      preparedPrefixItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}
