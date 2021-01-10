package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// Iterator is an iterator over items type of interface{}.
type Iterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() interface{}
	// Err contains first met error while Next.
	Err() error
}

type emptyIterator struct{}

func (emptyIterator) HasNext() bool            { return false }
func (emptyIterator) Next() (next interface{}) { return nil }
func (emptyIterator) Err() error               { return nil }

// EmptyIterator is a zero value for Iterator.
// It is not contains any item to iterate over it.
var EmptyIterator Iterator = emptyIterator{}

// IterMaker is a maker of Iterator.
type IterMaker interface {
	// MakeIter should return a new instance of Iterator to iterate over it.
	MakeIter() Iterator
}

// MakeIter is a shortcut implementation
// of Iterator based on a function.
type MakeIter func() Iterator

// MakeIter returns a new instance of Iterator to iterate over it.
func (m MakeIter) MakeIter() Iterator { return m() }

// MakeNoIter is a zero value for IterMaker.
// It always returns EmptyIterator and an empty error.
var MakeNoIter IterMaker = MakeIter(
	func() Iterator { return EmptyIterator })

// Discard just range over all items and do nothing with each of them.
func Discard(items Iterator) error {
	if items == nil {
		items = EmptyIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

// Checker is an object checking an item type of interface{}
// for some condition.
type Checker interface {
	// Check should check an item type of interface{} for some condition.
	// It is suggested to return EndOfIterator to stop iteration.
	Check(interface{}) (bool, error)
}

// Check is a shortcut implementation
// of Checker based on a function.
type Check func(interface{}) (bool, error)

// Check checks an item type of interface{} for some condition.
// It returns EndOfIterator to stop iteration.
func (ch Check) Check(item interface{}) (bool, error) { return ch(item) }

var (
	// AlwaysCheckTrue always returns true and empty error.
	AlwaysCheckTrue Checker = Check(
		func(item interface{}) (bool, error) { return true, nil })
	// AlwaysCheckFalse always returns false and empty error.
	AlwaysCheckFalse Checker = Check(
		func(item interface{}) (bool, error) { return false, nil })
)

// Not do an inversion for checker result.
// It is returns AlwaysCheckTrue if checker is nil.
func Not(checker Checker) Checker {
	if checker == nil {
		return AlwaysCheckTrue
	}
	return Check(func(item interface{}) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type and struct {
	lhs, rhs Checker
}

func (a and) Check(item interface{}) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// All combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func All(checkers ...Checker) Checker {
	var all = AlwaysCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = and{checkers[i], all}
	}
	return all
}

type or struct {
	lhs, rhs Checker
}

func (o or) Check(item interface{}) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// Any combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func Any(checkers ...Checker) Checker {
	var any = AlwaysCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = or{checkers[i], any}
	}
	return any
}

