package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// Int32Iterator is an iterator over items type of int32.
type Int32Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() int32
	// Err contains first met error while Next.
	Err() error
}

type emptyInt32Iterator struct{}

func (emptyInt32Iterator) HasNext() bool      { return false }
func (emptyInt32Iterator) Next() (next int32) { return 0 }
func (emptyInt32Iterator) Err() error         { return nil }

// EmptyInt32Iterator is a zero value for Int32Iterator.
// It is not contains any item to iterate over it.
var EmptyInt32Iterator Int32Iterator = emptyInt32Iterator{}

// Int32IterMaker is a maker of Int32Iterator.
type Int32IterMaker interface {
	// MakeIter should return a new instance of Int32Iterator to iterate over it.
	MakeIter() Int32Iterator
}

// MakeInt32Iter is a shortcut implementation
// of Int32Iterator based on a function.
type MakeInt32Iter func() Int32Iterator

// MakeIter returns a new instance of Int32Iterator to iterate over it.
func (m MakeInt32Iter) MakeIter() Int32Iterator { return m() }

// MakeNoInt32Iter is a zero value for Int32IterMaker.
// It always returns EmptyInt32Iterator and an empty error.
var MakeNoInt32Iter Int32IterMaker = MakeInt32Iter(
	func() Int32Iterator { return EmptyInt32Iterator })

