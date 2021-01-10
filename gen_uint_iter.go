package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// UintIterator is an iterator over items type of uint.
type UintIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() uint
	// Err contains first met error while Next.
	Err() error
}

type emptyUintIterator struct{}

func (emptyUintIterator) HasNext() bool     { return false }
func (emptyUintIterator) Next() (next uint) { return 0 }
func (emptyUintIterator) Err() error        { return nil }

// EmptyUintIterator is a zero value for UintIterator.
// It is not contains any item to iterate over it.
var EmptyUintIterator UintIterator = emptyUintIterator{}

// UintIterMaker is a maker of UintIterator.
type UintIterMaker interface {
	// MakeIter should return a new instance of UintIterator to iterate over it.
	MakeIter() UintIterator
}

// MakeUintIter is a shortcut implementation
// of UintIterator based on a function.
type MakeUintIter func() UintIterator

// MakeIter returns a new instance of UintIterator to iterate over it.
func (m MakeUintIter) MakeIter() UintIterator { return m() }

// MakeNoUintIter is a zero value for UintIterMaker.
// It always returns EmptyUintIterator and an empty error.
var MakeNoUintIter UintIterMaker = MakeUintIter(
	func() UintIterator { return EmptyUintIterator })

// UintDiscard just range over all items and do nothing with each of them.
func UintDiscard(items UintIterator) error {
	if items == nil {
		items = EmptyUintIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

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

// UintConverter is an object converting an item type of uint.
type UintConverter interface {
	// Convert should convert an item type of uint into another item of uint.
	// It is suggested to return EndOfUintIterator to stop iteration.
	Convert(uint) (uint, error)
}

// UintConvert is a shortcut implementation
// of UintConverter based on a function.
type UintConvert func(uint) (uint, error)

// Convert converts an item type of uint into another item of uint.
// It is suggested to return EndOfUintIterator to stop iteration.
func (c UintConvert) Convert(item uint) (uint, error) { return c(item) }

// NoUintConvert does nothing with item, just returns it as is.
var NoUintConvert UintConverter = UintConvert(
	func(item uint) (uint, error) { return item, nil })

type doubleUintConverter struct {
	lhs, rhs UintConverter
}

func (c doubleUintConverter) Convert(item uint) (uint, error) {
	item, err := c.lhs.Convert(item)
	if err != nil {
		return 0, errors.Wrap(err, "convert lhs")
	}
	item, err = c.rhs.Convert(item)
	if err != nil {
		return 0, errors.Wrap(err, "convert rhs")
	}
	return item, nil
}

// UintConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func UintConverterSeries(converters ...UintConverter) UintConverter {
	var series = NoUintConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUintConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingUintIterator does iteration with
// converting by previously set converter.
type ConvertingUintIterator struct {
	preparedUintItem
	converter UintConverter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUintItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfUintIterator(err) {
				err = errors.Wrap(err, "converting iterator: check")
			}
			it.err = err
			return false
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// UintConverting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func UintConverting(items UintIterator, converters ...UintConverter) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	return &ConvertingUintIterator{
		preparedUintItem{base: items}, UintConverterSeries(converters...)}
}

// UintEnumConverter is an object converting an item type of uint and its ordering number.
type UintEnumConverter interface {
	// Convert should convert an item type of uint into another item of uint.
	// It is suggested to return EndOfUintIterator to stop iteration.
	Convert(n int, val uint) (uint, error)
}

// UintEnumConvert is a shortcut implementation
// of UintEnumConverter based on a function.
type UintEnumConvert func(int, uint) (uint, error)

// Convert converts an item type of uint into another item of uint.
// It is suggested to return EndOfUintIterator to stop iteration.
func (c UintEnumConvert) Convert(n int, item uint) (uint, error) { return c(n, item) }

// NoUintEnumConvert does nothing with item, just returns it as is.
var NoUintEnumConvert UintEnumConverter = UintEnumConvert(
	func(_ int, item uint) (uint, error) { return item, nil })

type enumFromUintConverter struct {
	UintConverter
}

func (ch enumFromUintConverter) Convert(_ int, item uint) (uint, error) {
	return ch.UintConverter.Convert(item)
}

// EnumFromUintConverter adapts checker type of UintConverter
// to the interface UintEnumConverter.
// If converter is nil it is return based on NoUintConvert enum checker.
func EnumFromUintConverter(converter UintConverter) UintEnumConverter {
	if converter == nil {
		converter = NoUintConvert
	}
	return &enumFromUintConverter{converter}
}

type doubleUintEnumConverter struct {
	lhs, rhs UintEnumConverter
}

func (c doubleUintEnumConverter) Convert(n int, item uint) (uint, error) {
	item, err := c.lhs.Convert(n, item)
	if err != nil {
		return 0, errors.Wrap(err, "convert lhs")
	}
	item, err = c.rhs.Convert(n, item)
	if err != nil {
		return 0, errors.Wrap(err, "convert rhs")
	}
	return item, nil
}

// EnumUintConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumUintConverterSeries(converters ...UintEnumConverter) UintEnumConverter {
	var series = NoUintEnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUintEnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingUintIterator does iteration with
// converting by previously set converter.
type EnumConvertingUintIterator struct {
	preparedUintItem
	converter UintEnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUintItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfUintIterator(err) {
				err = errors.Wrap(err, "converting iterator: check")
			}
			it.err = err
			return false
		}
		it.count++

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// UintEnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func UintEnumConverting(items UintIterator, converters ...UintEnumConverter) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	return &EnumConvertingUintIterator{
		preparedUintItem{base: items}, EnumUintConverterSeries(converters...), 0}
}

