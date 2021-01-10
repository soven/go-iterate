package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// Uint8Iterator is an iterator over items type of uint8.
type Uint8Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() uint8
	// Err contains first met error while Next.
	Err() error
}

type emptyUint8Iterator struct{}

func (emptyUint8Iterator) HasNext() bool      { return false }
func (emptyUint8Iterator) Next() (next uint8) { return 0 }
func (emptyUint8Iterator) Err() error         { return nil }

// EmptyUint8Iterator is a zero value for Uint8Iterator.
// It is not contains any item to iterate over it.
var EmptyUint8Iterator Uint8Iterator = emptyUint8Iterator{}

// Uint8IterMaker is a maker of Uint8Iterator.
type Uint8IterMaker interface {
	// MakeIter should return a new instance of Uint8Iterator to iterate over it.
	MakeIter() Uint8Iterator
}

// MakeUint8Iter is a shortcut implementation
// of Uint8Iterator based on a function.
type MakeUint8Iter func() Uint8Iterator

// MakeIter returns a new instance of Uint8Iterator to iterate over it.
func (m MakeUint8Iter) MakeIter() Uint8Iterator { return m() }

// MakeNoUint8Iter is a zero value for Uint8IterMaker.
// It always returns EmptyUint8Iterator and an empty error.
var MakeNoUint8Iter Uint8IterMaker = MakeUint8Iter(
	func() Uint8Iterator { return EmptyUint8Iterator })

