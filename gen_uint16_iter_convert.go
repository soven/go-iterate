// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

// Uint16Converter is an object converting an item type of uint16.
type Uint16Converter interface {
	// Convert should convert an item type of uint16 into another item of uint16.
	// It is suggested to return EndOfUint16Iterator to stop iteration.
	Convert(uint16) (uint16, error)
}

// Uint16Convert is a shortcut implementation
// of Uint16Converter based on a function.
type Uint16Convert func(uint16) (uint16, error)

// Convert converts an item type of uint16 into another item of uint16.
// It is suggested to return EndOfUint16Iterator to stop iteration.
func (c Uint16Convert) Convert(item uint16) (uint16, error) { return c(item) }

// NoUint16Convert does nothing with item, just returns it as is.
var NoUint16Convert Uint16Converter = Uint16Convert(
	func(item uint16) (uint16, error) { return item, nil })

type doubleUint16Converter struct {
	lhs, rhs Uint16Converter
}

func (c doubleUint16Converter) Convert(item uint16) (uint16, error) {
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

// Uint16ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Uint16ConverterSeries(converters ...Uint16Converter) Uint16Converter {
	var series = NoUint16Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint16Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingUint16Iterator does iteration with
// converting by previously set converter.
type ConvertingUint16Iterator struct {
	preparedUint16Item
	converter Uint16Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
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

// Uint16Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Uint16Converting(items Uint16Iterator, converters ...Uint16Converter) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	return &ConvertingUint16Iterator{
		preparedUint16Item{base: items}, Uint16ConverterSeries(converters...)}
}

// Uint16EnumConverter is an object converting an item type of uint16 and its ordering number.
type Uint16EnumConverter interface {
	// Convert should convert an item type of uint16 into another item of uint16.
	// It is suggested to return EndOfUint16Iterator to stop iteration.
	Convert(n int, val uint16) (uint16, error)
}

// Uint16EnumConvert is a shortcut implementation
// of Uint16EnumConverter based on a function.
type Uint16EnumConvert func(int, uint16) (uint16, error)

// Convert converts an item type of uint16 into another item of uint16.
// It is suggested to return EndOfUint16Iterator to stop iteration.
func (c Uint16EnumConvert) Convert(n int, item uint16) (uint16, error) { return c(n, item) }

// NoUint16EnumConvert does nothing with item, just returns it as is.
var NoUint16EnumConvert Uint16EnumConverter = Uint16EnumConvert(
	func(_ int, item uint16) (uint16, error) { return item, nil })

type enumFromUint16Converter struct {
	Uint16Converter
}

func (ch enumFromUint16Converter) Convert(_ int, item uint16) (uint16, error) {
	return ch.Uint16Converter.Convert(item)
}

// EnumFromUint16Converter adapts checker type of Uint16Converter
// to the interface Uint16EnumConverter.
// If converter is nil it is return based on NoUint16Convert enum checker.
func EnumFromUint16Converter(converter Uint16Converter) Uint16EnumConverter {
	if converter == nil {
		converter = NoUint16Convert
	}
	return &enumFromUint16Converter{converter}
}

type doubleUint16EnumConverter struct {
	lhs, rhs Uint16EnumConverter
}

func (c doubleUint16EnumConverter) Convert(n int, item uint16) (uint16, error) {
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

// EnumUint16ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumUint16ConverterSeries(converters ...Uint16EnumConverter) Uint16EnumConverter {
	var series = NoUint16EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint16EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingUint16Iterator does iteration with
// converting by previously set converter.
type EnumConvertingUint16Iterator struct {
	preparedUint16Item
	converter Uint16EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingUint16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint16Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfUint16Iterator(err) {
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

// Uint16EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Uint16EnumConverting(items Uint16Iterator, converters ...Uint16EnumConverter) Uint16Iterator {
	if items == nil {
		return EmptyUint16Iterator
	}
	return &EnumConvertingUint16Iterator{
		preparedUint16Item{base: items}, EnumUint16ConverterSeries(converters...), 0}
}
