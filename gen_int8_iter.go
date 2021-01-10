package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// Int8Iterator is an iterator over items type of int8.
type Int8Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() int8
	// Err contains first met error while Next.
	Err() error
}

type emptyInt8Iterator struct{}

func (emptyInt8Iterator) HasNext() bool     { return false }
func (emptyInt8Iterator) Next() (next int8) { return 0 }
func (emptyInt8Iterator) Err() error        { return nil }

// EmptyInt8Iterator is a zero value for Int8Iterator.
// It is not contains any item to iterate over it.
var EmptyInt8Iterator Int8Iterator = emptyInt8Iterator{}

// Int8IterMaker is a maker of Int8Iterator.
type Int8IterMaker interface {
	// MakeIter should return a new instance of Int8Iterator to iterate over it.
	MakeIter() Int8Iterator
}

// MakeInt8Iter is a shortcut implementation
// of Int8Iterator based on a function.
type MakeInt8Iter func() Int8Iterator

// MakeIter returns a new instance of Int8Iterator to iterate over it.
func (m MakeInt8Iter) MakeIter() Int8Iterator { return m() }

// MakeNoInt8Iter is a zero value for Int8IterMaker.
// It always returns EmptyInt8Iterator and an empty error.
var MakeNoInt8Iter Int8IterMaker = MakeInt8Iter(
	func() Int8Iterator { return EmptyInt8Iterator })