// Int32Discard just range over all items and do nothing with each of them.
func Int32Discard(items Int32Iterator) error {
	if items == nil {
		items = EmptyInt32Iterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

// Int32Checker is an object checking an item type of int32
// for some condition.
type Int32Checker interface {
	// Check should check an item type of int32 for some condition.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Check(int32) (bool, error)
}

// Int32Check is a shortcut implementation
// of Int32Checker based on a function.
type Int32Check func(int32) (bool, error)

// Check checks an item type of int32 for some condition.
// It returns EndOfInt32Iterator to stop iteration.
func (ch Int32Check) Check(item int32) (bool, error) { return ch(item) }

var (
	// AlwaysInt32CheckTrue always returns true and empty error.
	AlwaysInt32CheckTrue Int32Checker = Int32Check(
		func(item int32) (bool, error) { return true, nil })
	// AlwaysInt32CheckFalse always returns false and empty error.
	AlwaysInt32CheckFalse Int32Checker = Int32Check(
		func(item int32) (bool, error) { return false, nil })
)

// NotInt32 do an inversion for checker result.
// It is returns AlwaysInt32CheckTrue if checker is nil.
func NotInt32(checker Int32Checker) Int32Checker {
	if checker == nil {
		return AlwaysInt32CheckTrue
	}
	return Int32Check(func(item int32) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andInt32 struct {
	lhs, rhs Int32Checker
}

func (a andInt32) Check(item int32) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllInt32 combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllInt32(checkers ...Int32Checker) Int32Checker {
	var all = AlwaysInt32CheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andInt32{checkers[i], all}
	}
	return all
}

type orInt32 struct {
	lhs, rhs Int32Checker
}

func (o orInt32) Check(item int32) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyInt32 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyInt32(checkers ...Int32Checker) Int32Checker {
	var any = AlwaysInt32CheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orInt32{checkers[i], any}
	}
	return any
}

// FilteringInt32Iterator does iteration with
// filtering by previously set checker.
type FilteringInt32Iterator struct {
	preparedInt32Item
	filter Int32Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
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

// Int32Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Int32Filtering(items Int32Iterator, filters ...Int32Checker) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &FilteringInt32Iterator{preparedInt32Item{base: items}, AllInt32(filters...)}
}

// DoingUntilInt32Iterator does iteration
// until previously set checker is passed.
type DoingUntilInt32Iterator struct {
	preparedInt32Item
	until Int32Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfInt32Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int32DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int32DoingUntil(items Int32Iterator, untilList ...Int32Checker) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	var until Int32Checker
	if len(untilList) > 0 {
		until = AllInt32(untilList...)
	} else {
		until = AlwaysInt32CheckFalse
	}
	return &DoingUntilInt32Iterator{preparedInt32Item{base: items}, until}
}

// Int32SkipUntil sets until conditions to skip few items.
func Int32SkipUntil(items Int32Iterator, untilList ...Int32Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int32Discard(Int32DoingUntil(items, untilList...))
}

// Int32EnumChecker is an object checking an item type of int32
// and its ordering number in for some condition.
type Int32EnumChecker interface {
	// Check checks an item type of int32 and its ordering number for some condition.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Check(int, int32) (bool, error)
}

// Int32EnumCheck is a shortcut implementation
// of Int32EnumChecker based on a function.
type Int32EnumCheck func(int, int32) (bool, error)

// Check checks an item type of int32 and its ordering number for some condition.
// It returns EndOfInt32Iterator to stop iteration.
func (ch Int32EnumCheck) Check(n int, item int32) (bool, error) { return ch(n, item) }

type enumFromInt32Checker struct {
	Int32Checker
}

func (ch enumFromInt32Checker) Check(_ int, item int32) (bool, error) {
	return ch.Int32Checker.Check(item)
}

// EnumFromInt32Checker adapts checker type of Int32Checker
// to the interface Int32EnumChecker.
// If checker is nil it is return based on AlwaysInt32CheckFalse enum checker.
func EnumFromInt32Checker(checker Int32Checker) Int32EnumChecker {
	if checker == nil {
		checker = AlwaysInt32CheckFalse
	}
	return &enumFromInt32Checker{checker}
}

var (
	// AlwaysInt32EnumCheckTrue always returns true and empty error.
	AlwaysInt32EnumCheckTrue = EnumFromInt32Checker(
		AlwaysInt32CheckTrue)
	// AlwaysInt32EnumCheckFalse always returns false and empty error.
	AlwaysInt32EnumCheckFalse = EnumFromInt32Checker(
		AlwaysInt32CheckFalse)
)

// EnumNotInt32 do an inversion for checker result.
// It is returns AlwaysInt32EnumCheckTrue if checker is nil.
func EnumNotInt32(checker Int32EnumChecker) Int32EnumChecker {
	if checker == nil {
		return AlwaysInt32EnumCheckTrue
	}
	return Int32EnumCheck(func(n int, item int32) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndInt32 struct {
	lhs, rhs Int32EnumChecker
}

func (a enumAndInt32) Check(n int, item int32) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllInt32 combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllInt32(checkers ...Int32EnumChecker) Int32EnumChecker {
	var all = AlwaysInt32EnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndInt32{checkers[i], all}
	}
	return all
}

type enumOrInt32 struct {
	lhs, rhs Int32EnumChecker
}

func (o enumOrInt32) Check(n int, item int32) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfInt32Iterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyInt32 combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyInt32(checkers ...Int32EnumChecker) Int32EnumChecker {
	var any = AlwaysInt32EnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrInt32{checkers[i], any}
	}
	return any
}

// EnumFilteringInt32Iterator does iteration with
// filtering by previously set checker.
type EnumFilteringInt32Iterator struct {
	preparedInt32Item
	filter Int32EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
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

// Int32EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func Int32EnumFiltering(items Int32Iterator, filters ...Int32EnumChecker) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &EnumFilteringInt32Iterator{preparedInt32Item{base: items}, EnumAllInt32(filters...), 0}
}

// EnumDoingUntilInt32Iterator does iteration
// until previously set checker is passed.
type EnumDoingUntilInt32Iterator struct {
	preparedInt32Item
	until Int32EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfInt32Iterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// Int32EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func Int32EnumDoingUntil(items Int32Iterator, untilList ...Int32EnumChecker) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	var until Int32EnumChecker
	if len(untilList) > 0 {
		until = EnumAllInt32(untilList...)
	} else {
		until = AlwaysInt32EnumCheckFalse
	}
	return &EnumDoingUntilInt32Iterator{preparedInt32Item{base: items}, until, 0}
}

// Int32EnumSkipUntil sets until conditions to skip few items.
func Int32EnumSkipUntil(items Int32Iterator, untilList ...Int32EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int32Discard(Int32EnumDoingUntil(items, untilList...))
}

// Int32GettingBatch returns the next batch from items.
func Int32GettingBatch(items Int32Iterator, batchSize int) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	if batchSize == 0 {
		return items
	}

	return Int32EnumDoingUntil(items, Int32EnumCheck(func(n int, item int32) (bool, error) {
		return n == batchSize-1, nil
	}))
}

// Int32Converter is an object converting an item type of int32.
type Int32Converter interface {
	// Convert should convert an item type of int32 into another item of int32.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Convert(int32) (int32, error)
}

// Int32Convert is a shortcut implementation
// of Int32Converter based on a function.
type Int32Convert func(int32) (int32, error)

// Convert converts an item type of int32 into another item of int32.
// It is suggested to return EndOfInt32Iterator to stop iteration.
func (c Int32Convert) Convert(item int32) (int32, error) { return c(item) }

// NoInt32Convert does nothing with item, just returns it as is.
var NoInt32Convert Int32Converter = Int32Convert(
	func(item int32) (int32, error) { return item, nil })

type doubleInt32Converter struct {
	lhs, rhs Int32Converter
}

func (c doubleInt32Converter) Convert(item int32) (int32, error) {
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

// Int32ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Int32ConverterSeries(converters ...Int32Converter) Int32Converter {
	var series = NoInt32Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt32Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingInt32Iterator does iteration with
// converting by previously set converter.
type ConvertingInt32Iterator struct {
	preparedInt32Item
	converter Int32Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
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

// Int32Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Int32Converting(items Int32Iterator, converters ...Int32Converter) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &ConvertingInt32Iterator{
		preparedInt32Item{base: items}, Int32ConverterSeries(converters...)}
}

// Int32EnumConverter is an object converting an item type of int32 and its ordering number.
type Int32EnumConverter interface {
	// Convert should convert an item type of int32 into another item of int32.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Convert(n int, val int32) (int32, error)
}

// Int32EnumConvert is a shortcut implementation
// of Int32EnumConverter based on a function.
type Int32EnumConvert func(int, int32) (int32, error)

// Convert converts an item type of int32 into another item of int32.
// It is suggested to return EndOfInt32Iterator to stop iteration.
func (c Int32EnumConvert) Convert(n int, item int32) (int32, error) { return c(n, item) }

// NoInt32EnumConvert does nothing with item, just returns it as is.
var NoInt32EnumConvert Int32EnumConverter = Int32EnumConvert(
	func(_ int, item int32) (int32, error) { return item, nil })

type enumFromInt32Converter struct {
	Int32Converter
}

func (ch enumFromInt32Converter) Convert(_ int, item int32) (int32, error) {
	return ch.Int32Converter.Convert(item)
}

// EnumFromInt32Converter adapts checker type of Int32Converter
// to the interface Int32EnumConverter.
// If converter is nil it is return based on NoInt32Convert enum checker.
func EnumFromInt32Converter(converter Int32Converter) Int32EnumConverter {
	if converter == nil {
		converter = NoInt32Convert
	}
	return &enumFromInt32Converter{converter}
}

type doubleInt32EnumConverter struct {
	lhs, rhs Int32EnumConverter
}

func (c doubleInt32EnumConverter) Convert(n int, item int32) (int32, error) {
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

// EnumInt32ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumInt32ConverterSeries(converters ...Int32EnumConverter) Int32EnumConverter {
	var series = NoInt32EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt32EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingInt32Iterator does iteration with
// converting by previously set converter.
type EnumConvertingInt32Iterator struct {
	preparedInt32Item
	converter Int32EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
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

// Int32EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Int32EnumConverting(items Int32Iterator, converters ...Int32EnumConverter) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &EnumConvertingInt32Iterator{
		preparedInt32Item{base: items}, EnumInt32ConverterSeries(converters...), 0}
}

// Int32Handler is an object handling an item type of int32.
type Int32Handler interface {
	// Handle should do something with item of int32.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Handle(int32) error
}

// Int32Handle is a shortcut implementation
// of Int32Handler based on a function.
type Int32Handle func(int32) error

// Handle does something with item of int32.
// It is suggested to return EndOfInt32Iterator to stop iteration.
func (h Int32Handle) Handle(item int32) error { return h(item) }

// Int32DoNothing does nothing.
var Int32DoNothing Int32Handler = Int32Handle(func(_ int32) error { return nil })

type doubleInt32Handler struct {
	lhs, rhs Int32Handler
}

func (h doubleInt32Handler) Handle(item int32) error {
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

// Int32HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int32HandlerSeries(handlers ...Int32Handler) Int32Handler {
	var series = Int32DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt32Handler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingInt32Iterator does iteration with
// handling by previously set handler.
type HandlingInt32Iterator struct {
	preparedInt32Item
	handler Int32Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
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

// Int32Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Int32Handling(items Int32Iterator, handlers ...Int32Handler) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &HandlingInt32Iterator{
		preparedInt32Item{base: items}, Int32HandlerSeries(handlers...)}
}

// Int32Range iterates over items and use handlers to each one.
func Int32Range(items Int32Iterator, handlers ...Int32Handler) error {
	return Int32Discard(Int32Handling(items, handlers...))
}

// Int32RangeIterator is an iterator over items.
type Int32RangeIterator interface {
	// Range should iterate over items.
	Range(...Int32Handler) error
}

type sInt32RangeIterator struct {
	iter Int32Iterator
}

// ToInt32RangeIterator constructs an instance implementing Int32RangeIterator
// based on Int32Iterator.
func ToInt32RangeIterator(iter Int32Iterator) Int32RangeIterator {
	if iter == nil {
		iter = EmptyInt32Iterator
	}
	return sInt32RangeIterator{iter: iter}
}

// MakeInt32RangeIterator constructs an instance implementing Int32RangeIterator
// based on Int32IterMaker.
func MakeInt32RangeIterator(maker Int32IterMaker) Int32RangeIterator {
	if maker == nil {
		maker = MakeNoInt32Iter
	}
	return ToInt32RangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sInt32RangeIterator) Range(handlers ...Int32Handler) error {
	return Int32Range(r.iter, handlers...)
}

// Int32EnumHandler is an object handling an item type of int32 and its ordered number.
type Int32EnumHandler interface {
	// Handle should do something with item of int32 and its ordered number.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Handle(int, int32) error
}

// Int32EnumHandle is a shortcut implementation
// of Int32EnumHandler based on a function.
type Int32EnumHandle func(int, int32) error

// Handle does something with item of int32 and its ordered number.
// It is suggested to return EndOfInt32Iterator to stop iteration.
func (h Int32EnumHandle) Handle(n int, item int32) error { return h(n, item) }

// Int32DoEnumNothing does nothing.
var Int32DoEnumNothing = Int32EnumHandle(func(_ int, _ int32) error { return nil })

type doubleInt32EnumHandler struct {
	lhs, rhs Int32EnumHandler
}

func (h doubleInt32EnumHandler) Handle(n int, item int32) error {
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

// Int32EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func Int32EnumHandlerSeries(handlers ...Int32EnumHandler) Int32EnumHandler {
	var series Int32EnumHandler = Int32DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleInt32EnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingInt32Iterator does iteration with
// handling by previously set handler.
type EnumHandlingInt32Iterator struct {
	preparedInt32Item
	handler Int32EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
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

// Int32EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func Int32EnumHandling(items Int32Iterator, handlers ...Int32EnumHandler) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &EnumHandlingInt32Iterator{
		preparedInt32Item{base: items}, Int32EnumHandlerSeries(handlers...), 0}
}

// Int32Enum iterates over items and their ordering numbers and use handlers to each one.
func Int32Enum(items Int32Iterator, handlers ...Int32EnumHandler) error {
	return Int32Discard(Int32EnumHandling(items, handlers...))
}

// Int32EnumIterator is an iterator over items and their ordering numbers.
type Int32EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...Int32EnumHandler) error
}

