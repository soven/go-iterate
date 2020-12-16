// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

// Uint64Checker is an object checking an item type of uint64
// for some condition.
type Uint64Checker interface {
	// Check should check an item type of uint64 for some condition.
	// It is suggested to return EndOfUint64Iterator to stop iteration.
	Check(uint64) (bool, error)
}

// Uint64Check is a shortcut implementation
// of Uint64Checker based on a function.
type Uint64Check func(uint64) (bool, error)

// Check checks an item type of uint64 for some condition.
// It returns EndOfUint64Iterator to stop iteration.
func (ch Uint64Check) Check(item uint64) (bool, error) { return ch(item) }

var (
	// AlwaysUint64CheckTrue always returns true and empty error.
	AlwaysUint64CheckTrue Uint64Checker = Uint64Check(
		func(item uint64) (bool, error) { return true, nil })
	// AlwaysUint64CheckFalse always returns false and empty error.
	AlwaysUint64CheckFalse Uint64Checker = Uint64Check(
		func(item uint64) (bool, error) { return false, nil })
)

// NotUint64 do an inversion for checker result.
// It is returns AlwaysUint64CheckTrue if checker is nil.
func NotUint64(checker Uint64Checker) Uint64Checker {
	if checker == nil {
		return AlwaysUint64CheckTrue
	}
	return Uint64Check(func(item uint64) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andUint64 struct {
	lhs, rhs Uint64Checker
}

func (a andUint64) Check(item uint64) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllUint64 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllUint64(checkers ...Uint64Checker) Uint64Checker {
	var all = AlwaysUint64CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andUint64{checkers[i], all}
	}
	return all
}

type orUint64 struct {
	lhs, rhs Uint64Checker
}

func (o orUint64) Check(item uint64) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyUint64 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyUint64(checkers ...Uint64Checker) Uint64Checker {
	var any = AlwaysUint64CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orUint64{checkers[i], any}
	}
	return any
}

// FilteringUint64Iterator does iteration with
// filtering by previously set checker.
type FilteringUint64Iterator struct {
	preparedUint64Item
	filter Uint64Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
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

// Uint64Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Uint64Filtering(items Uint64Iterator, filters ...Uint64Checker) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &FilteringUint64Iterator{preparedUint64Item{base: items}, AllUint64(filters...)}
}

// Uint64EnumChecker is an object checking an item type of uint64
// and its ordering number in for some condition.
type Uint64EnumChecker interface {
	// Check checks an item type of uint64 and its ordering number for some condition.
	// It is suggested to return EndOfUint64Iterator to stop iteration.
	Check(int, uint64) (bool, error)
}

// Uint64EnumCheck is a shortcut implementation
// of Uint64EnumChecker based on a function.
type Uint64EnumCheck func(int, uint64) (bool, error)

// Check checks an item type of uint64 and its ordering number for some condition.
// It returns EndOfUint64Iterator to stop iteration.
func (ch Uint64EnumCheck) Check(n int, item uint64) (bool, error) { return ch(n, item) }

type enumFromUint64Checker struct {
	Uint64Checker
}

func (ch enumFromUint64Checker) Check(_ int, item uint64) (bool, error) {
	return ch.Uint64Checker.Check(item)
}

// EnumFromUint64Checker adapts checker type of Uint64Checker
// to the interface Uint64EnumChecker.
// If checker is nil it is return based on AlwaysUint64CheckFalse enum checker.
func EnumFromUint64Checker(checker Uint64Checker) Uint64EnumChecker {
	if checker == nil {
		checker = AlwaysUint64CheckFalse
	}
	return &enumFromUint64Checker{checker}
}

var (
	// AlwaysUint64EnumCheckTrue always returns true and empty error.
	AlwaysUint64EnumCheckTrue Uint64EnumChecker = EnumFromUint64Checker(
		AlwaysUint64CheckTrue)
	// AlwaysUint64EnumCheckFalse always returns false and empty error.
	AlwaysUint64EnumCheckFalse Uint64EnumChecker = EnumFromUint64Checker(
		AlwaysUint64CheckFalse)
)

// EnumNotUint64 do an inversion for checker result.
// It is returns AlwaysUint64EnumCheckTrue if checker is nil.
func EnumNotUint64(checker Uint64EnumChecker) Uint64EnumChecker {
	if checker == nil {
		return AlwaysUint64EnumCheckTrue
	}
	return Uint64EnumCheck(func(n int, item uint64) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndUint64 struct {
	lhs, rhs Uint64EnumChecker
}

func (a enumAndUint64) Check(n int, item uint64) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllUint64 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllUint64(checkers ...Uint64EnumChecker) Uint64EnumChecker {
	var all = AlwaysUint64EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndUint64{checkers[i], all}
	}
	return all
}

type enumOrUint64 struct {
	lhs, rhs Uint64EnumChecker
}

func (o enumOrUint64) Check(n int, item uint64) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyUint64 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyUint64(checkers ...Uint64EnumChecker) Uint64EnumChecker {
	var any = AlwaysUint64EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrUint64{checkers[i], any}
	}
	return any
}

// EnumFilteringUint64Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringUint64Iterator struct {
	preparedUint64Item
	filter Uint64EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
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

// Uint64EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Uint64EnumFiltering(items Uint64Iterator, filters ...Uint64EnumChecker) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &EnumFilteringUint64Iterator{preparedUint64Item{base: items}, EnumAllUint64(filters...), 0}
}

// DoingUntilUint64Iterator does iteration
// until previously set checker is passed.
type DoingUntilUint64Iterator struct {
	preparedUint64Item
	until Uint64Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfUint64Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint64DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Uint64DoingUntil(items Uint64Iterator, untilList ...Uint64Checker) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &DoingUntilUint64Iterator{preparedUint64Item{base: items}, AllUint64(untilList...)}
}

// Uint64SkipUntil sets until conditions to skip few items.
func Uint64SkipUntil(items Uint64Iterator, untilList ...Uint64Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint64Discard(Uint64DoingUntil(items, untilList...))
}

// Uint64GettingBatch returns the next batch from items.
func Uint64GettingBatch(items Uint64Iterator, batchSize int) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	if batchSize == 0 {
		return items
	}

	size := 0
	return Uint64DoingUntil(items, Uint64Check(func(item uint64) (bool, error) {
		size++
		return size >= batchSize, nil
	}))
}