// Int8Discard just range over all items and do nothing with each of them.
func Int8Discard(items Int8Iterator) error {
	if items == nil {
		items = EmptyInt8Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

// Int8Checker is an object checking an item type of int8
// for some condition.
type Int8Checker interface {
	// Check should check an item type of int8 for some condition.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Check(int8) (bool, error)
}

// Int8Check is a shortcut implementation
// of Int8Checker based on a function.
type Int8Check func(int8) (bool, error)

// Check checks an item type of int8 for some condition.
// It returns EndOfInt8Iterator to stop iteration.
func (ch Int8Check) Check(item int8) (bool, error) { return ch(item) }

var (
	// AlwaysInt8CheckTrue always returns true and empty error.
	AlwaysInt8CheckTrue Int8Checker = Int8Check(
		func(item int8) (bool, error) { return true, nil })
	// AlwaysInt8CheckFalse always returns false and empty error.
	AlwaysInt8CheckFalse Int8Checker = Int8Check(
		func(item int8) (bool, error) { return false, nil })
)

// NotInt8 do an inversion for checker result.
// It is returns AlwaysInt8CheckTrue if checker is nil.
func NotInt8(checker Int8Checker) Int8Checker {
	if checker == nil {
		return AlwaysInt8CheckTrue
	}
	return Int8Check(func(item int8) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andInt8 struct {
	lhs, rhs Int8Checker
}

func (a andInt8) Check(item int8) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllInt8 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllInt8(checkers ...Int8Checker) Int8Checker {
	var all = AlwaysInt8CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andInt8{checkers[i], all}
	}
	return all
}

type orInt8 struct {
	lhs, rhs Int8Checker
}

func (o orInt8) Check(item int8) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyInt8 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyInt8(checkers ...Int8Checker) Int8Checker {
	var any = AlwaysInt8CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orInt8{checkers[i], any}
	}
	return any
}

// FilteringInt8Iterator does iteration with
// filtering by previously set checker.
type FilteringInt8Iterator struct {
	preparedInt8Item
	filter Int8Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Int8Filtering(items Int8Iterator, filters ...Int8Checker) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &FilteringInt8Iterator{preparedInt8Item{base: items}, AllInt8(filters...)}
}

// DoingUntilInt8Iterator does iteration
// until previously set checker is passed.
type DoingUntilInt8Iterator struct {
	preparedInt8Item
	until Int8Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfInt8Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int8DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int8DoingUntil(items Int8Iterator, untilList ...Int8Checker) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	var until Int8Checker
	if len(untilList) > 0 {
		until = AllInt8(untilList...)
	} else {
		until = AlwaysInt8CheckFalse
	}
	return &DoingUntilInt8Iterator{preparedInt8Item{base: items}, until}
}

// Int8SkipUntil sets until conditions to skip few items.
func Int8SkipUntil(items Int8Iterator, untilList ...Int8Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int8Discard(Int8DoingUntil(items, untilList...))
}

// Int8EnumChecker is an object checking an item type of int8
// and its ordering number in for some condition.
type Int8EnumChecker interface {
	// Check checks an item type of int8 and its ordering number for some condition.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Check(int, int8) (bool, error)
}

// Int8EnumCheck is a shortcut implementation
// of Int8EnumChecker based on a function.
type Int8EnumCheck func(int, int8) (bool, error)

// Check checks an item type of int8 and its ordering number for some condition.
// It returns EndOfInt8Iterator to stop iteration.
func (ch Int8EnumCheck) Check(n int, item int8) (bool, error) { return ch(n, item) }

type enumFromInt8Checker struct {
	Int8Checker
}

func (ch enumFromInt8Checker) Check(_ int, item int8) (bool, error) {
	return ch.Int8Checker.Check(item)
}

// EnumFromInt8Checker adapts checker type of Int8Checker
// to the interface Int8EnumChecker.
// If checker is nil it is return based on AlwaysInt8CheckFalse enum checker.
func EnumFromInt8Checker(checker Int8Checker) Int8EnumChecker {
	if checker == nil {
		checker = AlwaysInt8CheckFalse
	}
	return &enumFromInt8Checker{checker}
}

var (
	// AlwaysInt8EnumCheckTrue always returns true and empty error.
	AlwaysInt8EnumCheckTrue = EnumFromInt8Checker(
		AlwaysInt8CheckTrue)
	// AlwaysInt8EnumCheckFalse always returns false and empty error.
	AlwaysInt8EnumCheckFalse = EnumFromInt8Checker(
		AlwaysInt8CheckFalse)
)

// EnumNotInt8 do an inversion for checker result.
// It is returns AlwaysInt8EnumCheckTrue if checker is nil.
func EnumNotInt8(checker Int8EnumChecker) Int8EnumChecker {
	if checker == nil {
		return AlwaysInt8EnumCheckTrue
	}
	return Int8EnumCheck(func(n int, item int8) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndInt8 struct {
	lhs, rhs Int8EnumChecker
}

func (a enumAndInt8) Check(n int, item int8) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllInt8 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllInt8(checkers ...Int8EnumChecker) Int8EnumChecker {
	var all = AlwaysInt8EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndInt8{checkers[i], all}
	}
	return all
}

type enumOrInt8 struct {
	lhs, rhs Int8EnumChecker
}

func (o enumOrInt8) Check(n int, item int8) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt8Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyInt8 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyInt8(checkers ...Int8EnumChecker) Int8EnumChecker {
	var any = AlwaysInt8EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrInt8{checkers[i], any}
	}
	return any
}

// EnumFilteringInt8Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringInt8Iterator struct {
	preparedInt8Item
	filter Int8EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Int8EnumFiltering(items Int8Iterator, filters ...Int8EnumChecker) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &EnumFilteringInt8Iterator{preparedInt8Item{base: items}, EnumAllInt8(filters...), 0}
}

// EnumDoingUntilInt8Iterator does iteration
// until previously set checker is passed.
type EnumDoingUntilInt8Iterator struct {
	preparedInt8Item
	until Int8EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfInt8Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int8EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int8EnumDoingUntil(items Int8Iterator, untilList ...Int8EnumChecker) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	var until Int8EnumChecker
	if len(untilList) > 0 {
		until = EnumAllInt8(untilList...)
	} else {
		until = AlwaysInt8EnumCheckFalse
	}
	return &EnumDoingUntilInt8Iterator{preparedInt8Item{base: items}, until, 0}
}

// Int8EnumSkipUntil sets until conditions to skip few items.
func Int8EnumSkipUntil(items Int8Iterator, untilList ...Int8EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int8Discard(Int8EnumDoingUntil(items, untilList...))
}

// Int8GettingBatch returns the next batch from items.
func Int8GettingBatch(items Int8Iterator, batchSize int) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	if batchSize == 0 {
		return items
	}

	return Int8EnumDoingUntil(items, Int8EnumCheck(func(n int, item int8) (bool, error) {
		return n == batchSize-1, nil
	}))
}