type sInt32EnumIterator struct {
	iter Int32Iterator
}

// ToInt32EnumIterator constructs an instance implementing Int32EnumIterator
// based on Int32Iterator.
func ToInt32EnumIterator(iter Int32Iterator) Int32EnumIterator {
	if iter == nil {
		iter = EmptyInt32Iterator
	}
	return sInt32EnumIterator{iter: iter}
}

// MakeInt32EnumIterator constructs an instance implementing Int32EnumIterator
// based on Int32IterMaker.
func MakeInt32EnumIterator(maker Int32IterMaker) Int32EnumIterator {
	if maker == nil {
		maker = MakeNoInt32Iter
	}
	return ToInt32EnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sInt32EnumIterator) Enum(handlers ...Int32EnumHandler) error {
	return Int32Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sInt32EnumIterator) Range(handlers ...Int32Handler) error {
	return Int32Range(r.iter, handlers...)
}

type doubleInt32Iterator struct {
	lhs, rhs Int32Iterator
	inRHS    bool
}

func (it *doubleInt32Iterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleInt32Iterator) Next() int32 {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleInt32Iterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperInt32Iterator combines all iterators to one.
func SuperInt32Iterator(itemList ...Int32Iterator) Int32Iterator {
	var super = EmptyInt32Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleInt32Iterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// Int32EnumComparer is a strategy to compare two types.
type Int32Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs int32) bool
}

