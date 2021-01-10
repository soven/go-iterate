package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// Uint16Iterator is an iterator over items type of uint16.
type Uint16Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() uint16
	// Err contains first met error while Next.
	Err() error
}

type emptyUint16Iterator struct{}

func (emptyUint16Iterator) HasNext() bool       { return false }
func (emptyUint16Iterator) Next() (next uint16) { return 0 }
func (emptyUint16Iterator) Err() error          { return nil }

// EmptyUint16Iterator is a zero value for Uint16Iterator.
// It is not contains any item to iterate over it.
var EmptyUint16Iterator Uint16Iterator = emptyUint16Iterator{}

// Uint16IterMaker is a maker of Uint16Iterator.
type Uint16IterMaker interface {
	// MakeIter should return a new instance of Uint16Iterator to iterate over it.
	MakeIter() Uint16Iterator
}

// MakeUint16Iter is a shortcut implementation
// of Uint16Iterator based on a function.
type MakeUint16Iter func() Uint16Iterator

// MakeIter returns a new instance of Uint16Iterator to iterate over it.
func (m MakeUint16Iter) MakeIter() Uint16Iterator { return m() }

// MakeNoUint16Iter is a zero value for Uint16IterMaker.
// It always returns EmptyUint16Iterator and an empty error.
var MakeNoUint16Iter Uint16IterMaker = MakeUint16Iter(
	func() Uint16Iterator { return EmptyUint16Iterator })

