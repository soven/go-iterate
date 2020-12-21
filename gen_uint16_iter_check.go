package iter

import "github.com/pkg/errors"

// Uint16Checker is an object checking an item type of uint16
// for some condition.
type Uint16Checker interface {
	// Check should check an item type of uint16 for some condition.
	// It is suggested to return EndOfUint16Iterator to stop iteration.
	Check(uint16) (bool, error)
}

// Uint16Check is a shortcut implementation
// of Uint16Checker based on a function.
type Uint16Check func(uint16) (bool, error)

// Check checks an item type of uint16 for some condition.
// It returns EndOfUint16Iterator to stop iteration.
func (ch Uint16Check) Check(item uint16) (bool, error) { return ch(item) }

var (
	// AlwaysUint16CheckTrue always returns true and empty error.
	AlwaysUint16CheckTrue Uint16Checker = Uint16Check(
		func(item uint16) (bool, error) { return true, nil })
	// AlwaysUint16CheckFalse always returns false and empty error.
	AlwaysUint16CheckFalse Uint16Checker = Uint16Check(
		func(item uint16) (bool, error) { return false, nil })
)

// NotUint16 do an inversion for checker result.
// It is returns AlwaysUint16CheckTrue if checker is nil.
func NotUint16(checker Uint16Checker) Uint16Checker {
	if checker == nil {
		return AlwaysUint16CheckTrue
	}
	return Uint16Check(func(item uint16) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andUint16 struct {
	lhs, rhs Uint16Checker
}

func (a andUint16) Check(item uint16) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint16Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllUint16 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllUint16(checkers ...Uint16Checker) Uint16Checker {
	var all = AlwaysUint16CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andUint16{checkers[i], all}
	}
	return all
}

type orUint16 struct {
	lhs, rhs Uint16Checker
}

func (o orUint16) Check(item uint16) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint16Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyUint16 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyUint16(checkers ...Uint16Checker) Uint16Checker {
	var any = AlwaysUint16CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orUint16{checkers[i], any}
	}
	return any
}

// FilteringUint16Iterator does iteration with
// filtering by previously set checker.
type FilteringUint16Iterator struct {
	preparedUint16Item
	filter Uint16Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
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

// Uint16Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Uint16Filtering(items Uint16Iterator, filters ...Uint16Checker) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	return &FilteringUint16Iterator{preparedUint16Item{base: items}, AllUint16(filters...)}
}

// DoingUntilUint16Iterator does iteration
// until previously set checker is passed.
type DoingUntilUint16Iterator struct {
	preparedUint16Item
	until Uint16Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfUint16Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint16DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Uint16DoingUntil(items Uint16Iterator, untilList ...Uint16Checker) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	var until Uint16Checker
	if len(untilList) > 0 {
		until = AllUint16(untilList...)
	} else {
		until = AlwaysUint16CheckFalse
	}
	return &DoingUntilUint16Iterator{preparedUint16Item{base: items}, until}
}

// Uint16SkipUntil sets until conditions to skip few items.
func Uint16SkipUntil(items Uint16Iterator, untilList ...Uint16Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint16Discard(Uint16DoingUntil(items, untilList...))
}

// Uint16EnumChecker is an object checking an item type of uint16
// and its ordering number in for some condition.
type Uint16EnumChecker interface {
	// Check checks an item type of uint16 and its ordering number for some condition.
	// It is suggested to return EndOfUint16Iterator to stop iteration.
	Check(int, uint16) (bool, error)
}

// Uint16EnumCheck is a shortcut implementation
// of Uint16EnumChecker based on a function.
type Uint16EnumCheck func(int, uint16) (bool, error)

// Check checks an item type of uint16 and its ordering number for some condition.
// It returns EndOfUint16Iterator to stop iteration.
func (ch Uint16EnumCheck) Check(n int, item uint16) (bool, error) { return ch(n, item) }

type enumFromUint16Checker struct {
	Uint16Checker
}

func (ch enumFromUint16Checker) Check(_ int, item uint16) (bool, error) {
	return ch.Uint16Checker.Check(item)
}

// EnumFromUint16Checker adapts checker type of Uint16Checker
// to the interface Uint16EnumChecker.
// If checker is nil it is return based on AlwaysUint16CheckFalse enum checker.
func EnumFromUint16Checker(checker Uint16Checker) Uint16EnumChecker {
	if checker == nil {
		checker = AlwaysUint16CheckFalse
	}
	return &enumFromUint16Checker{checker}
}

var (
	// AlwaysUint16EnumCheckTrue always returns true and empty error.
	AlwaysUint16EnumCheckTrue = EnumFromUint16Checker(
		AlwaysUint16CheckTrue)
	// AlwaysUint16EnumCheckFalse always returns false and empty error.
	AlwaysUint16EnumCheckFalse = EnumFromUint16Checker(
		AlwaysUint16CheckFalse)
)

// EnumNotUint16 do an inversion for checker result.
// It is returns AlwaysUint16EnumCheckTrue if checker is nil.
func EnumNotUint16(checker Uint16EnumChecker) Uint16EnumChecker {
	if checker == nil {
		return AlwaysUint16EnumCheckTrue
	}
	return Uint16EnumCheck(func(n int, item uint16) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndUint16 struct {
	lhs, rhs Uint16EnumChecker
}

func (a enumAndUint16) Check(n int, item uint16) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint16Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllUint16 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllUint16(checkers ...Uint16EnumChecker) Uint16EnumChecker {
	var all = AlwaysUint16EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndUint16{checkers[i], all}
	}
	return all
}

type enumOrUint16 struct {
	lhs, rhs Uint16EnumChecker
}

func (o enumOrUint16) Check(n int, item uint16) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint16Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyUint16 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyUint16(checkers ...Uint16EnumChecker) Uint16EnumChecker {
	var any = AlwaysUint16EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrUint16{checkers[i], any}
	}
	return any
}

// EnumFilteringUint16Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringUint16Iterator struct {
	preparedUint16Item
	filter Uint16EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
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

// Uint16EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Uint16EnumFiltering(items Uint16Iterator, filters ...Uint16EnumChecker) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	return &EnumFilteringUint16Iterator{preparedUint16Item{base: items}, EnumAllUint16(filters...), 0}
}

// EnumDoingUntilUint16Iterator does iteration
// until previously set checker is passed.
type EnumDoingUntilUint16Iterator struct {
	preparedUint16Item
	until Uint16EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfUint16Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint16EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Uint16EnumDoingUntil(items Uint16Iterator, untilList ...Uint16EnumChecker) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	var until Uint16EnumChecker
	if len(untilList) > 0 {
		until = EnumAllUint16(untilList...)
	} else {
		until = AlwaysUint16EnumCheckFalse
	}
	return &EnumDoingUntilUint16Iterator{preparedUint16Item{base: items}, until, 0}
}

// Uint16EnumSkipUntil sets until conditions to skip few items.
func Uint16EnumSkipUntil(items Uint16Iterator, untilList ...Uint16EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint16Discard(Uint16EnumDoingUntil(items, untilList...))
}

// Uint16GettingBatch returns the next batch from items.
func Uint16GettingBatch(items Uint16Iterator, batchSize int) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	if batchSize == 0 {
		return items
	}

	return Uint16EnumDoingUntil(items, Uint16EnumCheck(func(n int, item uint16) (bool, error) {
		return n == batchSize-1, nil
	}))
}
