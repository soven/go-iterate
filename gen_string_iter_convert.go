package iter

import "github.com/pkg/errors"

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
