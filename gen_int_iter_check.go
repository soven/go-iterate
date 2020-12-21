package iter

import "github.com/pkg/errors"

// IntChecker is an object checking an item type of int
// for some condition.
type IntChecker interface {
	// Check should check an item type of int for some condition.
	// It is suggested to return EndOfIntIterator to stop iteration.
	Check(int) (bool, error)
}

// IntCheck is a shortcut implementation
// of IntChecker based on a function.
type IntCheck func(int) (bool, error)

// Check checks an item type of int for some condition.
// It returns EndOfIntIterator to stop iteration.
func (ch IntCheck) Check(item int) (bool, error) { return ch(item) }

var (
	// AlwaysIntCheckTrue always returns true and empty error.
	AlwaysIntCheckTrue IntChecker = IntCheck(
		func(item int) (bool, error) { return true, nil })
	// AlwaysIntCheckFalse always returns false and empty error.
	AlwaysIntCheckFalse IntChecker = IntCheck(
		func(item int) (bool, error) { return false, nil })
)

// NotInt do an inversion for checker result.
// It is returns AlwaysIntCheckTrue if checker is nil.
func NotInt(checker IntChecker) IntChecker {
	if checker == nil {
		return AlwaysIntCheckTrue
	}
	return IntCheck(func(item int) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andInt struct {
	lhs, rhs IntChecker
}

func (a andInt) Check(item int) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllInt combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllInt(checkers ...IntChecker) IntChecker {
	var all = AlwaysIntCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andInt{checkers[i], all}
	}
	return all
}

type orInt struct {
	lhs, rhs IntChecker
}

func (o orInt) Check(item int) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyInt combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyInt(checkers ...IntChecker) IntChecker {
	var any = AlwaysIntCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orInt{checkers[i], any}
	}
	return any
}

// FilteringIntIterator does iteration with
// filtering by previously set checker.
type FilteringIntIterator struct {
	preparedIntItem
	filter IntChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedIntItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfIntIterator(err) {
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

// IntFiltering sets filter while iterating over items.
// If filters is empty, so all items will return.
func IntFiltering(items IntIterator, filters ...IntChecker) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	return &FilteringIntIterator{preparedIntItem{base: items}, AllInt(filters...)}
}

// DoingUntilIntIterator does iteration
// until previously set checker is passed.
type DoingUntilIntIterator struct {
	preparedIntItem
	until IntChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedIntItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfIntIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfIntIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// IntDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func IntDoingUntil(items IntIterator, untilList ...IntChecker) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	var until IntChecker
	if len(untilList) > 0 {
		until = AllInt(untilList...)
	} else {
		until = AlwaysIntCheckFalse
	}
	return &DoingUntilIntIterator{preparedIntItem{base: items}, until}
}

// IntSkipUntil sets until conditions to skip few items.
func IntSkipUntil(items IntIterator, untilList ...IntChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return IntDiscard(IntDoingUntil(items, untilList...))
}

// IntEnumChecker is an object checking an item type of int
// and its ordering number in for some condition.
type IntEnumChecker interface {
	// Check checks an item type of int and its ordering number for some condition.
	// It is suggested to return EndOfIntIterator to stop iteration.
	Check(int, int) (bool, error)
}

// IntEnumCheck is a shortcut implementation
// of IntEnumChecker based on a function.
type IntEnumCheck func(int, int) (bool, error)

// Check checks an item type of int and its ordering number for some condition.
// It returns EndOfIntIterator to stop iteration.
func (ch IntEnumCheck) Check(n int, item int) (bool, error) { return ch(n, item) }

type enumFromIntChecker struct {
	IntChecker
}

func (ch enumFromIntChecker) Check(_ int, item int) (bool, error) {
	return ch.IntChecker.Check(item)
}

// EnumFromIntChecker adapts checker type of IntChecker
// to the interface IntEnumChecker.
// If checker is nil it is return based on AlwaysIntCheckFalse enum checker.
func EnumFromIntChecker(checker IntChecker) IntEnumChecker {
	if checker == nil {
		checker = AlwaysIntCheckFalse
	}
	return &enumFromIntChecker{checker}
}

var (
	// AlwaysIntEnumCheckTrue always returns true and empty error.
	AlwaysIntEnumCheckTrue = EnumFromIntChecker(
		AlwaysIntCheckTrue)
	// AlwaysIntEnumCheckFalse always returns false and empty error.
	AlwaysIntEnumCheckFalse = EnumFromIntChecker(
		AlwaysIntCheckFalse)
)

// EnumNotInt do an inversion for checker result.
// It is returns AlwaysIntEnumCheckTrue if checker is nil.
func EnumNotInt(checker IntEnumChecker) IntEnumChecker {
	if checker == nil {
		return AlwaysIntEnumCheckTrue
	}
	return IntEnumCheck(func(n int, item int) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndInt struct {
	lhs, rhs IntEnumChecker
}

func (a enumAndInt) Check(n int, item int) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllInt combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllInt(checkers ...IntEnumChecker) IntEnumChecker {
	var all = AlwaysIntEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndInt{checkers[i], all}
	}
	return all
}

type enumOrInt struct {
	lhs, rhs IntEnumChecker
}

func (o enumOrInt) Check(n int, item int) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyInt combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyInt(checkers ...IntEnumChecker) IntEnumChecker {
	var any = AlwaysIntEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrInt{checkers[i], any}
	}
	return any
}

// EnumFilteringIntIterator does iteration with
// filtering by previously set checker.
type EnumFilteringIntIterator struct {
	preparedIntItem
	filter IntEnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedIntItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfIntIterator(err) {
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

// IntEnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func IntEnumFiltering(items IntIterator, filters ...IntEnumChecker) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	return &EnumFilteringIntIterator{preparedIntItem{base: items}, EnumAllInt(filters...), 0}
}

// EnumDoingUntilIntIterator does iteration
// until previously set checker is passed.
type EnumDoingUntilIntIterator struct {
	preparedIntItem
	until IntEnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedIntItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfIntIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfIntIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// IntEnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func IntEnumDoingUntil(items IntIterator, untilList ...IntEnumChecker) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	var until IntEnumChecker
	if len(untilList) > 0 {
		until = EnumAllInt(untilList...)
	} else {
		until = AlwaysIntEnumCheckFalse
	}
	return &EnumDoingUntilIntIterator{preparedIntItem{base: items}, until, 0}
}

// IntEnumSkipUntil sets until conditions to skip few items.
func IntEnumSkipUntil(items IntIterator, untilList ...IntEnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return IntDiscard(IntEnumDoingUntil(items, untilList...))
}

// IntGettingBatch returns the next batch from items.
func IntGettingBatch(items IntIterator, batchSize int) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	if batchSize == 0 {
		return items
	}

	return IntEnumDoingUntil(items, IntEnumCheck(func(n int, item int) (bool, error) {
		return n == batchSize-1, nil
	}))
}
