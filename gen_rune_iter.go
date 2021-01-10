package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// RuneIterator is an iterator over items type of rune.
type RuneIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() rune
	// Err contains first met error while Next.
	Err() error
}

type emptyRuneIterator struct{}

func (emptyRuneIterator) HasNext() bool     { return false }
func (emptyRuneIterator) Next() (next rune) { return 0 }
func (emptyRuneIterator) Err() error        { return nil }

// EmptyRuneIterator is a zero value for RuneIterator.
// It is not contains any item to iterate over it.
var EmptyRuneIterator RuneIterator = emptyRuneIterator{}

// RuneIterMaker is a maker of RuneIterator.
type RuneIterMaker interface {
	// MakeIter should return a new instance of RuneIterator to iterate over it.
	MakeIter() RuneIterator
}

// MakeRuneIter is a shortcut implementation
// of RuneIterator based on a function.
type MakeRuneIter func() RuneIterator

// MakeIter returns a new instance of RuneIterator to iterate over it.
func (m MakeRuneIter) MakeIter() RuneIterator { return m() }

// MakeNoRuneIter is a zero value for RuneIterMaker.
// It always returns EmptyRuneIterator and an empty error.
var MakeNoRuneIter RuneIterMaker = MakeRuneIter(
	func() RuneIterator { return EmptyRuneIterator })

