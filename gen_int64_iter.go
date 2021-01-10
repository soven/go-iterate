package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// Int64Iterator is an iterator over items type of int64.
type Int64Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() int64
	// Err contains first met error while Next.
	Err() error
}

type emptyInt64Iterator struct{}

func (emptyInt64Iterator) HasNext() bool      { return false }
func (emptyInt64Iterator) Next() (next int64) { return 0 }
func (emptyInt64Iterator) Err() error         { return nil }

// EmptyInt64Iterator is a zero value for Int64Iterator.
// It is not contains any item to iterate over it.
var EmptyInt64Iterator Int64Iterator = emptyInt64Iterator{}

// Int64IterMaker is a maker of Int64Iterator.
type Int64IterMaker interface {
	// MakeIter should return a new instance of Int64Iterator to iterate over it.
	MakeIter() Int64Iterator
}

// MakeInt64Iter is a shortcut implementation
// of Int64Iterator based on a function.
type MakeInt64Iter func() Int64Iterator

// MakeIter returns a new instance of Int64Iterator to iterate over it.
func (m MakeInt64Iter) MakeIter() Int64Iterator { return m() }

// MakeNoInt64Iter is a zero value for Int64IterMaker.
// It always returns EmptyInt64Iterator and an empty error.
var MakeNoInt64Iter Int64IterMaker = MakeInt64Iter(
	func() Int64Iterator { return EmptyInt64Iterator })

