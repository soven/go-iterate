package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// Uint64Iterator is an iterator over items type of uint64.
type Uint64Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() uint64
	// Err contains first met error while Next.
	Err() error
}

type emptyUint64Iterator struct{}

func (emptyUint64Iterator) HasNext() bool       { return false }
func (emptyUint64Iterator) Next() (next uint64) { return 0 }
func (emptyUint64Iterator) Err() error          { return nil }

// EmptyUint64Iterator is a zero value for Uint64Iterator.
// It is not contains any item to iterate over it.
var EmptyUint64Iterator Uint64Iterator = emptyUint64Iterator{}

// Uint64IterMaker is a maker of Uint64Iterator.
type Uint64IterMaker interface {
	// MakeIter should return a new instance of Uint64Iterator to iterate over it.
	MakeIter() Uint64Iterator
}

// MakeUint64Iter is a shortcut implementation
// of Uint64Iterator based on a function.
type MakeUint64Iter func() Uint64Iterator

// MakeIter returns a new instance of Uint64Iterator to iterate over it.
func (m MakeUint64Iter) MakeIter() Uint64Iterator { return m() }

// MakeNoUint64Iter is a zero value for Uint64IterMaker.
// It always returns EmptyUint64Iterator and an empty error.
var MakeNoUint64Iter Uint64IterMaker = MakeUint64Iter(
	func() Uint64Iterator { return EmptyUint64Iterator })

