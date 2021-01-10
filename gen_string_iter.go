package iter

import "github.com/pkg/errors"

var _ = errors.New("") // hack to get import // TODO clearly

// StringIterator is an iterator over items type of string.
type StringIterator interface {
	// HasNext checks if there is the next item
	// in the iterator. HasNext should be idempotent.
	HasNext() bool
	// Next should return next item in the iterator.
	// It should be invoked after check HasNext.
	Next() string
	// Err contains first met error while Next.
	Err() error
}

type emptyStringIterator struct{}

func (emptyStringIterator) HasNext() bool       { return false }
func (emptyStringIterator) Next() (next string) { return "" }
func (emptyStringIterator) Err() error          { return nil }

// EmptyStringIterator is a zero value for StringIterator.
// It is not contains any item to iterate over it.
var EmptyStringIterator StringIterator = emptyStringIterator{}

// StringIterMaker is a maker of StringIterator.
type StringIterMaker interface {
	// MakeIter should return a new instance of StringIterator to iterate over it.
	MakeIter() StringIterator
}

// MakeStringIter is a shortcut implementation
// of StringIterator based on a function.
type MakeStringIter func() StringIterator

// MakeIter returns a new instance of StringIterator to iterate over it.
func (m MakeStringIter) MakeIter() StringIterator { return m() }

// MakeNoStringIter is a zero value for StringIterMaker.
// It always returns EmptyStringIterator and an empty error.
var MakeNoStringIter StringIterMaker = MakeStringIter(
	func() StringIterator { return EmptyStringIterator })

// StringDiscard just range over all items and do nothing with each of them.
func StringDiscard(items StringIterator) error {
	if items == nil {
		items = EmptyStringIterator
	}
	for items.HasNext() {
		_ = items.Next()
	}
	// no error wrapping since no additional context for the error; just return it.
	return items.Err()
}

// StringChecker is an object checking an item type of string
// for some condition.
type StringChecker interface {
	// Check should check an item type of string for some condition.
	// It is suggested to return EndOfStringIterator to stop iteration.
	Check(string) (bool, error)
}

// StringCheck is a shortcut implementation
// of StringChecker based on a function.
type StringCheck func(string) (bool, error)

// Check checks an item type of string for some condition.
// It returns EndOfStringIterator to stop iteration.
func (ch StringCheck) Check(item string) (bool, error) { return ch(item) }

var (
	// AlwaysStringCheckTrue always returns true and empty error.
	AlwaysStringCheckTrue StringChecker = StringCheck(
		func(item string) (bool, error) { return true, nil })
	// AlwaysStringCheckFalse always returns false and empty error.
	AlwaysStringCheckFalse StringChecker = StringCheck(
		func(item string) (bool, error) { return false, nil })
)