// Uint8Discard just range over all items and do nothing with each of them.
func Uint8Discard(items Uint8Iterator) error {
	if items == nil {
		items = EmptyUint8Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

// Uint8Checker is an object checking an item type of uint8
// for some condition.
type Uint8Checker interface {
	// Check should check an item type of uint8 for some condition.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Check(uint8) (bool, error)
}

// Uint8Check is a shortcut implementation
// of Uint8Checker based on a function.
type Uint8Check func(uint8) (bool, error)

// Check checks an item type of uint8 for some condition.
// It returns EndOfUint8Iterator to stop iteration.
func (ch Uint8Check) Check(item uint8) (bool, error) { return ch(item) }

var (
	// AlwaysUint8CheckTrue always returns true and empty error.
	AlwaysUint8CheckTrue Uint8Checker = Uint8Check(
		func(item uint8) (bool, error) { return true, nil })
	// AlwaysUint8CheckFalse always returns false and empty error.
	AlwaysUint8CheckFalse Uint8Checker = Uint8Check(
		func(item uint8) (bool, error) { return false, nil })
)

// NotUint8 do an inversion for checker result.
// It is returns AlwaysUint8CheckTrue if checker is nil.
func NotUint8(checker Uint8Checker) Uint8Checker {
	if checker == nil {
		return AlwaysUint8CheckTrue
	}
	return Uint8Check(func(item uint8) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andUint8 struct {
	lhs, rhs Uint8Checker
}

func (a andUint8) Check(item uint8) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllUint8 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllUint8(checkers ...Uint8Checker) Uint8Checker {
	var all = AlwaysUint8CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andUint8{checkers[i], all}
	}
	return all
}

type orUint8 struct {
	lhs, rhs Uint8Checker
}

func (o orUint8) Check(item uint8) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyUint8 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyUint8(checkers ...Uint8Checker) Uint8Checker {
	var any = AlwaysUint8CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orUint8{checkers[i], any}
	}
	return any
}

// FilteringUint8Iterator does iteration with
// filtering by previously set checker.
type FilteringUint8Iterator struct {
	preparedUint8Item
	filter Uint8Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
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

// Uint8Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Uint8Filtering(items Uint8Iterator, filters ...Uint8Checker) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &FilteringUint8Iterator{preparedUint8Item{base: items}, AllUint8(filters...)}
}

// DoingUntilUint8Iterator does iteration
// until previously set checker is passed.
type DoingUntilUint8Iterator struct {
	preparedUint8Item
	until Uint8Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfUint8Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint8DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Uint8DoingUntil(items Uint8Iterator, untilList ...Uint8Checker) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	var until Uint8Checker
	if len(untilList) > 0 {
		until = AllUint8(untilList...)
	} else {
		until = AlwaysUint8CheckFalse
	}
	return &DoingUntilUint8Iterator{preparedUint8Item{base: items}, until}
}

// Uint8SkipUntil sets until conditions to skip few items.
func Uint8SkipUntil(items Uint8Iterator, untilList ...Uint8Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint8Discard(Uint8DoingUntil(items, untilList...))
}

// Uint8EnumChecker is an object checking an item type of uint8
// and its ordering number in for some condition.
type Uint8EnumChecker interface {
	// Check checks an item type of uint8 and its ordering number for some condition.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Check(int, uint8) (bool, error)
}

// Uint8EnumCheck is a shortcut implementation
// of Uint8EnumChecker based on a function.
type Uint8EnumCheck func(int, uint8) (bool, error)

// Check checks an item type of uint8 and its ordering number for some condition.
// It returns EndOfUint8Iterator to stop iteration.
func (ch Uint8EnumCheck) Check(n int, item uint8) (bool, error) { return ch(n, item) }

type enumFromUint8Checker struct {
	Uint8Checker
}

func (ch enumFromUint8Checker) Check(_ int, item uint8) (bool, error) {
	return ch.Uint8Checker.Check(item)
}

// EnumFromUint8Checker adapts checker type of Uint8Checker
// to the interface Uint8EnumChecker.
// If checker is nil it is return based on AlwaysUint8CheckFalse enum checker.
func EnumFromUint8Checker(checker Uint8Checker) Uint8EnumChecker {
	if checker == nil {
		checker = AlwaysUint8CheckFalse
	}
	return &enumFromUint8Checker{checker}
}

var (
	// AlwaysUint8EnumCheckTrue always returns true and empty error.
	AlwaysUint8EnumCheckTrue = EnumFromUint8Checker(
		AlwaysUint8CheckTrue)
	// AlwaysUint8EnumCheckFalse always returns false and empty error.
	AlwaysUint8EnumCheckFalse = EnumFromUint8Checker(
		AlwaysUint8CheckFalse)
)

// EnumNotUint8 do an inversion for checker result.
// It is returns AlwaysUint8EnumCheckTrue if checker is nil.
func EnumNotUint8(checker Uint8EnumChecker) Uint8EnumChecker {
	if checker == nil {
		return AlwaysUint8EnumCheckTrue
	}
	return Uint8EnumCheck(func(n int, item uint8) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndUint8 struct {
	lhs, rhs Uint8EnumChecker
}

func (a enumAndUint8) Check(n int, item uint8) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllUint8 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllUint8(checkers ...Uint8EnumChecker) Uint8EnumChecker {
	var all = AlwaysUint8EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndUint8{checkers[i], all}
	}
	return all
}

type enumOrUint8 struct {
	lhs, rhs Uint8EnumChecker
}

func (o enumOrUint8) Check(n int, item uint8) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfUint8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyUint8 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyUint8(checkers ...Uint8EnumChecker) Uint8EnumChecker {
	var any = AlwaysUint8EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrUint8{checkers[i], any}
	}
	return any
}

// EnumFilteringUint8Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringUint8Iterator struct {
	preparedUint8Item
	filter Uint8EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
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

// Uint8EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Uint8EnumFiltering(items Uint8Iterator, filters ...Uint8EnumChecker) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &EnumFilteringUint8Iterator{preparedUint8Item{base: items}, EnumAllUint8(filters...), 0}
}

// EnumDoingUntilUint8Iterator does iteration
// until previously set checker is passed.
type EnumDoingUntilUint8Iterator struct {
	preparedUint8Item
	until Uint8EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfUint8Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Uint8EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Uint8EnumDoingUntil(items Uint8Iterator, untilList ...Uint8EnumChecker) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	var until Uint8EnumChecker
	if len(untilList) > 0 {
		until = EnumAllUint8(untilList...)
	} else {
		until = AlwaysUint8EnumCheckFalse
	}
	return &EnumDoingUntilUint8Iterator{preparedUint8Item{base: items}, until, 0}
}

// Uint8EnumSkipUntil sets until conditions to skip few items.
func Uint8EnumSkipUntil(items Uint8Iterator, untilList ...Uint8EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Uint8Discard(Uint8EnumDoingUntil(items, untilList...))
}

// Uint8GettingBatch returns the next batch from items.
func Uint8GettingBatch(items Uint8Iterator, batchSize int) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	if batchSize == 0 {
		return items
	}

	return Uint8EnumDoingUntil(items, Uint8EnumCheck(func(n int, item uint8) (bool, error) {
		return n == batchSize-1, nil
	}))
}

