package iter

import "github.com/pkg/errors"

// Uint8Checker is an object checking an item type of uint8
// for some condition.
type Uint8Checker interface {
	// Check should check an item type of uint8 for some condition.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Check(uint8) (bool, error)
}

// Uint8Check is a shortcut implementation
// of Uint8Checker based on a function.
type Uint8Check func(uint8) (bool, error)

// Check checks an item type of uint8 for some condition.
// It returns EndOfUint8Iterator to stop iteration.
func (ch Uint8Check) Check(item uint8) (bool, error) { return ch(item) }

var (
	// AlwaysUint8CheckTrue always returns true and empty error.
	AlwaysUint8CheckTrue Uint8Checker = Uint8Check(
		func(item uint8) (bool, error) { return true, nil })
	// AlwaysUint8CheckFalse always returns false and empty error.
	AlwaysUint8CheckFalse Uint8Checker = Uint8Check(
		func(item uint8) (bool, error) { return false, nil })
)

// NotUint8 do an inversion for checker result.
// It is returns AlwaysUint8CheckTrue if checker is nil.
func NotUint8(checker Uint8Checker) Uint8Checker {
	if checker == nil {
		return AlwaysUint8CheckTrue
	}
	return Uint8Check(func(item uint8) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andUint8 struct {
	lhs, rhs Uint8Checker
}

func (a andUint8) Check(item uint8) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllUint8 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllUint8(checkers ...Uint8Checker) Uint8Checker {
	var all = AlwaysUint8CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andUint8{checkers[i], all}
	}
	return all
}

type orUint8 struct {
	lhs, rhs Uint8Checker
}

func (o orUint8) Check(item uint8) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyUint8 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyUint8(checkers ...Uint8Checker) Uint8Checker {
	var any = AlwaysUint8CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orUint8{checkers[i], any}
	}
	return any
}

// FilteringUint8Iterator does iteration with
// filtering by previously set checker.
type FilteringUint8Iterator struct {
	preparedUint8Item
	filter Uint8Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
				err = errors.Wrap(err, "filtering iterator: check")
			}
			it.err = err
			return false
		}

		if !isFilterPassed {
			continue
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint8Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Uint8Filtering(items Uint8Iterator, filters ...Uint8Checker) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &FilteringUint8Iterator{preparedUint8Item{base: items}, AllUint8(filters...)}
}

// DoingUntilUint8Iterator does iteration
// until previously set checker is passed.
type DoingUntilUint8Iterator struct {
	preparedUint8Item
	until Uint8Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfUint8Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint8DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Uint8DoingUntil(items Uint8Iterator, untilList ...Uint8Checker) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	var until Uint8Checker
	if len(untilList) > 0 {
		until = AllUint8(untilList...)
	} else {
		until = AlwaysUint8CheckFalse
	}
	return &DoingUntilUint8Iterator{preparedUint8Item{base: items}, until}
}

// Uint8SkipUntil sets until conditions to skip few items.
func Uint8SkipUntil(items Uint8Iterator, untilList ...Uint8Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint8Discard(Uint8DoingUntil(items, untilList...))
}

// Uint8EnumChecker is an object checking an item type of uint8
// and its ordering number in for some condition.
type Uint8EnumChecker interface {
	// Check checks an item type of uint8 and its ordering number for some condition.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Check(int, uint8) (bool, error)
}

// Uint8EnumCheck is a shortcut implementation
// of Uint8EnumChecker based on a function.
type Uint8EnumCheck func(int, uint8) (bool, error)

// Check checks an item type of uint8 and its ordering number for some condition.
// It returns EndOfUint8Iterator to stop iteration.
func (ch Uint8EnumCheck) Check(n int, item uint8) (bool, error) { return ch(n, item) }

type enumFromUint8Checker struct {
	Uint8Checker
}

func (ch enumFromUint8Checker) Check(_ int, item uint8) (bool, error) {
	return ch.Uint8Checker.Check(item)
}

// EnumFromUint8Checker adapts checker type of Uint8Checker
// to the interface Uint8EnumChecker.
// If checker is nil it is return based on AlwaysUint8CheckFalse enum checker.
func EnumFromUint8Checker(checker Uint8Checker) Uint8EnumChecker {
	if checker == nil {
		checker = AlwaysUint8CheckFalse
	}
	return &enumFromUint8Checker{checker}
}

var (
	// AlwaysUint8EnumCheckTrue always returns true and empty error.
	AlwaysUint8EnumCheckTrue = EnumFromUint8Checker(
		AlwaysUint8CheckTrue)
	// AlwaysUint8EnumCheckFalse always returns false and empty error.
	AlwaysUint8EnumCheckFalse = EnumFromUint8Checker(
		AlwaysUint8CheckFalse)
)

// EnumNotUint8 do an inversion for checker result.
// It is returns AlwaysUint8EnumCheckTrue if checker is nil.
func EnumNotUint8(checker Uint8EnumChecker) Uint8EnumChecker {
	if checker == nil {
		return AlwaysUint8EnumCheckTrue
	}
	return Uint8EnumCheck(func(n int, item uint8) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndUint8 struct {
	lhs, rhs Uint8EnumChecker
}

func (a enumAndUint8) Check(n int, item uint8) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllUint8 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllUint8(checkers ...Uint8EnumChecker) Uint8EnumChecker {
	var all = AlwaysUint8EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndUint8{checkers[i], all}
	}
	return all
}

type enumOrUint8 struct {
	lhs, rhs Uint8EnumChecker
}

func (o enumOrUint8) Check(n int, item uint8) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyUint8 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyUint8(checkers ...Uint8EnumChecker) Uint8EnumChecker {
	var any = AlwaysUint8EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrUint8{checkers[i], any}
	}
	return any
}

// EnumFilteringUint8Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringUint8Iterator struct {
	preparedUint8Item
	filter Uint8EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
				err = errors.Wrap(err, "enum filtering iterator: check")
			}
			it.err = err
			return false
		}
		it.count++

		if !isFilterPassed {
			continue
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint8EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Uint8EnumFiltering(items Uint8Iterator, filters ...Uint8EnumChecker) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &EnumFilteringUint8Iterator{preparedUint8Item{base: items}, EnumAllUint8(filters...), 0}
}

// EnumDoingUntilUint8Iterator does iteration
// until previously set checker is passed.
type EnumDoingUntilUint8Iterator struct {
	preparedUint8Item
	until Uint8EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfUint8Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint8EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Uint8EnumDoingUntil(items Uint8Iterator, untilList ...Uint8EnumChecker) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	var until Uint8EnumChecker
	if len(untilList) > 0 {
		until = EnumAllUint8(untilList...)
	} else {
		until = AlwaysUint8EnumCheckFalse
	}
	return &EnumDoingUntilUint8Iterator{preparedUint8Item{base: items}, until, 0}
}

// Uint8EnumSkipUntil sets until conditions to skip few items.
func Uint8EnumSkipUntil(items Uint8Iterator, untilList ...Uint8EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint8Discard(Uint8EnumDoingUntil(items, untilList...))
}

// Uint8GettingBatch returns the next batch from items.
func Uint8GettingBatch(items Uint8Iterator, batchSize int) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	if batchSize == 0 {
		return items
	}

	return Uint8EnumDoingUntil(items, Uint8EnumCheck(func(n int, item uint8) (bool, error) {
		return n == batchSize-1, nil
	}))
}