// Uint64Discard just range over all items and do nothing with each of them.
func Uint64Discard(items Uint64Iterator) error {
	if items == nil {
		items = EmptyUint64Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

// Uint64Checker is an object checking an item type of uint64
// for some condition.
type Uint64Checker interface {
	// Check should check an item type of uint64 for some condition.
	// It is suggested to return EndOfUint64Iterator to stop iteration.
	Check(uint64) (bool, error)
}

// Uint64Check is a shortcut implementation
// of Uint64Checker based on a function.
type Uint64Check func(uint64) (bool, error)

// Check checks an item type of uint64 for some condition.
// It returns EndOfUint64Iterator to stop iteration.
func (ch Uint64Check) Check(item uint64) (bool, error) { return ch(item) }

var (
	// AlwaysUint64CheckTrue always returns true and empty error.
	AlwaysUint64CheckTrue Uint64Checker = Uint64Check(
		func(item uint64) (bool, error) { return true, nil })
	// AlwaysUint64CheckFalse always returns false and empty error.
	AlwaysUint64CheckFalse Uint64Checker = Uint64Check(
		func(item uint64) (bool, error) { return false, nil })
)

// NotUint64 do an inversion for checker result.
// It is returns AlwaysUint64CheckTrue if checker is nil.
func NotUint64(checker Uint64Checker) Uint64Checker {
	if checker == nil {
		return AlwaysUint64CheckTrue
	}
	return Uint64Check(func(item uint64) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andUint64 struct {
	lhs, rhs Uint64Checker
}

func (a andUint64) Check(item uint64) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllUint64 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllUint64(checkers ...Uint64Checker) Uint64Checker {
	var all = AlwaysUint64CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andUint64{checkers[i], all}
	}
	return all
}

type orUint64 struct {
	lhs, rhs Uint64Checker
}

func (o orUint64) Check(item uint64) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyUint64 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyUint64(checkers ...Uint64Checker) Uint64Checker {
	var any = AlwaysUint64CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orUint64{checkers[i], any}
	}
	return any
}

// FilteringUint64Iterator does iteration with
// filtering by previously set checker.
type FilteringUint64Iterator struct {
	preparedUint64Item
	filter Uint64Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
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

// Uint64Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Uint64Filtering(items Uint64Iterator, filters ...Uint64Checker) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &FilteringUint64Iterator{preparedUint64Item{base: items}, AllUint64(filters...)}
}

// DoingUntilUint64Iterator does iteration
// until previously set checker is passed.
type DoingUntilUint64Iterator struct {
	preparedUint64Item
	until Uint64Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfUint64Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint64DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Uint64DoingUntil(items Uint64Iterator, untilList ...Uint64Checker) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	var until Uint64Checker
	if len(untilList) > 0 {
		until = AllUint64(untilList...)
	} else {
		until = AlwaysUint64CheckFalse
	}
	return &DoingUntilUint64Iterator{preparedUint64Item{base: items}, until}
}

// Uint64SkipUntil sets until conditions to skip few items.
func Uint64SkipUntil(items Uint64Iterator, untilList ...Uint64Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint64Discard(Uint64DoingUntil(items, untilList...))
}

// Uint64EnumChecker is an object checking an item type of uint64
// and its ordering number in for some condition.
type Uint64EnumChecker interface {
	// Check checks an item type of uint64 and its ordering number for some condition.
	// It is suggested to return EndOfUint64Iterator to stop iteration.
	Check(int, uint64) (bool, error)
}

// Uint64EnumCheck is a shortcut implementation
// of Uint64EnumChecker based on a function.
type Uint64EnumCheck func(int, uint64) (bool, error)

// Check checks an item type of uint64 and its ordering number for some condition.
// It returns EndOfUint64Iterator to stop iteration.
func (ch Uint64EnumCheck) Check(n int, item uint64) (bool, error) { return ch(n, item) }

type enumFromUint64Checker struct {
	Uint64Checker
}

func (ch enumFromUint64Checker) Check(_ int, item uint64) (bool, error) {
	return ch.Uint64Checker.Check(item)
}

// EnumFromUint64Checker adapts checker type of Uint64Checker
// to the interface Uint64EnumChecker.
// If checker is nil it is return based on AlwaysUint64CheckFalse enum checker.
func EnumFromUint64Checker(checker Uint64Checker) Uint64EnumChecker {
	if checker == nil {
		checker = AlwaysUint64CheckFalse
	}
	return &enumFromUint64Checker{checker}
}

var (
	// AlwaysUint64EnumCheckTrue always returns true and empty error.
	AlwaysUint64EnumCheckTrue = EnumFromUint64Checker(
		AlwaysUint64CheckTrue)
	// AlwaysUint64EnumCheckFalse always returns false and empty error.
	AlwaysUint64EnumCheckFalse = EnumFromUint64Checker(
		AlwaysUint64CheckFalse)
)

// EnumNotUint64 do an inversion for checker result.
// It is returns AlwaysUint64EnumCheckTrue if checker is nil.
func EnumNotUint64(checker Uint64EnumChecker) Uint64EnumChecker {
	if checker == nil {
		return AlwaysUint64EnumCheckTrue
	}
	return Uint64EnumCheck(func(n int, item uint64) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndUint64 struct {
	lhs, rhs Uint64EnumChecker
}

func (a enumAndUint64) Check(n int, item uint64) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllUint64 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllUint64(checkers ...Uint64EnumChecker) Uint64EnumChecker {
	var all = AlwaysUint64EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndUint64{checkers[i], all}
	}
	return all
}

type enumOrUint64 struct {
	lhs, rhs Uint64EnumChecker
}

func (o enumOrUint64) Check(n int, item uint64) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyUint64 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyUint64(checkers ...Uint64EnumChecker) Uint64EnumChecker {
	var any = AlwaysUint64EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrUint64{checkers[i], any}
	}
	return any
}

// EnumFilteringUint64Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringUint64Iterator struct {
	preparedUint64Item
	filter Uint64EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
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

// Uint64EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Uint64EnumFiltering(items Uint64Iterator, filters ...Uint64EnumChecker) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &EnumFilteringUint64Iterator{preparedUint64Item{base: items}, EnumAllUint64(filters...), 0}
}

// EnumDoingUntilUint64Iterator does iteration
// until previously set checker is passed.
type EnumDoingUntilUint64Iterator struct {
	preparedUint64Item
	until Uint64EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfUint64Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint64EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Uint64EnumDoingUntil(items Uint64Iterator, untilList ...Uint64EnumChecker) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	var until Uint64EnumChecker
	if len(untilList) > 0 {
		until = EnumAllUint64(untilList...)
	} else {
		until = AlwaysUint64EnumCheckFalse
	}
	return &EnumDoingUntilUint64Iterator{preparedUint64Item{base: items}, until, 0}
}

// Uint64EnumSkipUntil sets until conditions to skip few items.
func Uint64EnumSkipUntil(items Uint64Iterator, untilList ...Uint64EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint64Discard(Uint64EnumDoingUntil(items, untilList...))
}

// Uint64GettingBatch returns the next batch from items.
func Uint64GettingBatch(items Uint64Iterator, batchSize int) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	if batchSize == 0 {
		return items
	}

	return Uint64EnumDoingUntil(items, Uint64EnumCheck(func(n int, item uint64) (bool, error) {
		return n == batchSize-1, nil
	}))
}