// RuneDiscard just range over all items and do nothing with each of them.
func RuneDiscard(items RuneIterator) error {
	if items == nil {
		items = EmptyRuneIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

// RuneChecker is an object checking an item type of rune
// for some condition.
type RuneChecker interface {
	// Check should check an item type of rune for some condition.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Check(rune) (bool, error)
}

// RuneCheck is a shortcut implementation
// of RuneChecker based on a function.
type RuneCheck func(rune) (bool, error)

// Check checks an item type of rune for some condition.
// It returns EndOfRuneIterator to stop iteration.
func (ch RuneCheck) Check(item rune) (bool, error) { return ch(item) }

var (
	// AlwaysRuneCheckTrue always returns true and empty error.
	AlwaysRuneCheckTrue RuneChecker = RuneCheck(
		func(item rune) (bool, error) { return true, nil })
	// AlwaysRuneCheckFalse always returns false and empty error.
	AlwaysRuneCheckFalse RuneChecker = RuneCheck(
		func(item rune) (bool, error) { return false, nil })
)

// NotRune do an inversion for checker result.
// It is returns AlwaysRuneCheckTrue if checker is nil.
func NotRune(checker RuneChecker) RuneChecker {
	if checker == nil {
		return AlwaysRuneCheckTrue
	}
	return RuneCheck(func(item rune) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andRune struct {
	lhs, rhs RuneChecker
}

func (a andRune) Check(item rune) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllRune combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllRune(checkers ...RuneChecker) RuneChecker {
	var all = AlwaysRuneCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andRune{checkers[i], all}
	}
	return all
}

type orRune struct {
	lhs, rhs RuneChecker
}

func (o orRune) Check(item rune) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyRune combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyRune(checkers ...RuneChecker) RuneChecker {
	var any = AlwaysRuneCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orRune{checkers[i], any}
	}
	return any
}

// FilteringRuneIterator does iteration with
// filtering by previously set checker.
type FilteringRuneIterator struct {
	preparedRuneItem
	filter RuneChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneFiltering sets filter while iterating over items.
// If filters is empty, so all items will return.
func RuneFiltering(items RuneIterator, filters ...RuneChecker) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &FilteringRuneIterator{preparedRuneItem{base: items}, AllRune(filters...)}
}

// DoingUntilRuneIterator does iteration
// until previously set checker is passed.
type DoingUntilRuneIterator struct {
	preparedRuneItem
	until RuneChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfRuneIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// RuneDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func RuneDoingUntil(items RuneIterator, untilList ...RuneChecker) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	var until RuneChecker
	if len(untilList) > 0 {
		until = AllRune(untilList...)
	} else {
		until = AlwaysRuneCheckFalse
	}
	return &DoingUntilRuneIterator{preparedRuneItem{base: items}, until}
}

// RuneSkipUntil sets until conditions to skip few items.
func RuneSkipUntil(items RuneIterator, untilList ...RuneChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return RuneDiscard(RuneDoingUntil(items, untilList...))
}

// RuneEnumChecker is an object checking an item type of rune
// and its ordering number in for some condition.
type RuneEnumChecker interface {
	// Check checks an item type of rune and its ordering number for some condition.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Check(int, rune) (bool, error)
}

// RuneEnumCheck is a shortcut implementation
// of RuneEnumChecker based on a function.
type RuneEnumCheck func(int, rune) (bool, error)

// Check checks an item type of rune and its ordering number for some condition.
// It returns EndOfRuneIterator to stop iteration.
func (ch RuneEnumCheck) Check(n int, item rune) (bool, error) { return ch(n, item) }

type enumFromRuneChecker struct {
	RuneChecker
}

func (ch enumFromRuneChecker) Check(_ int, item rune) (bool, error) {
	return ch.RuneChecker.Check(item)
}

// EnumFromRuneChecker adapts checker type of RuneChecker
// to the interface RuneEnumChecker.
// If checker is nil it is return based on AlwaysRuneCheckFalse enum checker.
func EnumFromRuneChecker(checker RuneChecker) RuneEnumChecker {
	if checker == nil {
		checker = AlwaysRuneCheckFalse
	}
	return &enumFromRuneChecker{checker}
}

var (
	// AlwaysRuneEnumCheckTrue always returns true and empty error.
	AlwaysRuneEnumCheckTrue = EnumFromRuneChecker(
		AlwaysRuneCheckTrue)
	// AlwaysRuneEnumCheckFalse always returns false and empty error.
	AlwaysRuneEnumCheckFalse = EnumFromRuneChecker(
		AlwaysRuneCheckFalse)
)

// EnumNotRune do an inversion for checker result.
// It is returns AlwaysRuneEnumCheckTrue if checker is nil.
func EnumNotRune(checker RuneEnumChecker) RuneEnumChecker {
	if checker == nil {
		return AlwaysRuneEnumCheckTrue
	}
	return RuneEnumCheck(func(n int, item rune) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndRune struct {
	lhs, rhs RuneEnumChecker
}

func (a enumAndRune) Check(n int, item rune) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllRune combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllRune(checkers ...RuneEnumChecker) RuneEnumChecker {
	var all = AlwaysRuneEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndRune{checkers[i], all}
	}
	return all
}

type enumOrRune struct {
	lhs, rhs RuneEnumChecker
}

func (o enumOrRune) Check(n int, item rune) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfRuneIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyRune combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyRune(checkers ...RuneEnumChecker) RuneEnumChecker {
	var any = AlwaysRuneEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrRune{checkers[i], any}
	}
	return any
}

// EnumFilteringRuneIterator does iteration with
// filtering by previously set checker.
type EnumFilteringRuneIterator struct {
	preparedRuneItem
	filter RuneEnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneEnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func RuneEnumFiltering(items RuneIterator, filters ...RuneEnumChecker) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &EnumFilteringRuneIterator{preparedRuneItem{base: items}, EnumAllRune(filters...), 0}
}

// EnumDoingUntilRuneIterator does iteration
// until previously set checker is passed.
type EnumDoingUntilRuneIterator struct {
	preparedRuneItem
	until RuneEnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfRuneIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// RuneEnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func RuneEnumDoingUntil(items RuneIterator, untilList ...RuneEnumChecker) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	var until RuneEnumChecker
	if len(untilList) > 0 {
		until = EnumAllRune(untilList...)
	} else {
		until = AlwaysRuneEnumCheckFalse
	}
	return &EnumDoingUntilRuneIterator{preparedRuneItem{base: items}, until, 0}
}

// RuneEnumSkipUntil sets until conditions to skip few items.
func RuneEnumSkipUntil(items RuneIterator, untilList ...RuneEnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return RuneDiscard(RuneEnumDoingUntil(items, untilList...))
}

// RuneGettingBatch returns the next batch from items.
func RuneGettingBatch(items RuneIterator, batchSize int) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	if batchSize == 0 {
		return items
	}

	return RuneEnumDoingUntil(items, RuneEnumCheck(func(n int, item rune) (bool, error) {
		return n == batchSize-1, nil
	}))
}

// RuneConverter is an object converting an item type of rune.
type RuneConverter interface {
	// Convert should convert an item type of rune into another item of rune.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Convert(rune) (rune, error)
}

// RuneConvert is a shortcut implementation
// of RuneConverter based on a function.
type RuneConvert func(rune) (rune, error)

// Convert converts an item type of rune into another item of rune.
// It is suggested to return EndOfRuneIterator to stop iteration.
func (c RuneConvert) Convert(item rune) (rune, error) { return c(item) }

// NoRuneConvert does nothing with item, just returns it as is.
var NoRuneConvert RuneConverter = RuneConvert(
	func(item rune) (rune, error) { return item, nil })

type doubleRuneConverter struct {
	lhs, rhs RuneConverter
}

func (c doubleRuneConverter) Convert(item rune) (rune, error) {
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

// RuneConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func RuneConverterSeries(converters ...RuneConverter) RuneConverter {
	var series = NoRuneConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleRuneConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingRuneIterator does iteration with
// converting by previously set converter.
type ConvertingRuneIterator struct {
	preparedRuneItem
	converter RuneConverter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneConverting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func RuneConverting(items RuneIterator, converters ...RuneConverter) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &ConvertingRuneIterator{
		preparedRuneItem{base: items}, RuneConverterSeries(converters...)}
}

// RuneEnumConverter is an object converting an item type of rune and its ordering number.
type RuneEnumConverter interface {
	// Convert should convert an item type of rune into another item of rune.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Convert(n int, val rune) (rune, error)
}

// RuneEnumConvert is a shortcut implementation
// of RuneEnumConverter based on a function.
type RuneEnumConvert func(int, rune) (rune, error)

// Convert converts an item type of rune into another item of rune.
// It is suggested to return EndOfRuneIterator to stop iteration.
func (c RuneEnumConvert) Convert(n int, item rune) (rune, error) { return c(n, item) }

// NoRuneEnumConvert does nothing with item, just returns it as is.
var NoRuneEnumConvert RuneEnumConverter = RuneEnumConvert(
	func(_ int, item rune) (rune, error) { return item, nil })

type enumFromRuneConverter struct {
	RuneConverter
}

func (ch enumFromRuneConverter) Convert(_ int, item rune) (rune, error) {
	return ch.RuneConverter.Convert(item)
}

// EnumFromRuneConverter adapts checker type of RuneConverter
// to the interface RuneEnumConverter.
// If converter is nil it is return based on NoRuneConvert enum checker.
func EnumFromRuneConverter(converter RuneConverter) RuneEnumConverter {
	if converter == nil {
		converter = NoRuneConvert
	}
	return &enumFromRuneConverter{converter}
}

type doubleRuneEnumConverter struct {
	lhs, rhs RuneEnumConverter
}

func (c doubleRuneEnumConverter) Convert(n int, item rune) (rune, error) {
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

// EnumRuneConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumRuneConverterSeries(converters ...RuneEnumConverter) RuneEnumConverter {
	var series = NoRuneEnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleRuneEnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingRuneIterator does iteration with
// converting by previously set converter.
type EnumConvertingRuneIterator struct {
	preparedRuneItem
	converter RuneEnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneEnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func RuneEnumConverting(items RuneIterator, converters ...RuneEnumConverter) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &EnumConvertingRuneIterator{
		preparedRuneItem{base: items}, EnumRuneConverterSeries(converters...), 0}
}

// RuneHandler is an object handling an item type of rune.
type RuneHandler interface {
	// Handle should do something with item of rune.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Handle(rune) error
}

// RuneHandle is a shortcut implementation
// of RuneHandler based on a function.
type RuneHandle func(rune) error

// Handle does something with item of rune.
// It is suggested to return EndOfRuneIterator to stop iteration.
func (h RuneHandle) Handle(item rune) error { return h(item) }

// RuneDoNothing does nothing.
var RuneDoNothing RuneHandler = RuneHandle(func(_ rune) error { return nil })

type doubleRuneHandler struct {
	lhs, rhs RuneHandler
}

func (h doubleRuneHandler) Handle(item rune) error {
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

// RuneHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func RuneHandlerSeries(handlers ...RuneHandler) RuneHandler {
	var series = RuneDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleRuneHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingRuneIterator does iteration with
// handling by previously set handler.
type HandlingRuneIterator struct {
	preparedRuneItem
	handler RuneHandler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneHandling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func RuneHandling(items RuneIterator, handlers ...RuneHandler) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &HandlingRuneIterator{
		preparedRuneItem{base: items}, RuneHandlerSeries(handlers...)}
}

// RuneRange iterates over items and use handlers to each one.
func RuneRange(items RuneIterator, handlers ...RuneHandler) error {
	return RuneDiscard(RuneHandling(items, handlers...))
}

// RuneRangeIterator is an iterator over items.
type RuneRangeIterator interface {
	// Range should iterate over items.
	Range(...RuneHandler) error
}

type sRuneRangeIterator struct {
	iter RuneIterator
}

// ToRuneRangeIterator constructs an instance implementing RuneRangeIterator
// based on RuneIterator.
func ToRuneRangeIterator(iter RuneIterator) RuneRangeIterator {
	if iter == nil {
		iter = EmptyRuneIterator
	}
	return sRuneRangeIterator{iter: iter}
}

// MakeRuneRangeIterator constructs an instance implementing RuneRangeIterator
// based on RuneIterMaker.
func MakeRuneRangeIterator(maker RuneIterMaker) RuneRangeIterator {
	if maker == nil {
		maker = MakeNoRuneIter
	}
	return ToRuneRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sRuneRangeIterator) Range(handlers ...RuneHandler) error {
	return RuneRange(r.iter, handlers...)
}

// RuneEnumHandler is an object handling an item type of rune and its ordered number.
type RuneEnumHandler interface {
	// Handle should do something with item of rune and its ordered number.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Handle(int, rune) error
}

// RuneEnumHandle is a shortcut implementation
// of RuneEnumHandler based on a function.
type RuneEnumHandle func(int, rune) error

// Handle does something with item of rune and its ordered number.
// It is suggested to return EndOfRuneIterator to stop iteration.
func (h RuneEnumHandle) Handle(n int, item rune) error { return h(n, item) }

// RuneDoEnumNothing does nothing.
var RuneDoEnumNothing = RuneEnumHandle(func(_ int, _ rune) error { return nil })

type doubleRuneEnumHandler struct {
	lhs, rhs RuneEnumHandler
}

func (h doubleRuneEnumHandler) Handle(n int, item rune) error {
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

// RuneEnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func RuneEnumHandlerSeries(handlers ...RuneEnumHandler) RuneEnumHandler {
	var series RuneEnumHandler = RuneDoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleRuneEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingRuneIterator does iteration with
// handling by previously set handler.
type EnumHandlingRuneIterator struct {
	preparedRuneItem
	handler RuneEnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneEnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func RuneEnumHandling(items RuneIterator, handlers ...RuneEnumHandler) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &EnumHandlingRuneIterator{
		preparedRuneItem{base: items}, RuneEnumHandlerSeries(handlers...), 0}
}

// RuneEnum iterates over items and their ordering numbers and use handlers to each one.
func RuneEnum(items RuneIterator, handlers ...RuneEnumHandler) error {
	return RuneDiscard(RuneEnumHandling(items, handlers...))
}

// RuneEnumIterator is an iterator over items and their ordering numbers.
type RuneEnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...RuneEnumHandler) error
}

type sRuneEnumIterator struct {
	iter RuneIterator
}

// ToRuneEnumIterator constructs an instance implementing RuneEnumIterator
// based on RuneIterator.
func ToRuneEnumIterator(iter RuneIterator) RuneEnumIterator {
	if iter == nil {
		iter = EmptyRuneIterator
	}
	return sRuneEnumIterator{iter: iter}
}

// MakeRuneEnumIterator constructs an instance implementing RuneEnumIterator
// based on RuneIterMaker.
func MakeRuneEnumIterator(maker RuneIterMaker) RuneEnumIterator {
	if maker == nil {
		maker = MakeNoRuneIter
	}
	return ToRuneEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sRuneEnumIterator) Enum(handlers ...RuneEnumHandler) error {
	return RuneEnum(r.iter, handlers...)
}

// Range iterates over items.
func (r sRuneEnumIterator) Range(handlers ...RuneHandler) error {
	return RuneRange(r.iter, handlers...)
}

type doubleRuneIterator struct {
	lhs, rhs RuneIterator
	inRHS    bool
}

func (it *doubleRuneIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleRuneIterator) Next() rune {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleRuneIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperRuneIterator combines all iterators to one.
func SuperRuneIterator(itemList ...RuneIterator) RuneIterator {
	var super = EmptyRuneIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleRuneIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// RuneComparer is a strategy to compare two types.
type RuneComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs rune) bool
}

// RuneCompare is a shortcut implementation
// of RuneComparer based on a function.
type RuneCompare func(lhs, rhs rune) bool

// IsLess is true if lhs is less than rhs.
func (c RuneCompare) IsLess(lhs, rhs rune) bool { return c(lhs, rhs) }

// RuneAlwaysLess is an implementation of RuneComparer returning always true.
var RuneAlwaysLess RuneComparer = RuneCompare(func(_, _ rune) bool { return true })

type priorityRuneIterator struct {
	lhs, rhs preparedRuneItem
	comparer RuneComparer
}

func (it *priorityRuneIterator) HasNext() bool {
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

func (it *priorityRuneIterator) Next() rune {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfRuneIteratorError(
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

func (it priorityRuneIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorRuneIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorRuneIterator(comparer RuneComparer, itemList ...RuneIterator) RuneIterator {
	if comparer == nil {
		comparer = RuneAlwaysLess
	}

	var prior = EmptyRuneIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityRuneIterator{
			lhs:      preparedRuneItem{base: itemList[i]},
			rhs:      preparedRuneItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// RuneSliceIterator is an iterator based on a slice of rune.
type RuneSliceIterator struct {
	slice []rune
	cur   int
}

// NewRuneSliceIterator returns a new instance of RuneSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use RuneUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewRuneSliceIterator(slice []rune) *RuneSliceIterator {
	it := &RuneSliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it RuneSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *RuneSliceIterator) Next() rune {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (RuneSliceIterator) Err() error { return nil }

// RuneSliceIterator is an iterator based on a slice of rune
// and doing iteration in back direction.
type InvertingRuneSliceIterator struct {
	slice []rune
	cur   int
}

// NewInvertingRuneSliceIterator returns a new instance of InvertingRuneSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingRuneSlice(RuneUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingRuneSliceIterator(slice []rune) *InvertingRuneSliceIterator {
	it := &InvertingRuneSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingRuneSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingRuneSliceIterator) Next() rune {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingRuneSliceIterator) Err() error { return nil }

// RuneUnroll unrolls items to slice of rune.
func RuneUnroll(items RuneIterator) RuneSlice {
	var slice RuneSlice
	panicIfRuneIteratorError(RuneDiscard(RuneHandling(items, RuneHandle(func(item rune) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// RuneSlice is a slice of rune.
type RuneSlice []rune

// MakeIter returns a new instance of RuneIterator to iterate over it.
// It returns EmptyRuneIterator if the error is not nil.
func (s RuneSlice) MakeIter() RuneIterator {
	return NewRuneSliceIterator(s)
}

// RuneSlice is a slice of rune which can make inverting iterator.
type InvertingRuneSlice []rune

// MakeIter returns a new instance of RuneIterator to iterate over it.
// It returns EmptyRuneIterator if the error is not nil.
func (s InvertingRuneSlice) MakeIter() RuneIterator {
	return NewInvertingRuneSliceIterator(s)
}

// RuneInvert unrolls items and make inverting iterator based on them.
func RuneInvert(items RuneIterator) RuneIterator {
	return InvertingRuneSlice(RuneUnroll(items)).MakeIter()
}

// EndOfRuneIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfRuneIterator = errors.New("end of rune iterator")

func isEndOfRuneIterator(err error) bool {
	return errors.Is(err, EndOfRuneIterator)
}

func wrapIfNotEndOfRuneIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfRuneIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfRuneIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedRuneItem struct {
	base    RuneIterator
	hasNext bool
	next    rune
	err     error
}

func (it *preparedRuneItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedRuneItem) Next() rune {
	if !it.hasNext {
		panicIfRuneIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = 0
	return next
}

func (it preparedRuneItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfRuneIterator) {
		return it.err
	}
	return it.base.Err()
}
