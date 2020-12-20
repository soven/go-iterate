// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

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
