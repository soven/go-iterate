package iter

import "github.com/pkg/errors"

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