// Int32Compare is a shortcut implementation
// of Int32EnumComparer based on a function.
type Int32Compare func(lhs, rhs int32) bool

// IsLess is true if lhs is less than rhs.
func (c Int32Compare) IsLess(lhs, rhs int32) bool { return c(lhs, rhs) }

// EnumInt32AlwaysLess is an implementation of Int32EnumComparer returning always true.
var Int32AlwaysLess Int32Comparer = Int32Compare(func(_, _ int32) bool { return true })

type priorityInt32Iterator struct {
	lhs, rhs preparedInt32Item
	comparer Int32Comparer
}

func (it *priorityInt32Iterator) HasNext() bool {
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

func (it *priorityInt32Iterator) Next() int32 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfInt32IteratorError(
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

func (it priorityInt32Iterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorInt32Iterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorInt32Iterator(comparer Int32Comparer, itemList ...Int32Iterator) Int32Iterator {
	if comparer == nil {
		comparer = Int32AlwaysLess
	}

	var prior = EmptyInt32Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityInt32Iterator{
			lhs:      preparedInt32Item{base: itemList[i]},
			rhs:      preparedInt32Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Int32EnumComparer is a strategy to compare two types and their order numbers.
type Int32EnumComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(nLHS int, lhs int32, nRHS int, rhs int32) bool
}

// Int32EnumCompare is a shortcut implementation
// of Int32EnumComparer based on a function.
type Int32EnumCompare func(nLHS int, lhs int32, nRHS int, rhs int32) bool

// IsLess is true if lhs is less than rhs.
func (c Int32EnumCompare) IsLess(nLHS int, lhs int32, nRHS int, rhs int32) bool {
	return c(nLHS, lhs, nRHS, rhs)
}

// EnumInt32AlwaysLess is an implementation of Int32EnumComparer returning always true.
var EnumInt32AlwaysLess Int32EnumComparer = Int32EnumCompare(
	func(_ int, _ int32, _ int, _ int32) bool { return true })

type priorityInt32EnumIterator struct {
	lhs, rhs           preparedInt32Item
	countLHS, countRHS int
	comparer           Int32EnumComparer
}

func (it *priorityInt32EnumIterator) HasNext() bool {
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

func (it *priorityInt32EnumIterator) Next() int32 {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfInt32IteratorError(
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

func (it priorityInt32EnumIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorInt32EnumIterator compare one by one items and their ordering numbers fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorInt32EnumIterator(comparer Int32EnumComparer, itemList ...Int32Iterator) Int32Iterator {
	if comparer == nil {
		comparer = EnumInt32AlwaysLess
	}

	var prior = EmptyInt32Iterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityInt32EnumIterator{
			lhs:      preparedInt32Item{base: itemList[i]},
			rhs:      preparedInt32Item{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// Int32SliceIterator is an iterator based on a slice of int32.
type Int32SliceIterator struct {
	slice []int32
	cur   int
}

// NewInt32SliceIterator returns a new instance of Int32SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Int32Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewInt32SliceIterator(slice []int32) *Int32SliceIterator {
	it := &Int32SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it Int32SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *Int32SliceIterator) Next() int32 {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (Int32SliceIterator) Err() error { return nil }

// Int32SliceIterator is an iterator based on a slice of int32
// and doing iteration in back direction.
type InvertingInt32SliceIterator struct {
	slice []int32
	cur   int
}

// NewInvertingInt32SliceIterator returns a new instance of InvertingInt32SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingInt32Slice(Int32Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingInt32SliceIterator(slice []int32) *InvertingInt32SliceIterator {
	it := &InvertingInt32SliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingInt32SliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingInt32SliceIterator) Next() int32 {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingInt32SliceIterator) Err() error { return nil }

// Int32Unroll unrolls items to slice of int32.
func Int32Unroll(items Int32Iterator) Int32Slice {
	var slice Int32Slice
	panicIfInt32IteratorError(Int32Discard(Int32Handling(items, Int32Handle(func(item int32) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Int32Slice is a slice of int32.
type Int32Slice []int32

// MakeIter returns a new instance of Int32Iterator to iterate over it.
// It returns EmptyInt32Iterator if the error is not nil.
func (s Int32Slice) MakeIter() Int32Iterator {
	return NewInt32SliceIterator(s)
}

// Int32Slice is a slice of int32 which can make inverting iterator.
type InvertingInt32Slice []int32

// MakeIter returns a new instance of Int32Iterator to iterate over it.
// It returns EmptyInt32Iterator if the error is not nil.
func (s InvertingInt32Slice) MakeIter() Int32Iterator {
	return NewInvertingInt32SliceIterator(s)
}

// Int32Invert unrolls items and make inverting iterator based on them.
func Int32Invert(items Int32Iterator) Int32Iterator {
	return InvertingInt32Slice(Int32Unroll(items)).MakeIter()
}

// EndOfInt32Iterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfInt32Iterator = errors.New("end of int32 iterator")

func isEndOfInt32Iterator(err error) bool {
	return errors.Is(err, EndOfInt32Iterator)
}

func wrapIfNotEndOfInt32Iterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfInt32Iterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfInt32IteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedInt32Item struct {
	base    Int32Iterator
	hasNext bool
	next    int32
	err     error
}

func (it *preparedInt32Item) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedInt32Item) Next() int32 {
	if !it.hasNext {
		panicIfInt32IteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedInt32Item) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfInt32Iterator) {
		return it.err
	}
	return it.base.Err()
}
