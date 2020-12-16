// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

// Int8Checker is an object checking an item type of int8
// for some condition.
type Int8Checker interface {
	// Check should check an item type of int8 for some condition.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Check(int8) (bool, error)
}

// Int8Check is a shortcut implementation
// of Int8Checker based on a function.
type Int8Check func(int8) (bool, error)

// Check checks an item type of int8 for some condition.
// It returns EndOfInt8Iterator to stop iteration.
func (ch Int8Check) Check(item int8) (bool, error) { return ch(item) }

var (
	// AlwaysInt8CheckTrue always returns true and empty error.
	AlwaysInt8CheckTrue Int8Checker = Int8Check(
		func(item int8) (bool, error) { return true, nil })
	// AlwaysInt8CheckFalse always returns false and empty error.
	AlwaysInt8CheckFalse Int8Checker = Int8Check(
		func(item int8) (bool, error) { return false, nil })
)

// NotInt8 do an inversion for checker result.
// It is returns AlwaysInt8CheckTrue if checker is nil.
func NotInt8(checker Int8Checker) Int8Checker {
	if checker == nil {
		return AlwaysInt8CheckTrue
	}
	return Int8Check(func(item int8) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andInt8 struct {
	lhs, rhs Int8Checker
}

func (a andInt8) Check(item int8) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllInt8 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllInt8(checkers ...Int8Checker) Int8Checker {
	var all = AlwaysInt8CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andInt8{checkers[i], all}
	}
	return all
}

type orInt8 struct {
	lhs, rhs Int8Checker
}

func (o orInt8) Check(item int8) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyInt8 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyInt8(checkers ...Int8Checker) Int8Checker {
	var any = AlwaysInt8CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orInt8{checkers[i], any}
	}
	return any
}

// FilteringInt8Iterator does iteration with
// filtering by previously set checker.
type FilteringInt8Iterator struct {
	preparedInt8Item
	filter Int8Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Int8Filtering(items Int8Iterator, filters ...Int8Checker) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &FilteringInt8Iterator{preparedInt8Item{base: items}, AllInt8(filters...)}
}

// Int8EnumChecker is an object checking an item type of int8
// and its ordering number in for some condition.
type Int8EnumChecker interface {
	// Check checks an item type of int8 and its ordering number for some condition.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Check(int, int8) (bool, error)
}

// Int8EnumCheck is a shortcut implementation
// of Int8EnumChecker based on a function.
type Int8EnumCheck func(int, int8) (bool, error)

// Check checks an item type of int8 and its ordering number for some condition.
// It returns EndOfInt8Iterator to stop iteration.
func (ch Int8EnumCheck) Check(n int, item int8) (bool, error) { return ch(n, item) }

type enumFromInt8Checker struct {
	Int8Checker
}

func (ch enumFromInt8Checker) Check(_ int, item int8) (bool, error) {
	return ch.Int8Checker.Check(item)
}

// EnumFromInt8Checker adapts checker type of Int8Checker
// to the interface Int8EnumChecker.
// If checker is nil it is return based on AlwaysInt8CheckFalse enum checker.
func EnumFromInt8Checker(checker Int8Checker) Int8EnumChecker {
	if checker == nil {
		checker = AlwaysInt8CheckFalse
	}
	return &enumFromInt8Checker{checker}
}

var (
	// AlwaysInt8EnumCheckTrue always returns true and empty error.
	AlwaysInt8EnumCheckTrue Int8EnumChecker = EnumFromInt8Checker(
		AlwaysInt8CheckTrue)
	// AlwaysInt8EnumCheckFalse always returns false and empty error.
	AlwaysInt8EnumCheckFalse Int8EnumChecker = EnumFromInt8Checker(
		AlwaysInt8CheckFalse)
)

// EnumNotInt8 do an inversion for checker result.
// It is returns AlwaysInt8EnumCheckTrue if checker is nil.
func EnumNotInt8(checker Int8EnumChecker) Int8EnumChecker {
	if checker == nil {
		return AlwaysInt8EnumCheckTrue
	}
	return Int8EnumCheck(func(n int, item int8) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndInt8 struct {
	lhs, rhs Int8EnumChecker
}

func (a enumAndInt8) Check(n int, item int8) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllInt8 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllInt8(checkers ...Int8EnumChecker) Int8EnumChecker {
	var all = AlwaysInt8EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndInt8{checkers[i], all}
	}
	return all
}

type enumOrInt8 struct {
	lhs, rhs Int8EnumChecker
}

func (o enumOrInt8) Check(n int, item int8) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyInt8 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyInt8(checkers ...Int8EnumChecker) Int8EnumChecker {
	var any = AlwaysInt8EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrInt8{checkers[i], any}
	}
	return any
}

// EnumFilteringInt8Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringInt8Iterator struct {
	preparedInt8Item
	filter Int8EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Int8EnumFiltering(items Int8Iterator, filters ...Int8EnumChecker) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &EnumFilteringInt8Iterator{preparedInt8Item{base: items}, EnumAllInt8(filters...), 0}
}

// DoingUntilInt8Iterator does iteration
// until previously set checker is passed.
type DoingUntilInt8Iterator struct {
	preparedInt8Item
	until Int8Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfInt8Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int8DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int8DoingUntil(items Int8Iterator, untilList ...Int8Checker) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &DoingUntilInt8Iterator{preparedInt8Item{base: items}, AllInt8(untilList...)}
}

// Int8SkipUntil sets until conditions to skip few items.
func Int8SkipUntil(items Int8Iterator, untilList ...Int8Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int8Discard(Int8DoingUntil(items, untilList...))
}

// Int8GettingBatch returns the next batch from items.
func Int8GettingBatch(items Int8Iterator, batchSize int) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	if batchSize == 0 {
		return items
	}

	size := 0
	return Int8DoingUntil(items, Int8Check(func(item int8) (bool, error) {
		size++
		return size >= batchSize, nil
	}))
}