// NotString do an inversion for checker result.
// It is returns AlwaysStringCheckTrue if checker is nil.
func NotString(checker StringChecker) StringChecker {
	if checker == nil {
		return AlwaysStringCheckTrue
	}
	return StringCheck(func(item string) (bool, error) {
		yes, err := checker.Check(item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type andString struct {
	lhs, rhs StringChecker
}

func (a andString) Check(item string) (bool, error) {
	isLHSPassed, err := a.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AllString combines all the given checkers to one
// checking if all checkers return true.
// It returns true checker if the list of checkers is empty.
func AllString(checkers ...StringChecker) StringChecker {
	var all = AlwaysStringCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = andString{checkers[i], all}
	}
	return all
}

type orString struct {
	lhs, rhs StringChecker
}

func (o orString) Check(item string) (bool, error) {
	isLHSPassed, err := o.lhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// AnyString combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func AnyString(checkers ...StringChecker) StringChecker {
	var any = AlwaysStringCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = orString{checkers[i], any}
	}
	return any
}

// FilteringStringIterator does iteration with
// filtering by previously set checker.
type FilteringStringIterator struct {
	preparedStringItem
	filter StringChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *FilteringStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedStringItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(next)
		if err != nil {
			if !isEndOfStringIterator(err) {
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

// StringFiltering sets filter while iterating over items.
// If filters is empty, so all items will return.
func StringFiltering(items StringIterator, filters ...StringChecker) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	return &FilteringStringIterator{preparedStringItem{base: items}, AllString(filters...)}
}

// DoingUntilStringIterator does iteration
// until previously set checker is passed.
type DoingUntilStringIterator struct {
	preparedStringItem
	until StringChecker
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *DoingUntilStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedStringItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(next)
		if err != nil {
			if !isEndOfStringIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}

		if isUntilPassed {
			it.err = EndOfStringIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// StringDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func StringDoingUntil(items StringIterator, untilList ...StringChecker) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	var until StringChecker
	if len(untilList) > 0 {
		until = AllString(untilList...)
	} else {
		until = AlwaysStringCheckFalse
	}
	return &DoingUntilStringIterator{preparedStringItem{base: items}, until}
}

// StringSkipUntil sets until conditions to skip few items.
func StringSkipUntil(items StringIterator, untilList ...StringChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return StringDiscard(StringDoingUntil(items, untilList...))
}

// StringEnumChecker is an object checking an item type of string
// and its ordering number in for some condition.
type StringEnumChecker interface {
	// Check checks an item type of string and its ordering number for some condition.
	// It is suggested to return EndOfStringIterator to stop iteration.
	Check(int, string) (bool, error)
}

// StringEnumCheck is a shortcut implementation
// of StringEnumChecker based on a function.
type StringEnumCheck func(int, string) (bool, error)

// Check checks an item type of string and its ordering number for some condition.
// It returns EndOfStringIterator to stop iteration.
func (ch StringEnumCheck) Check(n int, item string) (bool, error) { return ch(n, item) }

type enumFromStringChecker struct {
	StringChecker
}

func (ch enumFromStringChecker) Check(_ int, item string) (bool, error) {
	return ch.StringChecker.Check(item)
}

// EnumFromStringChecker adapts checker type of StringChecker
// to the interface StringEnumChecker.
// If checker is nil it is return based on AlwaysStringCheckFalse enum checker.
func EnumFromStringChecker(checker StringChecker) StringEnumChecker {
	if checker == nil {
		checker = AlwaysStringCheckFalse
	}
	return &enumFromStringChecker{checker}
}

var (
	// AlwaysStringEnumCheckTrue always returns true and empty error.
	AlwaysStringEnumCheckTrue = EnumFromStringChecker(
		AlwaysStringCheckTrue)
	// AlwaysStringEnumCheckFalse always returns false and empty error.
	AlwaysStringEnumCheckFalse = EnumFromStringChecker(
		AlwaysStringCheckFalse)
)

// EnumNotString do an inversion for checker result.
// It is returns AlwaysStringEnumCheckTrue if checker is nil.
func EnumNotString(checker StringEnumChecker) StringEnumChecker {
	if checker == nil {
		return AlwaysStringEnumCheckTrue
	}
	return StringEnumCheck(func(n int, item string) (bool, error) {
		yes, err := checker.Check(n, item)
		if err != nil {
			// No error wrapping since an error context is missing.
			return false, err
		}

		return !yes, nil
	})
}

type enumAndString struct {
	lhs, rhs StringEnumChecker
}

func (a enumAndString) Check(n int, item string) (bool, error) {
	isLHSPassed, err := a.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "lhs check")
	}
	if !isLHSPassed {
		return false, nil
	}

	isRHSPassed, err := a.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAllString combines all the given checkers to one
// checking if all checkers return true.
// It returns true if the list of checkers is empty.
func EnumAllString(checkers ...StringEnumChecker) StringEnumChecker {
	var all = AlwaysStringEnumCheckTrue
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		all = enumAndString{checkers[i], all}
	}
	return all
}

type enumOrString struct {
	lhs, rhs StringEnumChecker
}

func (o enumOrString) Check(n int, item string) (bool, error) {
	isLHSPassed, err := o.lhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "lhs check")
	}
	if isLHSPassed {
		return true, nil
	}

	isRHSPassed, err := o.rhs.Check(n, item)
	if err != nil {
		return false, wrapIfNotEndOfStringIterator(err, "rhs check")
	}
	return isRHSPassed, nil
}

// EnumAnyString combines all the given checkers to one.
// checking if any checker return true.
// It returns false if the list of checkers is empty.
func EnumAnyString(checkers ...StringEnumChecker) StringEnumChecker {
	var any = AlwaysStringEnumCheckFalse
	for i := len(checkers) - 1; i >= 0; i-- {
		if checkers[i] == nil {
			continue
		}
		any = enumOrString{checkers[i], any}
	}
	return any
}

// EnumFilteringStringIterator does iteration with
// filtering by previously set checker.
type EnumFilteringStringIterator struct {
	preparedStringItem
	filter StringEnumChecker
	count  int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumFilteringStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedStringItem.HasNext() {
		next := it.base.Next()
		isFilterPassed, err := it.filter.Check(it.count, next)
		if err != nil {
			if !isEndOfStringIterator(err) {
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

// StringEnumFiltering sets filter while iterating over items and their serial numbers.
// If filters is empty, so all items will return.
func StringEnumFiltering(items StringIterator, filters ...StringEnumChecker) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	return &EnumFilteringStringIterator{preparedStringItem{base: items}, EnumAllString(filters...), 0}
}

// EnumDoingUntilStringIterator does iteration
// until previously set checker is passed.
type EnumDoingUntilStringIterator struct {
	preparedStringItem
	until StringEnumChecker
	count int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumDoingUntilStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	for it.preparedStringItem.HasNext() {
		next := it.base.Next()
		isUntilPassed, err := it.until.Check(it.count, next)
		if err != nil {
			if !isEndOfStringIterator(err) {
				err = errors.Wrap(err, "doing until iterator: until")
			}
			it.err = err
			return false
		}
		it.count++

		if isUntilPassed {
			it.err = EndOfStringIterator
		}

		it.hasNext = true
		it.next = next
		return true
	}

	return false
}

// StringEnumDoingUntil sets until checker while iterating over items.
// If untilList is empty, so all items returned as is.
func StringEnumDoingUntil(items StringIterator, untilList ...StringEnumChecker) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	var until StringEnumChecker
	if len(untilList) > 0 {
		until = EnumAllString(untilList...)
	} else {
		until = AlwaysStringEnumCheckFalse
	}
	return &EnumDoingUntilStringIterator{preparedStringItem{base: items}, until, 0}
}

// StringEnumSkipUntil sets until conditions to skip few items.
func StringEnumSkipUntil(items StringIterator, untilList ...StringEnumChecker) error {
	// no error wrapping since no additional context for the error; just return it.
	return StringDiscard(StringEnumDoingUntil(items, untilList...))
}

// StringGettingBatch returns the next batch from items.
func StringGettingBatch(items StringIterator, batchSize int) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	if batchSize == 0 {
		return items
	}

	return StringEnumDoingUntil(items, StringEnumCheck(func(n int, item string) (bool, error) {
		return n == batchSize-1, nil
	}))
}

// StringConverter is an object converting an item type of string.
type StringConverter interface {
	// Convert should convert an item type of string into another item of string.
	// It is suggested to return EndOfStringIterator to stop iteration.
	Convert(string) (string, error)
}

// StringConvert is a shortcut implementation
// of StringConverter based on a function.
type StringConvert func(string) (string, error)

// Convert converts an item type of string into another item of string.
// It is suggested to return EndOfStringIterator to stop iteration.
func (c StringConvert) Convert(item string) (string, error) { return c(item) }

// NoStringConvert does nothing with item, just returns it as is.
var NoStringConvert StringConverter = StringConvert(
	func(item string) (string, error) { return item, nil })

type doubleStringConverter struct {
	lhs, rhs StringConverter
}

func (c doubleStringConverter) Convert(item string) (string, error) {
	item, err := c.lhs.Convert(item)
	if err != nil {
		return "", errors.Wrap(err, "convert lhs")
	}
	item, err = c.rhs.Convert(item)
	if err != nil {
		return "", errors.Wrap(err, "convert rhs")
	}
	return item, nil
}

// StringConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func StringConverterSeries(converters ...StringConverter) StringConverter {
	var series = NoStringConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleStringConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingStringIterator does iteration with
// converting by previously set converter.
type ConvertingStringIterator struct {
	preparedStringItem
	converter StringConverter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedStringItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfStringIterator(err) {
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

// StringConverting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func StringConverting(items StringIterator, converters ...StringConverter) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	return &ConvertingStringIterator{
		preparedStringItem{base: items}, StringConverterSeries(converters...)}
}

// StringEnumConverter is an object converting an item type of string and its ordering number.
type StringEnumConverter interface {
	// Convert should convert an item type of string into another item of string.
	// It is suggested to return EndOfStringIterator to stop iteration.
	Convert(n int, val string) (string, error)
}

// StringEnumConvert is a shortcut implementation
// of StringEnumConverter based on a function.
type StringEnumConvert func(int, string) (string, error)

// Convert converts an item type of string into another item of string.
// It is suggested to return EndOfStringIterator to stop iteration.
func (c StringEnumConvert) Convert(n int, item string) (string, error) { return c(n, item) }

// NoStringEnumConvert does nothing with item, just returns it as is.
var NoStringEnumConvert StringEnumConverter = StringEnumConvert(
	func(_ int, item string) (string, error) { return item, nil })

type enumFromStringConverter struct {
	StringConverter
}

func (ch enumFromStringConverter) Convert(_ int, item string) (string, error) {
	return ch.StringConverter.Convert(item)
}

// EnumFromStringConverter adapts checker type of StringConverter
// to the interface StringEnumConverter.
// If converter is nil it is return based on NoStringConvert enum checker.
func EnumFromStringConverter(converter StringConverter) StringEnumConverter {
	if converter == nil {
		converter = NoStringConvert
	}
	return &enumFromStringConverter{converter}
}

type doubleStringEnumConverter struct {
	lhs, rhs StringEnumConverter
}

func (c doubleStringEnumConverter) Convert(n int, item string) (string, error) {
	item, err := c.lhs.Convert(n, item)
	if err != nil {
		return "", errors.Wrap(err, "convert lhs")
	}
	item, err = c.rhs.Convert(n, item)
	if err != nil {
		return "", errors.Wrap(err, "convert rhs")
	}
	return item, nil
}

// EnumStringConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumStringConverterSeries(converters ...StringEnumConverter) StringEnumConverter {
	var series = NoStringEnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleStringEnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingStringIterator does iteration with
// converting by previously set converter.
type EnumConvertingStringIterator struct {
	preparedStringItem
	converter StringEnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedStringItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfStringIterator(err) {
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

// StringEnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func StringEnumConverting(items StringIterator, converters ...StringEnumConverter) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	return &EnumConvertingStringIterator{
		preparedStringItem{base: items}, EnumStringConverterSeries(converters...), 0}
}

// StringHandler is an object handling an item type of string.
type StringHandler interface {
	// Handle should do something with item of string.
	// It is suggested to return EndOfStringIterator to stop iteration.
	Handle(string) error
}

// StringHandle is a shortcut implementation
// of StringHandler based on a function.
type StringHandle func(string) error

// Handle does something with item of string.
// It is suggested to return EndOfStringIterator to stop iteration.
func (h StringHandle) Handle(item string) error { return h(item) }

// StringDoNothing does nothing.
var StringDoNothing StringHandler = StringHandle(func(_ string) error { return nil })

type doubleStringHandler struct {
	lhs, rhs StringHandler
}

func (h doubleStringHandler) Handle(item string) error {
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

// StringHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func StringHandlerSeries(handlers ...StringHandler) StringHandler {
	var series = StringDoNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleStringHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// HandlingStringIterator does iteration with
// handling by previously set handler.
type HandlingStringIterator struct {
	preparedStringItem
	handler StringHandler
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *HandlingStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedStringItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(next)
		if err != nil {
			if !isEndOfStringIterator(err) {
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

// StringHandling sets handler while iterating over items.
// If handlers is empty, so it will do nothing.
func StringHandling(items StringIterator, handlers ...StringHandler) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	return &HandlingStringIterator{
		preparedStringItem{base: items}, StringHandlerSeries(handlers...)}
}

// StringRange iterates over items and use handlers to each one.
func StringRange(items StringIterator, handlers ...StringHandler) error {
	return StringDiscard(StringHandling(items, handlers...))
}

// StringRangeIterator is an iterator over items.
type StringRangeIterator interface {
	// Range should iterate over items.
	Range(...StringHandler) error
}

type sStringRangeIterator struct {
	iter StringIterator
}

// ToStringRangeIterator constructs an instance implementing StringRangeIterator
// based on StringIterator.
func ToStringRangeIterator(iter StringIterator) StringRangeIterator {
	if iter == nil {
		iter = EmptyStringIterator
	}
	return sStringRangeIterator{iter: iter}
}

// MakeStringRangeIterator constructs an instance implementing StringRangeIterator
// based on StringIterMaker.
func MakeStringRangeIterator(maker StringIterMaker) StringRangeIterator {
	if maker == nil {
		maker = MakeNoStringIter
	}
	return ToStringRangeIterator(maker.MakeIter())
}

// Range iterates over items.
func (r sStringRangeIterator) Range(handlers ...StringHandler) error {
	return StringRange(r.iter, handlers...)
}

// StringEnumHandler is an object handling an item type of string and its ordered number.
type StringEnumHandler interface {
	// Handle should do something with item of string and its ordered number.
	// It is suggested to return EndOfStringIterator to stop iteration.
	Handle(int, string) error
}

// StringEnumHandle is a shortcut implementation
// of StringEnumHandler based on a function.
type StringEnumHandle func(int, string) error

// Handle does something with item of string and its ordered number.
// It is suggested to return EndOfStringIterator to stop iteration.
func (h StringEnumHandle) Handle(n int, item string) error { return h(n, item) }

// StringDoEnumNothing does nothing.
var StringDoEnumNothing = StringEnumHandle(func(_ int, _ string) error { return nil })

type doubleStringEnumHandler struct {
	lhs, rhs StringEnumHandler
}

func (h doubleStringEnumHandler) Handle(n int, item string) error {
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

// StringEnumHandlerSeries combines all the given handlers to sequenced one
// It returns do nothing handler if the list of handlers is empty.
func StringEnumHandlerSeries(handlers ...StringEnumHandler) StringEnumHandler {
	var series StringEnumHandler = StringDoEnumNothing
	for i := len(handlers) - 1; i >= 0; i-- {
		if handlers[i] == nil {
			continue
		}
		series = doubleStringEnumHandler{lhs: handlers[i], rhs: series}
	}
	return series
}

// EnumHandlingStringIterator does iteration with
// handling by previously set handler.
type EnumHandlingStringIterator struct {
	preparedStringItem
	handler StringEnumHandler
	count   int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumHandlingStringIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedStringItem.HasNext() {
		next := it.base.Next()
		err := it.handler.Handle(it.count, next)
		if err != nil {
			if !isEndOfStringIterator(err) {
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

// StringEnumHandling sets handler while iterating over items with their serial number.
// If handlers is empty, so it will do nothing.
func StringEnumHandling(items StringIterator, handlers ...StringEnumHandler) StringIterator {
	if items == nil {
		return EmptyStringIterator
	}
	return &EnumHandlingStringIterator{
		preparedStringItem{base: items}, StringEnumHandlerSeries(handlers...), 0}
}

// StringEnum iterates over items and their ordering numbers and use handlers to each one.
func StringEnum(items StringIterator, handlers ...StringEnumHandler) error {
	return StringDiscard(StringEnumHandling(items, handlers...))
}

// StringEnumIterator is an iterator over items and their ordering numbers.
type StringEnumIterator interface {
	// Enum should iterate over items and their ordering numbers.
	Enum(...StringEnumHandler) error
}

type sStringEnumIterator struct {
	iter StringIterator
}

// ToStringEnumIterator constructs an instance implementing StringEnumIterator
// based on StringIterator.
func ToStringEnumIterator(iter StringIterator) StringEnumIterator {
	if iter == nil {
		iter = EmptyStringIterator
	}
	return sStringEnumIterator{iter: iter}
}

// MakeStringEnumIterator constructs an instance implementing StringEnumIterator
// based on StringIterMaker.
func MakeStringEnumIterator(maker StringIterMaker) StringEnumIterator {
	if maker == nil {
		maker = MakeNoStringIter
	}
	return ToStringEnumIterator(maker.MakeIter())
}

// Enum iterates over items and their ordering numbers.
func (r sStringEnumIterator) Enum(handlers ...StringEnumHandler) error {
	return StringEnum(r.iter, handlers...)
}

// Range iterates over items.
func (r sStringEnumIterator) Range(handlers ...StringHandler) error {
	return StringRange(r.iter, handlers...)
}

type doubleStringIterator struct {
	lhs, rhs StringIterator
	inRHS    bool
}

func (it *doubleStringIterator) HasNext() bool {
	if !it.inRHS {
		if it.lhs.HasNext() {
			return true
		}
		it.inRHS = true
	}
	return it.rhs.HasNext()
}

func (it *doubleStringIterator) Next() string {
	if !it.inRHS {
		return it.lhs.Next()
	}
	return it.rhs.Next()
}

func (it *doubleStringIterator) Err() error {
	if !it.inRHS {
		return it.lhs.Err()
	}
	return it.rhs.Err()
}

// SuperStringIterator combines all iterators to one.
func SuperStringIterator(itemList ...StringIterator) StringIterator {
	var super = EmptyStringIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		super = &doubleStringIterator{lhs: itemList[i], rhs: super}
	}
	return super
}

// StringComparer is a strategy to compare two types.
type StringComparer interface {
	// IsLess should be true if lhs is less than rhs.
	IsLess(lhs, rhs string) bool
}

// StringCompare is a shortcut implementation
// of StringComparer based on a function.
type StringCompare func(lhs, rhs string) bool

// IsLess is true if lhs is less than rhs.
func (c StringCompare) IsLess(lhs, rhs string) bool { return c(lhs, rhs) }

// StringAlwaysLess is an implementation of StringComparer returning always true.
var StringAlwaysLess StringComparer = StringCompare(func(_, _ string) bool { return true })

type priorityStringIterator struct {
	lhs, rhs preparedStringItem
	comparer StringComparer
}

func (it *priorityStringIterator) HasNext() bool {
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

func (it *priorityStringIterator) Next() string {
	if !it.lhs.hasNext && !it.rhs.hasNext {
		panicIfStringIteratorError(
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

func (it priorityStringIterator) Err() error {
	if err := it.lhs.Err(); err != nil {
		return err
	}
	return it.rhs.Err()
}

// PriorStringIterator compare one by one items fetched from
// all iterators and choose smallest from them to return as next.
// If comparer is nil so more left iterator is considered had smallest item.
// It is recommended to use the iterator to order already ordered iterators.
func PriorStringIterator(comparer StringComparer, itemList ...StringIterator) StringIterator {
	if comparer == nil {
		comparer = StringAlwaysLess
	}

	var prior = EmptyStringIterator
	for i := len(itemList) - 1; i >= 0; i-- {
		if itemList[i] == nil {
			continue
		}
		prior = &priorityStringIterator{
			lhs:      preparedStringItem{base: itemList[i]},
			rhs:      preparedStringItem{base: prior},
			comparer: comparer,
		}
	}

	return prior
}

// StringSliceIterator is an iterator based on a slice of string.
type StringSliceIterator struct {
	slice []string
	cur   int
}

// NewStringSliceIterator returns a new instance of StringSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use StringUnroll(slice).MakeIter() instead of to iterate over copies of item in the items.
func NewStringSliceIterator(slice []string) *StringSliceIterator {
	it := &StringSliceIterator{slice: slice}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it StringSliceIterator) HasNext() bool {
	return it.cur < len(it.slice)
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *StringSliceIterator) Next() string {
	if it.cur >= len(it.slice) {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur++
	return item
}

// Err contains first met error while Next.
func (StringSliceIterator) Err() error { return nil }

// StringSliceIterator is an iterator based on a slice of string
// and doing iteration in back direction.
type InvertingStringSliceIterator struct {
	slice []string
	cur   int
}

// NewInvertingStringSliceIterator returns a new instance of InvertingStringSliceIterator.
// Note: any changes in slice will affect correspond items in the iterator.
// Use InvertingStringSlice(StringUnroll(slice)).MakeIter() instead of to iterate over copies of item in the items.
func NewInvertingStringSliceIterator(slice []string) *InvertingStringSliceIterator {
	it := &InvertingStringSliceIterator{slice: slice, cur: len(slice) - 1}
	return it
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it InvertingStringSliceIterator) HasNext() bool {
	return it.cur >= 0
}

// Next returns next item in the iterator.
// It should be invoked after check HasNext.
func (it *InvertingStringSliceIterator) Next() string {
	if it.cur < 0 {
		panic("iterator next: pointer out of range")
	}

	item := it.slice[it.cur]
	it.cur--
	return item
}

// Err contains first met error while Next.
func (InvertingStringSliceIterator) Err() error { return nil }

// StringUnroll unrolls items to slice of string.
func StringUnroll(items StringIterator) StringSlice {
	var slice StringSlice
	panicIfStringIteratorError(StringDiscard(StringHandling(items, StringHandle(func(item string) error {
		slice = append(slice, item)
		return nil
	}))), "unroll: discard")

	return slice
}

// StringSlice is a slice of string.
type StringSlice []string

// MakeIter returns a new instance of StringIterator to iterate over it.
// It returns EmptyStringIterator if the error is not nil.
func (s StringSlice) MakeIter() StringIterator {
	return NewStringSliceIterator(s)
}

// StringSlice is a slice of string which can make inverting iterator.
type InvertingStringSlice []string

// MakeIter returns a new instance of StringIterator to iterate over it.
// It returns EmptyStringIterator if the error is not nil.
func (s InvertingStringSlice) MakeIter() StringIterator {
	return NewInvertingStringSliceIterator(s)
}

// StringInvert unrolls items and make inverting iterator based on them.
func StringInvert(items StringIterator) StringIterator {
	return InvertingStringSlice(StringUnroll(items)).MakeIter()
}

// EndOfStringIterator is an error to stop iterating over an iterator.
// It is used in some method of the package.
var EndOfStringIterator = errors.New("end of string iterator")

func isEndOfStringIterator(err error) bool {
	return errors.Is(err, EndOfStringIterator)
}

func wrapIfNotEndOfStringIterator(err error, msg string) error {
	if err == nil {
		return nil
	}
	if !isEndOfStringIterator(err) {
		err = errors.Wrap(err, msg)
	}
	return err
}

func panicIfStringIteratorError(err error, block string) {
	msg := "something went wrong"
	if len(block) > 0 {
		msg = block + ": " + msg
	}
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}

type preparedStringItem struct {
	base    StringIterator
	hasNext bool
	next    string
	err     error
}

func (it *preparedStringItem) HasNext() bool {
	it.hasNext = it.err == nil && it.base.HasNext()
	return it.hasNext
}

func (it *preparedStringItem) Next() string {
	if !it.hasNext {
		panicIfStringIteratorError(
			errors.New("no next"), "prepared: next")
	}
	next := it.next
	it.hasNext = false
	it.next = ""
	return next
}

func (it preparedStringItem) Err() error {
	if it.err != nil && !errors.Is(it.err, EndOfStringIterator) {
		return it.err
	}
	return it.base.Err()
}
