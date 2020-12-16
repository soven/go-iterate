package resembled

import "github.com/pkg/errors"

// PrefixChecker is an object checking an item type of Type
// for some condition.
type PrefixChecker interface {
	// Check should check an item type of Type for some condition.
	// It is suggested to return EndOfPrefixIterator to stop iteration.
	Check(Type) (bool, error)
}

// PrefixCheck is a shortcut implementation
// of PrefixChecker based on a function.
type PrefixCheck func(Type) (bool, error)

// Check checks an item type of Type for some condition.
// It returns EndOfPrefixIterator to stop iteration.
func (ch PrefixCheck) Check(item Type) (bool, error) { return ch(item) }

var (
	// AlwaysPrefixCheckTrue always returns true and empty error.
	AlwaysPrefixCheckTrue PrefixChecker = PrefixCheck(
		func(item Type) (bool, error) { return true, nil })
	// AlwaysPrefixCheckFalse always returns false and empty error.
	AlwaysPrefixCheckFalse PrefixChecker = PrefixCheck(
		func(item Type) (bool, error) { return false, nil })
)

// NotPrefix do an inversion for checker result.
// It is returns AlwaysPrefixCheckTrue if checker is nil.
func NotPrefix(checker PrefixChecker) PrefixChecker {
	if checker == nil {
		return AlwaysPrefixCheckTrue
	}
	return PrefixCheck(func(item Type) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andPrefix struct {
	lhs, rhs PrefixChecker
}

func (a andPrefix) Check(item Type) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfPrefixIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfPrefixIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllPrefix combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllPrefix(checkers ...PrefixChecker) PrefixChecker {
	var all = AlwaysPrefixCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andPrefix{checkers[i], all}
	}
	return all
}

type orPrefix struct {
	lhs, rhs PrefixChecker
}

func (o orPrefix) Check(item Type) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfPrefixIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfPrefixIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyPrefix combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyPrefix(checkers ...PrefixChecker) PrefixChecker {
	var any = AlwaysPrefixCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orPrefix{checkers[i], any}
	}
	return any
}

// FilteringPrefixIterator does iteration with
// filtering by previously set checker.
type FilteringPrefixIterator struct {
	preparedPrefixItem
	filter PrefixChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringPrefixIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedPrefixItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfPrefixIterator(err) {
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

// PrefixFiltering sets filter while iterating over items.
// If filters is empty, so all items will return.
func PrefixFiltering(items PrefixIterator, filters ...PrefixChecker) PrefixIterator {
	if items == nil {
		return EmptyPrefixIterator
	}
	return &FilteringPrefixIterator{preparedPrefixItem{base: items}, AllPrefix(filters...)}
}

// PrefixEnumChecker is an object checking an item type of Type
// and its ordering number in for some condition.
type PrefixEnumChecker interface {
	// Check checks an item type of Type and its ordering number for some condition.
	// It is suggested to return EndOfPrefixIterator to stop iteration.
	Check(int, Type) (bool, error)
}

// PrefixEnumCheck is a shortcut implementation
// of PrefixEnumChecker based on a function.
type PrefixEnumCheck func(int, Type) (bool, error)

// Check checks an item type of Type and its ordering number for some condition.
// It returns EndOfPrefixIterator to stop iteration.
func (ch PrefixEnumCheck) Check(n int, item Type) (bool, error) { return ch(n, item) }

type enumFromPrefixChecker struct {
	PrefixChecker
}

func (ch enumFromPrefixChecker) Check(_ int, item Type) (bool, error) {
	return ch.PrefixChecker.Check(item)
}

// EnumFromPrefixChecker adapts checker type of PrefixChecker
// to the interface PrefixEnumChecker.
// If checker is nil it is return based on AlwaysPrefixCheckFalse enum checker.
func EnumFromPrefixChecker(checker PrefixChecker) PrefixEnumChecker {
	if checker == nil {
		checker = AlwaysPrefixCheckFalse
	}
	return &enumFromPrefixChecker{checker}
}

var (
	// AlwaysPrefixEnumCheckTrue always returns true and empty error.
	AlwaysPrefixEnumCheckTrue PrefixEnumChecker = EnumFromPrefixChecker(
		AlwaysPrefixCheckTrue)
	// AlwaysPrefixEnumCheckFalse always returns false and empty error.
	AlwaysPrefixEnumCheckFalse PrefixEnumChecker = EnumFromPrefixChecker(
		AlwaysPrefixCheckFalse)
)

// EnumNotPrefix do an inversion for checker result.
// It is returns AlwaysPrefixEnumCheckTrue if checker is nil.
func EnumNotPrefix(checker PrefixEnumChecker) PrefixEnumChecker {
	if checker == nil {
		return AlwaysPrefixEnumCheckTrue
	}
	return PrefixEnumCheck(func(n int, item Type) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndPrefix struct {
	lhs, rhs PrefixEnumChecker
}

func (a enumAndPrefix) Check(n int, item Type) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfPrefixIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfPrefixIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllPrefix combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllPrefix(checkers ...PrefixEnumChecker) PrefixEnumChecker {
	var all = AlwaysPrefixEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndPrefix{checkers[i], all}
	}
	return all
}

type enumOrPrefix struct {
	lhs, rhs PrefixEnumChecker
}

func (o enumOrPrefix) Check(n int, item Type) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfPrefixIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfPrefixIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyPrefix combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyPrefix(checkers ...PrefixEnumChecker) PrefixEnumChecker {
	var any = AlwaysPrefixEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrPrefix{checkers[i], any}
	}
	return any
}

// EnumFilteringPrefixIterator does iteration with
// filtering by previously set checker.
type EnumFilteringPrefixIterator struct {
	preparedPrefixItem
	filter PrefixEnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringPrefixIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedPrefixItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfPrefixIterator(err) {
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

// PrefixEnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func PrefixEnumFiltering(items PrefixIterator, filters ...PrefixEnumChecker) PrefixIterator {
	if items == nil {
		return EmptyPrefixIterator
	}
	return &EnumFilteringPrefixIterator{preparedPrefixItem{base: items}, EnumAllPrefix(filters...), 0}
}

// DoingUntilPrefixIterator does iteration
// until previously set checker is passed.
type DoingUntilPrefixIterator struct {
	preparedPrefixItem
	until PrefixChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilPrefixIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedPrefixItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfPrefixIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfPrefixIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// PrefixDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func PrefixDoingUntil(items PrefixIterator, untilList ...PrefixChecker) PrefixIterator {
	if items == nil {
		return EmptyPrefixIterator
	}
	return &DoingUntilPrefixIterator{preparedPrefixItem{base: items}, AllPrefix(untilList...)}
}

// PrefixSkipUntil sets until conditions to skip few items.
func PrefixSkipUntil(items PrefixIterator, untilList ...PrefixChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return PrefixDiscard(PrefixDoingUntil(items, untilList...))
}

// PrefixGettingBatch returns the next batch from items.
func PrefixGettingBatch(items PrefixIterator, batchSize int) PrefixIterator {
	if items == nil {
		return EmptyPrefixIterator
	}
	if batchSize == 0 {
		return items
	}

	size := 0
	return PrefixDoingUntil(items, PrefixCheck(func(item Type) (bool, error) {
		size++
		return size >= batchSize, nil
	}))
}
