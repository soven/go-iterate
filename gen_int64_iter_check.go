// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

type Int64Checker interface {
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Check(int64) (bool, error)
}

type Int64Check func(int64) (bool, error)

func (ch Int64Check) Check(item int64) (bool, error) { return ch(item) }

var (
	AlwaysInt64CheckTrue  = Int64Check(func(item int64) (bool, error) { return true, nil })
	AlwaysInt64CheckFalse = Int64Check(func(item int64) (bool, error) { return false, nil })
)

func NotInt64(checker Int64Checker) Int64Checker {
	if checker == nil {
		return AlwaysInt64CheckTrue
	}
	return Int64Check(func(item int64) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andInt64 struct {
	lhs, rhs Int64Checker
}

func (a andInt64) Check(item int64) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

func AllInt64(checkers ...Int64Checker) Int64Checker {
	var all Int64Checker = AlwaysInt64CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andInt64{checkers[i], all}
	}
	return all
}

type orInt64 struct {
	lhs, rhs Int64Checker
}

func (o orInt64) Check(item int64) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

func AnyInt64(checkers ...Int64Checker) Int64Checker {
	var any Int64Checker = AlwaysInt64CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orInt64{checkers[i], any}
	}
	return any
}

type FilteringInt64Iterator struct {
	preparedInt64Item
	filter Int64Checker
}

func (it *FilteringInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Int64Filtering(items Int64Iterator, filters ...Int64Checker) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &FilteringInt64Iterator{preparedInt64Item{base: items}, AllInt64(filters...)}
}

func Int64Filter(items Int64Iterator, checker ...Int64Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int64Discard(Int64Filtering(items, checker...))
}

type Int64EnumChecker interface {
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Check(int, int64) (bool, error)
}

type Int64EnumCheck func(int, int64) (bool, error)

func (ch Int64EnumCheck) Check(n int, item int64) (bool, error) { return ch(n, item) }

type enumFromInt64Checker struct {
	Int64Checker
}

func (ch enumFromInt64Checker) Check(_ int, item int64) (bool, error) {
	return ch.Int64Checker.Check(item)
}

func EnumFromInt64Checker(checker Int64Checker) Int64EnumChecker {
	return &enumFromInt64Checker{checker}
}

var (
	AlwaysInt64EnumCheckTrue  = EnumFromInt64Checker(AlwaysInt64CheckTrue)
	AlwaysInt64EnumCheckFalse = EnumFromInt64Checker(AlwaysInt64CheckFalse)
)

func EnumNotInt64(checker Int64EnumChecker) Int64EnumChecker {
	if checker == nil {
		return AlwaysInt64EnumCheckTrue
	}
	return Int64EnumCheck(func(n int, item int64) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndInt64 struct {
	lhs, rhs Int64EnumChecker
}

func (a enumAndInt64) Check(n int, item int64) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

func EnumAllInt64(checkers ...Int64EnumChecker) Int64EnumChecker {
	var all Int64EnumChecker = AlwaysInt64EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndInt64{checkers[i], all}
	}
	return all
}

type enumOrInt64 struct {
	lhs, rhs Int64EnumChecker
}

func (o enumOrInt64) Check(n int, item int64) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

func EnumAnyInt64(checkers ...Int64EnumChecker) Int64EnumChecker {
	var any Int64EnumChecker = AlwaysInt64EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrInt64{checkers[i], any}
	}
	return any
}

type EnumFilteringInt64Iterator struct {
	preparedInt64Item
	filter Int64EnumChecker
	count  int
}

func (it *EnumFilteringInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Int64EnumFiltering(items Int64Iterator, filters ...Int64EnumChecker) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &EnumFilteringInt64Iterator{preparedInt64Item{base: items}, EnumAllInt64(filters...), 0}
}

type DoingUntilInt64Iterator struct {
	preparedInt64Item
	until Int64Checker
}

func (it *DoingUntilInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfInt64Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int64DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int64DoingUntil(items Int64Iterator, untilList ...Int64Checker) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &DoingUntilInt64Iterator{preparedInt64Item{base: items}, AllInt64(untilList...)}
}

func Int64DoUntil(items Int64Iterator, untilList ...Int64Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int64Discard(Int64DoingUntil(items, untilList...))
}

// Int64GettingBatch returns the next batch from items.
func Int64GettingBatch(items Int64Iterator, batchSize int) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	if batchSize == 0 {
		return items
	}

	size := 0
	return Int64DoingUntil(items, Int64Check(func(item int64) (bool, error) {
		size++
		return size >= batchSize, nil
	}))
}
