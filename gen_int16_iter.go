package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// Int16Iterator is an iterator over items type of int16.
type Int16Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() int16
	// Err contains first met error while Next.
	Err() error
}

type emptyInt16Iterator struct{}

func (emptyInt16Iterator) HasNext() bool      { return false }
func (emptyInt16Iterator) Next() (next int16) { return 0 }
func (emptyInt16Iterator) Err() error         { return nil }

// EmptyInt16Iterator is a zero value for Int16Iterator.
// It is not contains any item to iterate over it.
var EmptyInt16Iterator Int16Iterator = emptyInt16Iterator{}

// Int16IterMaker is a maker of Int16Iterator.
type Int16IterMaker interface {
	// MakeIter should return a new instance of Int16Iterator to iterate over it.
	MakeIter() Int16Iterator
}

// MakeInt16Iter is a shortcut implementation
// of Int16Iterator based on a function.
type MakeInt16Iter func() Int16Iterator

// MakeIter returns a new instance of Int16Iterator to iterate over it.
func (m MakeInt16Iter) MakeIter() Int16Iterator { return m() }

// MakeNoInt16Iter is a zero value for Int16IterMaker.
// It always returns EmptyInt16Iterator and an empty error.
var MakeNoInt16Iter Int16IterMaker = MakeInt16Iter(
	func() Int16Iterator { return EmptyInt16Iterator })

