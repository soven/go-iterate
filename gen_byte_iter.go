package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// ByteIterator is an iterator over items type of byte.
type ByteIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() byte
	// Err contains first met error while Next.
	Err() error
}

type emptyByteIterator struct{}

func (emptyByteIterator) HasNext() bool     { return false }
func (emptyByteIterator) Next() (next byte) { return 0 }
func (emptyByteIterator) Err() error        { return nil }

// EmptyByteIterator is a zero value for ByteIterator.
// It is not contains any item to iterate over it.
var EmptyByteIterator ByteIterator = emptyByteIterator{}

// ByteIterMaker is a maker of ByteIterator.
type ByteIterMaker interface {
	// MakeIter should return a new instance of ByteIterator to iterate over it.
	MakeIter() ByteIterator
}

// MakeByteIter is a shortcut implementation
// of ByteIterator based on a function.
type MakeByteIter func() ByteIterator

// MakeIter returns a new instance of ByteIterator to iterate over it.
func (m MakeByteIter) MakeIter() ByteIterator { return m() }

// MakeNoByteIter is a zero value for ByteIterMaker.
// It always returns EmptyByteIterator and an empty error.
var MakeNoByteIter ByteIterMaker = MakeByteIter(
	func() ByteIterator { return EmptyByteIterator })