// Int64Discard just range over all items and do nothing with each of them.
func Int64Discard(items Int64Iterator) error {
	if items == nil {
		items = EmptyInt64Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

// Int64Checker is an object checking an item type of int64
// for some condition.
type Int64Checker interface {
	// Check should check an item type of int64 for some condition.
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Check(int64) (bool, error)
}

// Int64Check is a shortcut implementation
// of Int64Checker based on a function.
type Int64Check func(int64) (bool, error)

// Check checks an item type of int64 for some condition.
// It returns EndOfInt64Iterator to stop iteration.
func (ch Int64Check) Check(item int64) (bool, error) { return ch(item) }

var (
	// AlwaysInt64CheckTrue always returns true and empty error.
	AlwaysInt64CheckTrue Int64Checker = Int64Check(
		func(item int64) (bool, error) { return true, nil })
	// AlwaysInt64CheckFalse always returns false and empty error.
	AlwaysInt64CheckFalse Int64Checker = Int64Check(
		func(item int64) (bool, error) { return false, nil })
)

// NotInt64 do an inversion for checker result.
// It is returns AlwaysInt64CheckTrue if checker is nil.
func NotInt64(checker Int64Checker) Int64Checker {
	if checker == nil {
		return AlwaysInt64CheckTrue
	}
	return Int64Check(func(item int64) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andInt64 struct {
	lhs, rhs Int64Checker
}

func (a andInt64) Check(item int64) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllInt64 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllInt64(checkers ...Int64Checker) Int64Checker {
	var all = AlwaysInt64CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andInt64{checkers[i], all}
	}
	return all
}

type orInt64 struct {
	lhs, rhs Int64Checker
}

func (o orInt64) Check(item int64) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyInt64 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyInt64(checkers ...Int64Checker) Int64Checker {
	var any = AlwaysInt64CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orInt64{checkers[i], any}
	}
	return any
}

// FilteringInt64Iterator does iteration with
// filtering by previously set checker.
type FilteringInt64Iterator struct {
	preparedInt64Item
	filter Int64Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Int64Filtering(items Int64Iterator, filters ...Int64Checker) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &FilteringInt64Iterator{preparedInt64Item{base: items}, AllInt64(filters...)}
}

// DoingUntilInt64Iterator does iteration
// until previously set checker is passed.
type DoingUntilInt64Iterator struct {
	preparedInt64Item
	until Int64Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfInt64Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int64DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int64DoingUntil(items Int64Iterator, untilList ...Int64Checker) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	var until Int64Checker
	if len(untilList) > 0 {
		until = AllInt64(untilList...)
	} else {
		until = AlwaysInt64CheckFalse
	}
	return &DoingUntilInt64Iterator{preparedInt64Item{base: items}, until}
}

// Int64SkipUntil sets until conditions to skip few items.
func Int64SkipUntil(items Int64Iterator, untilList ...Int64Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int64Discard(Int64DoingUntil(items, untilList...))
}

// Int64EnumChecker is an object checking an item type of int64
// and its ordering number in for some condition.
type Int64EnumChecker interface {
	// Check checks an item type of int64 and its ordering number for some condition.
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Check(int, int64) (bool, error)
}

// Int64EnumCheck is a shortcut implementation
// of Int64EnumChecker based on a function.
type Int64EnumCheck func(int, int64) (bool, error)

// Check checks an item type of int64 and its ordering number for some condition.
// It returns EndOfInt64Iterator to stop iteration.
func (ch Int64EnumCheck) Check(n int, item int64) (bool, error) { return ch(n, item) }

type enumFromInt64Checker struct {
	Int64Checker
}

func (ch enumFromInt64Checker) Check(_ int, item int64) (bool, error) {
	return ch.Int64Checker.Check(item)
}

// EnumFromInt64Checker adapts checker type of Int64Checker
// to the interface Int64EnumChecker.
// If checker is nil it is return based on AlwaysInt64CheckFalse enum checker.
func EnumFromInt64Checker(checker Int64Checker) Int64EnumChecker {
	if checker == nil {
		checker = AlwaysInt64CheckFalse
	}
	return &enumFromInt64Checker{checker}
}

var (
	// AlwaysInt64EnumCheckTrue always returns true and empty error.
	AlwaysInt64EnumCheckTrue = EnumFromInt64Checker(
		AlwaysInt64CheckTrue)
	// AlwaysInt64EnumCheckFalse always returns false and empty error.
	AlwaysInt64EnumCheckFalse = EnumFromInt64Checker(
		AlwaysInt64CheckFalse)
)

// EnumNotInt64 do an inversion for checker result.
// It is returns AlwaysInt64EnumCheckTrue if checker is nil.
func EnumNotInt64(checker Int64EnumChecker) Int64EnumChecker {
	if checker == nil {
		return AlwaysInt64EnumCheckTrue
	}
	return Int64EnumCheck(func(n int, item int64) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndInt64 struct {
	lhs, rhs Int64EnumChecker
}

func (a enumAndInt64) Check(n int, item int64) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllInt64 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllInt64(checkers ...Int64EnumChecker) Int64EnumChecker {
	var all = AlwaysInt64EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndInt64{checkers[i], all}
	}
	return all
}

type enumOrInt64 struct {
	lhs, rhs Int64EnumChecker
}

func (o enumOrInt64) Check(n int, item int64) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt64Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyInt64 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyInt64(checkers ...Int64EnumChecker) Int64EnumChecker {
	var any = AlwaysInt64EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrInt64{checkers[i], any}
	}
	return any
}

// EnumFilteringInt64Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringInt64Iterator struct {
	preparedInt64Item
	filter Int64EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Int64EnumFiltering(items Int64Iterator, filters ...Int64EnumChecker) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &EnumFilteringInt64Iterator{preparedInt64Item{base: items}, EnumAllInt64(filters...), 0}
}

// EnumDoingUntilInt64Iterator does iteration
// until previously set checker is passed.
type EnumDoingUntilInt64Iterator struct {
	preparedInt64Item
	until Int64EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfInt64Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int64EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int64EnumDoingUntil(items Int64Iterator, untilList ...Int64EnumChecker) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	var until Int64EnumChecker
	if len(untilList) > 0 {
		until = EnumAllInt64(untilList...)
	} else {
		until = AlwaysInt64EnumCheckFalse
	}
	return &EnumDoingUntilInt64Iterator{preparedInt64Item{base: items}, until, 0}
}

// Int64EnumSkipUntil sets until conditions to skip few items.
func Int64EnumSkipUntil(items Int64Iterator, untilList ...Int64EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int64Discard(Int64EnumDoingUntil(items, untilList...))
}

// Int64GettingBatch returns the next batch from items.
func Int64GettingBatch(items Int64Iterator, batchSize int) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	if batchSize == 0 {
		return items
	}

	return Int64EnumDoingUntil(items, Int64EnumCheck(func(n int, item int64) (bool, error) {
		return n == batchSize-1, nil
	}))
}