// Uint16Discard just range over all items and do nothing with each of them.
func Uint16Discard(items Uint16Iterator) error {
	if items == nil {
		items = EmptyUint16Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

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

// Uint16Converter is an object converting an item type of uint16.
type Uint16Converter interface {
	// Convert should convert an item type of uint16 into another item of uint16.
	// It is suggested to return EndOfUint16Iterator to stop iteration.
	Convert(uint16) (uint16, error)
}

// Uint16Convert is a shortcut implementation
// of Uint16Converter based on a function.
type Uint16Convert func(uint16) (uint16, error)

// Convert converts an item type of uint16 into another item of uint16.
// It is suggested to return EndOfUint16Iterator to stop iteration.
func (c Uint16Convert) Convert(item uint16) (uint16, error) { return c(item) }

// NoUint16Convert does nothing with item, just returns it as is.
var NoUint16Convert Uint16Converter = Uint16Convert(
	func(item uint16) (uint16, error) { return item, nil })

type doubleUint16Converter struct {
	lhs, rhs Uint16Converter
}

func (c doubleUint16Converter) Convert(item uint16) (uint16, error) {
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

// Uint16ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Uint16ConverterSeries(converters ...Uint16Converter) Uint16Converter {
	var series = NoUint16Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint16Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingUint16Iterator does iteration with
// converting by previously set converter.
type ConvertingUint16Iterator struct {
	preparedUint16Item
	converter Uint16Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
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

// Uint16Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Uint16Converting(items Uint16Iterator, converters ...Uint16Converter) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	return &ConvertingUint16Iterator{
		preparedUint16Item{base: items}, Uint16ConverterSeries(converters...)}
}

// Uint16EnumConverter is an object converting an item type of uint16 and its ordering number.
type Uint16EnumConverter interface {
	// Convert should convert an item type of uint16 into another item of uint16.
	// It is suggested to return EndOfUint16Iterator to stop iteration.
	Convert(n int, val uint16) (uint16, error)
}

// Uint16EnumConvert is a shortcut implementation
// of Uint16EnumConverter based on a function.
type Uint16EnumConvert func(int, uint16) (uint16, error)

// Convert converts an item type of uint16 into another item of uint16.
// It is suggested to return EndOfUint16Iterator to stop iteration.
func (c Uint16EnumConvert) Convert(n int, item uint16) (uint16, error) { return c(n, item) }

// NoUint16EnumConvert does nothing with item, just returns it as is.
var NoUint16EnumConvert Uint16EnumConverter = Uint16EnumConvert(
	func(_ int, item uint16) (uint16, error) { return item, nil })

type enumFromUint16Converter struct {
	Uint16Converter
}

func (ch enumFromUint16Converter) Convert(_ int, item uint16) (uint16, error) {
	return ch.Uint16Converter.Convert(item)
}

// EnumFromUint16Converter adapts checker type of Uint16Converter
// to the interface Uint16EnumConverter.
// If converter is nil it is return based on NoUint16Convert enum checker.
func EnumFromUint16Converter(converter Uint16Converter) Uint16EnumConverter {
	if converter == nil {
		converter = NoUint16Convert
	}
	return &enumFromUint16Converter{converter}
}

type doubleUint16EnumConverter struct {
	lhs, rhs Uint16EnumConverter
}

func (c doubleUint16EnumConverter) Convert(n int, item uint16) (uint16, error) {
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

// EnumUint16ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumUint16ConverterSeries(converters ...Uint16EnumConverter) Uint16EnumConverter {
	var series = NoUint16EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint16EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingUint16Iterator does iteration with
// converting by previously set converter.
type EnumConvertingUint16Iterator struct {
	preparedUint16Item
	converter Uint16EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
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

// Uint16EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Uint16EnumConverting(items Uint16Iterator, converters ...Uint16EnumConverter) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	return &EnumConvertingUint16Iterator{
		preparedUint16Item{base: items}, EnumUint16ConverterSeries(converters...), 0}
}

// Uint16Handler is an object handling an item type of uint16.
type Uint16Handler interface {
	// Handle should do something with item of uint16.
	// It is suggested to return EndOfUint16Iterator to stop iteration.
	Handle(uint16) error
}

// Uint16Handle is a shortcut implementation
// of Uint16Handler based on a function.
type Uint16Handle func(uint16) error

// Handle does something with item of uint16.
// It is suggested to return EndOfUint16Iterator to stop iteration.
func (h Uint16Handle) Handle(item uint16) error { return h(item) }

// Uint16DoNothing does nothing.
var Uint16DoNothing Uint16Handler = Uint16Handle(func(_ uint16) error { return nil })

type doubleUint16Handler struct {
	lhs, rhs Uint16Handler
}

func (h doubleUint16Handler) Handle(item uint16) error {
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

// Uint16HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint16HandlerSeries(handlers ...Uint16Handler) Uint16Handler {
	var series = Uint16DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint16Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingUint16Iterator does iteration with
// handling by previously set handler.
type HandlingUint16Iterator struct {
	preparedUint16Item
	handler Uint16Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
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

// Uint16Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Uint16Handling(items Uint16Iterator, handlers ...Uint16Handler) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	return &HandlingUint16Iterator{
		preparedUint16Item{base: items}, Uint16HandlerSeries(handlers...)}
}

// Uint16Range iterates over items and use handlers to each one.
func Uint16Range(items Uint16Iterator, handlers ...Uint16Handler) error {
	return Uint16Discard(Uint16Handling(items, handlers...))
}

// Uint16RangeIterator is an iterator over items.
type Uint16RangeIterator interface {
	// Range should iterate over items.
	Range(...Uint16Handler) error
}

type sUint16RangeIterator struct {
	iter Uint16Iterator
}

// ToUint16RangeIterator constructs an instance implementing Uint16RangeIterator
// based on Uint16Iterator.
func ToUint16RangeIterator(iter Uint16Iterator) Uint16RangeIterator {
	if iter == nil {
		iter = EmptyUint16Iterator
	}
	return sUint16RangeIterator{iter: iter}
}

// MakeUint16RangeIterator constructs an instance implementing Uint16RangeIterator
// based on Uint16IterMaker.
func MakeUint16RangeIterator(maker Uint16IterMaker) Uint16RangeIterator {
	if maker == nil {
		maker = MakeNoUint16Iter
	}
	return ToUint16RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sUint16RangeIterator) Range(handlers ...Uint16Handler) error {
	return Uint16Range(r.iter, handlers...)
}

// Uint16EnumHandler is an object handling an item type of uint16 and its ordered number.
type Uint16EnumHandler interface {
	// Handle should do something with item of uint16 and its ordered number.
	// It is suggested to return EndOfUint16Iterator to stop iteration.
	Handle(int, uint16) error
}

// Uint16EnumHandle is a shortcut implementation
// of Uint16EnumHandler based on a function.
type Uint16EnumHandle func(int, uint16) error

// Handle does something with item of uint16 and its ordered number.
// It is suggested to return EndOfUint16Iterator to stop iteration.
func (h Uint16EnumHandle) Handle(n int, item uint16) error { return h(n, item) }

// Uint16DoEnumNothing does nothing.
var Uint16DoEnumNothing = Uint16EnumHandle(func(_ int, _ uint16) error { return nil })

type doubleUint16EnumHandler struct {
	lhs, rhs Uint16EnumHandler
}

func (h doubleUint16EnumHandler) Handle(n int, item uint16) error {
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

// Uint16EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint16EnumHandlerSeries(handlers ...Uint16EnumHandler) Uint16EnumHandler {
	var series Uint16EnumHandler = Uint16DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint16EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingUint16Iterator does iteration with
// handling by previously set handler.
type EnumHandlingUint16Iterator struct {
	preparedUint16Item
	handler Uint16EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
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

// Uint16EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Uint16EnumHandling(items Uint16Iterator, handlers ...Uint16EnumHandler) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	return &EnumHandlingUint16Iterator{
		preparedUint16Item{base: items}, Uint16EnumHandlerSeries(handlers...), 0}
}

// Uint16Enum iterates over items and their ordering numbers and use handlers to each one.
func Uint16Enum(items Uint16Iterator, handlers ...Uint16EnumHandler) error {
	return Uint16Discard(Uint16EnumHandling(items, handlers...))
}

// Uint16EnumIterator is an iterator over items and their ordering numbers.
type Uint16EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Uint16EnumHandler) error
}

