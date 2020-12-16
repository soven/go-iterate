// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

// RuneChecker is an object checking an item type of rune
// for some condition.
type RuneChecker interface {
	// Check should check an item type of rune for some condition.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Check(rune) (bool, error)
}

// RuneCheck is a shortcut implementation
// of RuneChecker based on a function.
type RuneCheck func(rune) (bool, error)

// Check checks an item type of rune for some condition.
// It returns EndOfRuneIterator to stop iteration.
func (ch RuneCheck) Check(item rune) (bool, error) { return ch(item) }

var (
	// AlwaysRuneCheckTrue always returns true and empty error.
	AlwaysRuneCheckTrue RuneChecker = RuneCheck(
		func(item rune) (bool, error) { return true, nil })
	// AlwaysRuneCheckFalse always returns false and empty error.
	AlwaysRuneCheckFalse RuneChecker = RuneCheck(
		func(item rune) (bool, error) { return false, nil })
)

// NotRune do an inversion for checker result.
// It is returns AlwaysRuneCheckTrue if checker is nil.
func NotRune(checker RuneChecker) RuneChecker {
	if checker == nil {
		return AlwaysRuneCheckTrue
	}
	return RuneCheck(func(item rune) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andRune struct {
	lhs, rhs RuneChecker
}

func (a andRune) Check(item rune) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllRune combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllRune(checkers ...RuneChecker) RuneChecker {
	var all = AlwaysRuneCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andRune{checkers[i], all}
	}
	return all
}

type orRune struct {
	lhs, rhs RuneChecker
}

func (o orRune) Check(item rune) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyRune combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyRune(checkers ...RuneChecker) RuneChecker {
	var any = AlwaysRuneCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orRune{checkers[i], any}
	}
	return any
}

// FilteringRuneIterator does iteration with
// filtering by previously set checker.
type FilteringRuneIterator struct {
	preparedRuneItem
	filter RuneChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneFiltering sets filter while iterating over items.
// If filters is empty, so all items will return.
func RuneFiltering(items RuneIterator, filters ...RuneChecker) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &FilteringRuneIterator{preparedRuneItem{base: items}, AllRune(filters...)}
}

// RuneEnumChecker is an object checking an item type of rune
// and its ordering number in for some condition.
type RuneEnumChecker interface {
	// Check checks an item type of rune and its ordering number for some condition.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Check(int, rune) (bool, error)
}

// RuneEnumCheck is a shortcut implementation
// of RuneEnumChecker based on a function.
type RuneEnumCheck func(int, rune) (bool, error)

// Check checks an item type of rune and its ordering number for some condition.
// It returns EndOfRuneIterator to stop iteration.
func (ch RuneEnumCheck) Check(n int, item rune) (bool, error) { return ch(n, item) }

type enumFromRuneChecker struct {
	RuneChecker
}

func (ch enumFromRuneChecker) Check(_ int, item rune) (bool, error) {
	return ch.RuneChecker.Check(item)
}

// EnumFromRuneChecker adapts checker type of RuneChecker
// to the interface RuneEnumChecker.
// If checker is nil it is return based on AlwaysRuneCheckFalse enum checker.
func EnumFromRuneChecker(checker RuneChecker) RuneEnumChecker {
	if checker == nil {
		checker = AlwaysRuneCheckFalse
	}
	return &enumFromRuneChecker{checker}
}

var (
	// AlwaysRuneEnumCheckTrue always returns true and empty error.
	AlwaysRuneEnumCheckTrue RuneEnumChecker = EnumFromRuneChecker(
		AlwaysRuneCheckTrue)
	// AlwaysRuneEnumCheckFalse always returns false and empty error.
	AlwaysRuneEnumCheckFalse RuneEnumChecker = EnumFromRuneChecker(
		AlwaysRuneCheckFalse)
)

// EnumNotRune do an inversion for checker result.
// It is returns AlwaysRuneEnumCheckTrue if checker is nil.
func EnumNotRune(checker RuneEnumChecker) RuneEnumChecker {
	if checker == nil {
		return AlwaysRuneEnumCheckTrue
	}
	return RuneEnumCheck(func(n int, item rune) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndRune struct {
	lhs, rhs RuneEnumChecker
}

func (a enumAndRune) Check(n int, item rune) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllRune combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllRune(checkers ...RuneEnumChecker) RuneEnumChecker {
	var all = AlwaysRuneEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndRune{checkers[i], all}
	}
	return all
}

type enumOrRune struct {
	lhs, rhs RuneEnumChecker
}

func (o enumOrRune) Check(n int, item rune) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyRune combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyRune(checkers ...RuneEnumChecker) RuneEnumChecker {
	var any = AlwaysRuneEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrRune{checkers[i], any}
	}
	return any
}

// EnumFilteringRuneIterator does iteration with
// filtering by previously set checker.
type EnumFilteringRuneIterator struct {
	preparedRuneItem
	filter RuneEnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneEnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func RuneEnumFiltering(items RuneIterator, filters ...RuneEnumChecker) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &EnumFilteringRuneIterator{preparedRuneItem{base: items}, EnumAllRune(filters...), 0}
}

// DoingUntilRuneIterator does iteration
// until previously set checker is passed.
type DoingUntilRuneIterator struct {
	preparedRuneItem
	until RuneChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfRuneIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// RuneDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func RuneDoingUntil(items RuneIterator, untilList ...RuneChecker) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &DoingUntilRuneIterator{preparedRuneItem{base: items}, AllRune(untilList...)}
}

// RuneSkipUntil sets until conditions to skip few items.
func RuneSkipUntil(items RuneIterator, untilList ...RuneChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return RuneDiscard(RuneDoingUntil(items, untilList...))
}

// RuneGettingBatch returns the next batch from items.
func RuneGettingBatch(items RuneIterator, batchSize int) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	if batchSize == 0 {
		return items
	}

	size := 0
	return RuneDoingUntil(items, RuneCheck(func(item rune) (bool, error) {
		size++
		return size >= batchSize, nil
	}))
}