// Uint64Converter is an object converting an item type of uint64.
type Uint64Converter interface {
	// Convert should convert an item type of uint64 into another item of uint64.
	// It is suggested to return EndOfUint64Iterator to stop iteration.
	Convert(uint64) (uint64, error)
}

// Uint64Convert is a shortcut implementation
// of Uint64Converter based on a function.
type Uint64Convert func(uint64) (uint64, error)

// Convert converts an item type of uint64 into another item of uint64.
// It is suggested to return EndOfUint64Iterator to stop iteration.
func (c Uint64Convert) Convert(item uint64) (uint64, error) { return c(item) }

// NoUint64Convert does nothing with item, just returns it as is.
var NoUint64Convert Uint64Converter = Uint64Convert(
	func(item uint64) (uint64, error) { return item, nil })

type doubleUint64Converter struct {
	lhs, rhs Uint64Converter
}

func (c doubleUint64Converter) Convert(item uint64) (uint64, error) {
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

// Uint64ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Uint64ConverterSeries(converters ...Uint64Converter) Uint64Converter {
	var series = NoUint64Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint64Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingUint64Iterator does iteration with
// converting by previously set converter.
type ConvertingUint64Iterator struct {
	preparedUint64Item
	converter Uint64Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
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

// Uint64Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Uint64Converting(items Uint64Iterator, converters ...Uint64Converter) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &ConvertingUint64Iterator{
		preparedUint64Item{base: items}, Uint64ConverterSeries(converters...)}
}

// Uint64EnumConverter is an object converting an item type of uint64 and its ordering number.
type Uint64EnumConverter interface {
	// Convert should convert an item type of uint64 into another item of uint64.
	// It is suggested to return EndOfUint64Iterator to stop iteration.
	Convert(n int, val uint64) (uint64, error)
}

// Uint64EnumConvert is a shortcut implementation
// of Uint64EnumConverter based on a function.
type Uint64EnumConvert func(int, uint64) (uint64, error)

// Convert converts an item type of uint64 into another item of uint64.
// It is suggested to return EndOfUint64Iterator to stop iteration.
func (c Uint64EnumConvert) Convert(n int, item uint64) (uint64, error) { return c(n, item) }

// NoUint64EnumConvert does nothing with item, just returns it as is.
var NoUint64EnumConvert Uint64EnumConverter = Uint64EnumConvert(
	func(_ int, item uint64) (uint64, error) { return item, nil })

type enumFromUint64Converter struct {
	Uint64Converter
}

func (ch enumFromUint64Converter) Convert(_ int, item uint64) (uint64, error) {
	return ch.Uint64Converter.Convert(item)
}

// EnumFromUint64Converter adapts checker type of Uint64Converter
// to the interface Uint64EnumConverter.
// If converter is nil it is return based on NoUint64Convert enum checker.
func EnumFromUint64Converter(converter Uint64Converter) Uint64EnumConverter {
	if converter == nil {
		converter = NoUint64Convert
	}
	return &enumFromUint64Converter{converter}
}

type doubleUint64EnumConverter struct {
	lhs, rhs Uint64EnumConverter
}

func (c doubleUint64EnumConverter) Convert(n int, item uint64) (uint64, error) {
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

// EnumUint64ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumUint64ConverterSeries(converters ...Uint64EnumConverter) Uint64EnumConverter {
	var series = NoUint64EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint64EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingUint64Iterator does iteration with
// converting by previously set converter.
type EnumConvertingUint64Iterator struct {
	preparedUint64Item
	converter Uint64EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
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

// Uint64EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Uint64EnumConverting(items Uint64Iterator, converters ...Uint64EnumConverter) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &EnumConvertingUint64Iterator{
		preparedUint64Item{base: items}, EnumUint64ConverterSeries(converters...), 0}
}