type sUint16EnumIterator struct {
	iter Uint16Iterator
}

// ToUint16EnumIterator constructs an instance implementing Uint16EnumIterator
// based on Uint16Iterator.
func ToUint16EnumIterator(iter Uint16Iterator) Uint16EnumIterator {
	if iter == nil {
		iter = EmptyUint16Iterator
	}
	return sUint16EnumIterator{iter: iter}
}

// MakeUint16EnumIterator constructs an instance implementing Uint16EnumIterator
// based on Uint16IterMaker.
func MakeUint16EnumIterator(maker Uint16IterMaker) Uint16EnumIterator {
	if maker == nil {
		maker = MakeNoUint16Iter
	}
	return ToUint16EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sUint16EnumIterator) Enum(handlers ...Uint16EnumHandler) error {
	return Uint16Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sUint16EnumIterator) Range(handlers ...Uint16Handler) error {
	return Uint16Range(r.iter, handlers...)
}

type doubleUint16Iterator struct {
	lhs, rhs Uint16Iterator
	inRHS    bool
}

func (it *doubleUint16Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint16Iterator) Next() uint16 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint16Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUint16Iterator combines all iterators to one.
func SuperUint16Iterator(itemList ...Uint16Iterator) Uint16Iterator {
	var super = EmptyUint16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint16Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Uint16EnumComparer is a strategy to compare two types.
type Uint16Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs uint16) bool
}

// Uint16Compare is a shortcut implementation
// of Uint16EnumComparer based on a function.
type Uint16Compare func(lhs, rhs uint16) bool

// IsLess is true if lhs is less than rhs.
func (c Uint16Compare) IsLess(lhs, rhs uint16) bool { return c(lhs, rhs) }

// EnumUint16AlwaysLess is an implementation of Uint16EnumComparer returning always true.
var Uint16AlwaysLess Uint16Comparer = Uint16Compare(func(_, _ uint16) bool { return true })

type priorityUint16Iterator struct {
	lhs, rhs preparedUint16Item
	comparer Uint16Comparer
}