// ByteDiscard just range over all items and do nothing with each of them.
func ByteDiscard(items ByteIterator) error {
	if items == nil {
		items = EmptyByteIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

// ByteChecker is an object checking an item type of byte
// for some condition.
type ByteChecker interface {
	// Check should check an item type of byte for some condition.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Check(byte) (bool, error)
}

// ByteCheck is a shortcut implementation
// of ByteChecker based on a function.
type ByteCheck func(byte) (bool, error)

// Check checks an item type of byte for some condition.
// It returns EndOfByteIterator to stop iteration.
func (ch ByteCheck) Check(item byte) (bool, error) { return ch(item) }

var (
	// AlwaysByteCheckTrue always returns true and empty error.
	AlwaysByteCheckTrue ByteChecker = ByteCheck(
		func(item byte) (bool, error) { return true, nil })
	// AlwaysByteCheckFalse always returns false and empty error.
	AlwaysByteCheckFalse ByteChecker = ByteCheck(
		func(item byte) (bool, error) { return false, nil })
)

// NotByte do an inversion for checker result.
// It is returns AlwaysByteCheckTrue if checker is nil.
func NotByte(checker ByteChecker) ByteChecker {
	if checker == nil {
		return AlwaysByteCheckTrue
	}
	return ByteCheck(func(item byte) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andByte struct {
	lhs, rhs ByteChecker
}

func (a andByte) Check(item byte) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllByte combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllByte(checkers ...ByteChecker) ByteChecker {
	var all = AlwaysByteCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andByte{checkers[i], all}
	}
	return all
}

type orByte struct {
	lhs, rhs ByteChecker
}

func (o orByte) Check(item byte) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyByte combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyByte(checkers ...ByteChecker) ByteChecker {
	var any = AlwaysByteCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orByte{checkers[i], any}
	}
	return any
}

// FilteringByteIterator does iteration with
// filtering by previously set checker.
type FilteringByteIterator struct {
	preparedByteItem
	filter ByteChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedByteItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteFiltering sets filter while iterating over items.
// If filters is empty, so all items will return.
func ByteFiltering(items ByteIterator, filters ...ByteChecker) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &FilteringByteIterator{preparedByteItem{base: items}, AllByte(filters...)}
}

// DoingUntilByteIterator does iteration
// until previously set checker is passed.
type DoingUntilByteIterator struct {
	preparedByteItem
	until ByteChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedByteItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfByteIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfByteIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// ByteDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func ByteDoingUntil(items ByteIterator, untilList ...ByteChecker) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	var until ByteChecker
	if len(untilList) > 0 {
		until = AllByte(untilList...)
	} else {
		until = AlwaysByteCheckFalse
	}
	return &DoingUntilByteIterator{preparedByteItem{base: items}, until}
}

// ByteSkipUntil sets until conditions to skip few items.
func ByteSkipUntil(items ByteIterator, untilList ...ByteChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return ByteDiscard(ByteDoingUntil(items, untilList...))
}

// ByteEnumChecker is an object checking an item type of byte
// and its ordering number in for some condition.
type ByteEnumChecker interface {
	// Check checks an item type of byte and its ordering number for some condition.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Check(int, byte) (bool, error)
}

// ByteEnumCheck is a shortcut implementation
// of ByteEnumChecker based on a function.
type ByteEnumCheck func(int, byte) (bool, error)

// Check checks an item type of byte and its ordering number for some condition.
// It returns EndOfByteIterator to stop iteration.
func (ch ByteEnumCheck) Check(n int, item byte) (bool, error) { return ch(n, item) }

type enumFromByteChecker struct {
	ByteChecker
}

func (ch enumFromByteChecker) Check(_ int, item byte) (bool, error) {
	return ch.ByteChecker.Check(item)
}

// EnumFromByteChecker adapts checker type of ByteChecker
// to the interface ByteEnumChecker.
// If checker is nil it is return based on AlwaysByteCheckFalse enum checker.
func EnumFromByteChecker(checker ByteChecker) ByteEnumChecker {
	if checker == nil {
		checker = AlwaysByteCheckFalse
	}
	return &enumFromByteChecker{checker}
}

var (
	// AlwaysByteEnumCheckTrue always returns true and empty error.
	AlwaysByteEnumCheckTrue = EnumFromByteChecker(
		AlwaysByteCheckTrue)
	// AlwaysByteEnumCheckFalse always returns false and empty error.
	AlwaysByteEnumCheckFalse = EnumFromByteChecker(
		AlwaysByteCheckFalse)
)

// EnumNotByte do an inversion for checker result.
// It is returns AlwaysByteEnumCheckTrue if checker is nil.
func EnumNotByte(checker ByteEnumChecker) ByteEnumChecker {
	if checker == nil {
		return AlwaysByteEnumCheckTrue
	}
	return ByteEnumCheck(func(n int, item byte) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndByte struct {
	lhs, rhs ByteEnumChecker
}

func (a enumAndByte) Check(n int, item byte) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllByte combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllByte(checkers ...ByteEnumChecker) ByteEnumChecker {
	var all = AlwaysByteEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndByte{checkers[i], all}
	}
	return all
}

type enumOrByte struct {
	lhs, rhs ByteEnumChecker
}

func (o enumOrByte) Check(n int, item byte) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfByteIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyByte combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyByte(checkers ...ByteEnumChecker) ByteEnumChecker {
	var any = AlwaysByteEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrByte{checkers[i], any}
	}
	return any
}

// EnumFilteringByteIterator does iteration with
// filtering by previously set checker.
type EnumFilteringByteIterator struct {
	preparedByteItem
	filter ByteEnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedByteItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteEnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func ByteEnumFiltering(items ByteIterator, filters ...ByteEnumChecker) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &EnumFilteringByteIterator{preparedByteItem{base: items}, EnumAllByte(filters...), 0}
}

// EnumDoingUntilByteIterator does iteration
// until previously set checker is passed.
type EnumDoingUntilByteIterator struct {
	preparedByteItem
	until ByteEnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedByteItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfByteIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfByteIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// ByteEnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func ByteEnumDoingUntil(items ByteIterator, untilList ...ByteEnumChecker) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	var until ByteEnumChecker
	if len(untilList) > 0 {
		until = EnumAllByte(untilList...)
	} else {
		until = AlwaysByteEnumCheckFalse
	}
	return &EnumDoingUntilByteIterator{preparedByteItem{base: items}, until, 0}
}

// ByteEnumSkipUntil sets until conditions to skip few items.
func ByteEnumSkipUntil(items ByteIterator, untilList ...ByteEnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return ByteDiscard(ByteEnumDoingUntil(items, untilList...))
}

// ByteGettingBatch returns the next batch from items.
func ByteGettingBatch(items ByteIterator, batchSize int) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	if batchSize == 0 {
		return items
	}

	return ByteEnumDoingUntil(items, ByteEnumCheck(func(n int, item byte) (bool, error) {
		return n == batchSize-1, nil
	}))
}

// ByteConverter is an object converting an item type of byte.
type ByteConverter interface {
	// Convert should convert an item type of byte into another item of byte.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Convert(byte) (byte, error)
}

// ByteConvert is a shortcut implementation
// of ByteConverter based on a function.
type ByteConvert func(byte) (byte, error)

// Convert converts an item type of byte into another item of byte.
// It is suggested to return EndOfByteIterator to stop iteration.
func (c ByteConvert) Convert(item byte) (byte, error) { return c(item) }

// NoByteConvert does nothing with item, just returns it as is.
var NoByteConvert ByteConverter = ByteConvert(
	func(item byte) (byte, error) { return item, nil })

type doubleByteConverter struct {
	lhs, rhs ByteConverter
}

func (c doubleByteConverter) Convert(item byte) (byte, error) {
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

// ByteConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func ByteConverterSeries(converters ...ByteConverter) ByteConverter {
	var series = NoByteConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleByteConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingByteIterator does iteration with
// converting by previously set converter.
type ConvertingByteIterator struct {
	preparedByteItem
	converter ByteConverter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedByteItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteConverting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func ByteConverting(items ByteIterator, converters ...ByteConverter) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &ConvertingByteIterator{
		preparedByteItem{base: items}, ByteConverterSeries(converters...)}
}

// ByteEnumConverter is an object converting an item type of byte and its ordering number.
type ByteEnumConverter interface {
	// Convert should convert an item type of byte into another item of byte.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Convert(n int, val byte) (byte, error)
}

// ByteEnumConvert is a shortcut implementation
// of ByteEnumConverter based on a function.
type ByteEnumConvert func(int, byte) (byte, error)

// Convert converts an item type of byte into another item of byte.
// It is suggested to return EndOfByteIterator to stop iteration.
func (c ByteEnumConvert) Convert(n int, item byte) (byte, error) { return c(n, item) }

// NoByteEnumConvert does nothing with item, just returns it as is.
var NoByteEnumConvert ByteEnumConverter = ByteEnumConvert(
	func(_ int, item byte) (byte, error) { return item, nil })

type enumFromByteConverter struct {
	ByteConverter
}

func (ch enumFromByteConverter) Convert(_ int, item byte) (byte, error) {
	return ch.ByteConverter.Convert(item)
}

// EnumFromByteConverter adapts checker type of ByteConverter
// to the interface ByteEnumConverter.
// If converter is nil it is return based on NoByteConvert enum checker.
func EnumFromByteConverter(converter ByteConverter) ByteEnumConverter {
	if converter == nil {
		converter = NoByteConvert
	}
	return &enumFromByteConverter{converter}
}

type doubleByteEnumConverter struct {
	lhs, rhs ByteEnumConverter
}

func (c doubleByteEnumConverter) Convert(n int, item byte) (byte, error) {
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

// EnumByteConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumByteConverterSeries(converters ...ByteEnumConverter) ByteEnumConverter {
	var series = NoByteEnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleByteEnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingByteIterator does iteration with
// converting by previously set converter.
type EnumConvertingByteIterator struct {
	preparedByteItem
	converter ByteEnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedByteItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteEnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func ByteEnumConverting(items ByteIterator, converters ...ByteEnumConverter) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &EnumConvertingByteIterator{
		preparedByteItem{base: items}, EnumByteConverterSeries(converters...), 0}
}

// ByteHandler is an object handling an item type of byte.
type ByteHandler interface {
	// Handle should do something with item of byte.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Handle(byte) error
}

// ByteHandle is a shortcut implementation
// of ByteHandler based on a function.
type ByteHandle func(byte) error

// Handle does something with item of byte.
// It is suggested to return EndOfByteIterator to stop iteration.
func (h ByteHandle) Handle(item byte) error { return h(item) }

// ByteDoNothing does nothing.
var ByteDoNothing ByteHandler = ByteHandle(func(_ byte) error { return nil })

type doubleByteHandler struct {
	lhs, rhs ByteHandler
}

func (h doubleByteHandler) Handle(item byte) error {
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

// ByteHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func ByteHandlerSeries(handlers ...ByteHandler) ByteHandler {
	var series = ByteDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleByteHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingByteIterator does iteration with
// handling by previously set handler.
type HandlingByteIterator struct {
	preparedByteItem
	handler ByteHandler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedByteItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteHandling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func ByteHandling(items ByteIterator, handlers ...ByteHandler) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &HandlingByteIterator{
		preparedByteItem{base: items}, ByteHandlerSeries(handlers...)}
}

// ByteRange iterates over items and use handlers to each one.
func ByteRange(items ByteIterator, handlers ...ByteHandler) error {
	return ByteDiscard(ByteHandling(items, handlers...))
}

// ByteRangeIterator is an iterator over items.
type ByteRangeIterator interface {
	// Range should iterate over items.
	Range(...ByteHandler) error
}

type sByteRangeIterator struct {
	iter ByteIterator
}

// ToByteRangeIterator constructs an instance implementing ByteRangeIterator
// based on ByteIterator.
func ToByteRangeIterator(iter ByteIterator) ByteRangeIterator {
	if iter == nil {
		iter = EmptyByteIterator
	}
	return sByteRangeIterator{iter: iter}
}

// MakeByteRangeIterator constructs an instance implementing ByteRangeIterator
// based on ByteIterMaker.
func MakeByteRangeIterator(maker ByteIterMaker) ByteRangeIterator {
	if maker == nil {
		maker = MakeNoByteIter
	}
	return ToByteRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sByteRangeIterator) Range(handlers ...ByteHandler) error {
	return ByteRange(r.iter, handlers...)
}

// ByteEnumHandler is an object handling an item type of byte and its ordered number.
type ByteEnumHandler interface {
	// Handle should do something with item of byte and its ordered number.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Handle(int, byte) error
}

// ByteEnumHandle is a shortcut implementation
// of ByteEnumHandler based on a function.
type ByteEnumHandle func(int, byte) error

// Handle does something with item of byte and its ordered number.
// It is suggested to return EndOfByteIterator to stop iteration.
func (h ByteEnumHandle) Handle(n int, item byte) error { return h(n, item) }

// ByteDoEnumNothing does nothing.
var ByteDoEnumNothing = ByteEnumHandle(func(_ int, _ byte) error { return nil })

type doubleByteEnumHandler struct {
	lhs, rhs ByteEnumHandler
}

func (h doubleByteEnumHandler) Handle(n int, item byte) error {
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

// ByteEnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func ByteEnumHandlerSeries(handlers ...ByteEnumHandler) ByteEnumHandler {
	var series ByteEnumHandler = ByteDoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleByteEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingByteIterator does iteration with
// handling by previously set handler.
type EnumHandlingByteIterator struct {
	preparedByteItem
	handler ByteEnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedByteItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteEnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func ByteEnumHandling(items ByteIterator, handlers ...ByteEnumHandler) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &EnumHandlingByteIterator{
		preparedByteItem{base: items}, ByteEnumHandlerSeries(handlers...), 0}
}

// ByteEnum iterates over items and their ordering numbers and use handlers to each one.
func ByteEnum(items ByteIterator, handlers ...ByteEnumHandler) error {
	return ByteDiscard(ByteEnumHandling(items, handlers...))
}

// ByteEnumIterator is an iterator over items and their ordering numbers.
type ByteEnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...ByteEnumHandler) error
}

type sByteEnumIterator struct {
	iter ByteIterator
}

// ToByteEnumIterator constructs an instance implementing ByteEnumIterator
// based on ByteIterator.
func ToByteEnumIterator(iter ByteIterator) ByteEnumIterator {
	if iter == nil {
		iter = EmptyByteIterator
	}
	return sByteEnumIterator{iter: iter}
}

// MakeByteEnumIterator constructs an instance implementing ByteEnumIterator
// based on ByteIterMaker.
func MakeByteEnumIterator(maker ByteIterMaker) ByteEnumIterator {
	if maker == nil {
		maker = MakeNoByteIter
	}
	return ToByteEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sByteEnumIterator) Enum(handlers ...ByteEnumHandler) error {
	return ByteEnum(r.iter, handlers...)
}

// Range iterates over items.
func (r sByteEnumIterator) Range(handlers ...ByteHandler) error {
	return ByteRange(r.iter, handlers...)
}

type doubleByteIterator struct {
	lhs, rhs ByteIterator
	inRHS    bool
}

func (it *doubleByteIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleByteIterator) Next() byte {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleByteIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperByteIterator combines all iterators to one.
func SuperByteIterator(itemList ...ByteIterator) ByteIterator {
	var super = EmptyByteIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleByteIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// ByteEnumComparer is a strategy to compare two types.
type ByteComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs byte) bool
}

// ByteCompare is a shortcut implementation
// of ByteEnumComparer based on a function.
type ByteCompare func(lhs, rhs byte) bool

// IsLess is true if lhs is less than rhs.
func (c ByteCompare) IsLess(lhs, rhs byte) bool { return c(lhs, rhs) }

// EnumByteAlwaysLess is an implementation of ByteEnumComparer returning always true.
var ByteAlwaysLess ByteComparer = ByteCompare(func(_, _ byte) bool { return true })

type priorityByteIterator struct {
	lhs, rhs preparedByteItem
	comparer ByteComparer
}

func (it *priorityByteIterator) HasNext() bool {
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

func (it *priorityByteIterator) Next() byte {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfByteIteratorError(
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

func (it priorityByteIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorByteIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorByteIterator(comparer ByteComparer, itemList ...ByteIterator) ByteIterator {
	if comparer == nil {
		comparer = ByteAlwaysLess
	}

	var prior = EmptyByteIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityByteIterator{
			lhs:      preparedByteItem{base: itemList[i]},
			rhs:      preparedByteItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// ByteEnumComparer is a strategy to compare two types and their order numbers.
type ByteEnumComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(nLHS int, lhs byte, nRHS int, rhs byte) bool
}

// ByteEnumCompare is a shortcut implementation
// of ByteEnumComparer based on a function.
type ByteEnumCompare func(nLHS int, lhs byte, nRHS int, rhs byte) bool

// IsLess is true if lhs is less than rhs.
func (c ByteEnumCompare) IsLess(nLHS int, lhs byte, nRHS int, rhs byte) bool {
	return c(nLHS, lhs, nRHS, rhs)
}

// EnumByteAlwaysLess is an implementation of ByteEnumComparer returning always true.
var EnumByteAlwaysLess ByteEnumComparer = ByteEnumCompare(
	func(_ int, _ byte, _ int, _ byte) bool { return true })

type priorityByteEnumIterator struct {
	lhs, rhs           preparedByteItem
	countLHS, countRHS int
	comparer           ByteEnumComparer
}

func (it *priorityByteEnumIterator) HasNext() bool {
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

func (it *priorityByteEnumIterator) Next() byte {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfByteIteratorError(
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

func (it priorityByteEnumIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorByteEnumIterator compare one by one items and their ordering numbers fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorByteEnumIterator(comparer ByteEnumComparer, itemList ...ByteIterator) ByteIterator {
	if comparer == nil {
		comparer = EnumByteAlwaysLess
	}

	var prior = EmptyByteIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityByteEnumIterator{
			lhs:      preparedByteItem{base: itemList[i]},
			rhs:      preparedByteItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// ByteSliceIterator is an iterator based on a slice of byte.
type ByteSliceIterator struct {
	slice []byte
	cur   int
}

// NewByteSliceIterator returns a new instance of ByteSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use ByteUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewByteSliceIterator(slice []byte) *ByteSliceIterator {
	it := &ByteSliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it ByteSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *ByteSliceIterator) Next() byte {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (ByteSliceIterator) Err() error { return nil }

// ByteSliceIterator is an iterator based on a slice of byte
// and doing iteration in back direction.
type InvertingByteSliceIterator struct {
	slice []byte
	cur   int
}

// NewInvertingByteSliceIterator returns a new instance of InvertingByteSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingByteSlice(ByteUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingByteSliceIterator(slice []byte) *InvertingByteSliceIterator {
	it := &InvertingByteSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingByteSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingByteSliceIterator) Next() byte {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingByteSliceIterator) Err() error { return nil }

// ByteUnroll unrolls items to slice of byte.
func ByteUnroll(items ByteIterator) ByteSlice {
	var slice ByteSlice
	panicIfByteIteratorError(ByteDiscard(ByteHandling(items, ByteHandle(func(item byte) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// ByteSlice is a slice of byte.
type ByteSlice []byte

// MakeIter returns a new instance of ByteIterator to iterate over it.
// It returns EmptyByteIterator if the error is not nil.
func (s ByteSlice) MakeIter() ByteIterator {
	return NewByteSliceIterator(s)
}

// ByteSlice is a slice of byte which can make inverting iterator.
type InvertingByteSlice []byte

// MakeIter returns a new instance of ByteIterator to iterate over it.
// It returns EmptyByteIterator if the error is not nil.
func (s InvertingByteSlice) MakeIter() ByteIterator {
	return NewInvertingByteSliceIterator(s)
}

// ByteInvert unrolls items and make inverting iterator based on them.
func ByteInvert(items ByteIterator) ByteIterator {
	return InvertingByteSlice(ByteUnroll(items)).MakeIter()
}

// EndOfByteIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfByteIterator = errors.New("end of byte iterator")

func isEndOfByteIterator(err error) bool {
	return errors.Is(err, EndOfByteIterator)
}

func wrapIfNotEndOfByteIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfByteIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfByteIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedByteItem struct {
	base    ByteIterator
	hasNext bool
	next    byte
	err     error
}

func (it *preparedByteItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedByteItem) Next() byte {
	if !it.hasNext {
		panicIfByteIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedByteItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfByteIterator) {
		return it.err
	}
	return it.base.Err()
}