// Uint8Converter is an object converting an item type of uint8.
type Uint8Converter interface {
	// Convert should convert an item type of uint8 into another item of uint8.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Convert(uint8) (uint8, error)
}

// Uint8Convert is a shortcut implementation
// of Uint8Converter based on a function.
type Uint8Convert func(uint8) (uint8, error)

// Convert converts an item type of uint8 into another item of uint8.
// It is suggested to return EndOfUint8Iterator to stop iteration.
func (c Uint8Convert) Convert(item uint8) (uint8, error) { return c(item) }

// NoUint8Convert does nothing with item, just returns it as is.
var NoUint8Convert Uint8Converter = Uint8Convert(
	func(item uint8) (uint8, error) { return item, nil })

type doubleUint8Converter struct {
	lhs, rhs Uint8Converter
}

func (c doubleUint8Converter) Convert(item uint8) (uint8, error) {
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

// Uint8ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Uint8ConverterSeries(converters ...Uint8Converter) Uint8Converter {
	var series = NoUint8Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint8Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingUint8Iterator does iteration with
// converting by previously set converter.
type ConvertingUint8Iterator struct {
	preparedUint8Item
	converter Uint8Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
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

// Uint8Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Uint8Converting(items Uint8Iterator, converters ...Uint8Converter) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &ConvertingUint8Iterator{
		preparedUint8Item{base: items}, Uint8ConverterSeries(converters...)}
}

// Uint8EnumConverter is an object converting an item type of uint8 and its ordering number.
type Uint8EnumConverter interface {
	// Convert should convert an item type of uint8 into another item of uint8.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Convert(n int, val uint8) (uint8, error)
}

// Uint8EnumConvert is a shortcut implementation
// of Uint8EnumConverter based on a function.
type Uint8EnumConvert func(int, uint8) (uint8, error)

// Convert converts an item type of uint8 into another item of uint8.
// It is suggested to return EndOfUint8Iterator to stop iteration.
func (c Uint8EnumConvert) Convert(n int, item uint8) (uint8, error) { return c(n, item) }

// NoUint8EnumConvert does nothing with item, just returns it as is.
var NoUint8EnumConvert Uint8EnumConverter = Uint8EnumConvert(
	func(_ int, item uint8) (uint8, error) { return item, nil })

type enumFromUint8Converter struct {
	Uint8Converter
}

func (ch enumFromUint8Converter) Convert(_ int, item uint8) (uint8, error) {
	return ch.Uint8Converter.Convert(item)
}

// EnumFromUint8Converter adapts checker type of Uint8Converter
// to the interface Uint8EnumConverter.
// If converter is nil it is return based on NoUint8Convert enum checker.
func EnumFromUint8Converter(converter Uint8Converter) Uint8EnumConverter {
	if converter == nil {
		converter = NoUint8Convert
	}
	return &enumFromUint8Converter{converter}
}

type doubleUint8EnumConverter struct {
	lhs, rhs Uint8EnumConverter
}