// Uint64Handler is an object handling an item type of uint64.
type Uint64Handler interface {
	// Handle should do something with item of uint64.
	// It is suggested to return EndOfUint64Iterator to stop iteration.
	Handle(uint64) error
}

// Uint64Handle is a shortcut implementation
// of Uint64Handler based on a function.
type Uint64Handle func(uint64) error

// Handle does something with item of uint64.
// It is suggested to return EndOfUint64Iterator to stop iteration.
func (h Uint64Handle) Handle(item uint64) error { return h(item) }

// Uint64DoNothing does nothing.
var Uint64DoNothing Uint64Handler = Uint64Handle(func(_ uint64) error { return nil })

type doubleUint64Handler struct {
	lhs, rhs Uint64Handler
}

func (h doubleUint64Handler) Handle(item uint64) error {
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

// Uint64HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint64HandlerSeries(handlers ...Uint64Handler) Uint64Handler {
	var series = Uint64DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint64Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingUint64Iterator does iteration with
// handling by previously set handler.
type HandlingUint64Iterator struct {
	preparedUint64Item
	handler Uint64Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
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

// Uint64Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Uint64Handling(items Uint64Iterator, handlers ...Uint64Handler) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &HandlingUint64Iterator{
		preparedUint64Item{base: items}, Uint64HandlerSeries(handlers...)}
}

// Uint64Range iterates over items and use handlers to each one.
func Uint64Range(items Uint64Iterator, handlers ...Uint64Handler) error {
	return Uint64Discard(Uint64Handling(items, handlers...))
}

// Uint64RangeIterator is an iterator over items.
type Uint64RangeIterator interface {
	// Range should iterate over items.
	Range(...Uint64Handler) error
}

type sUint64RangeIterator struct {
	iter Uint64Iterator
}

// ToUint64RangeIterator constructs an instance implementing Uint64RangeIterator
// based on Uint64Iterator.
func ToUint64RangeIterator(iter Uint64Iterator) Uint64RangeIterator {
	if iter == nil {
		iter = EmptyUint64Iterator
	}
	return sUint64RangeIterator{iter: iter}
}

// MakeUint64RangeIterator constructs an instance implementing Uint64RangeIterator
// based on Uint64IterMaker.
func MakeUint64RangeIterator(maker Uint64IterMaker) Uint64RangeIterator {
	if maker == nil {
		maker = MakeNoUint64Iter
	}
	return ToUint64RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sUint64RangeIterator) Range(handlers ...Uint64Handler) error {
	return Uint64Range(r.iter, handlers...)
}

// Uint64EnumHandler is an object handling an item type of uint64 and its ordered number.
type Uint64EnumHandler interface {
	// Handle should do something with item of uint64 and its ordered number.
	// It is suggested to return EndOfUint64Iterator to stop iteration.
	Handle(int, uint64) error
}

// Uint64EnumHandle is a shortcut implementation
// of Uint64EnumHandler based on a function.
type Uint64EnumHandle func(int, uint64) error

// Handle does something with item of uint64 and its ordered number.
// It is suggested to return EndOfUint64Iterator to stop iteration.
func (h Uint64EnumHandle) Handle(n int, item uint64) error { return h(n, item) }

// Uint64DoEnumNothing does nothing.
var Uint64DoEnumNothing = Uint64EnumHandle(func(_ int, _ uint64) error { return nil })

type doubleUint64EnumHandler struct {
	lhs, rhs Uint64EnumHandler
}

func (h doubleUint64EnumHandler) Handle(n int, item uint64) error {
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

// Uint64EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint64EnumHandlerSeries(handlers ...Uint64EnumHandler) Uint64EnumHandler {
	var series Uint64EnumHandler = Uint64DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint64EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingUint64Iterator does iteration with
// handling by previously set handler.
type EnumHandlingUint64Iterator struct {
	preparedUint64Item
	handler Uint64EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
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

// Uint64EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Uint64EnumHandling(items Uint64Iterator, handlers ...Uint64EnumHandler) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &EnumHandlingUint64Iterator{
		preparedUint64Item{base: items}, Uint64EnumHandlerSeries(handlers...), 0}
}

