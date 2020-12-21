package iter

import "github.com/pkg/errors"

// Uint8Converter is an object converting an item type of uint8.
type Uint8Converter interface {
	// Convert should convert an item type of uint8 into another item of uint8.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Convert(uint8) (uint8, error)
}

// Uint8Convert is a shortcut implementation
// of Uint8Converter based on a function.
type Uint8Convert func(uint8) (uint8, error)

// Convert converts an item type of uint8 into another item of uint8.
// It is suggested to return EndOfUint8Iterator to stop iteration.
func (c Uint8Convert) Convert(item uint8) (uint8, error) { return c(item) }

// NoUint8Convert does nothing with item, just returns it as is.
var NoUint8Convert Uint8Converter = Uint8Convert(
	func(item uint8) (uint8, error) { return item, nil })

type doubleUint8Converter struct {
	lhs, rhs Uint8Converter
}

func (c doubleUint8Converter) Convert(item uint8) (uint8, error) {
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

// Uint8ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Uint8ConverterSeries(converters ...Uint8Converter) Uint8Converter {
	var series = NoUint8Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint8Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingUint8Iterator does iteration with
// converting by previously set converter.
type ConvertingUint8Iterator struct {
	preparedUint8Item
	converter Uint8Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
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

// Uint8Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Uint8Converting(items Uint8Iterator, converters ...Uint8Converter) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &ConvertingUint8Iterator{
		preparedUint8Item{base: items}, Uint8ConverterSeries(converters...)}
}

// Uint8EnumConverter is an object converting an item type of uint8 and its ordering number.
type Uint8EnumConverter interface {
	// Convert should convert an item type of uint8 into another item of uint8.
	// It is suggested to return EndOfUint8Iterator to stop iteration.
	Convert(n int, val uint8) (uint8, error)
}

// Uint8EnumConvert is a shortcut implementation
// of Uint8EnumConverter based on a function.
type Uint8EnumConvert func(int, uint8) (uint8, error)

// Convert converts an item type of uint8 into another item of uint8.
// It is suggested to return EndOfUint8Iterator to stop iteration.
func (c Uint8EnumConvert) Convert(n int, item uint8) (uint8, error) { return c(n, item) }

// NoUint8EnumConvert does nothing with item, just returns it as is.
var NoUint8EnumConvert Uint8EnumConverter = Uint8EnumConvert(
	func(_ int, item uint8) (uint8, error) { return item, nil })

type enumFromUint8Converter struct {
	Uint8Converter
}

func (ch enumFromUint8Converter) Convert(_ int, item uint8) (uint8, error) {
	return ch.Uint8Converter.Convert(item)
}

// EnumFromUint8Converter adapts checker type of Uint8Converter
// to the interface Uint8EnumConverter.
// If converter is nil it is return based on NoUint8Convert enum checker.
func EnumFromUint8Converter(converter Uint8Converter) Uint8EnumConverter {
	if converter == nil {
		converter = NoUint8Convert
	}
	return &enumFromUint8Converter{converter}
}

type doubleUint8EnumConverter struct {
	lhs, rhs Uint8EnumConverter
}

func (c doubleUint8EnumConverter) Convert(n int, item uint8) (uint8, error) {
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

// EnumUint8ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumUint8ConverterSeries(converters ...Uint8EnumConverter) Uint8EnumConverter {
	var series = NoUint8EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint8EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingUint8Iterator does iteration with
// converting by previously set converter.
type EnumConvertingUint8Iterator struct {
	preparedUint8Item
	converter Uint8EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingUint8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint8Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfUint8Iterator(err) {
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

// Uint8EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Uint8EnumConverting(items Uint8Iterator, converters ...Uint8EnumConverter) Uint8Iterator {
	if items == nil {
		return EmptyUint8Iterator
	}
	return &EnumConvertingUint8Iterator{
		preparedUint8Item{base: items}, EnumUint8ConverterSeries(converters...), 0}
}