func (c doubleUint8EnumConverter) Convert(n int, item uint8) (uint8, error) {
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

// EnumUint8ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumUint8ConverterSeries(converters ...Uint8EnumConverter) Uint8EnumConverter {
	var series = NoUint8EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint8EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingUint8Iterator does iteration with
// converting by previously set converter.
type EnumConvertingUint8Iterator struct {
	preparedUint8Item
	converter Uint8EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
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

// Uint8EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Uint8EnumConverting(items Uint8Iterator, converters ...Uint8EnumConverter) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &EnumConvertingUint8Iterator{
		preparedUint8Item{base: items}, EnumUint8ConverterSeries(converters...), 0}
}

// Uint8Handler is an object handling an item type of uint8.
type Uint8Handler interface {
	// Handle should do something with item of uint8.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Handle(uint8) error
}

// Uint8Handle is a shortcut implementation
// of Uint8Handler based on a function.
type Uint8Handle func(uint8) error

// Handle does something with item of uint8.
// It is suggested to return EndOfUint8Iterator to stop iteration.
func (h Uint8Handle) Handle(item uint8) error { return h(item) }

// Uint8DoNothing does nothing.
var Uint8DoNothing Uint8Handler = Uint8Handle(func(_ uint8) error { return nil })

type doubleUint8Handler struct {
	lhs, rhs Uint8Handler
}

func (h doubleUint8Handler) Handle(item uint8) error {
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

// Uint8HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint8HandlerSeries(handlers ...Uint8Handler) Uint8Handler {
	var series = Uint8DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint8Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingUint8Iterator does iteration with
// handling by previously set handler.
type HandlingUint8Iterator struct {
	preparedUint8Item
	handler Uint8Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
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

// Uint8Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Uint8Handling(items Uint8Iterator, handlers ...Uint8Handler) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &HandlingUint8Iterator{
		preparedUint8Item{base: items}, Uint8HandlerSeries(handlers...)}
}

// Uint8Range iterates over items and use handlers to each one.
func Uint8Range(items Uint8Iterator, handlers ...Uint8Handler) error {
	return Uint8Discard(Uint8Handling(items, handlers...))
}

// Uint8RangeIterator is an iterator over items.
type Uint8RangeIterator interface {
	// Range should iterate over items.
	Range(...Uint8Handler) error
}

type sUint8RangeIterator struct {
	iter Uint8Iterator
}

// ToUint8RangeIterator constructs an instance implementing Uint8RangeIterator
// based on Uint8Iterator.
func ToUint8RangeIterator(iter Uint8Iterator) Uint8RangeIterator {
	if iter == nil {
		iter = EmptyUint8Iterator
	}
	return sUint8RangeIterator{iter: iter}
}

// MakeUint8RangeIterator constructs an instance implementing Uint8RangeIterator
// based on Uint8IterMaker.
func MakeUint8RangeIterator(maker Uint8IterMaker) Uint8RangeIterator {
	if maker == nil {
		maker = MakeNoUint8Iter
	}
	return ToUint8RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sUint8RangeIterator) Range(handlers ...Uint8Handler) error {
	return Uint8Range(r.iter, handlers...)
}

// Uint8EnumHandler is an object handling an item type of uint8 and its ordered number.
type Uint8EnumHandler interface {
	// Handle should do something with item of uint8 and its ordered number.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Handle(int, uint8) error
}

// Uint8EnumHandle is a shortcut implementation
// of Uint8EnumHandler based on a function.
type Uint8EnumHandle func(int, uint8) error

// Handle does something with item of uint8 and its ordered number.
// It is suggested to return EndOfUint8Iterator to stop iteration.
func (h Uint8EnumHandle) Handle(n int, item uint8) error { return h(n, item) }

// Uint8DoEnumNothing does nothing.
var Uint8DoEnumNothing = Uint8EnumHandle(func(_ int, _ uint8) error { return nil })

type doubleUint8EnumHandler struct {
	lhs, rhs Uint8EnumHandler
}