// Uint64Enum iterates over items and their ordering numbers and use handlers to each one.
func Uint64Enum(items Uint64Iterator, handlers ...Uint64EnumHandler) error {
	return Uint64Discard(Uint64EnumHandling(items, handlers...))
}

// Uint64EnumIterator is an iterator over items and their ordering numbers.
type Uint64EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Uint64EnumHandler) error
}

type sUint64EnumIterator struct {
	iter Uint64Iterator
}

// ToUint64EnumIterator constructs an instance implementing Uint64EnumIterator
// based on Uint64Iterator.
func ToUint64EnumIterator(iter Uint64Iterator) Uint64EnumIterator {
	if iter == nil {
		iter = EmptyUint64Iterator
	}
	return sUint64EnumIterator{iter: iter}
}

// MakeUint64EnumIterator constructs an instance implementing Uint64EnumIterator
// based on Uint64IterMaker.
func MakeUint64EnumIterator(maker Uint64IterMaker) Uint64EnumIterator {
	if maker == nil {
		maker = MakeNoUint64Iter
	}
	return ToUint64EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sUint64EnumIterator) Enum(handlers ...Uint64EnumHandler) error {
	return Uint64Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sUint64EnumIterator) Range(handlers ...Uint64Handler) error {
	return Uint64Range(r.iter, handlers...)
}

type doubleUint64Iterator struct {
	lhs, rhs Uint64Iterator
	inRHS    bool
}

func (it *doubleUint64Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint64Iterator) Next() uint64 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint64Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUint64Iterator combines all iterators to one.
func SuperUint64Iterator(itemList ...Uint64Iterator) Uint64Iterator {
	var super = EmptyUint64Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint64Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Uint64EnumComparer is a strategy to compare two types.
type Uint64Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs uint64) bool
}

// Uint64Compare is a shortcut implementation
// of Uint64EnumComparer based on a function.
type Uint64Compare func(lhs, rhs uint64) bool

// IsLess is true if lhs is less than rhs.
func (c Uint64Compare) IsLess(lhs, rhs uint64) bool { return c(lhs, rhs) }

// EnumUint64AlwaysLess is an implementation of Uint64EnumComparer returning always true.
var Uint64AlwaysLess Uint64Comparer = Uint64Compare(func(_, _ uint64) bool { return true })

type priorityUint64Iterator struct {
	lhs, rhs preparedUint64Item
	comparer Uint64Comparer
}

