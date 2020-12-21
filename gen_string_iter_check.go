package iter

import "github.com/pkg/errors"

// StringChecker is an object checking an item type of string
// for some condition.
type StringChecker interface {
	// Check should check an item type of string for some condition.
	// It is suggested to return EndOfStringIterator to stop iteration.
	Check(string) (bool, error)
}

// StringCheck is a shortcut implementation
// of StringChecker based on a function.
type StringCheck func(string) (bool, error)

// Check checks an item type of string for some condition.
// It returns EndOfStringIterator to stop iteration.
func (ch StringCheck) Check(item string) (bool, error) { return ch(item) }

var (
	// AlwaysStringCheckTrue always returns true and empty error.
	AlwaysStringCheckTrue StringChecker = StringCheck(
		func(item string) (bool, error) { return true, nil })
	// AlwaysStringCheckFalse always returns false and empty error.
	AlwaysStringCheckFalse StringChecker = StringCheck(
		func(item string) (bool, error) { return false, nil })
)

// NotString do an inversion for checker result.
// It is returns AlwaysStringCheckTrue if checker is nil.
func NotString(checker StringChecker) StringChecker {
	if checker == nil {
		return AlwaysStringCheckTrue
	}
	return StringCheck(func(item string) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andString struct {
	lhs, rhs StringChecker
}

func (a andString) Check(item string) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllString combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllString(checkers ...StringChecker) StringChecker {
	var all = AlwaysStringCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andString{checkers[i], all}
	}
	return all
}

type orString struct {
	lhs, rhs StringChecker
}

func (o orString) Check(item string) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyString combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyString(checkers ...StringChecker) StringChecker {
	var any = AlwaysStringCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orString{checkers[i], any}
	}
	return any
}

// FilteringStringIterator does iteration with
// filtering by previously set checker.
type FilteringStringIterator struct {
	preparedStringItem
	filter StringChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedStringItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfStringIterator(err) {
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

// StringFiltering sets filter while iterating over items.
// If filters is empty, so all items will return.
func StringFiltering(items StringIterator, filters ...StringChecker) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	return &FilteringStringIterator{preparedStringItem{base: items}, AllString(filters...)}
}

// DoingUntilStringIterator does iteration
// until previously set checker is passed.
type DoingUntilStringIterator struct {
	preparedStringItem
	until StringChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedStringItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfStringIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfStringIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// StringDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func StringDoingUntil(items StringIterator, untilList ...StringChecker) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	var until StringChecker
	if len(untilList) > 0 {
		until = AllString(untilList...)
	} else {
		until = AlwaysStringCheckFalse
	}
	return &DoingUntilStringIterator{preparedStringItem{base: items}, until}
}

// StringSkipUntil sets until conditions to skip few items.
func StringSkipUntil(items StringIterator, untilList ...StringChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return StringDiscard(StringDoingUntil(items, untilList...))
}

// StringEnumChecker is an object checking an item type of string
// and its ordering number in for some condition.
type StringEnumChecker interface {
	// Check checks an item type of string and its ordering number for some condition.
	// It is suggested to return EndOfStringIterator to stop iteration.
	Check(int, string) (bool, error)
}

// StringEnumCheck is a shortcut implementation
// of StringEnumChecker based on a function.
type StringEnumCheck func(int, string) (bool, error)

// Check checks an item type of string and its ordering number for some condition.
// It returns EndOfStringIterator to stop iteration.
func (ch StringEnumCheck) Check(n int, item string) (bool, error) { return ch(n, item) }

type enumFromStringChecker struct {
	StringChecker
}

func (ch enumFromStringChecker) Check(_ int, item string) (bool, error) {
	return ch.StringChecker.Check(item)
}

// EnumFromStringChecker adapts checker type of StringChecker
// to the interface StringEnumChecker.
// If checker is nil it is return based on AlwaysStringCheckFalse enum checker.
func EnumFromStringChecker(checker StringChecker) StringEnumChecker {
	if checker == nil {
		checker = AlwaysStringCheckFalse
	}
	return &enumFromStringChecker{checker}
}

var (
	// AlwaysStringEnumCheckTrue always returns true and empty error.
	AlwaysStringEnumCheckTrue = EnumFromStringChecker(
		AlwaysStringCheckTrue)
	// AlwaysStringEnumCheckFalse always returns false and empty error.
	AlwaysStringEnumCheckFalse = EnumFromStringChecker(
		AlwaysStringCheckFalse)
)

// EnumNotString do an inversion for checker result.
// It is returns AlwaysStringEnumCheckTrue if checker is nil.
func EnumNotString(checker StringEnumChecker) StringEnumChecker {
	if checker == nil {
		return AlwaysStringEnumCheckTrue
	}
	return StringEnumCheck(func(n int, item string) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndString struct {
	lhs, rhs StringEnumChecker
}

func (a enumAndString) Check(n int, item string) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllString combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllString(checkers ...StringEnumChecker) StringEnumChecker {
	var all = AlwaysStringEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndString{checkers[i], all}
	}
	return all
}

type enumOrString struct {
	lhs, rhs StringEnumChecker
}

func (o enumOrString) Check(n int, item string) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyString combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyString(checkers ...StringEnumChecker) StringEnumChecker {
	var any = AlwaysStringEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrString{checkers[i], any}
	}
	return any
}

// EnumFilteringStringIterator does iteration with
// filtering by previously set checker.
type EnumFilteringStringIterator struct {
	preparedStringItem
	filter StringEnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedStringItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfStringIterator(err) {
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

// StringEnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func StringEnumFiltering(items StringIterator, filters ...StringEnumChecker) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	return &EnumFilteringStringIterator{preparedStringItem{base: items}, EnumAllString(filters...), 0}
}

// EnumDoingUntilStringIterator does iteration
// until previously set checker is passed.
type EnumDoingUntilStringIterator struct {
	preparedStringItem
	until StringEnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedStringItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfStringIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfStringIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// StringEnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func StringEnumDoingUntil(items StringIterator, untilList ...StringEnumChecker) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	var until StringEnumChecker
	if len(untilList) > 0 {
		until = EnumAllString(untilList...)
	} else {
		until = AlwaysStringEnumCheckFalse
	}
	return &EnumDoingUntilStringIterator{preparedStringItem{base: items}, until, 0}
}

// StringEnumSkipUntil sets until conditions to skip few items.
func StringEnumSkipUntil(items StringIterator, untilList ...StringEnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return StringDiscard(StringEnumDoingUntil(items, untilList...))
}

// StringGettingBatch returns the next batch from items.
func StringGettingBatch(items StringIterator, batchSize int) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	if batchSize == 0 {
		return items
	}

	return StringEnumDoingUntil(items, StringEnumCheck(func(n int, item string) (bool, error) {
		return n == batchSize-1, nil
	}))
}