// UintHandler is an object handling an item type of uint.
type UintHandler interface {
	// Handle should do something with item of uint.
	// It is suggested to return EndOfUintIterator to stop iteration.
	Handle(uint) error
}

// UintHandle is a shortcut implementation
// of UintHandler based on a function.
type UintHandle func(uint) error

// Handle does something with item of uint.
// It is suggested to return EndOfUintIterator to stop iteration.
func (h UintHandle) Handle(item uint) error { return h(item) }

// UintDoNothing does nothing.
var UintDoNothing UintHandler = UintHandle(func(_ uint) error { return nil })

type doubleUintHandler struct {
	lhs, rhs UintHandler
}

func (h doubleUintHandler) Handle(item uint) error {
	err := h.lhs.Handle(item)
	if err != nil {
		return errors.Wrap(err, "handle lhs")
	}
	err = h.rhs.Handle(item)
	if err != nil {
		return errors.Wrap(err, "handle rhs")
	}
	return nil
}

// UintHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func UintHandlerSeries(handlers ...UintHandler) UintHandler {
	var series = UintDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUintHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingUintIterator does iteration with
// handling by previously set handler.
type HandlingUintIterator struct {
	preparedUintItem
	handler UintHandler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUintItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfUintIterator(err) {
				err = errors.Wrap(err, "handling iterator: check")
			}
			it.err = err
			return false
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// UintHandling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func UintHandling(items UintIterator, handlers ...UintHandler) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	return &HandlingUintIterator{
		preparedUintItem{base: items}, UintHandlerSeries(handlers...)}
}

// UintRange iterates over items and use handlers to each one.
func UintRange(items UintIterator, handlers ...UintHandler) error {
	return UintDiscard(UintHandling(items, handlers...))
}

// UintRangeIterator is an iterator over items.
type UintRangeIterator interface {
	// Range should iterate over items.
	Range(...UintHandler) error
}

type sUintRangeIterator struct {
	iter UintIterator
}

// ToUintRangeIterator constructs an instance implementing UintRangeIterator
// based on UintIterator.
func ToUintRangeIterator(iter UintIterator) UintRangeIterator {
	if iter == nil {
		iter = EmptyUintIterator
	}
	return sUintRangeIterator{iter: iter}
}

// MakeUintRangeIterator constructs an instance implementing UintRangeIterator
// based on UintIterMaker.
func MakeUintRangeIterator(maker UintIterMaker) UintRangeIterator {
	if maker == nil {
		maker = MakeNoUintIter
	}
	return ToUintRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sUintRangeIterator) Range(handlers ...UintHandler) error {
	return UintRange(r.iter, handlers...)
}