func (it *priorityUint64Iterator) HasNext() bool {
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

func (it *priorityUint64Iterator) Next() uint64 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUint64IteratorError(
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

func (it priorityUint64Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUint64Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUint64Iterator(comparer Uint64Comparer, itemList ...Uint64Iterator) Uint64Iterator {
	if comparer == nil {
		comparer = Uint64AlwaysLess
	}

	var prior = EmptyUint64Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUint64Iterator{
			lhs:      preparedUint64Item{base: itemList[i]},
			rhs:      preparedUint64Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Uint64EnumComparer is a strategy to compare two types and their order numbers.
type Uint64EnumComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(nLHS int, lhs uint64, nRHS int, rhs uint64) bool
}

// Uint64EnumCompare is a shortcut implementation
// of Uint64EnumComparer based on a function.
type Uint64EnumCompare func(nLHS int, lhs uint64, nRHS int, rhs uint64) bool

// IsLess is true if lhs is less than rhs.
func (c Uint64EnumCompare) IsLess(nLHS int, lhs uint64, nRHS int, rhs uint64) bool {
	return c(nLHS, lhs, nRHS, rhs)
}

// EnumUint64AlwaysLess is an implementation of Uint64EnumComparer returning always true.
var EnumUint64AlwaysLess Uint64EnumComparer = Uint64EnumCompare(
	func(_ int, _ uint64, _ int, _ uint64) bool { return true })

type priorityUint64EnumIterator struct {
	lhs, rhs           preparedUint64Item
	countLHS, countRHS int
	comparer           Uint64EnumComparer
}

func (it *priorityUint64EnumIterator) HasNext() bool {
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

func (it *priorityUint64EnumIterator) Next() uint64 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUint64IteratorError(
			errors.New("no next"), "priority enum: next")
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
	if it.comparer.IsLess(it.countLHS, lhsNext, it.countRHS, rhsNext) {
		// remember rhsNext
		it.rhs.hasNext = true
		it.rhs.next = rhsNext
		it.countLHS++
		return lhsNext
	}

	// rhsNext is less than or equal to lhsNext.
	// remember lhsNext
	it.lhs.hasNext = true
	it.lhs.next = lhsNext
	it.countRHS++
	return rhsNext
}

func (it priorityUint64EnumIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUint64EnumIterator compare one by one items and their ordering numbers fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUint64EnumIterator(comparer Uint64EnumComparer, itemList ...Uint64Iterator) Uint64Iterator {
	if comparer == nil {
		comparer = EnumUint64AlwaysLess
	}

	var prior = EmptyUint64Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUint64EnumIterator{
			lhs:      preparedUint64Item{base: itemList[i]},
			rhs:      preparedUint64Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Uint64SliceIterator is an iterator based on a slice of uint64.
type Uint64SliceIterator struct {
	slice []uint64
	cur   int
}

// NewUint64SliceIterator returns a new instance of Uint64SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Uint64Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewUint64SliceIterator(slice []uint64) *Uint64SliceIterator {
	it := &Uint64SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Uint64SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Uint64SliceIterator) Next() uint64 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Uint64SliceIterator) Err() error { return nil }

// Uint64SliceIterator is an iterator based on a slice of uint64
// and doing iteration in back direction.
type InvertingUint64SliceIterator struct {
	slice []uint64
	cur   int
}

// NewInvertingUint64SliceIterator returns a new instance of InvertingUint64SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingUint64Slice(Uint64Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingUint64SliceIterator(slice []uint64) *InvertingUint64SliceIterator {
	it := &InvertingUint64SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingUint64SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingUint64SliceIterator) Next() uint64 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingUint64SliceIterator) Err() error { return nil }

// Uint64Unroll unrolls items to slice of uint64.
func Uint64Unroll(items Uint64Iterator) Uint64Slice {
	var slice Uint64Slice
	panicIfUint64IteratorError(Uint64Discard(Uint64Handling(items, Uint64Handle(func(item uint64) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Uint64Slice is a slice of uint64.
type Uint64Slice []uint64

// MakeIter returns a new instance of Uint64Iterator to iterate over it.
// It returns EmptyUint64Iterator if the error is not nil.
func (s Uint64Slice) MakeIter() Uint64Iterator {
	return NewUint64SliceIterator(s)
}

// Uint64Slice is a slice of uint64 which can make inverting iterator.
type InvertingUint64Slice []uint64

// MakeIter returns a new instance of Uint64Iterator to iterate over it.
// It returns EmptyUint64Iterator if the error is not nil.
func (s InvertingUint64Slice) MakeIter() Uint64Iterator {
	return NewInvertingUint64SliceIterator(s)
}

// Uint64Invert unrolls items and make inverting iterator based on them.
func Uint64Invert(items Uint64Iterator) Uint64Iterator {
	return InvertingUint64Slice(Uint64Unroll(items)).MakeIter()
}

// EndOfUint64Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfUint64Iterator = errors.New("end of uint64 iterator")

func isEndOfUint64Iterator(err error) bool {
	return errors.Is(err, EndOfUint64Iterator)
}

func wrapIfNotEndOfUint64Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfUint64Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfUint64IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedUint64Item struct {
	base    Uint64Iterator
	hasNext bool
	next    uint64
	err     error
}

func (it *preparedUint64Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedUint64Item) Next() uint64 {
	if !it.hasNext {
		panicIfUint64IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedUint64Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfUint64Iterator) {
		return it.err
	}
	return it.base.Err()
}