// Int16Discard just range over all items and do nothing with each of them.
func Int16Discard(items Int16Iterator) error {
	if items == nil {
		items = EmptyInt16Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

// Int16Checker is an object checking an item type of int16
// for some condition.
type Int16Checker interface {
	// Check should check an item type of int16 for some condition.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Check(int16) (bool, error)
}

// Int16Check is a shortcut implementation
// of Int16Checker based on a function.
type Int16Check func(int16) (bool, error)

// Check checks an item type of int16 for some condition.
// It returns EndOfInt16Iterator to stop iteration.
func (ch Int16Check) Check(item int16) (bool, error) { return ch(item) }

var (
	// AlwaysInt16CheckTrue always returns true and empty error.
	AlwaysInt16CheckTrue Int16Checker = Int16Check(
		func(item int16) (bool, error) { return true, nil })
	// AlwaysInt16CheckFalse always returns false and empty error.
	AlwaysInt16CheckFalse Int16Checker = Int16Check(
		func(item int16) (bool, error) { return false, nil })
)

// NotInt16 do an inversion for checker result.
// It is returns AlwaysInt16CheckTrue if checker is nil.
func NotInt16(checker Int16Checker) Int16Checker {
	if checker == nil {
		return AlwaysInt16CheckTrue
	}
	return Int16Check(func(item int16) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andInt16 struct {
	lhs, rhs Int16Checker
}

func (a andInt16) Check(item int16) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllInt16 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllInt16(checkers ...Int16Checker) Int16Checker {
	var all = AlwaysInt16CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andInt16{checkers[i], all}
	}
	return all
}

type orInt16 struct {
	lhs, rhs Int16Checker
}

func (o orInt16) Check(item int16) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyInt16 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyInt16(checkers ...Int16Checker) Int16Checker {
	var any = AlwaysInt16CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orInt16{checkers[i], any}
	}
	return any
}

// FilteringInt16Iterator does iteration with
// filtering by previously set checker.
type FilteringInt16Iterator struct {
	preparedInt16Item
	filter Int16Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Int16Filtering(items Int16Iterator, filters ...Int16Checker) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &FilteringInt16Iterator{preparedInt16Item{base: items}, AllInt16(filters...)}
}

// DoingUntilInt16Iterator does iteration
// until previously set checker is passed.
type DoingUntilInt16Iterator struct {
	preparedInt16Item
	until Int16Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfInt16Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int16DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int16DoingUntil(items Int16Iterator, untilList ...Int16Checker) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	var until Int16Checker
	if len(untilList) > 0 {
		until = AllInt16(untilList...)
	} else {
		until = AlwaysInt16CheckFalse
	}
	return &DoingUntilInt16Iterator{preparedInt16Item{base: items}, until}
}

// Int16SkipUntil sets until conditions to skip few items.
func Int16SkipUntil(items Int16Iterator, untilList ...Int16Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int16Discard(Int16DoingUntil(items, untilList...))
}

// Int16EnumChecker is an object checking an item type of int16
// and its ordering number in for some condition.
type Int16EnumChecker interface {
	// Check checks an item type of int16 and its ordering number for some condition.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Check(int, int16) (bool, error)
}

// Int16EnumCheck is a shortcut implementation
// of Int16EnumChecker based on a function.
type Int16EnumCheck func(int, int16) (bool, error)

// Check checks an item type of int16 and its ordering number for some condition.
// It returns EndOfInt16Iterator to stop iteration.
func (ch Int16EnumCheck) Check(n int, item int16) (bool, error) { return ch(n, item) }

type enumFromInt16Checker struct {
	Int16Checker
}

func (ch enumFromInt16Checker) Check(_ int, item int16) (bool, error) {
	return ch.Int16Checker.Check(item)
}

// EnumFromInt16Checker adapts checker type of Int16Checker
// to the interface Int16EnumChecker.
// If checker is nil it is return based on AlwaysInt16CheckFalse enum checker.
func EnumFromInt16Checker(checker Int16Checker) Int16EnumChecker {
	if checker == nil {
		checker = AlwaysInt16CheckFalse
	}
	return &enumFromInt16Checker{checker}
}

var (
	// AlwaysInt16EnumCheckTrue always returns true and empty error.
	AlwaysInt16EnumCheckTrue = EnumFromInt16Checker(
		AlwaysInt16CheckTrue)
	// AlwaysInt16EnumCheckFalse always returns false and empty error.
	AlwaysInt16EnumCheckFalse = EnumFromInt16Checker(
		AlwaysInt16CheckFalse)
)

// EnumNotInt16 do an inversion for checker result.
// It is returns AlwaysInt16EnumCheckTrue if checker is nil.
func EnumNotInt16(checker Int16EnumChecker) Int16EnumChecker {
	if checker == nil {
		return AlwaysInt16EnumCheckTrue
	}
	return Int16EnumCheck(func(n int, item int16) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndInt16 struct {
	lhs, rhs Int16EnumChecker
}

func (a enumAndInt16) Check(n int, item int16) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllInt16 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllInt16(checkers ...Int16EnumChecker) Int16EnumChecker {
	var all = AlwaysInt16EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndInt16{checkers[i], all}
	}
	return all
}

type enumOrInt16 struct {
	lhs, rhs Int16EnumChecker
}

func (o enumOrInt16) Check(n int, item int16) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt16Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyInt16 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyInt16(checkers ...Int16EnumChecker) Int16EnumChecker {
	var any = AlwaysInt16EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrInt16{checkers[i], any}
	}
	return any
}

// EnumFilteringInt16Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringInt16Iterator struct {
	preparedInt16Item
	filter Int16EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Int16EnumFiltering(items Int16Iterator, filters ...Int16EnumChecker) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &EnumFilteringInt16Iterator{preparedInt16Item{base: items}, EnumAllInt16(filters...), 0}
}

// EnumDoingUntilInt16Iterator does iteration
// until previously set checker is passed.
type EnumDoingUntilInt16Iterator struct {
	preparedInt16Item
	until Int16EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfInt16Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int16EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int16EnumDoingUntil(items Int16Iterator, untilList ...Int16EnumChecker) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	var until Int16EnumChecker
	if len(untilList) > 0 {
		until = EnumAllInt16(untilList...)
	} else {
		until = AlwaysInt16EnumCheckFalse
	}
	return &EnumDoingUntilInt16Iterator{preparedInt16Item{base: items}, until, 0}
}

// Int16EnumSkipUntil sets until conditions to skip few items.
func Int16EnumSkipUntil(items Int16Iterator, untilList ...Int16EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int16Discard(Int16EnumDoingUntil(items, untilList...))
}

// Int16GettingBatch returns the next batch from items.
func Int16GettingBatch(items Int16Iterator, batchSize int) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	if batchSize == 0 {
		return items
	}

	return Int16EnumDoingUntil(items, Int16EnumCheck(func(n int, item int16) (bool, error) {
		return n == batchSize-1, nil
	}))
}

// Int16Converter is an object converting an item type of int16.
type Int16Converter interface {
	// Convert should convert an item type of int16 into another item of int16.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Convert(int16) (int16, error)
}

// Int16Convert is a shortcut implementation
// of Int16Converter based on a function.
type Int16Convert func(int16) (int16, error)

// Convert converts an item type of int16 into another item of int16.
// It is suggested to return EndOfInt16Iterator to stop iteration.
func (c Int16Convert) Convert(item int16) (int16, error) { return c(item) }

// NoInt16Convert does nothing with item, just returns it as is.
var NoInt16Convert Int16Converter = Int16Convert(
	func(item int16) (int16, error) { return item, nil })