func (h doubleUint8EnumHandler) Handle(n int, item uint8) error {
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

// Uint8EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint8EnumHandlerSeries(handlers ...Uint8EnumHandler) Uint8EnumHandler {
	var series Uint8EnumHandler = Uint8DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint8EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingUint8Iterator does iteration with
// handling by previously set handler.
type EnumHandlingUint8Iterator struct {
	preparedUint8Item
	handler Uint8EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
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

// Uint8EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Uint8EnumHandling(items Uint8Iterator, handlers ...Uint8EnumHandler) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &EnumHandlingUint8Iterator{
		preparedUint8Item{base: items}, Uint8EnumHandlerSeries(handlers...), 0}
}

// Uint8Enum iterates over items and their ordering numbers and use handlers to each one.
func Uint8Enum(items Uint8Iterator, handlers ...Uint8EnumHandler) error {
	return Uint8Discard(Uint8EnumHandling(items, handlers...))
}

// Uint8EnumIterator is an iterator over items and their ordering numbers.
type Uint8EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Uint8EnumHandler) error
}

type sUint8EnumIterator struct {
	iter Uint8Iterator
}

// ToUint8EnumIterator constructs an instance implementing Uint8EnumIterator
// based on Uint8Iterator.
func ToUint8EnumIterator(iter Uint8Iterator) Uint8EnumIterator {
	if iter == nil {
		iter = EmptyUint8Iterator
	}
	return sUint8EnumIterator{iter: iter}
}

// MakeUint8EnumIterator constructs an instance implementing Uint8EnumIterator
// based on Uint8IterMaker.
func MakeUint8EnumIterator(maker Uint8IterMaker) Uint8EnumIterator {
	if maker == nil {
		maker = MakeNoUint8Iter
	}
	return ToUint8EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sUint8EnumIterator) Enum(handlers ...Uint8EnumHandler) error {
	return Uint8Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sUint8EnumIterator) Range(handlers ...Uint8Handler) error {
	return Uint8Range(r.iter, handlers...)
}

type doubleUint8Iterator struct {
	lhs, rhs Uint8Iterator
	inRHS    bool
}

func (it *doubleUint8Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint8Iterator) Next() uint8 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint8Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUint8Iterator combines all iterators to one.
func SuperUint8Iterator(itemList ...Uint8Iterator) Uint8Iterator {
	var super = EmptyUint8Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint8Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Uint8EnumComparer is a strategy to compare two types.
type Uint8Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs uint8) bool
}

// Uint8Compare is a shortcut implementation
// of Uint8EnumComparer based on a function.
type Uint8Compare func(lhs, rhs uint8) bool

// IsLess is true if lhs is less than rhs.
func (c Uint8Compare) IsLess(lhs, rhs uint8) bool { return c(lhs, rhs) }

// EnumUint8AlwaysLess is an implementation of Uint8EnumComparer returning always true.
var Uint8AlwaysLess Uint8Comparer = Uint8Compare(func(_, _ uint8) bool { return true })

type priorityUint8Iterator struct {
	lhs, rhs preparedUint8Item
	comparer Uint8Comparer
}