// Int64Converter is an object converting an item type of int64.
type Int64Converter interface {
	// Convert should convert an item type of int64 into another item of int64.
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Convert(int64) (int64, error)
}

// Int64Convert is a shortcut implementation
// of Int64Converter based on a function.
type Int64Convert func(int64) (int64, error)

// Convert converts an item type of int64 into another item of int64.
// It is suggested to return EndOfInt64Iterator to stop iteration.
func (c Int64Convert) Convert(item int64) (int64, error) { return c(item) }

// NoInt64Convert does nothing with item, just returns it as is.
var NoInt64Convert Int64Converter = Int64Convert(
	func(item int64) (int64, error) { return item, nil })

type doubleInt64Converter struct {
	lhs, rhs Int64Converter
}

func (c doubleInt64Converter) Convert(item int64) (int64, error) {
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

// Int64ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Int64ConverterSeries(converters ...Int64Converter) Int64Converter {
	var series = NoInt64Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt64Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingInt64Iterator does iteration with
// converting by previously set converter.
type ConvertingInt64Iterator struct {
	preparedInt64Item
	converter Int64Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Int64Converting(items Int64Iterator, converters ...Int64Converter) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &ConvertingInt64Iterator{
		preparedInt64Item{base: items}, Int64ConverterSeries(converters...)}
}

// Int64EnumConverter is an object converting an item type of int64 and its ordering number.
type Int64EnumConverter interface {
	// Convert should convert an item type of int64 into another item of int64.
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Convert(n int, val int64) (int64, error)
}

// Int64EnumConvert is a shortcut implementation
// of Int64EnumConverter based on a function.
type Int64EnumConvert func(int, int64) (int64, error)

// Convert converts an item type of int64 into another item of int64.
// It is suggested to return EndOfInt64Iterator to stop iteration.
func (c Int64EnumConvert) Convert(n int, item int64) (int64, error) { return c(n, item) }

// NoInt64EnumConvert does nothing with item, just returns it as is.
var NoInt64EnumConvert Int64EnumConverter = Int64EnumConvert(
	func(_ int, item int64) (int64, error) { return item, nil })

type enumFromInt64Converter struct {
	Int64Converter
}

func (ch enumFromInt64Converter) Convert(_ int, item int64) (int64, error) {
	return ch.Int64Converter.Convert(item)
}

// EnumFromInt64Converter adapts checker type of Int64Converter
// to the interface Int64EnumConverter.
// If converter is nil it is return based on NoInt64Convert enum checker.
func EnumFromInt64Converter(converter Int64Converter) Int64EnumConverter {
	if converter == nil {
		converter = NoInt64Convert
	}
	return &enumFromInt64Converter{converter}
}

type doubleInt64EnumConverter struct {
	lhs, rhs Int64EnumConverter
}

func (c doubleInt64EnumConverter) Convert(n int, item int64) (int64, error) {
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

// EnumInt64ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumInt64ConverterSeries(converters ...Int64EnumConverter) Int64EnumConverter {
	var series = NoInt64EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt64EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingInt64Iterator does iteration with
// converting by previously set converter.
type EnumConvertingInt64Iterator struct {
	preparedInt64Item
	converter Int64EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Int64EnumConverting(items Int64Iterator, converters ...Int64EnumConverter) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &EnumConvertingInt64Iterator{
		preparedInt64Item{base: items}, EnumInt64ConverterSeries(converters...), 0}
}

// Int64Handler is an object handling an item type of int64.
type Int64Handler interface {
	// Handle should do something with item of int64.
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Handle(int64) error
}

// Int64Handle is a shortcut implementation
// of Int64Handler based on a function.
type Int64Handle func(int64) error

// Handle does something with item of int64.
// It is suggested to return EndOfInt64Iterator to stop iteration.
func (h Int64Handle) Handle(item int64) error { return h(item) }

// Int64DoNothing does nothing.
var Int64DoNothing Int64Handler = Int64Handle(func(_ int64) error { return nil })

type doubleInt64Handler struct {
	lhs, rhs Int64Handler
}

func (h doubleInt64Handler) Handle(item int64) error {
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

// Int64HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int64HandlerSeries(handlers ...Int64Handler) Int64Handler {
	var series = Int64DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt64Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingInt64Iterator does iteration with
// handling by previously set handler.
type HandlingInt64Iterator struct {
	preparedInt64Item
	handler Int64Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Int64Handling(items Int64Iterator, handlers ...Int64Handler) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &HandlingInt64Iterator{
		preparedInt64Item{base: items}, Int64HandlerSeries(handlers...)}
}

// Int64Range iterates over items and use handlers to each one.
func Int64Range(items Int64Iterator, handlers ...Int64Handler) error {
	return Int64Discard(Int64Handling(items, handlers...))
}

// Int64RangeIterator is an iterator over items.
type Int64RangeIterator interface {
	// Range should iterate over items.
	Range(...Int64Handler) error
}

type sInt64RangeIterator struct {
	iter Int64Iterator
}

// ToInt64RangeIterator constructs an instance implementing Int64RangeIterator
// based on Int64Iterator.
func ToInt64RangeIterator(iter Int64Iterator) Int64RangeIterator {
	if iter == nil {
		iter = EmptyInt64Iterator
	}
	return sInt64RangeIterator{iter: iter}
}

// MakeInt64RangeIterator constructs an instance implementing Int64RangeIterator
// based on Int64IterMaker.
func MakeInt64RangeIterator(maker Int64IterMaker) Int64RangeIterator {
	if maker == nil {
		maker = MakeNoInt64Iter
	}
	return ToInt64RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sInt64RangeIterator) Range(handlers ...Int64Handler) error {
	return Int64Range(r.iter, handlers...)
}

// Int64EnumHandler is an object handling an item type of int64 and its ordered number.
type Int64EnumHandler interface {
	// Handle should do something with item of int64 and its ordered number.
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Handle(int, int64) error
}

// Int64EnumHandle is a shortcut implementation
// of Int64EnumHandler based on a function.
type Int64EnumHandle func(int, int64) error

// Handle does something with item of int64 and its ordered number.
// It is suggested to return EndOfInt64Iterator to stop iteration.
func (h Int64EnumHandle) Handle(n int, item int64) error { return h(n, item) }

// Int64DoEnumNothing does nothing.
var Int64DoEnumNothing = Int64EnumHandle(func(_ int, _ int64) error { return nil })

type doubleInt64EnumHandler struct {
	lhs, rhs Int64EnumHandler
}

func (h doubleInt64EnumHandler) Handle(n int, item int64) error {
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

// Int64EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int64EnumHandlerSeries(handlers ...Int64EnumHandler) Int64EnumHandler {
	var series Int64EnumHandler = Int64DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt64EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingInt64Iterator does iteration with
// handling by previously set handler.
type EnumHandlingInt64Iterator struct {
	preparedInt64Item
	handler Int64EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Int64EnumHandling(items Int64Iterator, handlers ...Int64EnumHandler) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &EnumHandlingInt64Iterator{
		preparedInt64Item{base: items}, Int64EnumHandlerSeries(handlers...), 0}
}

// Int64Enum iterates over items and their ordering numbers and use handlers to each one.
func Int64Enum(items Int64Iterator, handlers ...Int64EnumHandler) error {
	return Int64Discard(Int64EnumHandling(items, handlers...))
}

// Int64EnumIterator is an iterator over items and their ordering numbers.
type Int64EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Int64EnumHandler) error
}

type sInt64EnumIterator struct {
	iter Int64Iterator
}

// ToInt64EnumIterator constructs an instance implementing Int64EnumIterator
// based on Int64Iterator.
func ToInt64EnumIterator(iter Int64Iterator) Int64EnumIterator {
	if iter == nil {
		iter = EmptyInt64Iterator
	}
	return sInt64EnumIterator{iter: iter}
}

// MakeInt64EnumIterator constructs an instance implementing Int64EnumIterator
// based on Int64IterMaker.
func MakeInt64EnumIterator(maker Int64IterMaker) Int64EnumIterator {
	if maker == nil {
		maker = MakeNoInt64Iter
	}
	return ToInt64EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sInt64EnumIterator) Enum(handlers ...Int64EnumHandler) error {
	return Int64Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sInt64EnumIterator) Range(handlers ...Int64Handler) error {
	return Int64Range(r.iter, handlers...)
}

type doubleInt64Iterator struct {
	lhs, rhs Int64Iterator
	inRHS    bool
}

func (it *doubleInt64Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleInt64Iterator) Next() int64 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleInt64Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperInt64Iterator combines all iterators to one.
func SuperInt64Iterator(itemList ...Int64Iterator) Int64Iterator {
	var super = EmptyInt64Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleInt64Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Int64Comparer is a strategy to compare two types.
type Int64Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs int64) bool
}

// Int64Compare is a shortcut implementation
// of Int64Comparer based on a function.
type Int64Compare func(lhs, rhs int64) bool

// IsLess is true if lhs is less than rhs.
func (c Int64Compare) IsLess(lhs, rhs int64) bool { return c(lhs, rhs) }

// Int64AlwaysLess is an implementation of Int64Comparer returning always true.
var Int64AlwaysLess Int64Comparer = Int64Compare(func(_, _ int64) bool { return true })

type priorityInt64Iterator struct {
	lhs, rhs preparedInt64Item
	comparer Int64Comparer
}

func (it *priorityInt64Iterator) HasNext() bool {
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

func (it *priorityInt64Iterator) Next() int64 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfInt64IteratorError(
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

func (it priorityInt64Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorInt64Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorInt64Iterator(comparer Int64Comparer, itemList ...Int64Iterator) Int64Iterator {
	if comparer == nil {
		comparer = Int64AlwaysLess
	}

	var prior = EmptyInt64Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityInt64Iterator{
			lhs:      preparedInt64Item{base: itemList[i]},
			rhs:      preparedInt64Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Int64SliceIterator is an iterator based on a slice of int64.
type Int64SliceIterator struct {
	slice []int64
	cur   int
}

// NewInt64SliceIterator returns a new instance of Int64SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Int64Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewInt64SliceIterator(slice []int64) *Int64SliceIterator {
	it := &Int64SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Int64SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Int64SliceIterator) Next() int64 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Int64SliceIterator) Err() error { return nil }

// Int64SliceIterator is an iterator based on a slice of int64
// and doing iteration in back direction.
type InvertingInt64SliceIterator struct {
	slice []int64
	cur   int
}

// NewInvertingInt64SliceIterator returns a new instance of InvertingInt64SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingInt64Slice(Int64Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingInt64SliceIterator(slice []int64) *InvertingInt64SliceIterator {
	it := &InvertingInt64SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingInt64SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingInt64SliceIterator) Next() int64 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingInt64SliceIterator) Err() error { return nil }

// Int64Unroll unrolls items to slice of int64.
func Int64Unroll(items Int64Iterator) Int64Slice {
	var slice Int64Slice
	panicIfInt64IteratorError(Int64Discard(Int64Handling(items, Int64Handle(func(item int64) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Int64Slice is a slice of int64.
type Int64Slice []int64

// MakeIter returns a new instance of Int64Iterator to iterate over it.
// It returns EmptyInt64Iterator if the error is not nil.
func (s Int64Slice) MakeIter() Int64Iterator {
	return NewInt64SliceIterator(s)
}

// Int64Slice is a slice of int64 which can make inverting iterator.
type InvertingInt64Slice []int64

// MakeIter returns a new instance of Int64Iterator to iterate over it.
// It returns EmptyInt64Iterator if the error is not nil.
func (s InvertingInt64Slice) MakeIter() Int64Iterator {
	return NewInvertingInt64SliceIterator(s)
}

// Int64Invert unrolls items and make inverting iterator based on them.
func Int64Invert(items Int64Iterator) Int64Iterator {
	return InvertingInt64Slice(Int64Unroll(items)).MakeIter()
}

// EndOfInt64Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfInt64Iterator = errors.New("end of int64 iterator")

func isEndOfInt64Iterator(err error) bool {
	return errors.Is(err, EndOfInt64Iterator)
}

func wrapIfNotEndOfInt64Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfInt64Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfInt64IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedInt64Item struct {
	base    Int64Iterator
	hasNext bool
	next    int64
	err     error
}

func (it *preparedInt64Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedInt64Item) Next() int64 {
	if !it.hasNext {
		panicIfInt64IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedInt64Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfInt64Iterator) {
		return it.err
	}
	return it.base.Err()
}
