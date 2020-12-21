package iter

import "github.com/pkg/errors"

// UintChecker is an object checking an item type of uint
// for some condition.
type UintChecker interface {
	// Check should check an item type of uint for some condition.
	// It is suggested to return EndOfUintIterator to stop iteration.
	Check(uint) (bool, error)
}

// UintCheck is a shortcut implementation
// of UintChecker based on a function.
type UintCheck func(uint) (bool, error)

// Check checks an item type of uint for some condition.
// It returns EndOfUintIterator to stop iteration.
func (ch UintCheck) Check(item uint) (bool, error) { return ch(item) }

var (
	// AlwaysUintCheckTrue always returns true and empty error.
	AlwaysUintCheckTrue UintChecker = UintCheck(
		func(item uint) (bool, error) { return true, nil })
	// AlwaysUintCheckFalse always returns false and empty error.
	AlwaysUintCheckFalse UintChecker = UintCheck(
		func(item uint) (bool, error) { return false, nil })
)

// NotUint do an inversion for checker result.
// It is returns AlwaysUintCheckTrue if checker is nil.
func NotUint(checker UintChecker) UintChecker {
	if checker == nil {
		return AlwaysUintCheckTrue
	}
	return UintCheck(func(item uint) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andUint struct {
	lhs, rhs UintChecker
}

func (a andUint) Check(item uint) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUintIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUintIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllUint combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllUint(checkers ...UintChecker) UintChecker {
	var all = AlwaysUintCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andUint{checkers[i], all}
	}
	return all
}

type orUint struct {
	lhs, rhs UintChecker
}

func (o orUint) Check(item uint) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUintIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUintIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyUint combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyUint(checkers ...UintChecker) UintChecker {
	var any = AlwaysUintCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orUint{checkers[i], any}
	}
	return any
}

// FilteringUintIterator does iteration with
// filtering by previously set checker.
type FilteringUintIterator struct {
	preparedUintItem
	filter UintChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUintItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfUintIterator(err) {
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

// UintFiltering sets filter while iterating over items.
// If filters is empty, so all items will return.
func UintFiltering(items UintIterator, filters ...UintChecker) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	return &FilteringUintIterator{preparedUintItem{base: items}, AllUint(filters...)}
}

// DoingUntilUintIterator does iteration
// until previously set checker is passed.
type DoingUntilUintIterator struct {
	preparedUintItem
	until UintChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUintItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfUintIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfUintIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// UintDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func UintDoingUntil(items UintIterator, untilList ...UintChecker) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	var until UintChecker
	if len(untilList) > 0 {
		until = AllUint(untilList...)
	} else {
		until = AlwaysUintCheckFalse
	}
	return &DoingUntilUintIterator{preparedUintItem{base: items}, until}
}

// UintSkipUntil sets until conditions to skip few items.
func UintSkipUntil(items UintIterator, untilList ...UintChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return UintDiscard(UintDoingUntil(items, untilList...))
}

// UintEnumChecker is an object checking an item type of uint
// and its ordering number in for some condition.
type UintEnumChecker interface {
	// Check checks an item type of uint and its ordering number for some condition.
	// It is suggested to return EndOfUintIterator to stop iteration.
	Check(int, uint) (bool, error)
}

// UintEnumCheck is a shortcut implementation
// of UintEnumChecker based on a function.
type UintEnumCheck func(int, uint) (bool, error)

// Check checks an item type of uint and its ordering number for some condition.
// It returns EndOfUintIterator to stop iteration.
func (ch UintEnumCheck) Check(n int, item uint) (bool, error) { return ch(n, item) }

type enumFromUintChecker struct {
	UintChecker
}

func (ch enumFromUintChecker) Check(_ int, item uint) (bool, error) {
	return ch.UintChecker.Check(item)
}

// EnumFromUintChecker adapts checker type of UintChecker
// to the interface UintEnumChecker.
// If checker is nil it is return based on AlwaysUintCheckFalse enum checker.
func EnumFromUintChecker(checker UintChecker) UintEnumChecker {
	if checker == nil {
		checker = AlwaysUintCheckFalse
	}
	return &enumFromUintChecker{checker}
}

var (
	// AlwaysUintEnumCheckTrue always returns true and empty error.
	AlwaysUintEnumCheckTrue = EnumFromUintChecker(
		AlwaysUintCheckTrue)
	// AlwaysUintEnumCheckFalse always returns false and empty error.
	AlwaysUintEnumCheckFalse = EnumFromUintChecker(
		AlwaysUintCheckFalse)
)

// EnumNotUint do an inversion for checker result.
// It is returns AlwaysUintEnumCheckTrue if checker is nil.
func EnumNotUint(checker UintEnumChecker) UintEnumChecker {
	if checker == nil {
		return AlwaysUintEnumCheckTrue
	}
	return UintEnumCheck(func(n int, item uint) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndUint struct {
	lhs, rhs UintEnumChecker
}

func (a enumAndUint) Check(n int, item uint) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUintIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUintIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllUint combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllUint(checkers ...UintEnumChecker) UintEnumChecker {
	var all = AlwaysUintEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndUint{checkers[i], all}
	}
	return all
}

type enumOrUint struct {
	lhs, rhs UintEnumChecker
}

func (o enumOrUint) Check(n int, item uint) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUintIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUintIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyUint combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyUint(checkers ...UintEnumChecker) UintEnumChecker {
	var any = AlwaysUintEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrUint{checkers[i], any}
	}
	return any
}

// EnumFilteringUintIterator does iteration with
// filtering by previously set checker.
type EnumFilteringUintIterator struct {
	preparedUintItem
	filter UintEnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUintItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfUintIterator(err) {
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

// UintEnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func UintEnumFiltering(items UintIterator, filters ...UintEnumChecker) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	return &EnumFilteringUintIterator{preparedUintItem{base: items}, EnumAllUint(filters...), 0}
}

// EnumDoingUntilUintIterator does iteration
// until previously set checker is passed.
type EnumDoingUntilUintIterator struct {
	preparedUintItem
	until UintEnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUintItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfUintIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfUintIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// UintEnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func UintEnumDoingUntil(items UintIterator, untilList ...UintEnumChecker) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	var until UintEnumChecker
	if len(untilList) > 0 {
		until = EnumAllUint(untilList...)
	} else {
		until = AlwaysUintEnumCheckFalse
	}
	return &EnumDoingUntilUintIterator{preparedUintItem{base: items}, until, 0}
}

// UintEnumSkipUntil sets until conditions to skip few items.
func UintEnumSkipUntil(items UintIterator, untilList ...UintEnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return UintDiscard(UintEnumDoingUntil(items, untilList...))
}

// UintGettingBatch returns the next batch from items.
func UintGettingBatch(items UintIterator, batchSize int) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	if batchSize == 0 {
		return items
	}

	return UintEnumDoingUntil(items, UintEnumCheck(func(n int, item uint) (bool, error) {
		return n == batchSize-1, nil
	}))
}