func (it *priorityUint16Iterator) HasNext() bool {
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

func (it *priorityUint16Iterator) Next() uint16 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUint16IteratorError(
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

func (it priorityUint16Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUint16Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUint16Iterator(comparer Uint16Comparer, itemList ...Uint16Iterator) Uint16Iterator {
	if comparer == nil {
		comparer = Uint16AlwaysLess
	}

	var prior = EmptyUint16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUint16Iterator{
			lhs:      preparedUint16Item{base: itemList[i]},
			rhs:      preparedUint16Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Uint16EnumComparer is a strategy to compare two types and their order numbers.
type Uint16EnumComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(nLHS int, lhs uint16, nRHS int, rhs uint16) bool
}

// Uint16EnumCompare is a shortcut implementation
// of Uint16EnumComparer based on a function.
type Uint16EnumCompare func(nLHS int, lhs uint16, nRHS int, rhs uint16) bool

// IsLess is true if lhs is less than rhs.
func (c Uint16EnumCompare) IsLess(nLHS int, lhs uint16, nRHS int, rhs uint16) bool {
	return c(nLHS, lhs, nRHS, rhs)
}

// EnumUint16AlwaysLess is an implementation of Uint16EnumComparer returning always true.
var EnumUint16AlwaysLess Uint16EnumComparer = Uint16EnumCompare(
	func(_ int, _ uint16, _ int, _ uint16) bool { return true })

type priorityUint16EnumIterator struct {
	lhs, rhs           preparedUint16Item
	countLHS, countRHS int
	comparer           Uint16EnumComparer
}

func (it *priorityUint16EnumIterator) HasNext() bool {
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

func (it *priorityUint16EnumIterator) Next() uint16 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUint16IteratorError(
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

func (it priorityUint16EnumIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUint16EnumIterator compare one by one items and their ordering numbers fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUint16EnumIterator(comparer Uint16EnumComparer, itemList ...Uint16Iterator) Uint16Iterator {
	if comparer == nil {
		comparer = EnumUint16AlwaysLess
	}

	var prior = EmptyUint16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUint16EnumIterator{
			lhs:      preparedUint16Item{base: itemList[i]},
			rhs:      preparedUint16Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Uint16SliceIterator is an iterator based on a slice of uint16.
type Uint16SliceIterator struct {
	slice []uint16
	cur   int
}

// NewUint16SliceIterator returns a new instance of Uint16SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Uint16Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewUint16SliceIterator(slice []uint16) *Uint16SliceIterator {
	it := &Uint16SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Uint16SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Uint16SliceIterator) Next() uint16 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Uint16SliceIterator) Err() error { return nil }

// Uint16SliceIterator is an iterator based on a slice of uint16
// and doing iteration in back direction.
type InvertingUint16SliceIterator struct {
	slice []uint16
	cur   int
}

// NewInvertingUint16SliceIterator returns a new instance of InvertingUint16SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingUint16Slice(Uint16Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingUint16SliceIterator(slice []uint16) *InvertingUint16SliceIterator {
	it := &InvertingUint16SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingUint16SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingUint16SliceIterator) Next() uint16 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingUint16SliceIterator) Err() error { return nil }

// Uint16Unroll unrolls items to slice of uint16.
func Uint16Unroll(items Uint16Iterator) Uint16Slice {
	var slice Uint16Slice
	panicIfUint16IteratorError(Uint16Discard(Uint16Handling(items, Uint16Handle(func(item uint16) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Uint16Slice is a slice of uint16.
type Uint16Slice []uint16

// MakeIter returns a new instance of Uint16Iterator to iterate over it.
// It returns EmptyUint16Iterator if the error is not nil.
func (s Uint16Slice) MakeIter() Uint16Iterator {
	return NewUint16SliceIterator(s)
}

// Uint16Slice is a slice of uint16 which can make inverting iterator.
type InvertingUint16Slice []uint16

// MakeIter returns a new instance of Uint16Iterator to iterate over it.
// It returns EmptyUint16Iterator if the error is not nil.
func (s InvertingUint16Slice) MakeIter() Uint16Iterator {
	return NewInvertingUint16SliceIterator(s)
}

// Uint16Invert unrolls items and make inverting iterator based on them.
func Uint16Invert(items Uint16Iterator) Uint16Iterator {
	return InvertingUint16Slice(Uint16Unroll(items)).MakeIter()
}

// EndOfUint16Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfUint16Iterator = errors.New("end of uint16 iterator")

func isEndOfUint16Iterator(err error) bool {
	return errors.Is(err, EndOfUint16Iterator)
}

func wrapIfNotEndOfUint16Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfUint16Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfUint16IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedUint16Item struct {
	base    Uint16Iterator
	hasNext bool
	next    uint16
	err     error
}

func (it *preparedUint16Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedUint16Item) Next() uint16 {
	if !it.hasNext {
		panicIfUint16IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedUint16Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfUint16Iterator) {
		return it.err
	}
	return it.base.Err()
}