// UintEnumHandler is an object handling an item type of uint and its ordered number.
type UintEnumHandler interface {
	// Handle should do something with item of uint and its ordered number.
	// It is suggested to return EndOfUintIterator to stop iteration.
	Handle(int, uint) error
}

// UintEnumHandle is a shortcut implementation
// of UintEnumHandler based on a function.
type UintEnumHandle func(int, uint) error

// Handle does something with item of uint and its ordered number.
// It is suggested to return EndOfUintIterator to stop iteration.
func (h UintEnumHandle) Handle(n int, item uint) error { return h(n, item) }

// UintDoEnumNothing does nothing.
var UintDoEnumNothing = UintEnumHandle(func(_ int, _ uint) error { return nil })

type doubleUintEnumHandler struct {
	lhs, rhs UintEnumHandler
}

func (h doubleUintEnumHandler) Handle(n int, item uint) error {
	err := h.lhs.Handle(n, item)
	if err != nil {
		return errors.Wrap(err, "handle lhs")
	}
	err = h.rhs.Handle(n, item)
	if err != nil {
		return errors.Wrap(err, "handle rhs")
	}
	return nil
}

// UintEnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func UintEnumHandlerSeries(handlers ...UintEnumHandler) UintEnumHandler {
	var series UintEnumHandler = UintDoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUintEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingUintIterator does iteration with
// handling by previously set handler.
type EnumHandlingUintIterator struct {
	preparedUintItem
	handler UintEnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUintItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfUintIterator(err) {
				err = errors.Wrap(err, "enum handling iterator: check")
			}
			it.err = err
			return false
		}
		it.count++

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// UintEnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func UintEnumHandling(items UintIterator, handlers ...UintEnumHandler) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	return &EnumHandlingUintIterator{
		preparedUintItem{base: items}, UintEnumHandlerSeries(handlers...), 0}
}

// UintEnum iterates over items and their ordering numbers and use handlers to each one.
func UintEnum(items UintIterator, handlers ...UintEnumHandler) error {
	return UintDiscard(UintEnumHandling(items, handlers...))
}

// UintEnumIterator is an iterator over items and their ordering numbers.
type UintEnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...UintEnumHandler) error
}

type sUintEnumIterator struct {
	iter UintIterator
}

// ToUintEnumIterator constructs an instance implementing UintEnumIterator
// based on UintIterator.
func ToUintEnumIterator(iter UintIterator) UintEnumIterator {
	if iter == nil {
		iter = EmptyUintIterator
	}
	return sUintEnumIterator{iter: iter}
}

// MakeUintEnumIterator constructs an instance implementing UintEnumIterator
// based on UintIterMaker.
func MakeUintEnumIterator(maker UintIterMaker) UintEnumIterator {
	if maker == nil {
		maker = MakeNoUintIter
	}
	return ToUintEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sUintEnumIterator) Enum(handlers ...UintEnumHandler) error {
	return UintEnum(r.iter, handlers...)
}

// Range iterates over items.
func (r sUintEnumIterator) Range(handlers ...UintHandler) error {
	return UintRange(r.iter, handlers...)
}

type doubleUintIterator struct {
	lhs, rhs UintIterator
	inRHS    bool
}

func (it *doubleUintIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUintIterator) Next() uint {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUintIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUintIterator combines all iterators to one.
func SuperUintIterator(itemList ...UintIterator) UintIterator {
	var super = EmptyUintIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUintIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// UintComparer is a strategy to compare two types.
type UintComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs uint) bool
}

// UintCompare is a shortcut implementation
// of UintComparer based on a function.
type UintCompare func(lhs, rhs uint) bool

// IsLess is true if lhs is less than rhs.
func (c UintCompare) IsLess(lhs, rhs uint) bool { return c(lhs, rhs) }

// UintAlwaysLess is an implementation of UintComparer returning always true.
var UintAlwaysLess UintComparer = UintCompare(func(_, _ uint) bool { return true })

type priorityUintIterator struct {
	lhs, rhs preparedUintItem
	comparer UintComparer
}

