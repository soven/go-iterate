package iter

import "github.com/pkg/errors"

// ByteChecker is an object checking an item type of byte
// for some condition.
type ByteChecker interface {
	// Check should check an item type of byte for some condition.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Check(byte) (bool, error)
}

// ByteCheck is a shortcut implementation
// of ByteChecker based on a function.
type ByteCheck func(byte) (bool, error)

// Check checks an item type of byte for some condition.
// It returns EndOfByteIterator to stop iteration.
func (ch ByteCheck) Check(item byte) (bool, error) { return ch(item) }

var (
	// AlwaysByteCheckTrue always returns true and empty error.
	AlwaysByteCheckTrue ByteChecker = ByteCheck(
		func(item byte) (bool, error) { return true, nil })
	// AlwaysByteCheckFalse always returns false and empty error.
	AlwaysByteCheckFalse ByteChecker = ByteCheck(
		func(item byte) (bool, error) { return false, nil })
)

// NotByte do an inversion for checker result.
// It is returns AlwaysByteCheckTrue if checker is nil.
func NotByte(checker ByteChecker) ByteChecker {
	if checker == nil {
		return AlwaysByteCheckTrue
	}
	return ByteCheck(func(item byte) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andByte struct {
	lhs, rhs ByteChecker
}

func (a andByte) Check(item byte) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllByte combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllByte(checkers ...ByteChecker) ByteChecker {
	var all = AlwaysByteCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andByte{checkers[i], all}
	}
	return all
}

type orByte struct {
	lhs, rhs ByteChecker
}

func (o orByte) Check(item byte) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyByte combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyByte(checkers ...ByteChecker) ByteChecker {
	var any = AlwaysByteCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orByte{checkers[i], any}
	}
	return any
}

// FilteringByteIterator does iteration with
// filtering by previously set checker.
type FilteringByteIterator struct {
	preparedByteItem
	filter ByteChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedByteItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteFiltering sets filter while iterating over items.
// If filters is empty, so all items will return.
func ByteFiltering(items ByteIterator, filters ...ByteChecker) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &FilteringByteIterator{preparedByteItem{base: items}, AllByte(filters...)}
}

// DoingUntilByteIterator does iteration
// until previously set checker is passed.
type DoingUntilByteIterator struct {
	preparedByteItem
	until ByteChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedByteItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfByteIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfByteIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// ByteDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func ByteDoingUntil(items ByteIterator, untilList ...ByteChecker) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	var until ByteChecker
	if len(untilList) > 0 {
		until = AllByte(untilList...)
	} else {
		until = AlwaysByteCheckFalse
	}
	return &DoingUntilByteIterator{preparedByteItem{base: items}, until}
}

// ByteSkipUntil sets until conditions to skip few items.
func ByteSkipUntil(items ByteIterator, untilList ...ByteChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return ByteDiscard(ByteDoingUntil(items, untilList...))
}

// ByteEnumChecker is an object checking an item type of byte
// and its ordering number in for some condition.
type ByteEnumChecker interface {
	// Check checks an item type of byte and its ordering number for some condition.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Check(int, byte) (bool, error)
}

// ByteEnumCheck is a shortcut implementation
// of ByteEnumChecker based on a function.
type ByteEnumCheck func(int, byte) (bool, error)

// Check checks an item type of byte and its ordering number for some condition.
// It returns EndOfByteIterator to stop iteration.
func (ch ByteEnumCheck) Check(n int, item byte) (bool, error) { return ch(n, item) }

type enumFromByteChecker struct {
	ByteChecker
}

func (ch enumFromByteChecker) Check(_ int, item byte) (bool, error) {
	return ch.ByteChecker.Check(item)
}

// EnumFromByteChecker adapts checker type of ByteChecker
// to the interface ByteEnumChecker.
// If checker is nil it is return based on AlwaysByteCheckFalse enum checker.
func EnumFromByteChecker(checker ByteChecker) ByteEnumChecker {
	if checker == nil {
		checker = AlwaysByteCheckFalse
	}
	return &enumFromByteChecker{checker}
}

var (
	// AlwaysByteEnumCheckTrue always returns true and empty error.
	AlwaysByteEnumCheckTrue = EnumFromByteChecker(
		AlwaysByteCheckTrue)
	// AlwaysByteEnumCheckFalse always returns false and empty error.
	AlwaysByteEnumCheckFalse = EnumFromByteChecker(
		AlwaysByteCheckFalse)
)

// EnumNotByte do an inversion for checker result.
// It is returns AlwaysByteEnumCheckTrue if checker is nil.
func EnumNotByte(checker ByteEnumChecker) ByteEnumChecker {
	if checker == nil {
		return AlwaysByteEnumCheckTrue
	}
	return ByteEnumCheck(func(n int, item byte) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndByte struct {
	lhs, rhs ByteEnumChecker
}

func (a enumAndByte) Check(n int, item byte) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllByte combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllByte(checkers ...ByteEnumChecker) ByteEnumChecker {
	var all = AlwaysByteEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndByte{checkers[i], all}
	}
	return all
}

type enumOrByte struct {
	lhs, rhs ByteEnumChecker
}

func (o enumOrByte) Check(n int, item byte) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyByte combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyByte(checkers ...ByteEnumChecker) ByteEnumChecker {
	var any = AlwaysByteEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrByte{checkers[i], any}
	}
	return any
}

// EnumFilteringByteIterator does iteration with
// filtering by previously set checker.
type EnumFilteringByteIterator struct {
	preparedByteItem
	filter ByteEnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedByteItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteEnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func ByteEnumFiltering(items ByteIterator, filters ...ByteEnumChecker) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &EnumFilteringByteIterator{preparedByteItem{base: items}, EnumAllByte(filters...), 0}
}

// EnumDoingUntilByteIterator does iteration
// until previously set checker is passed.
type EnumDoingUntilByteIterator struct {
	preparedByteItem
	until ByteEnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedByteItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfByteIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfByteIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// ByteEnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func ByteEnumDoingUntil(items ByteIterator, untilList ...ByteEnumChecker) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	var until ByteEnumChecker
	if len(untilList) > 0 {
		until = EnumAllByte(untilList...)
	} else {
		until = AlwaysByteEnumCheckFalse
	}
	return &EnumDoingUntilByteIterator{preparedByteItem{base: items}, until, 0}
}

// ByteEnumSkipUntil sets until conditions to skip few items.
func ByteEnumSkipUntil(items ByteIterator, untilList ...ByteEnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return ByteDiscard(ByteEnumDoingUntil(items, untilList...))
}

// ByteGettingBatch returns the next batch from items.
func ByteGettingBatch(items ByteIterator, batchSize int) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	if batchSize == 0 {
		return items
	}

	return ByteEnumDoingUntil(items, ByteEnumCheck(func(n int, item byte) (bool, error) {
		return n == batchSize-1, nil
	}))
}