type doubleInt16Converter struct {
	lhs, rhs Int16Converter
}

func (c doubleInt16Converter) Convert(item int16) (int16, error) {
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

// Int16ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Int16ConverterSeries(converters ...Int16Converter) Int16Converter {
	var series = NoInt16Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt16Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingInt16Iterator does iteration with
// converting by previously set converter.
type ConvertingInt16Iterator struct {
	preparedInt16Item
	converter Int16Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Int16Converting(items Int16Iterator, converters ...Int16Converter) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &ConvertingInt16Iterator{
		preparedInt16Item{base: items}, Int16ConverterSeries(converters...)}
}

// Int16EnumConverter is an object converting an item type of int16 and its ordering number.
type Int16EnumConverter interface {
	// Convert should convert an item type of int16 into another item of int16.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Convert(n int, val int16) (int16, error)
}

// Int16EnumConvert is a shortcut implementation
// of Int16EnumConverter based on a function.
type Int16EnumConvert func(int, int16) (int16, error)

// Convert converts an item type of int16 into another item of int16.
// It is suggested to return EndOfInt16Iterator to stop iteration.
func (c Int16EnumConvert) Convert(n int, item int16) (int16, error) { return c(n, item) }

// NoInt16EnumConvert does nothing with item, just returns it as is.
var NoInt16EnumConvert Int16EnumConverter = Int16EnumConvert(
	func(_ int, item int16) (int16, error) { return item, nil })

type enumFromInt16Converter struct {
	Int16Converter
}

func (ch enumFromInt16Converter) Convert(_ int, item int16) (int16, error) {
	return ch.Int16Converter.Convert(item)
}

// EnumFromInt16Converter adapts checker type of Int16Converter
// to the interface Int16EnumConverter.
// If converter is nil it is return based on NoInt16Convert enum checker.
func EnumFromInt16Converter(converter Int16Converter) Int16EnumConverter {
	if converter == nil {
		converter = NoInt16Convert
	}
	return &enumFromInt16Converter{converter}
}

type doubleInt16EnumConverter struct {
	lhs, rhs Int16EnumConverter
}

func (c doubleInt16EnumConverter) Convert(n int, item int16) (int16, error) {
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

// EnumInt16ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumInt16ConverterSeries(converters ...Int16EnumConverter) Int16EnumConverter {
	var series = NoInt16EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt16EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingInt16Iterator does iteration with
// converting by previously set converter.
type EnumConvertingInt16Iterator struct {
	preparedInt16Item
	converter Int16EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Int16EnumConverting(items Int16Iterator, converters ...Int16EnumConverter) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &EnumConvertingInt16Iterator{
		preparedInt16Item{base: items}, EnumInt16ConverterSeries(converters...), 0}
}

// Int16Handler is an object handling an item type of int16.
type Int16Handler interface {
	// Handle should do something with item of int16.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Handle(int16) error
}

// Int16Handle is a shortcut implementation
// of Int16Handler based on a function.
type Int16Handle func(int16) error

// Handle does something with item of int16.
// It is suggested to return EndOfInt16Iterator to stop iteration.
func (h Int16Handle) Handle(item int16) error { return h(item) }

// Int16DoNothing does nothing.
var Int16DoNothing Int16Handler = Int16Handle(func(_ int16) error { return nil })

type doubleInt16Handler struct {
	lhs, rhs Int16Handler
}

func (h doubleInt16Handler) Handle(item int16) error {
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

// Int16HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int16HandlerSeries(handlers ...Int16Handler) Int16Handler {
	var series = Int16DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt16Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingInt16Iterator does iteration with
// handling by previously set handler.
type HandlingInt16Iterator struct {
	preparedInt16Item
	handler Int16Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Int16Handling(items Int16Iterator, handlers ...Int16Handler) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &HandlingInt16Iterator{
		preparedInt16Item{base: items}, Int16HandlerSeries(handlers...)}
}

// Int16Range iterates over items and use handlers to each one.
func Int16Range(items Int16Iterator, handlers ...Int16Handler) error {
	return Int16Discard(Int16Handling(items, handlers...))
}

// Int16RangeIterator is an iterator over items.
type Int16RangeIterator interface {
	// Range should iterate over items.
	Range(...Int16Handler) error
}

type sInt16RangeIterator struct {
	iter Int16Iterator
}

// ToInt16RangeIterator constructs an instance implementing Int16RangeIterator
// based on Int16Iterator.
func ToInt16RangeIterator(iter Int16Iterator) Int16RangeIterator {
	if iter == nil {
		iter = EmptyInt16Iterator
	}
	return sInt16RangeIterator{iter: iter}
}

// MakeInt16RangeIterator constructs an instance implementing Int16RangeIterator
// based on Int16IterMaker.
func MakeInt16RangeIterator(maker Int16IterMaker) Int16RangeIterator {
	if maker == nil {
		maker = MakeNoInt16Iter
	}
	return ToInt16RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sInt16RangeIterator) Range(handlers ...Int16Handler) error {
	return Int16Range(r.iter, handlers...)
}

// Int16EnumHandler is an object handling an item type of int16 and its ordered number.
type Int16EnumHandler interface {
	// Handle should do something with item of int16 and its ordered number.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Handle(int, int16) error
}

// Int16EnumHandle is a shortcut implementation
// of Int16EnumHandler based on a function.
type Int16EnumHandle func(int, int16) error

// Handle does something with item of int16 and its ordered number.
// It is suggested to return EndOfInt16Iterator to stop iteration.
func (h Int16EnumHandle) Handle(n int, item int16) error { return h(n, item) }

// Int16DoEnumNothing does nothing.
var Int16DoEnumNothing = Int16EnumHandle(func(_ int, _ int16) error { return nil })

type doubleInt16EnumHandler struct {
	lhs, rhs Int16EnumHandler
}

func (h doubleInt16EnumHandler) Handle(n int, item int16) error {
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

// Int16EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int16EnumHandlerSeries(handlers ...Int16EnumHandler) Int16EnumHandler {
	var series Int16EnumHandler = Int16DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt16EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingInt16Iterator does iteration with
// handling by previously set handler.
type EnumHandlingInt16Iterator struct {
	preparedInt16Item
	handler Int16EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Int16EnumHandling(items Int16Iterator, handlers ...Int16EnumHandler) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &EnumHandlingInt16Iterator{
		preparedInt16Item{base: items}, Int16EnumHandlerSeries(handlers...), 0}
}

// Int16Enum iterates over items and their ordering numbers and use handlers to each one.
func Int16Enum(items Int16Iterator, handlers ...Int16EnumHandler) error {
	return Int16Discard(Int16EnumHandling(items, handlers...))
}

// Int16EnumIterator is an iterator over items and their ordering numbers.
type Int16EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Int16EnumHandler) error
}

type sInt16EnumIterator struct {
	iter Int16Iterator
}

// ToInt16EnumIterator constructs an instance implementing Int16EnumIterator
// based on Int16Iterator.
func ToInt16EnumIterator(iter Int16Iterator) Int16EnumIterator {
	if iter == nil {
		iter = EmptyInt16Iterator
	}
	return sInt16EnumIterator{iter: iter}
}

// MakeInt16EnumIterator constructs an instance implementing Int16EnumIterator
// based on Int16IterMaker.
func MakeInt16EnumIterator(maker Int16IterMaker) Int16EnumIterator {
	if maker == nil {
		maker = MakeNoInt16Iter
	}
	return ToInt16EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sInt16EnumIterator) Enum(handlers ...Int16EnumHandler) error {
	return Int16Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sInt16EnumIterator) Range(handlers ...Int16Handler) error {
	return Int16Range(r.iter, handlers...)
}

type doubleInt16Iterator struct {
	lhs, rhs Int16Iterator
	inRHS    bool
}

func (it *doubleInt16Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleInt16Iterator) Next() int16 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleInt16Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperInt16Iterator combines all iterators to one.
func SuperInt16Iterator(itemList ...Int16Iterator) Int16Iterator {
	var super = EmptyInt16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleInt16Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Int16EnumComparer is a strategy to compare two types.
type Int16Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs int16) bool
}

// Int16Compare is a shortcut implementation
// of Int16EnumComparer based on a function.
type Int16Compare func(lhs, rhs int16) bool

// IsLess is true if lhs is less than rhs.
func (c Int16Compare) IsLess(lhs, rhs int16) bool { return c(lhs, rhs) }

// EnumInt16AlwaysLess is an implementation of Int16EnumComparer returning always true.
var Int16AlwaysLess Int16Comparer = Int16Compare(func(_, _ int16) bool { return true })

type priorityInt16Iterator struct {
	lhs, rhs preparedInt16Item
	comparer Int16Comparer
}

func (it *priorityInt16Iterator) HasNext() bool {
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

func (it *priorityInt16Iterator) Next() int16 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfInt16IteratorError(
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

func (it priorityInt16Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorInt16Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorInt16Iterator(comparer Int16Comparer, itemList ...Int16Iterator) Int16Iterator {
	if comparer == nil {
		comparer = Int16AlwaysLess
	}

	var prior = EmptyInt16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityInt16Iterator{
			lhs:      preparedInt16Item{base: itemList[i]},
			rhs:      preparedInt16Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Int16EnumComparer is a strategy to compare two types and their order numbers.
type Int16EnumComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(nLHS int, lhs int16, nRHS int, rhs int16) bool
}

// Int16EnumCompare is a shortcut implementation
// of Int16EnumComparer based on a function.
type Int16EnumCompare func(nLHS int, lhs int16, nRHS int, rhs int16) bool

// IsLess is true if lhs is less than rhs.
func (c Int16EnumCompare) IsLess(nLHS int, lhs int16, nRHS int, rhs int16) bool {
	return c(nLHS, lhs, nRHS, rhs)
}

// EnumInt16AlwaysLess is an implementation of Int16EnumComparer returning always true.
var EnumInt16AlwaysLess Int16EnumComparer = Int16EnumCompare(
	func(_ int, _ int16, _ int, _ int16) bool { return true })

type priorityInt16EnumIterator struct {
	lhs, rhs           preparedInt16Item
	countLHS, countRHS int
	comparer           Int16EnumComparer
}

func (it *priorityInt16EnumIterator) HasNext() bool {
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

func (it *priorityInt16EnumIterator) Next() int16 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfInt16IteratorError(
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

func (it priorityInt16EnumIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorInt16EnumIterator compare one by one items and their ordering numbers fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorInt16EnumIterator(comparer Int16EnumComparer, itemList ...Int16Iterator) Int16Iterator {
	if comparer == nil {
		comparer = EnumInt16AlwaysLess
	}

	var prior = EmptyInt16Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityInt16EnumIterator{
			lhs:      preparedInt16Item{base: itemList[i]},
			rhs:      preparedInt16Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Int16SliceIterator is an iterator based on a slice of int16.
type Int16SliceIterator struct {
	slice []int16
	cur   int
}

// NewInt16SliceIterator returns a new instance of Int16SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Int16Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewInt16SliceIterator(slice []int16) *Int16SliceIterator {
	it := &Int16SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Int16SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Int16SliceIterator) Next() int16 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Int16SliceIterator) Err() error { return nil }

// Int16SliceIterator is an iterator based on a slice of int16
// and doing iteration in back direction.
type InvertingInt16SliceIterator struct {
	slice []int16
	cur   int
}

// NewInvertingInt16SliceIterator returns a new instance of InvertingInt16SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingInt16Slice(Int16Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingInt16SliceIterator(slice []int16) *InvertingInt16SliceIterator {
	it := &InvertingInt16SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingInt16SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingInt16SliceIterator) Next() int16 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingInt16SliceIterator) Err() error { return nil }

// Int16Unroll unrolls items to slice of int16.
func Int16Unroll(items Int16Iterator) Int16Slice {
	var slice Int16Slice
	panicIfInt16IteratorError(Int16Discard(Int16Handling(items, Int16Handle(func(item int16) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Int16Slice is a slice of int16.
type Int16Slice []int16

// MakeIter returns a new instance of Int16Iterator to iterate over it.
// It returns EmptyInt16Iterator if the error is not nil.
func (s Int16Slice) MakeIter() Int16Iterator {
	return NewInt16SliceIterator(s)
}

// Int16Slice is a slice of int16 which can make inverting iterator.
type InvertingInt16Slice []int16

// MakeIter returns a new instance of Int16Iterator to iterate over it.
// It returns EmptyInt16Iterator if the error is not nil.
func (s InvertingInt16Slice) MakeIter() Int16Iterator {
	return NewInvertingInt16SliceIterator(s)
}

// Int16Invert unrolls items and make inverting iterator based on them.
func Int16Invert(items Int16Iterator) Int16Iterator {
	return InvertingInt16Slice(Int16Unroll(items)).MakeIter()
}

// EndOfInt16Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfInt16Iterator = errors.New("end of int16 iterator")

func isEndOfInt16Iterator(err error) bool {
	return errors.Is(err, EndOfInt16Iterator)
}

func wrapIfNotEndOfInt16Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfInt16Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfInt16IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedInt16Item struct {
	base    Int16Iterator
	hasNext bool
	next    int16
	err     error
}

func (it *preparedInt16Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedInt16Item) Next() int16 {
	if !it.hasNext {
		panicIfInt16IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedInt16Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfInt16Iterator) {
		return it.err
	}
	return it.base.Err()
}
