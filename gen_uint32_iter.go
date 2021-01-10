package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// Uint32Iterator is an iterator over items type of uint32.
type Uint32Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() uint32
	// Err contains first met error while Next.
	Err() error
}

type emptyUint32Iterator struct{}

func (emptyUint32Iterator) HasNext() bool       { return false }
func (emptyUint32Iterator) Next() (next uint32) { return 0 }
func (emptyUint32Iterator) Err() error          { return nil }

// EmptyUint32Iterator is a zero value for Uint32Iterator.
// It is not contains any item to iterate over it.
var EmptyUint32Iterator Uint32Iterator = emptyUint32Iterator{}

// Uint32IterMaker is a maker of Uint32Iterator.
type Uint32IterMaker interface {
	// MakeIter should return a new instance of Uint32Iterator to iterate over it.
	MakeIter() Uint32Iterator
}

// MakeUint32Iter is a shortcut implementation
// of Uint32Iterator based on a function.
type MakeUint32Iter func() Uint32Iterator

// MakeIter returns a new instance of Uint32Iterator to iterate over it.
func (m MakeUint32Iter) MakeIter() Uint32Iterator { return m() }

// MakeNoUint32Iter is a zero value for Uint32IterMaker.
// It always returns EmptyUint32Iterator and an empty error.
var MakeNoUint32Iter Uint32IterMaker = MakeUint32Iter(
	func() Uint32Iterator { return EmptyUint32Iterator })

