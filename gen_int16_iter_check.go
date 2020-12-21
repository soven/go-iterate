package iter

import "github.com/pkg/errors"

// Int16Checker is an object checking an item type of int16
// for some condition.
type Int16Checker interface {
	// Check should check an item type of int16 for some condition.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Check(int16) (bool, error)
}

// Int16Check is a shortcut implementation
// of Int16Checker based on a function.
type Int16Check func(int16) (bool, error)

// Check checks an item type of int16 for some condition.
// It returns EndOfInt16Iterator to stop iteration.
func (ch Int16Check) Check(item int16) (bool, error) { return ch(item) }

var (
	// AlwaysInt16CheckTrue always returns true and empty error.
	AlwaysInt16CheckTrue Int16Checker = Int16Check(
		func(item int16) (bool, error) { return true, nil })
	// AlwaysInt16CheckFalse always returns false and empty error.
	AlwaysInt16CheckFalse Int16Checker = Int16Check(
		func(item int16) (bool, error) { return false, nil })
)

// NotInt16 do an inversion for checker result.
// It is returns AlwaysInt16CheckTrue if checker is nil.
func NotInt16(checker Int16Checker) Int16Checker {
	if checker == nil {
		return AlwaysInt16CheckTrue
	}
	return Int16Check(func(item int16) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andInt16 struct {
	lhs, rhs Int16Checker
}

func (a andInt16) Check(item int16) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllInt16 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllInt16(checkers ...Int16Checker) Int16Checker {
	var all = AlwaysInt16CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andInt16{checkers[i], all}
	}
	return all
}

type orInt16 struct {
	lhs, rhs Int16Checker
}

func (o orInt16) Check(item int16) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyInt16 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyInt16(checkers ...Int16Checker) Int16Checker {
	var any = AlwaysInt16CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orInt16{checkers[i], any}
	}
	return any
}

// FilteringInt16Iterator does iteration with
// filtering by previously set checker.
type FilteringInt16Iterator struct {
	preparedInt16Item
	filter Int16Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Int16Filtering(items Int16Iterator, filters ...Int16Checker) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &FilteringInt16Iterator{preparedInt16Item{base: items}, AllInt16(filters...)}
}

// DoingUntilInt16Iterator does iteration
// until previously set checker is passed.
type DoingUntilInt16Iterator struct {
	preparedInt16Item
	until Int16Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfInt16Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int16DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int16DoingUntil(items Int16Iterator, untilList ...Int16Checker) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	var until Int16Checker
	if len(untilList) > 0 {
		until = AllInt16(untilList...)
	} else {
		until = AlwaysInt16CheckFalse
	}
	return &DoingUntilInt16Iterator{preparedInt16Item{base: items}, until}
}

// Int16SkipUntil sets until conditions to skip few items.
func Int16SkipUntil(items Int16Iterator, untilList ...Int16Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int16Discard(Int16DoingUntil(items, untilList...))
}

// Int16EnumChecker is an object checking an item type of int16
// and its ordering number in for some condition.
type Int16EnumChecker interface {
	// Check checks an item type of int16 and its ordering number for some condition.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Check(int, int16) (bool, error)
}

// Int16EnumCheck is a shortcut implementation
// of Int16EnumChecker based on a function.
type Int16EnumCheck func(int, int16) (bool, error)

// Check checks an item type of int16 and its ordering number for some condition.
// It returns EndOfInt16Iterator to stop iteration.
func (ch Int16EnumCheck) Check(n int, item int16) (bool, error) { return ch(n, item) }

type enumFromInt16Checker struct {
	Int16Checker
}

func (ch enumFromInt16Checker) Check(_ int, item int16) (bool, error) {
	return ch.Int16Checker.Check(item)
}

// EnumFromInt16Checker adapts checker type of Int16Checker
// to the interface Int16EnumChecker.
// If checker is nil it is return based on AlwaysInt16CheckFalse enum checker.
func EnumFromInt16Checker(checker Int16Checker) Int16EnumChecker {
	if checker == nil {
		checker = AlwaysInt16CheckFalse
	}
	return &enumFromInt16Checker{checker}
}

var (
	// AlwaysInt16EnumCheckTrue always returns true and empty error.
	AlwaysInt16EnumCheckTrue = EnumFromInt16Checker(
		AlwaysInt16CheckTrue)
	// AlwaysInt16EnumCheckFalse always returns false and empty error.
	AlwaysInt16EnumCheckFalse = EnumFromInt16Checker(
		AlwaysInt16CheckFalse)
)

// EnumNotInt16 do an inversion for checker result.
// It is returns AlwaysInt16EnumCheckTrue if checker is nil.
func EnumNotInt16(checker Int16EnumChecker) Int16EnumChecker {
	if checker == nil {
		return AlwaysInt16EnumCheckTrue
	}
	return Int16EnumCheck(func(n int, item int16) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndInt16 struct {
	lhs, rhs Int16EnumChecker
}

func (a enumAndInt16) Check(n int, item int16) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllInt16 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllInt16(checkers ...Int16EnumChecker) Int16EnumChecker {
	var all = AlwaysInt16EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndInt16{checkers[i], all}
	}
	return all
}

type enumOrInt16 struct {
	lhs, rhs Int16EnumChecker
}

func (o enumOrInt16) Check(n int, item int16) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyInt16 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyInt16(checkers ...Int16EnumChecker) Int16EnumChecker {
	var any = AlwaysInt16EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrInt16{checkers[i], any}
	}
	return any
}

// EnumFilteringInt16Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringInt16Iterator struct {
	preparedInt16Item
	filter Int16EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Int16EnumFiltering(items Int16Iterator, filters ...Int16EnumChecker) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &EnumFilteringInt16Iterator{preparedInt16Item{base: items}, EnumAllInt16(filters...), 0}
}

// EnumDoingUntilInt16Iterator does iteration
// until previously set checker is passed.
type EnumDoingUntilInt16Iterator struct {
	preparedInt16Item
	until Int16EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfInt16Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int16EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int16EnumDoingUntil(items Int16Iterator, untilList ...Int16EnumChecker) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	var until Int16EnumChecker
	if len(untilList) > 0 {
		until = EnumAllInt16(untilList...)
	} else {
		until = AlwaysInt16EnumCheckFalse
	}
	return &EnumDoingUntilInt16Iterator{preparedInt16Item{base: items}, until, 0}
}

// Int16EnumSkipUntil sets until conditions to skip few items.
func Int16EnumSkipUntil(items Int16Iterator, untilList ...Int16EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int16Discard(Int16EnumDoingUntil(items, untilList...))
}

// Int16GettingBatch returns the next batch from items.
func Int16GettingBatch(items Int16Iterator, batchSize int) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	if batchSize == 0 {
		return items
	}

	return Int16EnumDoingUntil(items, Int16EnumCheck(func(n int, item int16) (bool, error) {
		return n == batchSize-1, nil
	}))
}