func (it *priorityUintIterator) HasNext() bool {
	if it.lhs.hasNext && it.rhs.hasNext {
		return true
	}
	if !it.lhs.hasNext && it.lhs.HasNext() {
		next := it.lhs.base.Next()
		it.lhs.hasNext = true
		it.lhs.next = next
	}
	if !it.rhs.hasNext && it.rhs.HasNext() {
		next := it.rhs.base.Next()
		it.rhs.hasNext = true
		it.rhs.next = next
	}

	return it.lhs.hasNext || it.rhs.hasNext
}

func (it *priorityUintIterator) Next() uint {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUintIteratorError(
			errors.New("no next"), "priority: next")
	}

	if !it.lhs.hasNext {
		// it.rhs.hasNext == true
		return it.rhs.Next()
	}
	if !it.rhs.hasNext {
		// it.lhs.hasNext == true
		return it.lhs.Next()
	}

	// both have next
	lhsNext := it.lhs.Next()
	rhsNext := it.rhs.Next()
	if it.comparer.IsLess(lhsNext, rhsNext) {
		// remember rhsNext
		it.rhs.hasNext = true
		it.rhs.next = rhsNext
		return lhsNext
	}

	// rhsNext is less than or equal to lhsNext.
	// remember lhsNext
	it.lhs.hasNext = true
	it.lhs.next = lhsNext
	return rhsNext
}

func (it priorityUintIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUintIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUintIterator(comparer UintComparer, itemList ...UintIterator) UintIterator {
	if comparer == nil {
		comparer = UintAlwaysLess
	}

	var prior = EmptyUintIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUintIterator{
			lhs:      preparedUintItem{base: itemList[i]},
			rhs:      preparedUintItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// UintSliceIterator is an iterator based on a slice of uint.
type UintSliceIterator struct {
	slice []uint
	cur   int
}

// NewUintSliceIterator returns a new instance of UintSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use UintUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewUintSliceIterator(slice []uint) *UintSliceIterator {
	it := &UintSliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it UintSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *UintSliceIterator) Next() uint {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (UintSliceIterator) Err() error { return nil }

// UintSliceIterator is an iterator based on a slice of uint
// and doing iteration in back direction.
type InvertingUintSliceIterator struct {
	slice []uint
	cur   int
}

// NewInvertingUintSliceIterator returns a new instance of InvertingUintSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingUintSlice(UintUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingUintSliceIterator(slice []uint) *InvertingUintSliceIterator {
	it := &InvertingUintSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingUintSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingUintSliceIterator) Next() uint {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingUintSliceIterator) Err() error { return nil }

// UintUnroll unrolls items to slice of uint.
func UintUnroll(items UintIterator) UintSlice {
	var slice UintSlice
	panicIfUintIteratorError(UintDiscard(UintHandling(items, UintHandle(func(item uint) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// UintSlice is a slice of uint.
type UintSlice []uint

// MakeIter returns a new instance of UintIterator to iterate over it.
// It returns EmptyUintIterator if the error is not nil.
func (s UintSlice) MakeIter() UintIterator {
	return NewUintSliceIterator(s)
}

// UintSlice is a slice of uint which can make inverting iterator.
type InvertingUintSlice []uint

// MakeIter returns a new instance of UintIterator to iterate over it.
// It returns EmptyUintIterator if the error is not nil.
func (s InvertingUintSlice) MakeIter() UintIterator {
	return NewInvertingUintSliceIterator(s)
}

// UintInvert unrolls items and make inverting iterator based on them.
func UintInvert(items UintIterator) UintIterator {
	return InvertingUintSlice(UintUnroll(items)).MakeIter()
}

// EndOfUintIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfUintIterator = errors.New("end of uint iterator")

func isEndOfUintIterator(err error) bool {
	return errors.Is(err, EndOfUintIterator)
}

func wrapIfNotEndOfUintIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfUintIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfUintIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedUintItem struct {
	base    UintIterator
	hasNext bool
	next    uint
	err     error
}

func (it *preparedUintItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedUintItem) Next() uint {
	if !it.hasNext {
		panicIfUintIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedUintItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfUintIterator) {
		return it.err
	}
	return it.base.Err()
}
