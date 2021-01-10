package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// IntIterator is an iterator over items type of int.
type IntIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() int
	// Err contains first met error while Next.
	Err() error
}

type emptyIntIterator struct{}

func (emptyIntIterator) HasNext() bool    { return false }
func (emptyIntIterator) Next() (next int) { return 0 }
func (emptyIntIterator) Err() error       { return nil }

// EmptyIntIterator is a zero value for IntIterator.
// It is not contains any item to iterate over it.
var EmptyIntIterator IntIterator = emptyIntIterator{}

// IntIterMaker is a maker of IntIterator.
type IntIterMaker interface {
	// MakeIter should return a new instance of IntIterator to iterate over it.
	MakeIter() IntIterator
}

// MakeIntIter is a shortcut implementation
// of IntIterator based on a function.
type MakeIntIter func() IntIterator

// MakeIter returns a new instance of IntIterator to iterate over it.
func (m MakeIntIter) MakeIter() IntIterator { return m() }

// MakeNoIntIter is a zero value for IntIterMaker.
// It always returns EmptyIntIterator and an empty error.
var MakeNoIntIter IntIterMaker = MakeIntIter(
	func() IntIterator { return EmptyIntIterator })

// IntDiscard just range over all items and do nothing with each of them.
func IntDiscard(items IntIterator) error {
	if items == nil {
		items = EmptyIntIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

// IntChecker is an object checking an item type of int
// for some condition.
type IntChecker interface {
	// Check should check an item type of int for some condition.
	// It is suggested to return EndOfIntIterator to stop iteration.
	Check(int) (bool, error)
}

// IntCheck is a shortcut implementation
// of IntChecker based on a function.
type IntCheck func(int) (bool, error)

// Check checks an item type of int for some condition.
// It returns EndOfIntIterator to stop iteration.
func (ch IntCheck) Check(item int) (bool, error) { return ch(item) }

var (
	// AlwaysIntCheckTrue always returns true and empty error.
	AlwaysIntCheckTrue IntChecker = IntCheck(
		func(item int) (bool, error) { return true, nil })
	// AlwaysIntCheckFalse always returns false and empty error.
	AlwaysIntCheckFalse IntChecker = IntCheck(
		func(item int) (bool, error) { return false, nil })
)

// NotInt do an inversion for checker result.
// It is returns AlwaysIntCheckTrue if checker is nil.
func NotInt(checker IntChecker) IntChecker {
	if checker == nil {
		return AlwaysIntCheckTrue
	}
	return IntCheck(func(item int) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andInt struct {
	lhs, rhs IntChecker
}

func (a andInt) Check(item int) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllInt combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllInt(checkers ...IntChecker) IntChecker {
	var all = AlwaysIntCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andInt{checkers[i], all}
	}
	return all
}

type orInt struct {
	lhs, rhs IntChecker
}

func (o orInt) Check(item int) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyInt combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyInt(checkers ...IntChecker) IntChecker {
	var any = AlwaysIntCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orInt{checkers[i], any}
	}
	return any
}

// FilteringIntIterator does iteration with
// filtering by previously set checker.
type FilteringIntIterator struct {
	preparedIntItem
	filter IntChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedIntItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfIntIterator(err) {
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

// IntFiltering sets filter while iterating over items.
// If filters is empty, so all items will return.
func IntFiltering(items IntIterator, filters ...IntChecker) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	return &FilteringIntIterator{preparedIntItem{base: items}, AllInt(filters...)}
}

// DoingUntilIntIterator does iteration
// until previously set checker is passed.
type DoingUntilIntIterator struct {
	preparedIntItem
	until IntChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedIntItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfIntIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfIntIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// IntDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func IntDoingUntil(items IntIterator, untilList ...IntChecker) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	var until IntChecker
	if len(untilList) > 0 {
		until = AllInt(untilList...)
	} else {
		until = AlwaysIntCheckFalse
	}
	return &DoingUntilIntIterator{preparedIntItem{base: items}, until}
}

// IntSkipUntil sets until conditions to skip few items.
func IntSkipUntil(items IntIterator, untilList ...IntChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return IntDiscard(IntDoingUntil(items, untilList...))
}

// IntEnumChecker is an object checking an item type of int
// and its ordering number in for some condition.
type IntEnumChecker interface {
	// Check checks an item type of int and its ordering number for some condition.
	// It is suggested to return EndOfIntIterator to stop iteration.
	Check(int, int) (bool, error)
}

// IntEnumCheck is a shortcut implementation
// of IntEnumChecker based on a function.
type IntEnumCheck func(int, int) (bool, error)

// Check checks an item type of int and its ordering number for some condition.
// It returns EndOfIntIterator to stop iteration.
func (ch IntEnumCheck) Check(n int, item int) (bool, error) { return ch(n, item) }

type enumFromIntChecker struct {
	IntChecker
}

func (ch enumFromIntChecker) Check(_ int, item int) (bool, error) {
	return ch.IntChecker.Check(item)
}

// EnumFromIntChecker adapts checker type of IntChecker
// to the interface IntEnumChecker.
// If checker is nil it is return based on AlwaysIntCheckFalse enum checker.
func EnumFromIntChecker(checker IntChecker) IntEnumChecker {
	if checker == nil {
		checker = AlwaysIntCheckFalse
	}
	return &enumFromIntChecker{checker}
}

var (
	// AlwaysIntEnumCheckTrue always returns true and empty error.
	AlwaysIntEnumCheckTrue = EnumFromIntChecker(
		AlwaysIntCheckTrue)
	// AlwaysIntEnumCheckFalse always returns false and empty error.
	AlwaysIntEnumCheckFalse = EnumFromIntChecker(
		AlwaysIntCheckFalse)
)

// EnumNotInt do an inversion for checker result.
// It is returns AlwaysIntEnumCheckTrue if checker is nil.
func EnumNotInt(checker IntEnumChecker) IntEnumChecker {
	if checker == nil {
		return AlwaysIntEnumCheckTrue
	}
	return IntEnumCheck(func(n int, item int) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndInt struct {
	lhs, rhs IntEnumChecker
}

func (a enumAndInt) Check(n int, item int) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllInt combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllInt(checkers ...IntEnumChecker) IntEnumChecker {
	var all = AlwaysIntEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndInt{checkers[i], all}
	}
	return all
}

type enumOrInt struct {
	lhs, rhs IntEnumChecker
}

func (o enumOrInt) Check(n int, item int) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfIntIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyInt combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyInt(checkers ...IntEnumChecker) IntEnumChecker {
	var any = AlwaysIntEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrInt{checkers[i], any}
	}
	return any
}

// EnumFilteringIntIterator does iteration with
// filtering by previously set checker.
type EnumFilteringIntIterator struct {
	preparedIntItem
	filter IntEnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedIntItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfIntIterator(err) {
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

// IntEnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func IntEnumFiltering(items IntIterator, filters ...IntEnumChecker) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	return &EnumFilteringIntIterator{preparedIntItem{base: items}, EnumAllInt(filters...), 0}
}

// EnumDoingUntilIntIterator does iteration
// until previously set checker is passed.
type EnumDoingUntilIntIterator struct {
	preparedIntItem
	until IntEnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedIntItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfIntIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfIntIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// IntEnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func IntEnumDoingUntil(items IntIterator, untilList ...IntEnumChecker) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	var until IntEnumChecker
	if len(untilList) > 0 {
		until = EnumAllInt(untilList...)
	} else {
		until = AlwaysIntEnumCheckFalse
	}
	return &EnumDoingUntilIntIterator{preparedIntItem{base: items}, until, 0}
}

// IntEnumSkipUntil sets until conditions to skip few items.
func IntEnumSkipUntil(items IntIterator, untilList ...IntEnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return IntDiscard(IntEnumDoingUntil(items, untilList...))
}

// IntGettingBatch returns the next batch from items.
func IntGettingBatch(items IntIterator, batchSize int) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	if batchSize == 0 {
		return items
	}

	return IntEnumDoingUntil(items, IntEnumCheck(func(n int, item int) (bool, error) {
		return n == batchSize-1, nil
	}))
}

// IntConverter is an object converting an item type of int.
type IntConverter interface {
	// Convert should convert an item type of int into another item of int.
	// It is suggested to return EndOfIntIterator to stop iteration.
	Convert(int) (int, error)
}

// IntConvert is a shortcut implementation
// of IntConverter based on a function.
type IntConvert func(int) (int, error)

// Convert converts an item type of int into another item of int.
// It is suggested to return EndOfIntIterator to stop iteration.
func (c IntConvert) Convert(item int) (int, error) { return c(item) }

// NoIntConvert does nothing with item, just returns it as is.
var NoIntConvert IntConverter = IntConvert(
	func(item int) (int, error) { return item, nil })

type doubleIntConverter struct {
	lhs, rhs IntConverter
}

func (c doubleIntConverter) Convert(item int) (int, error) {
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

// IntConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func IntConverterSeries(converters ...IntConverter) IntConverter {
	var series = NoIntConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleIntConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingIntIterator does iteration with
// converting by previously set converter.
type ConvertingIntIterator struct {
	preparedIntItem
	converter IntConverter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedIntItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfIntIterator(err) {
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

// IntConverting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func IntConverting(items IntIterator, converters ...IntConverter) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	return &ConvertingIntIterator{
		preparedIntItem{base: items}, IntConverterSeries(converters...)}
}

// IntEnumConverter is an object converting an item type of int and its ordering number.
type IntEnumConverter interface {
	// Convert should convert an item type of int into another item of int.
	// It is suggested to return EndOfIntIterator to stop iteration.
	Convert(n int, val int) (int, error)
}

// IntEnumConvert is a shortcut implementation
// of IntEnumConverter based on a function.
type IntEnumConvert func(int, int) (int, error)

// Convert converts an item type of int into another item of int.
// It is suggested to return EndOfIntIterator to stop iteration.
func (c IntEnumConvert) Convert(n int, item int) (int, error) { return c(n, item) }

// NoIntEnumConvert does nothing with item, just returns it as is.
var NoIntEnumConvert IntEnumConverter = IntEnumConvert(
	func(_ int, item int) (int, error) { return item, nil })

type enumFromIntConverter struct {
	IntConverter
}

func (ch enumFromIntConverter) Convert(_ int, item int) (int, error) {
	return ch.IntConverter.Convert(item)
}

// EnumFromIntConverter adapts checker type of IntConverter
// to the interface IntEnumConverter.
// If converter is nil it is return based on NoIntConvert enum checker.
func EnumFromIntConverter(converter IntConverter) IntEnumConverter {
	if converter == nil {
		converter = NoIntConvert
	}
	return &enumFromIntConverter{converter}
}

type doubleIntEnumConverter struct {
	lhs, rhs IntEnumConverter
}

func (c doubleIntEnumConverter) Convert(n int, item int) (int, error) {
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

// EnumIntConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumIntConverterSeries(converters ...IntEnumConverter) IntEnumConverter {
	var series = NoIntEnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleIntEnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingIntIterator does iteration with
// converting by previously set converter.
type EnumConvertingIntIterator struct {
	preparedIntItem
	converter IntEnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedIntItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfIntIterator(err) {
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

// IntEnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func IntEnumConverting(items IntIterator, converters ...IntEnumConverter) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	return &EnumConvertingIntIterator{
		preparedIntItem{base: items}, EnumIntConverterSeries(converters...), 0}
}

// IntHandler is an object handling an item type of int.
type IntHandler interface {
	// Handle should do something with item of int.
	// It is suggested to return EndOfIntIterator to stop iteration.
	Handle(int) error
}

// IntHandle is a shortcut implementation
// of IntHandler based on a function.
type IntHandle func(int) error

// Handle does something with item of int.
// It is suggested to return EndOfIntIterator to stop iteration.
func (h IntHandle) Handle(item int) error { return h(item) }

// IntDoNothing does nothing.
var IntDoNothing IntHandler = IntHandle(func(_ int) error { return nil })

type doubleIntHandler struct {
	lhs, rhs IntHandler
}

func (h doubleIntHandler) Handle(item int) error {
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

// IntHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func IntHandlerSeries(handlers ...IntHandler) IntHandler {
	var series = IntDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleIntHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingIntIterator does iteration with
// handling by previously set handler.
type HandlingIntIterator struct {
	preparedIntItem
	handler IntHandler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedIntItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfIntIterator(err) {
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

// IntHandling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func IntHandling(items IntIterator, handlers ...IntHandler) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	return &HandlingIntIterator{
		preparedIntItem{base: items}, IntHandlerSeries(handlers...)}
}

// IntRange iterates over items and use handlers to each one.
func IntRange(items IntIterator, handlers ...IntHandler) error {
	return IntDiscard(IntHandling(items, handlers...))
}

// IntRangeIterator is an iterator over items.
type IntRangeIterator interface {
	// Range should iterate over items.
	Range(...IntHandler) error
}

type sIntRangeIterator struct {
	iter IntIterator
}

// ToIntRangeIterator constructs an instance implementing IntRangeIterator
// based on IntIterator.
func ToIntRangeIterator(iter IntIterator) IntRangeIterator {
	if iter == nil {
		iter = EmptyIntIterator
	}
	return sIntRangeIterator{iter: iter}
}

// MakeIntRangeIterator constructs an instance implementing IntRangeIterator
// based on IntIterMaker.
func MakeIntRangeIterator(maker IntIterMaker) IntRangeIterator {
	if maker == nil {
		maker = MakeNoIntIter
	}
	return ToIntRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sIntRangeIterator) Range(handlers ...IntHandler) error {
	return IntRange(r.iter, handlers...)
}

// IntEnumHandler is an object handling an item type of int and its ordered number.
type IntEnumHandler interface {
	// Handle should do something with item of int and its ordered number.
	// It is suggested to return EndOfIntIterator to stop iteration.
	Handle(int, int) error
}

// IntEnumHandle is a shortcut implementation
// of IntEnumHandler based on a function.
type IntEnumHandle func(int, int) error

// Handle does something with item of int and its ordered number.
// It is suggested to return EndOfIntIterator to stop iteration.
func (h IntEnumHandle) Handle(n int, item int) error { return h(n, item) }

// IntDoEnumNothing does nothing.
var IntDoEnumNothing = IntEnumHandle(func(_ int, _ int) error { return nil })

type doubleIntEnumHandler struct {
	lhs, rhs IntEnumHandler
}

func (h doubleIntEnumHandler) Handle(n int, item int) error {
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

// IntEnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func IntEnumHandlerSeries(handlers ...IntEnumHandler) IntEnumHandler {
	var series IntEnumHandler = IntDoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleIntEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingIntIterator does iteration with
// handling by previously set handler.
type EnumHandlingIntIterator struct {
	preparedIntItem
	handler IntEnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingIntIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedIntItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfIntIterator(err) {
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

// IntEnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func IntEnumHandling(items IntIterator, handlers ...IntEnumHandler) IntIterator {
	if items == nil {
		return EmptyIntIterator
	}
	return &EnumHandlingIntIterator{
		preparedIntItem{base: items}, IntEnumHandlerSeries(handlers...), 0}
}

// IntEnum iterates over items and their ordering numbers and use handlers to each one.
func IntEnum(items IntIterator, handlers ...IntEnumHandler) error {
	return IntDiscard(IntEnumHandling(items, handlers...))
}

// IntEnumIterator is an iterator over items and their ordering numbers.
type IntEnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...IntEnumHandler) error
}

type sIntEnumIterator struct {
	iter IntIterator
}

// ToIntEnumIterator constructs an instance implementing IntEnumIterator
// based on IntIterator.
func ToIntEnumIterator(iter IntIterator) IntEnumIterator {
	if iter == nil {
		iter = EmptyIntIterator
	}
	return sIntEnumIterator{iter: iter}
}

// MakeIntEnumIterator constructs an instance implementing IntEnumIterator
// based on IntIterMaker.
func MakeIntEnumIterator(maker IntIterMaker) IntEnumIterator {
	if maker == nil {
		maker = MakeNoIntIter
	}
	return ToIntEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sIntEnumIterator) Enum(handlers ...IntEnumHandler) error {
	return IntEnum(r.iter, handlers...)
}

// Range iterates over items.
func (r sIntEnumIterator) Range(handlers ...IntHandler) error {
	return IntRange(r.iter, handlers...)
}

type doubleIntIterator struct {
	lhs, rhs IntIterator
	inRHS    bool
}

func (it *doubleIntIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleIntIterator) Next() int {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleIntIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperIntIterator combines all iterators to one.
func SuperIntIterator(itemList ...IntIterator) IntIterator {
	var super = EmptyIntIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleIntIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// IntEnumComparer is a strategy to compare two types.
type IntComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs int) bool
}

// IntCompare is a shortcut implementation
// of IntEnumComparer based on a function.
type IntCompare func(lhs, rhs int) bool

// IsLess is true if lhs is less than rhs.
func (c IntCompare) IsLess(lhs, rhs int) bool { return c(lhs, rhs) }

// EnumIntAlwaysLess is an implementation of IntEnumComparer returning always true.
var IntAlwaysLess IntComparer = IntCompare(func(_, _ int) bool { return true })

type priorityIntIterator struct {
	lhs, rhs preparedIntItem
	comparer IntComparer
}

func (it *priorityIntIterator) HasNext() bool {
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

func (it *priorityIntIterator) Next() int {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfIntIteratorError(
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

func (it priorityIntIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorIntIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorIntIterator(comparer IntComparer, itemList ...IntIterator) IntIterator {
	if comparer == nil {
		comparer = IntAlwaysLess
	}

	var prior = EmptyIntIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityIntIterator{
			lhs:      preparedIntItem{base: itemList[i]},
			rhs:      preparedIntItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// IntEnumComparer is a strategy to compare two types and their order numbers.
type IntEnumComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(nLHS int, lhs int, nRHS int, rhs int) bool
}

// IntEnumCompare is a shortcut implementation
// of IntEnumComparer based on a function.
type IntEnumCompare func(nLHS int, lhs int, nRHS int, rhs int) bool

// IsLess is true if lhs is less than rhs.
func (c IntEnumCompare) IsLess(nLHS int, lhs int, nRHS int, rhs int) bool {
	return c(nLHS, lhs, nRHS, rhs)
}

// EnumIntAlwaysLess is an implementation of IntEnumComparer returning always true.
var EnumIntAlwaysLess IntEnumComparer = IntEnumCompare(
	func(_ int, _ int, _ int, _ int) bool { return true })

type priorityIntEnumIterator struct {
	lhs, rhs           preparedIntItem
	countLHS, countRHS int
	comparer           IntEnumComparer
}

func (it *priorityIntEnumIterator) HasNext() bool {
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

func (it *priorityIntEnumIterator) Next() int {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfIntIteratorError(
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

func (it priorityIntEnumIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorIntEnumIterator compare one by one items and their ordering numbers fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorIntEnumIterator(comparer IntEnumComparer, itemList ...IntIterator) IntIterator {
	if comparer == nil {
		comparer = EnumIntAlwaysLess
	}

	var prior = EmptyIntIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityIntEnumIterator{
			lhs:      preparedIntItem{base: itemList[i]},
			rhs:      preparedIntItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// IntSliceIterator is an iterator based on a slice of int.
type IntSliceIterator struct {
	slice []int
	cur   int
}

// NewIntSliceIterator returns a new instance of IntSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use IntUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewIntSliceIterator(slice []int) *IntSliceIterator {
	it := &IntSliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it IntSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *IntSliceIterator) Next() int {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (IntSliceIterator) Err() error { return nil }

// IntSliceIterator is an iterator based on a slice of int
// and doing iteration in back direction.
type InvertingIntSliceIterator struct {
	slice []int
	cur   int
}

// NewInvertingIntSliceIterator returns a new instance of InvertingIntSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingIntSlice(IntUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingIntSliceIterator(slice []int) *InvertingIntSliceIterator {
	it := &InvertingIntSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingIntSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingIntSliceIterator) Next() int {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingIntSliceIterator) Err() error { return nil }

// IntUnroll unrolls items to slice of int.
func IntUnroll(items IntIterator) IntSlice {
	var slice IntSlice
	panicIfIntIteratorError(IntDiscard(IntHandling(items, IntHandle(func(item int) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// IntSlice is a slice of int.
type IntSlice []int

// MakeIter returns a new instance of IntIterator to iterate over it.
// It returns EmptyIntIterator if the error is not nil.
func (s IntSlice) MakeIter() IntIterator {
	return NewIntSliceIterator(s)
}

// IntSlice is a slice of int which can make inverting iterator.
type InvertingIntSlice []int

// MakeIter returns a new instance of IntIterator to iterate over it.
// It returns EmptyIntIterator if the error is not nil.
func (s InvertingIntSlice) MakeIter() IntIterator {
	return NewInvertingIntSliceIterator(s)
}

// IntInvert unrolls items and make inverting iterator based on them.
func IntInvert(items IntIterator) IntIterator {
	return InvertingIntSlice(IntUnroll(items)).MakeIter()
}

// EndOfIntIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfIntIterator = errors.New("end of int iterator")

func isEndOfIntIterator(err error) bool {
	return errors.Is(err, EndOfIntIterator)
}

func wrapIfNotEndOfIntIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfIntIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfIntIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedIntItem struct {
	base    IntIterator
	hasNext bool
	next    int
	err     error
}

func (it *preparedIntItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedIntItem) Next() int {
	if !it.hasNext {
		panicIfIntIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedIntItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfIntIterator) {
		return it.err
	}
	return it.base.Err()
}
