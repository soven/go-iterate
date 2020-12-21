package iter

import "github.com/pkg/errors"

// Uint32Checker is an object checking an item type of uint32
// for some condition.
type Uint32Checker interface {
	// Check should check an item type of uint32 for some condition.
	// It is suggested to return EndOfUint32Iterator to stop iteration.
	Check(uint32) (bool, error)
}

// Uint32Check is a shortcut implementation
// of Uint32Checker based on a function.
type Uint32Check func(uint32) (bool, error)

// Check checks an item type of uint32 for some condition.
// It returns EndOfUint32Iterator to stop iteration.
func (ch Uint32Check) Check(item uint32) (bool, error) { return ch(item) }

var (
	// AlwaysUint32CheckTrue always returns true and empty error.
	AlwaysUint32CheckTrue Uint32Checker = Uint32Check(
		func(item uint32) (bool, error) { return true, nil })
	// AlwaysUint32CheckFalse always returns false and empty error.
	AlwaysUint32CheckFalse Uint32Checker = Uint32Check(
		func(item uint32) (bool, error) { return false, nil })
)

// NotUint32 do an inversion for checker result.
// It is returns AlwaysUint32CheckTrue if checker is nil.
func NotUint32(checker Uint32Checker) Uint32Checker {
	if checker == nil {
		return AlwaysUint32CheckTrue
	}
	return Uint32Check(func(item uint32) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andUint32 struct {
	lhs, rhs Uint32Checker
}

func (a andUint32) Check(item uint32) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint32Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllUint32 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllUint32(checkers ...Uint32Checker) Uint32Checker {
	var all = AlwaysUint32CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andUint32{checkers[i], all}
	}
	return all
}

type orUint32 struct {
	lhs, rhs Uint32Checker
}

func (o orUint32) Check(item uint32) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint32Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyUint32 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyUint32(checkers ...Uint32Checker) Uint32Checker {
	var any = AlwaysUint32CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orUint32{checkers[i], any}
	}
	return any
}

// FilteringUint32Iterator does iteration with
// filtering by previously set checker.
type FilteringUint32Iterator struct {
	preparedUint32Item
	filter Uint32Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringUint32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint32Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfUint32Iterator(err) {
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

// Uint32Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Uint32Filtering(items Uint32Iterator, filters ...Uint32Checker) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	return &FilteringUint32Iterator{preparedUint32Item{base: items}, AllUint32(filters...)}
}

// DoingUntilUint32Iterator does iteration
// until previously set checker is passed.
type DoingUntilUint32Iterator struct {
	preparedUint32Item
	until Uint32Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilUint32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint32Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfUint32Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfUint32Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint32DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Uint32DoingUntil(items Uint32Iterator, untilList ...Uint32Checker) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	var until Uint32Checker
	if len(untilList) > 0 {
		until = AllUint32(untilList...)
	} else {
		until = AlwaysUint32CheckFalse
	}
	return &DoingUntilUint32Iterator{preparedUint32Item{base: items}, until}
}

// Uint32SkipUntil sets until conditions to skip few items.
func Uint32SkipUntil(items Uint32Iterator, untilList ...Uint32Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint32Discard(Uint32DoingUntil(items, untilList...))
}

// Uint32EnumChecker is an object checking an item type of uint32
// and its ordering number in for some condition.
type Uint32EnumChecker interface {
	// Check checks an item type of uint32 and its ordering number for some condition.
	// It is suggested to return EndOfUint32Iterator to stop iteration.
	Check(int, uint32) (bool, error)
}

// Uint32EnumCheck is a shortcut implementation
// of Uint32EnumChecker based on a function.
type Uint32EnumCheck func(int, uint32) (bool, error)

// Check checks an item type of uint32 and its ordering number for some condition.
// It returns EndOfUint32Iterator to stop iteration.
func (ch Uint32EnumCheck) Check(n int, item uint32) (bool, error) { return ch(n, item) }

type enumFromUint32Checker struct {
	Uint32Checker
}

func (ch enumFromUint32Checker) Check(_ int, item uint32) (bool, error) {
	return ch.Uint32Checker.Check(item)
}

// EnumFromUint32Checker adapts checker type of Uint32Checker
// to the interface Uint32EnumChecker.
// If checker is nil it is return based on AlwaysUint32CheckFalse enum checker.
func EnumFromUint32Checker(checker Uint32Checker) Uint32EnumChecker {
	if checker == nil {
		checker = AlwaysUint32CheckFalse
	}
	return &enumFromUint32Checker{checker}
}

var (
	// AlwaysUint32EnumCheckTrue always returns true and empty error.
	AlwaysUint32EnumCheckTrue = EnumFromUint32Checker(
		AlwaysUint32CheckTrue)
	// AlwaysUint32EnumCheckFalse always returns false and empty error.
	AlwaysUint32EnumCheckFalse = EnumFromUint32Checker(
		AlwaysUint32CheckFalse)
)

// EnumNotUint32 do an inversion for checker result.
// It is returns AlwaysUint32EnumCheckTrue if checker is nil.
func EnumNotUint32(checker Uint32EnumChecker) Uint32EnumChecker {
	if checker == nil {
		return AlwaysUint32EnumCheckTrue
	}
	return Uint32EnumCheck(func(n int, item uint32) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndUint32 struct {
	lhs, rhs Uint32EnumChecker
}

func (a enumAndUint32) Check(n int, item uint32) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint32Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllUint32 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllUint32(checkers ...Uint32EnumChecker) Uint32EnumChecker {
	var all = AlwaysUint32EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndUint32{checkers[i], all}
	}
	return all
}

type enumOrUint32 struct {
	lhs, rhs Uint32EnumChecker
}

func (o enumOrUint32) Check(n int, item uint32) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint32Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyUint32 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyUint32(checkers ...Uint32EnumChecker) Uint32EnumChecker {
	var any = AlwaysUint32EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrUint32{checkers[i], any}
	}
	return any
}

// EnumFilteringUint32Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringUint32Iterator struct {
	preparedUint32Item
	filter Uint32EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringUint32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint32Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfUint32Iterator(err) {
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

// Uint32EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Uint32EnumFiltering(items Uint32Iterator, filters ...Uint32EnumChecker) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	return &EnumFilteringUint32Iterator{preparedUint32Item{base: items}, EnumAllUint32(filters...), 0}
}

// EnumDoingUntilUint32Iterator does iteration
// until previously set checker is passed.
type EnumDoingUntilUint32Iterator struct {
	preparedUint32Item
	until Uint32EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilUint32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint32Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfUint32Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfUint32Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint32EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Uint32EnumDoingUntil(items Uint32Iterator, untilList ...Uint32EnumChecker) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	var until Uint32EnumChecker
	if len(untilList) > 0 {
		until = EnumAllUint32(untilList...)
	} else {
		until = AlwaysUint32EnumCheckFalse
	}
	return &EnumDoingUntilUint32Iterator{preparedUint32Item{base: items}, until, 0}
}

// Uint32EnumSkipUntil sets until conditions to skip few items.
func Uint32EnumSkipUntil(items Uint32Iterator, untilList ...Uint32EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint32Discard(Uint32EnumDoingUntil(items, untilList...))
}

// Uint32GettingBatch returns the next batch from items.
func Uint32GettingBatch(items Uint32Iterator, batchSize int) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	if batchSize == 0 {
		return items
	}

	return Uint32EnumDoingUntil(items, Uint32EnumCheck(func(n int, item uint32) (bool, error) {
		return n == batchSize-1, nil
	}))
}
