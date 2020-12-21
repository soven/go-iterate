package iter

import "github.com/pkg/errors"

// Checker is an object checking an item type of interface{}
// for some condition.
type Checker interface {
	// Check should check an item type of interface{} for some condition.
	// It is suggested to return EndOfIterator to stop iteration.
	Check(interface{}) (bool, error)
}

// Check is a shortcut implementation
// of Checker based on a function.
type Check func(interface{}) (bool, error)

// Check checks an item type of interface{} for some condition.
// It returns EndOfIterator to stop iteration.
func (ch Check) Check(item interface{}) (bool, error) { return ch(item) }

var (
	// AlwaysCheckTrue always returns true and empty error.
	AlwaysCheckTrue Checker = Check(
		func(item interface{}) (bool, error) { return true, nil })
	// AlwaysCheckFalse always returns false and empty error.
	AlwaysCheckFalse Checker = Check(
		func(item interface{}) (bool, error) { return false, nil })
)

// Not do an inversion for checker result.
// It is returns AlwaysCheckTrue if checker is nil.
func Not(checker Checker) Checker {
	if checker == nil {
		return AlwaysCheckTrue
	}
	return Check(func(item interface{}) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type and struct {
	lhs, rhs Checker
}

func (a and) Check(item interface{}) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// All combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func All(checkers ...Checker) Checker {
	var all = AlwaysCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = and{checkers[i], all}
	}
	return all
}

type or struct {
	lhs, rhs Checker
}

func (o or) Check(item interface{}) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// Any combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func Any(checkers ...Checker) Checker {
	var any = AlwaysCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = or{checkers[i], any}
	}
	return any
}

// FilteringIterator does iteration with
// filtering by previously set checker.
type FilteringIterator struct {
	preparedItem
	filter Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfIterator(err) {
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

// Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Filtering(items Iterator, filters ...Checker) Iterator {
	if items == nil {
		return EmptyIterator
	}
	return &FilteringIterator{preparedItem{base: items}, All(filters...)}
}

// DoingUntilIterator does iteration
// until previously set checker is passed.
type DoingUntilIterator struct {
	preparedItem
	until Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func DoingUntil(items Iterator, untilList ...Checker) Iterator {
	if items == nil {
		return EmptyIterator
	}
	var until Checker
	if len(untilList) > 0 {
		until = All(untilList...)
	} else {
		until = AlwaysCheckFalse
	}
	return &DoingUntilIterator{preparedItem{base: items}, until}
}

// SkipUntil sets until conditions to skip few items.
func SkipUntil(items Iterator, untilList ...Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Discard(DoingUntil(items, untilList...))
}

// EnumChecker is an object checking an item type of interface{}
// and its ordering number in for some condition.
type EnumChecker interface {
	// Check checks an item type of interface{} and its ordering number for some condition.
	// It is suggested to return EndOfIterator to stop iteration.
	Check(int, interface{}) (bool, error)
}

// EnumCheck is a shortcut implementation
// of EnumChecker based on a function.
type EnumCheck func(int, interface{}) (bool, error)

// Check checks an item type of interface{} and its ordering number for some condition.
// It returns EndOfIterator to stop iteration.
func (ch EnumCheck) Check(n int, item interface{}) (bool, error) { return ch(n, item) }

type enumFromChecker struct {
	Checker
}

func (ch enumFromChecker) Check(_ int, item interface{}) (bool, error) {
	return ch.Checker.Check(item)
}

// EnumFromChecker adapts checker type of Checker
// to the interface EnumChecker.
// If checker is nil it is return based on AlwaysCheckFalse enum checker.
func EnumFromChecker(checker Checker) EnumChecker {
	if checker == nil {
		checker = AlwaysCheckFalse
	}
	return &enumFromChecker{checker}
}

var (
	// AlwaysEnumCheckTrue always returns true and empty error.
	AlwaysEnumCheckTrue = EnumFromChecker(
		AlwaysCheckTrue)
	// AlwaysEnumCheckFalse always returns false and empty error.
	AlwaysEnumCheckFalse = EnumFromChecker(
		AlwaysCheckFalse)
)

// EnumNot do an inversion for checker result.
// It is returns AlwaysEnumCheckTrue if checker is nil.
func EnumNot(checker EnumChecker) EnumChecker {
	if checker == nil {
		return AlwaysEnumCheckTrue
	}
	return EnumCheck(func(n int, item interface{}) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAnd struct {
	lhs, rhs EnumChecker
}

func (a enumAnd) Check(n int, item interface{}) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAll combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAll(checkers ...EnumChecker) EnumChecker {
	var all = AlwaysEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAnd{checkers[i], all}
	}
	return all
}

type enumOr struct {
	lhs, rhs EnumChecker
}

func (o enumOr) Check(n int, item interface{}) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAny combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAny(checkers ...EnumChecker) EnumChecker {
	var any = AlwaysEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOr{checkers[i], any}
	}
	return any
}

// EnumFilteringIterator does iteration with
// filtering by previously set checker.
type EnumFilteringIterator struct {
	preparedItem
	filter EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfIterator(err) {
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

// EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func EnumFiltering(items Iterator, filters ...EnumChecker) Iterator {
	if items == nil {
		return EmptyIterator
	}
	return &EnumFilteringIterator{preparedItem{base: items}, EnumAll(filters...), 0}
}

// EnumDoingUntilIterator does iteration
// until previously set checker is passed.
type EnumDoingUntilIterator struct {
	preparedItem
	until EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func EnumDoingUntil(items Iterator, untilList ...EnumChecker) Iterator {
	if items == nil {
		return EmptyIterator
	}
	var until EnumChecker
	if len(untilList) > 0 {
		until = EnumAll(untilList...)
	} else {
		until = AlwaysEnumCheckFalse
	}
	return &EnumDoingUntilIterator{preparedItem{base: items}, until, 0}
}

// EnumSkipUntil sets until conditions to skip few items.
func EnumSkipUntil(items Iterator, untilList ...EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Discard(EnumDoingUntil(items, untilList...))
}

// GettingBatch returns the next batch from items.
func GettingBatch(items Iterator, batchSize int) Iterator {
	if items == nil {
		return EmptyIterator
	}
	if batchSize == 0 {
		return items
	}

	return EnumDoingUntil(items, EnumCheck(func(n int, item interface{}) (bool, error) {
		return n == batchSize-1, nil
	}))
}