// Uint32Discard just range over all items and do nothing with each of them.
func Uint32Discard(items Uint32Iterator) error {
	if items == nil {
		items = EmptyUint32Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

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

// Uint32Converter is an object converting an item type of uint32.
type Uint32Converter interface {
	// Convert should convert an item type of uint32 into another item of uint32.
	// It is suggested to return EndOfUint32Iterator to stop iteration.
	Convert(uint32) (uint32, error)
}

// Uint32Convert is a shortcut implementation
// of Uint32Converter based on a function.
type Uint32Convert func(uint32) (uint32, error)

// Convert converts an item type of uint32 into another item of uint32.
// It is suggested to return EndOfUint32Iterator to stop iteration.
func (c Uint32Convert) Convert(item uint32) (uint32, error) { return c(item) }

// NoUint32Convert does nothing with item, just returns it as is.
var NoUint32Convert Uint32Converter = Uint32Convert(
	func(item uint32) (uint32, error) { return item, nil })

type doubleUint32Converter struct {
	lhs, rhs Uint32Converter
}

func (c doubleUint32Converter) Convert(item uint32) (uint32, error) {
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

// Uint32ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Uint32ConverterSeries(converters ...Uint32Converter) Uint32Converter {
	var series = NoUint32Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint32Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingUint32Iterator does iteration with
// converting by previously set converter.
type ConvertingUint32Iterator struct {
	preparedUint32Item
	converter Uint32Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingUint32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint32Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfUint32Iterator(err) {
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

// Uint32Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Uint32Converting(items Uint32Iterator, converters ...Uint32Converter) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	return &ConvertingUint32Iterator{
		preparedUint32Item{base: items}, Uint32ConverterSeries(converters...)}
}

// Uint32EnumConverter is an object converting an item type of uint32 and its ordering number.
type Uint32EnumConverter interface {
	// Convert should convert an item type of uint32 into another item of uint32.
	// It is suggested to return EndOfUint32Iterator to stop iteration.
	Convert(n int, val uint32) (uint32, error)
}

// Uint32EnumConvert is a shortcut implementation
// of Uint32EnumConverter based on a function.
type Uint32EnumConvert func(int, uint32) (uint32, error)

// Convert converts an item type of uint32 into another item of uint32.
// It is suggested to return EndOfUint32Iterator to stop iteration.
func (c Uint32EnumConvert) Convert(n int, item uint32) (uint32, error) { return c(n, item) }

// NoUint32EnumConvert does nothing with item, just returns it as is.
var NoUint32EnumConvert Uint32EnumConverter = Uint32EnumConvert(
	func(_ int, item uint32) (uint32, error) { return item, nil })

type enumFromUint32Converter struct {
	Uint32Converter
}

func (ch enumFromUint32Converter) Convert(_ int, item uint32) (uint32, error) {
	return ch.Uint32Converter.Convert(item)
}

// EnumFromUint32Converter adapts checker type of Uint32Converter
// to the interface Uint32EnumConverter.
// If converter is nil it is return based on NoUint32Convert enum checker.
func EnumFromUint32Converter(converter Uint32Converter) Uint32EnumConverter {
	if converter == nil {
		converter = NoUint32Convert
	}
	return &enumFromUint32Converter{converter}
}

type doubleUint32EnumConverter struct {
	lhs, rhs Uint32EnumConverter
}

func (c doubleUint32EnumConverter) Convert(n int, item uint32) (uint32, error) {
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

// EnumUint32ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumUint32ConverterSeries(converters ...Uint32EnumConverter) Uint32EnumConverter {
	var series = NoUint32EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint32EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingUint32Iterator does iteration with
// converting by previously set converter.
type EnumConvertingUint32Iterator struct {
	preparedUint32Item
	converter Uint32EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingUint32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint32Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfUint32Iterator(err) {
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

// Uint32EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Uint32EnumConverting(items Uint32Iterator, converters ...Uint32EnumConverter) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	return &EnumConvertingUint32Iterator{
		preparedUint32Item{base: items}, EnumUint32ConverterSeries(converters...), 0}
}

// Uint32Handler is an object handling an item type of uint32.
type Uint32Handler interface {
	// Handle should do something with item of uint32.
	// It is suggested to return EndOfUint32Iterator to stop iteration.
	Handle(uint32) error
}

// Uint32Handle is a shortcut implementation
// of Uint32Handler based on a function.
type Uint32Handle func(uint32) error

// Handle does something with item of uint32.
// It is suggested to return EndOfUint32Iterator to stop iteration.
func (h Uint32Handle) Handle(item uint32) error { return h(item) }

// Uint32DoNothing does nothing.
var Uint32DoNothing Uint32Handler = Uint32Handle(func(_ uint32) error { return nil })

type doubleUint32Handler struct {
	lhs, rhs Uint32Handler
}

func (h doubleUint32Handler) Handle(item uint32) error {
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

// Uint32HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint32HandlerSeries(handlers ...Uint32Handler) Uint32Handler {
	var series = Uint32DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint32Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingUint32Iterator does iteration with
// handling by previously set handler.
type HandlingUint32Iterator struct {
	preparedUint32Item
	handler Uint32Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingUint32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint32Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfUint32Iterator(err) {
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

// Uint32Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Uint32Handling(items Uint32Iterator, handlers ...Uint32Handler) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	return &HandlingUint32Iterator{
		preparedUint32Item{base: items}, Uint32HandlerSeries(handlers...)}
}

// Uint32Range iterates over items and use handlers to each one.
func Uint32Range(items Uint32Iterator, handlers ...Uint32Handler) error {
	return Uint32Discard(Uint32Handling(items, handlers...))
}

// Uint32RangeIterator is an iterator over items.
type Uint32RangeIterator interface {
	// Range should iterate over items.
	Range(...Uint32Handler) error
}

type sUint32RangeIterator struct {
	iter Uint32Iterator
}

// ToUint32RangeIterator constructs an instance implementing Uint32RangeIterator
// based on Uint32Iterator.
func ToUint32RangeIterator(iter Uint32Iterator) Uint32RangeIterator {
	if iter == nil {
		iter = EmptyUint32Iterator
	}
	return sUint32RangeIterator{iter: iter}
}

// MakeUint32RangeIterator constructs an instance implementing Uint32RangeIterator
// based on Uint32IterMaker.
func MakeUint32RangeIterator(maker Uint32IterMaker) Uint32RangeIterator {
	if maker == nil {
		maker = MakeNoUint32Iter
	}
	return ToUint32RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sUint32RangeIterator) Range(handlers ...Uint32Handler) error {
	return Uint32Range(r.iter, handlers...)
}

// Uint32EnumHandler is an object handling an item type of uint32 and its ordered number.
type Uint32EnumHandler interface {
	// Handle should do something with item of uint32 and its ordered number.
	// It is suggested to return EndOfUint32Iterator to stop iteration.
	Handle(int, uint32) error
}

// Uint32EnumHandle is a shortcut implementation
// of Uint32EnumHandler based on a function.
type Uint32EnumHandle func(int, uint32) error

// Handle does something with item of uint32 and its ordered number.
// It is suggested to return EndOfUint32Iterator to stop iteration.
func (h Uint32EnumHandle) Handle(n int, item uint32) error { return h(n, item) }

// Uint32DoEnumNothing does nothing.
var Uint32DoEnumNothing = Uint32EnumHandle(func(_ int, _ uint32) error { return nil })

type doubleUint32EnumHandler struct {
	lhs, rhs Uint32EnumHandler
}

func (h doubleUint32EnumHandler) Handle(n int, item uint32) error {
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

// Uint32EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Uint32EnumHandlerSeries(handlers ...Uint32EnumHandler) Uint32EnumHandler {
	var series Uint32EnumHandler = Uint32DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleUint32EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingUint32Iterator does iteration with
// handling by previously set handler.
type EnumHandlingUint32Iterator struct {
	preparedUint32Item
	handler Uint32EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingUint32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint32Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfUint32Iterator(err) {
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

// Uint32EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Uint32EnumHandling(items Uint32Iterator, handlers ...Uint32EnumHandler) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	return &EnumHandlingUint32Iterator{
		preparedUint32Item{base: items}, Uint32EnumHandlerSeries(handlers...), 0}
}

// Uint32Enum iterates over items and their ordering numbers and use handlers to each one.
func Uint32Enum(items Uint32Iterator, handlers ...Uint32EnumHandler) error {
	return Uint32Discard(Uint32EnumHandling(items, handlers...))
}

// Uint32EnumIterator is an iterator over items and their ordering numbers.
type Uint32EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Uint32EnumHandler) error
}

type sUint32EnumIterator struct {
	iter Uint32Iterator
}

// ToUint32EnumIterator constructs an instance implementing Uint32EnumIterator
// based on Uint32Iterator.
func ToUint32EnumIterator(iter Uint32Iterator) Uint32EnumIterator {
	if iter == nil {
		iter = EmptyUint32Iterator
	}
	return sUint32EnumIterator{iter: iter}
}

// MakeUint32EnumIterator constructs an instance implementing Uint32EnumIterator
// based on Uint32IterMaker.
func MakeUint32EnumIterator(maker Uint32IterMaker) Uint32EnumIterator {
	if maker == nil {
		maker = MakeNoUint32Iter
	}
	return ToUint32EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sUint32EnumIterator) Enum(handlers ...Uint32EnumHandler) error {
	return Uint32Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sUint32EnumIterator) Range(handlers ...Uint32Handler) error {
	return Uint32Range(r.iter, handlers...)
}

type doubleUint32Iterator struct {
	lhs, rhs Uint32Iterator
	inRHS    bool
}

func (it *doubleUint32Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleUint32Iterator) Next() uint32 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleUint32Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperUint32Iterator combines all iterators to one.
func SuperUint32Iterator(itemList ...Uint32Iterator) Uint32Iterator {
	var super = EmptyUint32Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleUint32Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Uint32Comparer is a strategy to compare two types.
type Uint32Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs uint32) bool
}

// Uint32Compare is a shortcut implementation
// of Uint32Comparer based on a function.
type Uint32Compare func(lhs, rhs uint32) bool

// IsLess is true if lhs is less than rhs.
func (c Uint32Compare) IsLess(lhs, rhs uint32) bool { return c(lhs, rhs) }

// Uint32AlwaysLess is an implementation of Uint32Comparer returning always true.
var Uint32AlwaysLess Uint32Comparer = Uint32Compare(func(_, _ uint32) bool { return true })

type priorityUint32Iterator struct {
	lhs, rhs preparedUint32Item
	comparer Uint32Comparer
}

func (it *priorityUint32Iterator) HasNext() bool {
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

func (it *priorityUint32Iterator) Next() uint32 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfUint32IteratorError(
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

func (it priorityUint32Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorUint32Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorUint32Iterator(comparer Uint32Comparer, itemList ...Uint32Iterator) Uint32Iterator {
	if comparer == nil {
		comparer = Uint32AlwaysLess
	}

	var prior = EmptyUint32Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityUint32Iterator{
			lhs:      preparedUint32Item{base: itemList[i]},
			rhs:      preparedUint32Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Uint32SliceIterator is an iterator based on a slice of uint32.
type Uint32SliceIterator struct {
	slice []uint32
	cur   int
}

// NewUint32SliceIterator returns a new instance of Uint32SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Uint32Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewUint32SliceIterator(slice []uint32) *Uint32SliceIterator {
	it := &Uint32SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Uint32SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Uint32SliceIterator) Next() uint32 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Uint32SliceIterator) Err() error { return nil }

// Uint32SliceIterator is an iterator based on a slice of uint32
// and doing iteration in back direction.
type InvertingUint32SliceIterator struct {
	slice []uint32
	cur   int
}

// NewInvertingUint32SliceIterator returns a new instance of InvertingUint32SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingUint32Slice(Uint32Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingUint32SliceIterator(slice []uint32) *InvertingUint32SliceIterator {
	it := &InvertingUint32SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingUint32SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingUint32SliceIterator) Next() uint32 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingUint32SliceIterator) Err() error { return nil }

// Uint32Unroll unrolls items to slice of uint32.
func Uint32Unroll(items Uint32Iterator) Uint32Slice {
	var slice Uint32Slice
	panicIfUint32IteratorError(Uint32Discard(Uint32Handling(items, Uint32Handle(func(item uint32) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Uint32Slice is a slice of uint32.
type Uint32Slice []uint32

// MakeIter returns a new instance of Uint32Iterator to iterate over it.
// It returns EmptyUint32Iterator if the error is not nil.
func (s Uint32Slice) MakeIter() Uint32Iterator {
	return NewUint32SliceIterator(s)
}

// Uint32Slice is a slice of uint32 which can make inverting iterator.
type InvertingUint32Slice []uint32

// MakeIter returns a new instance of Uint32Iterator to iterate over it.
// It returns EmptyUint32Iterator if the error is not nil.
func (s InvertingUint32Slice) MakeIter() Uint32Iterator {
	return NewInvertingUint32SliceIterator(s)
}

// Uint32Invert unrolls items and make inverting iterator based on them.
func Uint32Invert(items Uint32Iterator) Uint32Iterator {
	return InvertingUint32Slice(Uint32Unroll(items)).MakeIter()
}

// EndOfUint32Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfUint32Iterator = errors.New("end of uint32 iterator")

func isEndOfUint32Iterator(err error) bool {
	return errors.Is(err, EndOfUint32Iterator)
}

func wrapIfNotEndOfUint32Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfUint32Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfUint32IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedUint32Item struct {
	base    Uint32Iterator
	hasNext bool
	next    uint32
	err     error
}

func (it *preparedUint32Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedUint32Item) Next() uint32 {
	if !it.hasNext {
		panicIfUint32IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedUint32Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfUint32Iterator) {
		return it.err
	}
	return it.base.Err()
}