func (it *priorityUint8Iterator) HasNext() bool {
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

func (it *priorityUint8Iterator) Next() uint8 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUint8IteratorError(
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

func (it priorityUint8Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUint8Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUint8Iterator(comparer Uint8Comparer, itemList ...Uint8Iterator) Uint8Iterator {
	if comparer == nil {
		comparer = Uint8AlwaysLess
	}

	var prior = EmptyUint8Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUint8Iterator{
			lhs:      preparedUint8Item{base: itemList[i]},
			rhs:      preparedUint8Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Uint8EnumComparer is a strategy to compare two types and their order numbers.
type Uint8EnumComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(nLHS int, lhs uint8, nRHS int, rhs uint8) bool
}

// Uint8EnumCompare is a shortcut implementation
// of Uint8EnumComparer based on a function.
type Uint8EnumCompare func(nLHS int, lhs uint8, nRHS int, rhs uint8) bool

// IsLess is true if lhs is less than rhs.
func (c Uint8EnumCompare) IsLess(nLHS int, lhs uint8, nRHS int, rhs uint8) bool {
	return c(nLHS, lhs, nRHS, rhs)
}

// EnumUint8AlwaysLess is an implementation of Uint8EnumComparer returning always true.
var EnumUint8AlwaysLess Uint8EnumComparer = Uint8EnumCompare(
	func(_ int, _ uint8, _ int, _ uint8) bool { return true })

type priorityUint8EnumIterator struct {
	lhs, rhs           preparedUint8Item
	countLHS, countRHS int
	comparer           Uint8EnumComparer
}

func (it *priorityUint8EnumIterator) HasNext() bool {
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

func (it *priorityUint8EnumIterator) Next() uint8 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUint8IteratorError(
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

func (it priorityUint8EnumIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUint8EnumIterator compare one by one items and their ordering numbers fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUint8EnumIterator(comparer Uint8EnumComparer, itemList ...Uint8Iterator) Uint8Iterator {
	if comparer == nil {
		comparer = EnumUint8AlwaysLess
	}

	var prior = EmptyUint8Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUint8EnumIterator{
			lhs:      preparedUint8Item{base: itemList[i]},
			rhs:      preparedUint8Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Uint8SliceIterator is an iterator based on a slice of uint8.
type Uint8SliceIterator struct {
	slice []uint8
	cur   int
}

// NewUint8SliceIterator returns a new instance of Uint8SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Uint8Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewUint8SliceIterator(slice []uint8) *Uint8SliceIterator {
	it := &Uint8SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Uint8SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Uint8SliceIterator) Next() uint8 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Uint8SliceIterator) Err() error { return nil }

// Uint8SliceIterator is an iterator based on a slice of uint8
// and doing iteration in back direction.
type InvertingUint8SliceIterator struct {
	slice []uint8
	cur   int
}

// NewInvertingUint8SliceIterator returns a new instance of InvertingUint8SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingUint8Slice(Uint8Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingUint8SliceIterator(slice []uint8) *InvertingUint8SliceIterator {
	it := &InvertingUint8SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingUint8SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingUint8SliceIterator) Next() uint8 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingUint8SliceIterator) Err() error { return nil }

// Uint8Unroll unrolls items to slice of uint8.
func Uint8Unroll(items Uint8Iterator) Uint8Slice {
	var slice Uint8Slice
	panicIfUint8IteratorError(Uint8Discard(Uint8Handling(items, Uint8Handle(func(item uint8) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Uint8Slice is a slice of uint8.
type Uint8Slice []uint8

// MakeIter returns a new instance of Uint8Iterator to iterate over it.
// It returns EmptyUint8Iterator if the error is not nil.
func (s Uint8Slice) MakeIter() Uint8Iterator {
	return NewUint8SliceIterator(s)
}

// Uint8Slice is a slice of uint8 which can make inverting iterator.
type InvertingUint8Slice []uint8

// MakeIter returns a new instance of Uint8Iterator to iterate over it.
// It returns EmptyUint8Iterator if the error is not nil.
func (s InvertingUint8Slice) MakeIter() Uint8Iterator {
	return NewInvertingUint8SliceIterator(s)
}

// Uint8Invert unrolls items and make inverting iterator based on them.
func Uint8Invert(items Uint8Iterator) Uint8Iterator {
	return InvertingUint8Slice(Uint8Unroll(items)).MakeIter()
}

// EndOfUint8Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfUint8Iterator = errors.New("end of uint8 iterator")

func isEndOfUint8Iterator(err error) bool {
	return errors.Is(err, EndOfUint8Iterator)
}

func wrapIfNotEndOfUint8Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfUint8Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfUint8IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedUint8Item struct {
	base    Uint8Iterator
	hasNext bool
	next    uint8
	err     error
}

func (it *preparedUint8Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedUint8Item) Next() uint8 {
	if !it.hasNext {
		panicIfUint8IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedUint8Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfUint8Iterator) {
		return it.err
	}
	return it.base.Err()
}
