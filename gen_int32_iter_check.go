package iter

import "github.com/pkg/errors"

// Int32Checker is an object checking an item type of int32
// for some condition.
type Int32Checker interface {
	// Check should check an item type of int32 for some condition.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Check(int32) (bool, error)
}

// Int32Check is a shortcut implementation
// of Int32Checker based on a function.
type Int32Check func(int32) (bool, error)

// Check checks an item type of int32 for some condition.
// It returns EndOfInt32Iterator to stop iteration.
func (ch Int32Check) Check(item int32) (bool, error) { return ch(item) }

var (
	// AlwaysInt32CheckTrue always returns true and empty error.
	AlwaysInt32CheckTrue Int32Checker = Int32Check(
		func(item int32) (bool, error) { return true, nil })
	// AlwaysInt32CheckFalse always returns false and empty error.
	AlwaysInt32CheckFalse Int32Checker = Int32Check(
		func(item int32) (bool, error) { return false, nil })
)

// NotInt32 do an inversion for checker result.
// It is returns AlwaysInt32CheckTrue if checker is nil.
func NotInt32(checker Int32Checker) Int32Checker {
	if checker == nil {
		return AlwaysInt32CheckTrue
	}
	return Int32Check(func(item int32) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andInt32 struct {
	lhs, rhs Int32Checker
}

func (a andInt32) Check(item int32) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllInt32 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllInt32(checkers ...Int32Checker) Int32Checker {
	var all = AlwaysInt32CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andInt32{checkers[i], all}
	}
	return all
}

type orInt32 struct {
	lhs, rhs Int32Checker
}

func (o orInt32) Check(item int32) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyInt32 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyInt32(checkers ...Int32Checker) Int32Checker {
	var any = AlwaysInt32CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orInt32{checkers[i], any}
	}
	return any
}

// FilteringInt32Iterator does iteration with
// filtering by previously set checker.
type FilteringInt32Iterator struct {
	preparedInt32Item
	filter Int32Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
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

// Int32Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Int32Filtering(items Int32Iterator, filters ...Int32Checker) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &FilteringInt32Iterator{preparedInt32Item{base: items}, AllInt32(filters...)}
}

// DoingUntilInt32Iterator does iteration
// until previously set checker is passed.
type DoingUntilInt32Iterator struct {
	preparedInt32Item
	until Int32Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfInt32Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int32DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int32DoingUntil(items Int32Iterator, untilList ...Int32Checker) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	var until Int32Checker
	if len(untilList) > 0 {
		until = AllInt32(untilList...)
	} else {
		until = AlwaysInt32CheckFalse
	}
	return &DoingUntilInt32Iterator{preparedInt32Item{base: items}, until}
}

// Int32SkipUntil sets until conditions to skip few items.
func Int32SkipUntil(items Int32Iterator, untilList ...Int32Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int32Discard(Int32DoingUntil(items, untilList...))
}

// Int32EnumChecker is an object checking an item type of int32
// and its ordering number in for some condition.
type Int32EnumChecker interface {
	// Check checks an item type of int32 and its ordering number for some condition.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Check(int, int32) (bool, error)
}

// Int32EnumCheck is a shortcut implementation
// of Int32EnumChecker based on a function.
type Int32EnumCheck func(int, int32) (bool, error)

// Check checks an item type of int32 and its ordering number for some condition.
// It returns EndOfInt32Iterator to stop iteration.
func (ch Int32EnumCheck) Check(n int, item int32) (bool, error) { return ch(n, item) }

type enumFromInt32Checker struct {
	Int32Checker
}

func (ch enumFromInt32Checker) Check(_ int, item int32) (bool, error) {
	return ch.Int32Checker.Check(item)
}

// EnumFromInt32Checker adapts checker type of Int32Checker
// to the interface Int32EnumChecker.
// If checker is nil it is return based on AlwaysInt32CheckFalse enum checker.
func EnumFromInt32Checker(checker Int32Checker) Int32EnumChecker {
	if checker == nil {
		checker = AlwaysInt32CheckFalse
	}
	return &enumFromInt32Checker{checker}
}

var (
	// AlwaysInt32EnumCheckTrue always returns true and empty error.
	AlwaysInt32EnumCheckTrue = EnumFromInt32Checker(
		AlwaysInt32CheckTrue)
	// AlwaysInt32EnumCheckFalse always returns false and empty error.
	AlwaysInt32EnumCheckFalse = EnumFromInt32Checker(
		AlwaysInt32CheckFalse)
)

// EnumNotInt32 do an inversion for checker result.
// It is returns AlwaysInt32EnumCheckTrue if checker is nil.
func EnumNotInt32(checker Int32EnumChecker) Int32EnumChecker {
	if checker == nil {
		return AlwaysInt32EnumCheckTrue
	}
	return Int32EnumCheck(func(n int, item int32) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndInt32 struct {
	lhs, rhs Int32EnumChecker
}

func (a enumAndInt32) Check(n int, item int32) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllInt32 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllInt32(checkers ...Int32EnumChecker) Int32EnumChecker {
	var all = AlwaysInt32EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndInt32{checkers[i], all}
	}
	return all
}

type enumOrInt32 struct {
	lhs, rhs Int32EnumChecker
}

func (o enumOrInt32) Check(n int, item int32) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyInt32 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyInt32(checkers ...Int32EnumChecker) Int32EnumChecker {
	var any = AlwaysInt32EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrInt32{checkers[i], any}
	}
	return any
}

// EnumFilteringInt32Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringInt32Iterator struct {
	preparedInt32Item
	filter Int32EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
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

// Int32EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Int32EnumFiltering(items Int32Iterator, filters ...Int32EnumChecker) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &EnumFilteringInt32Iterator{preparedInt32Item{base: items}, EnumAllInt32(filters...), 0}
}

// EnumDoingUntilInt32Iterator does iteration
// until previously set checker is passed.
type EnumDoingUntilInt32Iterator struct {
	preparedInt32Item
	until Int32EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfInt32Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int32EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int32EnumDoingUntil(items Int32Iterator, untilList ...Int32EnumChecker) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	var until Int32EnumChecker
	if len(untilList) > 0 {
		until = EnumAllInt32(untilList...)
	} else {
		until = AlwaysInt32EnumCheckFalse
	}
	return &EnumDoingUntilInt32Iterator{preparedInt32Item{base: items}, until, 0}
}

// Int32EnumSkipUntil sets until conditions to skip few items.
func Int32EnumSkipUntil(items Int32Iterator, untilList ...Int32EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int32Discard(Int32EnumDoingUntil(items, untilList...))
}

// Int32GettingBatch returns the next batch from items.
func Int32GettingBatch(items Int32Iterator, batchSize int) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	if batchSize == 0 {
		return items
	}

	return Int32EnumDoingUntil(items, Int32EnumCheck(func(n int, item int32) (bool, error) {
		return n == batchSize-1, nil
	}))
}