// FilteringIterator does iteration with
// filtering by previously set checker.
type FilteringIterator struct {
	preparedItem
	filter Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfIterator(err) {
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

// Filtering sets filter while iterating over items.
// If filters is empty, so all items will return.
func Filtering(items Iterator, filters ...Checker) Iterator {
	if items == nil {
		return EmptyIterator
	}
	return &FilteringIterator{preparedItem{base: items}, All(filters...)}
}

// DoingUntilIterator does iteration
// until previously set checker is passed.
type DoingUntilIterator struct {
	preparedItem
	until Checker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// DoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func DoingUntil(items Iterator, untilList ...Checker) Iterator {
	if items == nil {
		return EmptyIterator
	}
	var until Checker
	if len(untilList) > 0 {
		until = All(untilList...)
	} else {
		until = AlwaysCheckFalse
	}
	return &DoingUntilIterator{preparedItem{base: items}, until}
}

// SkipUntil sets until conditions to skip few items.
func SkipUntil(items Iterator, untilList ...Checker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Discard(DoingUntil(items, untilList...))
}

// EnumChecker is an object checking an item type of interface{}
// and its ordering number in for some condition.
type EnumChecker interface {
	// Check checks an item type of interface{} and its ordering number for some condition.
	// It is suggested to return EndOfIterator to stop iteration.
	Check(int, interface{}) (bool, error)
}

// EnumCheck is a shortcut implementation
// of EnumChecker based on a function.
type EnumCheck func(int, interface{}) (bool, error)

// Check checks an item type of interface{} and its ordering number for some condition.
// It returns EndOfIterator to stop iteration.
func (ch EnumCheck) Check(n int, item interface{}) (bool, error) { return ch(n, item) }

type enumFromChecker struct {
	Checker
}

func (ch enumFromChecker) Check(_ int, item interface{}) (bool, error) {
	return ch.Checker.Check(item)
}

// EnumFromChecker adapts checker type of Checker
// to the interface EnumChecker.
// If checker is nil it is return based on AlwaysCheckFalse enum checker.
func EnumFromChecker(checker Checker) EnumChecker {
	if checker == nil {
		checker = AlwaysCheckFalse
	}
	return &enumFromChecker{checker}
}

var (
	// AlwaysEnumCheckTrue always returns true and empty error.
	AlwaysEnumCheckTrue = EnumFromChecker(
		AlwaysCheckTrue)
	// AlwaysEnumCheckFalse always returns false and empty error.
	AlwaysEnumCheckFalse = EnumFromChecker(
		AlwaysCheckFalse)
)

// EnumNot do an inversion for checker result.
// It is returns AlwaysEnumCheckTrue if checker is nil.
func EnumNot(checker EnumChecker) EnumChecker {
	if checker == nil {
		return AlwaysEnumCheckTrue
	}
	return EnumCheck(func(n int, item interface{}) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAnd struct {
	lhs, rhs EnumChecker
}

func (a enumAnd) Check(n int, item interface{}) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAll combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAll(checkers ...EnumChecker) EnumChecker {
	var all = AlwaysEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAnd{checkers[i], all}
	}
	return all
}

type enumOr struct {
	lhs, rhs EnumChecker
}

func (o enumOr) Check(n int, item interface{}) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAny combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAny(checkers ...EnumChecker) EnumChecker {
	var any = AlwaysEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOr{checkers[i], any}
	}
	return any
}

// EnumFilteringIterator does iteration with
// filtering by previously set checker.
type EnumFilteringIterator struct {
	preparedItem
	filter EnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfIterator(err) {
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

// EnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func EnumFiltering(items Iterator, filters ...EnumChecker) Iterator {
	if items == nil {
		return EmptyIterator
	}
	return &EnumFilteringIterator{preparedItem{base: items}, EnumAll(filters...), 0}
}

// EnumDoingUntilIterator does iteration
// until previously set checker is passed.
type EnumDoingUntilIterator struct {
	preparedItem
	until EnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// EnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func EnumDoingUntil(items Iterator, untilList ...EnumChecker) Iterator {
	if items == nil {
		return EmptyIterator
	}
	var until EnumChecker
	if len(untilList) > 0 {
		until = EnumAll(untilList...)
	} else {
		until = AlwaysEnumCheckFalse
	}
	return &EnumDoingUntilIterator{preparedItem{base: items}, until, 0}
}

// EnumSkipUntil sets until conditions to skip few items.
func EnumSkipUntil(items Iterator, untilList ...EnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return Discard(EnumDoingUntil(items, untilList...))
}

// GettingBatch returns the next batch from items.
func GettingBatch(items Iterator, batchSize int) Iterator {
	if items == nil {
		return EmptyIterator
	}
	if batchSize == 0 {
		return items
	}

	return EnumDoingUntil(items, EnumCheck(func(n int, item interface{}) (bool, error) {
		return n == batchSize-1, nil
	}))
}

// Converter is an object converting an item type of interface{}.
type Converter interface {
	// Convert should convert an item type of interface{} into another item of interface{}.
	// It is suggested to return EndOfIterator to stop iteration.
	Convert(interface{}) (interface{}, error)
}

// Convert is a shortcut implementation
// of Converter based on a function.
type Convert func(interface{}) (interface{}, error)

// Convert converts an item type of interface{} into another item of interface{}.
// It is suggested to return EndOfIterator to stop iteration.
func (c Convert) Convert(item interface{}) (interface{}, error) { return c(item) }

// NoConvert does nothing with item, just returns it as is.
var NoConvert Converter = Convert(
	func(item interface{}) (interface{}, error) { return item, nil })

type doubleConverter struct {
	lhs, rhs Converter
}

func (c doubleConverter) Convert(item interface{}) (interface{}, error) {
	item, err := c.lhs.Convert(item)
	if err != nil {
		return nil, errors.Wrap(err, "convert lhs")
	}
	item, err = c.rhs.Convert(item)
	if err != nil {
		return nil, errors.Wrap(err, "convert rhs")
	}
	return item, nil
}

// ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func ConverterSeries(converters ...Converter) Converter {
	var series = NoConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingIterator does iteration with
// converting by previously set converter.
type ConvertingIterator struct {
	preparedItem
	converter Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfIterator(err) {
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

// Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Converting(items Iterator, converters ...Converter) Iterator {
	if items == nil {
		return EmptyIterator
	}
	return &ConvertingIterator{
		preparedItem{base: items}, ConverterSeries(converters...)}
}

// EnumConverter is an object converting an item type of interface{} and its ordering number.
type EnumConverter interface {
	// Convert should convert an item type of interface{} into another item of interface{}.
	// It is suggested to return EndOfIterator to stop iteration.
	Convert(n int, val interface{}) (interface{}, error)
}

// EnumConvert is a shortcut implementation
// of EnumConverter based on a function.
type EnumConvert func(int, interface{}) (interface{}, error)

// Convert converts an item type of interface{} into another item of interface{}.
// It is suggested to return EndOfIterator to stop iteration.
func (c EnumConvert) Convert(n int, item interface{}) (interface{}, error) { return c(n, item) }

// NoEnumConvert does nothing with item, just returns it as is.
var NoEnumConvert EnumConverter = EnumConvert(
	func(_ int, item interface{}) (interface{}, error) { return item, nil })

type enumFromConverter struct {
	Converter
}

func (ch enumFromConverter) Convert(_ int, item interface{}) (interface{}, error) {
	return ch.Converter.Convert(item)
}

// EnumFromConverter adapts checker type of Converter
// to the interface EnumConverter.
// If converter is nil it is return based on NoConvert enum checker.
func EnumFromConverter(converter Converter) EnumConverter {
	if converter == nil {
		converter = NoConvert
	}
	return &enumFromConverter{converter}
}

type doubleEnumConverter struct {
	lhs, rhs EnumConverter
}

func (c doubleEnumConverter) Convert(n int, item interface{}) (interface{}, error) {
	item, err := c.lhs.Convert(n, item)
	if err != nil {
		return nil, errors.Wrap(err, "convert lhs")
	}
	item, err = c.rhs.Convert(n, item)
	if err != nil {
		return nil, errors.Wrap(err, "convert rhs")
	}
	return item, nil
}

// EnumConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumConverterSeries(converters ...EnumConverter) EnumConverter {
	var series = NoEnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleEnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingIterator does iteration with
// converting by previously set converter.
type EnumConvertingIterator struct {
	preparedItem
	converter EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfIterator(err) {
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

// EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func EnumConverting(items Iterator, converters ...EnumConverter) Iterator {
	if items == nil {
		return EmptyIterator
	}
	return &EnumConvertingIterator{
		preparedItem{base: items}, EnumConverterSeries(converters...), 0}
}

// Handler is an object handling an item type of interface{}.
type Handler interface {
	// Handle should do something with item of interface{}.
	// It is suggested to return EndOfIterator to stop iteration.
	Handle(interface{}) error
}

// Handle is a shortcut implementation
// of Handler based on a function.
type Handle func(interface{}) error

// Handle does something with item of interface{}.
// It is suggested to return EndOfIterator to stop iteration.
func (h Handle) Handle(item interface{}) error { return h(item) }

// DoNothing does nothing.
var DoNothing Handler = Handle(func(_ interface{}) error { return nil })

type doubleHandler struct {
	lhs, rhs Handler
}

func (h doubleHandler) Handle(item interface{}) error {
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

// HandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func HandlerSeries(handlers ...Handler) Handler {
	var series = DoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingIterator does iteration with
// handling by previously set handler.
type HandlingIterator struct {
	preparedItem
	handler Handler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfIterator(err) {
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

// Handling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func Handling(items Iterator, handlers ...Handler) Iterator {
	if items == nil {
		return EmptyIterator
	}
	return &HandlingIterator{
		preparedItem{base: items}, HandlerSeries(handlers...)}
}

// Range iterates over items and use handlers to each one.
func Range(items Iterator, handlers ...Handler) error {
	return Discard(Handling(items, handlers...))
}

// RangeIterator is an iterator over items.
type RangeIterator interface {
	// Range should iterate over items.
	Range(...Handler) error
}

type sRangeIterator struct {
	iter Iterator
}

// ToRangeIterator constructs an instance implementing RangeIterator
// based on Iterator.
func ToRangeIterator(iter Iterator) RangeIterator {
	if iter == nil {
		iter = EmptyIterator
	}
	return sRangeIterator{iter: iter}
}

// MakeRangeIterator constructs an instance implementing RangeIterator
// based on IterMaker.
func MakeRangeIterator(maker IterMaker) RangeIterator {
	if maker == nil {
		maker = MakeNoIter
	}
	return ToRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sRangeIterator) Range(handlers ...Handler) error {
	return Range(r.iter, handlers...)
}

// EnumHandler is an object handling an item type of interface{} and its ordered number.
type EnumHandler interface {
	// Handle should do something with item of interface{} and its ordered number.
	// It is suggested to return EndOfIterator to stop iteration.
	Handle(int, interface{}) error
}

// EnumHandle is a shortcut implementation
// of EnumHandler based on a function.
type EnumHandle func(int, interface{}) error

// Handle does something with item of interface{} and its ordered number.
// It is suggested to return EndOfIterator to stop iteration.
func (h EnumHandle) Handle(n int, item interface{}) error { return h(n, item) }

// DoEnumNothing does nothing.
var DoEnumNothing = EnumHandle(func(_ int, _ interface{}) error { return nil })

type doubleEnumHandler struct {
	lhs, rhs EnumHandler
}

func (h doubleEnumHandler) Handle(n int, item interface{}) error {
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

// EnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func EnumHandlerSeries(handlers ...EnumHandler) EnumHandler {
	var series EnumHandler = DoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingIterator does iteration with
// handling by previously set handler.
type EnumHandlingIterator struct {
	preparedItem
	handler EnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfIterator(err) {
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

// EnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func EnumHandling(items Iterator, handlers ...EnumHandler) Iterator {
	if items == nil {
		return EmptyIterator
	}
	return &EnumHandlingIterator{
		preparedItem{base: items}, EnumHandlerSeries(handlers...), 0}
}

// Enum iterates over items and their ordering numbers and use handlers to each one.
func Enum(items Iterator, handlers ...EnumHandler) error {
	return Discard(EnumHandling(items, handlers...))
}

// EnumIterator is an iterator over items and their ordering numbers.
type EnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...EnumHandler) error
}

type sEnumIterator struct {
	iter Iterator
}

// ToEnumIterator constructs an instance implementing EnumIterator
// based on Iterator.
func ToEnumIterator(iter Iterator) EnumIterator {
	if iter == nil {
		iter = EmptyIterator
	}
	return sEnumIterator{iter: iter}
}

// MakeEnumIterator constructs an instance implementing EnumIterator
// based on IterMaker.
func MakeEnumIterator(maker IterMaker) EnumIterator {
	if maker == nil {
		maker = MakeNoIter
	}
	return ToEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sEnumIterator) Enum(handlers ...EnumHandler) error {
	return Enum(r.iter, handlers...)
}

// Range iterates over items.
func (r sEnumIterator) Range(handlers ...Handler) error {
	return Range(r.iter, handlers...)
}

type doubleIterator struct {
	lhs, rhs Iterator
	inRHS    bool
}

func (it *doubleIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleIterator) Next() interface{} {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperIterator combines all iterators to one.
func SuperIterator(itemList ...Iterator) Iterator {
	var super = EmptyIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// EnumComparer is a strategy to compare two types.
type Comparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs interface{}) bool
}

// Compare is a shortcut implementation
// of EnumComparer based on a function.
type Compare func(lhs, rhs interface{}) bool

// IsLess is true if lhs is less than rhs.
func (c Compare) IsLess(lhs, rhs interface{}) bool { return c(lhs, rhs) }

// EnumAlwaysLess is an implementation of EnumComparer returning always true.
var AlwaysLess Comparer = Compare(func(_, _ interface{}) bool { return true })

type priorityIterator struct {
	lhs, rhs preparedItem
	comparer Comparer
}

func (it *priorityIterator) HasNext() bool {
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

func (it *priorityIterator) Next() interface{} {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfIteratorError(
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

func (it priorityIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorIterator(comparer Comparer, itemList ...Iterator) Iterator {
	if comparer == nil {
		comparer = AlwaysLess
	}

	var prior = EmptyIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityIterator{
			lhs:      preparedItem{base: itemList[i]},
			rhs:      preparedItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// EnumComparer is a strategy to compare two types and their order numbers.
type EnumComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(nLHS int, lhs interface{}, nRHS int, rhs interface{}) bool
}

// EnumCompare is a shortcut implementation
// of EnumComparer based on a function.
type EnumCompare func(nLHS int, lhs interface{}, nRHS int, rhs interface{}) bool

// IsLess is true if lhs is less than rhs.
func (c EnumCompare) IsLess(nLHS int, lhs interface{}, nRHS int, rhs interface{}) bool {
	return c(nLHS, lhs, nRHS, rhs)
}

// EnumAlwaysLess is an implementation of EnumComparer returning always true.
var EnumAlwaysLess EnumComparer = EnumCompare(
	func(_ int, _ interface{}, _ int, _ interface{}) bool { return true })

type priorityEnumIterator struct {
	lhs, rhs           preparedItem
	countLHS, countRHS int
	comparer           EnumComparer
}

func (it *priorityEnumIterator) HasNext() bool {
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

func (it *priorityEnumIterator) Next() interface{} {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfIteratorError(
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

func (it priorityEnumIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorEnumIterator compare one by one items and their ordering numbers fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorEnumIterator(comparer EnumComparer, itemList ...Iterator) Iterator {
	if comparer == nil {
		comparer = EnumAlwaysLess
	}

	var prior = EmptyIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityEnumIterator{
			lhs:      preparedItem{base: itemList[i]},
			rhs:      preparedItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// SliceIterator is an iterator based on a slice of interface{}.
type SliceIterator struct {
	slice []interface{}
	cur   int
}

// NewSliceIterator returns a new instance of SliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use Unroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewSliceIterator(slice []interface{}) *SliceIterator {
	it := &SliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it SliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *SliceIterator) Next() interface{} {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (SliceIterator) Err() error { return nil }

// SliceIterator is an iterator based on a slice of interface{}
// and doing iteration in back direction.
type InvertingSliceIterator struct {
	slice []interface{}
	cur   int
}

// NewInvertingSliceIterator returns a new instance of InvertingSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingSlice(Unroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingSliceIterator(slice []interface{}) *InvertingSliceIterator {
	it := &InvertingSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingSliceIterator) Next() interface{} {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingSliceIterator) Err() error { return nil }

// Unroll unrolls items to slice of interface{}.
func Unroll(items Iterator) Slice {
	var slice Slice
	panicIfIteratorError(Discard(Handling(items, Handle(func(item interface{}) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// Slice is a slice of interface{}.
type Slice []interface{}

// MakeIter returns a new instance of Iterator to iterate over it.
// It returns EmptyIterator if the error is not nil.
func (s Slice) MakeIter() Iterator {
	return NewSliceIterator(s)
}

// Slice is a slice of interface{} which can make inverting iterator.
type InvertingSlice []interface{}

// MakeIter returns a new instance of Iterator to iterate over it.
// It returns EmptyIterator if the error is not nil.
func (s InvertingSlice) MakeIter() Iterator {
	return NewInvertingSliceIterator(s)
}

// Invert unrolls items and make inverting iterator based on them.
func Invert(items Iterator) Iterator {
	return InvertingSlice(Unroll(items)).MakeIter()
}

// EndOfIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfIterator = errors.New("end of interface{} iterator")

func isEndOfIterator(err error) bool {
	return errors.Is(err, EndOfIterator)
}

func wrapIfNotEndOfIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedItem struct {
	base    Iterator
	hasNext bool
	next    interface{}
	err     error
}

func (it *preparedItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedItem) Next() interface{} {
	if !it.hasNext {
		panicIfIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = nil
	return next
}

func (it preparedItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfIterator) {
		return it.err
	}
	return it.base.Err()
}