// Int8Converter is an object converting an item type of int8.
type Int8Converter interface {
	// Convert should convert an item type of int8 into another item of int8.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Convert(int8) (int8, error)
}

// Int8Convert is a shortcut implementation
// of Int8Converter based on a function.
type Int8Convert func(int8) (int8, error)

// Convert converts an item type of int8 into another item of int8.
// It is suggested to return EndOfInt8Iterator to stop iteration.
func (c Int8Convert) Convert(item int8) (int8, error) { return c(item) }

// NoInt8Convert does nothing with item, just returns it as is.
var NoInt8Convert Int8Converter = Int8Convert(
	func(item int8) (int8, error) { return item, nil })

type doubleInt8Converter struct {
	lhs, rhs Int8Converter
}

func (c doubleInt8Converter) Convert(item int8) (int8, error) {
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

// Int8ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Int8ConverterSeries(converters ...Int8Converter) Int8Converter {
	var series = NoInt8Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt8Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingInt8Iterator does iteration with
// converting by previously set converter.
type ConvertingInt8Iterator struct {
	preparedInt8Item
	converter Int8Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Int8Converting(items Int8Iterator, converters ...Int8Converter) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &ConvertingInt8Iterator{
		preparedInt8Item{base: items}, Int8ConverterSeries(converters...)}
}

// Int8EnumConverter is an object converting an item type of int8 and its ordering number.
type Int8EnumConverter interface {
	// Convert should convert an item type of int8 into another item of int8.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Convert(n int, val int8) (int8, error)
}

// Int8EnumConvert is a shortcut implementation
// of Int8EnumConverter based on a function.
type Int8EnumConvert func(int, int8) (int8, error)

// Convert converts an item type of int8 into another item of int8.
// It is suggested to return EndOfInt8Iterator to stop iteration.
func (c Int8EnumConvert) Convert(n int, item int8) (int8, error) { return c(n, item) }

// NoInt8EnumConvert does nothing with item, just returns it as is.
var NoInt8EnumConvert Int8EnumConverter = Int8EnumConvert(
	func(_ int, item int8) (int8, error) { return item, nil })

type enumFromInt8Converter struct {
	Int8Converter
}

func (ch enumFromInt8Converter) Convert(_ int, item int8) (int8, error) {
	return ch.Int8Converter.Convert(item)
}

// EnumFromInt8Converter adapts checker type of Int8Converter
// to the interface Int8EnumConverter.
// If converter is nil it is return based on NoInt8Convert enum checker.
func EnumFromInt8Converter(converter Int8Converter) Int8EnumConverter {
	if converter == nil {
		converter = NoInt8Convert
	}
	return &enumFromInt8Converter{converter}
}

type doubleInt8EnumConverter struct {
	lhs, rhs Int8EnumConverter
}

func (c doubleInt8EnumConverter) Convert(n int, item int8) (int8, error) {
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

// EnumInt8ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumInt8ConverterSeries(converters ...Int8EnumConverter) Int8EnumConverter {
	var series = NoInt8EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt8EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingInt8Iterator does iteration with
// converting by previously set converter.
type EnumConvertingInt8Iterator struct {
	preparedInt8Item
	converter Int8EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Int8EnumConverting(items Int8Iterator, converters ...Int8EnumConverter) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &EnumConvertingInt8Iterator{
		preparedInt8Item{base: items}, EnumInt8ConverterSeries(converters...), 0}
}

// Int8Handler is an object handling an item type of int8.
type Int8Handler interface {
	// Handle should do something with item of int8.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Handle(int8) error
}

// Int8Handle is a shortcut implementation
// of Int8Handler based on a function.
type Int8Handle func(int8) error

// Handle does something with item of int8.
// It is suggested to return EndOfInt8Iterator to stop iteration.
func (h Int8Handle) Handle(item int8) error { return h(item) }

// Int8DoNothing does nothing.
var Int8DoNothing Int8Handler = Int8Handle(func(_ int8) error { return nil })

type doubleInt8Handler struct {
	lhs, rhs Int8Handler
}

func (h doubleInt8Handler) Handle(item int8) error {
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

// Int8HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int8HandlerSeries(handlers ...Int8Handler) Int8Handler {
	var series = Int8DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt8Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingInt8Iterator does iteration with
// handling by previously set handler.
type HandlingInt8Iterator struct {
	preparedInt8Item
	handler Int8Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Int8Handling(items Int8Iterator, handlers ...Int8Handler) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &HandlingInt8Iterator{
		preparedInt8Item{base: items}, Int8HandlerSeries(handlers...)}
}

// Int8Range iterates over items and use handlers to each one.
func Int8Range(items Int8Iterator, handlers ...Int8Handler) error {
	return Int8Discard(Int8Handling(items, handlers...))
}

// Int8RangeIterator is an iterator over items.
type Int8RangeIterator interface {
	// Range should iterate over items.
	Range(...Int8Handler) error
}

type sInt8RangeIterator struct {
	iter Int8Iterator
}

// ToInt8RangeIterator constructs an instance implementing Int8RangeIterator
// based on Int8Iterator.
func ToInt8RangeIterator(iter Int8Iterator) Int8RangeIterator {
	if iter == nil {
		iter = EmptyInt8Iterator
	}
	return sInt8RangeIterator{iter: iter}
}

// MakeInt8RangeIterator constructs an instance implementing Int8RangeIterator
// based on Int8IterMaker.
func MakeInt8RangeIterator(maker Int8IterMaker) Int8RangeIterator {
	if maker == nil {
		maker = MakeNoInt8Iter
	}
	return ToInt8RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sInt8RangeIterator) Range(handlers ...Int8Handler) error {
	return Int8Range(r.iter, handlers...)
}

// Int8EnumHandler is an object handling an item type of int8 and its ordered number.
type Int8EnumHandler interface {
	// Handle should do something with item of int8 and its ordered number.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Handle(int, int8) error
}

// Int8EnumHandle is a shortcut implementation
// of Int8EnumHandler based on a function.
type Int8EnumHandle func(int, int8) error

// Handle does something with item of int8 and its ordered number.
// It is suggested to return EndOfInt8Iterator to stop iteration.
func (h Int8EnumHandle) Handle(n int, item int8) error { return h(n, item) }

// Int8DoEnumNothing does nothing.
var Int8DoEnumNothing = Int8EnumHandle(func(_ int, _ int8) error { return nil })

type doubleInt8EnumHandler struct {
	lhs, rhs Int8EnumHandler
}

func (h doubleInt8EnumHandler) Handle(n int, item int8) error {
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

// Int8EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int8EnumHandlerSeries(handlers ...Int8EnumHandler) Int8EnumHandler {
	var series Int8EnumHandler = Int8DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt8EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingInt8Iterator does iteration with
// handling by previously set handler.
type EnumHandlingInt8Iterator struct {
	preparedInt8Item
	handler Int8EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Int8EnumHandling(items Int8Iterator, handlers ...Int8EnumHandler) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &EnumHandlingInt8Iterator{
		preparedInt8Item{base: items}, Int8EnumHandlerSeries(handlers...), 0}
}

// Int8Enum iterates over items and their ordering numbers and use handlers to each one.
func Int8Enum(items Int8Iterator, handlers ...Int8EnumHandler) error {
	return Int8Discard(Int8EnumHandling(items, handlers...))
}

// Int8EnumIterator is an iterator over items and their ordering numbers.
type Int8EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Int8EnumHandler) error
}

type sInt8EnumIterator struct {
	iter Int8Iterator
}

// ToInt8EnumIterator constructs an instance implementing Int8EnumIterator
// based on Int8Iterator.
func ToInt8EnumIterator(iter Int8Iterator) Int8EnumIterator {
	if iter == nil {
		iter = EmptyInt8Iterator
	}
	return sInt8EnumIterator{iter: iter}
}

// MakeInt8EnumIterator constructs an instance implementing Int8EnumIterator
// based on Int8IterMaker.
func MakeInt8EnumIterator(maker Int8IterMaker) Int8EnumIterator {
	if maker == nil {
		maker = MakeNoInt8Iter
	}
	return ToInt8EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sInt8EnumIterator) Enum(handlers ...Int8EnumHandler) error {
	return Int8Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sInt8EnumIterator) Range(handlers ...Int8Handler) error {
	return Int8Range(r.iter, handlers...)
}

type doubleInt8Iterator struct {
	lhs, rhs Int8Iterator
	inRHS    bool
}

func (it *doubleInt8Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleInt8Iterator) Next() int8 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleInt8Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperInt8Iterator combines all iterators to one.
func SuperInt8Iterator(itemList ...Int8Iterator) Int8Iterator {
	var super = EmptyInt8Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleInt8Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Int8Comparer is a strategy to compare two types.
type Int8Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs int8) bool
}

// Int8Compare is a shortcut implementation
// of Int8Comparer based on a function.
type Int8Compare func(lhs, rhs int8) bool

// IsLess is true if lhs is less than rhs.
func (c Int8Compare) IsLess(lhs, rhs int8) bool { return c(lhs, rhs) }

// Int8AlwaysLess is an implementation of Int8Comparer returning always true.
var Int8AlwaysLess Int8Comparer = Int8Compare(func(_, _ int8) bool { return true })

type priorityInt8Iterator struct {
	lhs, rhs preparedInt8Item
	comparer Int8Comparer
}

func (it *priorityInt8Iterator) HasNext() bool {
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

func (it *priorityInt8Iterator) Next() int8 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfInt8IteratorError(
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

func (it priorityInt8Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorInt8Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorInt8Iterator(comparer Int8Comparer, itemList ...Int8Iterator) Int8Iterator {
	if comparer == nil {
		comparer = Int8AlwaysLess
	}

	var prior = EmptyInt8Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityInt8Iterator{
			lhs:      preparedInt8Item{base: itemList[i]},
			rhs:      preparedInt8Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Int8SliceIterator is an iterator based on a slice of int8.
type Int8SliceIterator struct {
	slice []int8
	cur   int
}

// NewInt8SliceIterator returns a new instance of Int8SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Int8Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewInt8SliceIterator(slice []int8) *Int8SliceIterator {
	it := &Int8SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Int8SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Int8SliceIterator) Next() int8 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Int8SliceIterator) Err() error { return nil }

// Int8SliceIterator is an iterator based on a slice of int8
// and doing iteration in back direction.
type InvertingInt8SliceIterator struct {
	slice []int8
	cur   int
}

// NewInvertingInt8SliceIterator returns a new instance of InvertingInt8SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingInt8Slice(Int8Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingInt8SliceIterator(slice []int8) *InvertingInt8SliceIterator {
	it := &InvertingInt8SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingInt8SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingInt8SliceIterator) Next() int8 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingInt8SliceIterator) Err() error { return nil }

// Int8Unroll unrolls items to slice of int8.
func Int8Unroll(items Int8Iterator) Int8Slice {
	var slice Int8Slice
	panicIfInt8IteratorError(Int8Discard(Int8Handling(items, Int8Handle(func(item int8) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Int8Slice is a slice of int8.
type Int8Slice []int8

// MakeIter returns a new instance of Int8Iterator to iterate over it.
// It returns EmptyInt8Iterator if the error is not nil.
func (s Int8Slice) MakeIter() Int8Iterator {
	return NewInt8SliceIterator(s)
}

// Int8Slice is a slice of int8 which can make inverting iterator.
type InvertingInt8Slice []int8

// MakeIter returns a new instance of Int8Iterator to iterate over it.
// It returns EmptyInt8Iterator if the error is not nil.
func (s InvertingInt8Slice) MakeIter() Int8Iterator {
	return NewInvertingInt8SliceIterator(s)
}

// Int8Invert unrolls items and make inverting iterator based on them.
func Int8Invert(items Int8Iterator) Int8Iterator {
	return InvertingInt8Slice(Int8Unroll(items)).MakeIter()
}

// EndOfInt8Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfInt8Iterator = errors.New("end of int8 iterator")

func isEndOfInt8Iterator(err error) bool {
	return errors.Is(err, EndOfInt8Iterator)
}

func wrapIfNotEndOfInt8Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfInt8Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfInt8IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedInt8Item struct {
	base    Int8Iterator
	hasNext bool
	next    int8
	err     error
}

func (it *preparedInt8Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedInt8Item) Next() int8 {
	if !it.hasNext {
		panicIfInt8IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedInt8Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfInt8Iterator) {
		return it.err
	}
	return it.base.Err()
}
