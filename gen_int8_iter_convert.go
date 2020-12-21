package iter

import "github.com/pkg/errors"

// Int8Converter is an object converting an item type of int8.
type Int8Converter interface {
	// Convert should convert an item type of int8 into another item of int8.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Convert(int8) (int8, error)
}

// Int8Convert is a shortcut implementation
// of Int8Converter based on a function.
type Int8Convert func(int8) (int8, error)

// Convert converts an item type of int8 into another item of int8.
// It is suggested to return EndOfInt8Iterator to stop iteration.
func (c Int8Convert) Convert(item int8) (int8, error) { return c(item) }

// NoInt8Convert does nothing with item, just returns it as is.
var NoInt8Convert Int8Converter = Int8Convert(
	func(item int8) (int8, error) { return item, nil })

type doubleInt8Converter struct {
	lhs, rhs Int8Converter
}

func (c doubleInt8Converter) Convert(item int8) (int8, error) {
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

// Int8ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Int8ConverterSeries(converters ...Int8Converter) Int8Converter {
	var series = NoInt8Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt8Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingInt8Iterator does iteration with
// converting by previously set converter.
type ConvertingInt8Iterator struct {
	preparedInt8Item
	converter Int8Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Int8Converting(items Int8Iterator, converters ...Int8Converter) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &ConvertingInt8Iterator{
		preparedInt8Item{base: items}, Int8ConverterSeries(converters...)}
}

// Int8EnumConverter is an object converting an item type of int8 and its ordering number.
type Int8EnumConverter interface {
	// Convert should convert an item type of int8 into another item of int8.
	// It is suggested to return EndOfInt8Iterator to stop iteration.
	Convert(n int, val int8) (int8, error)
}

// Int8EnumConvert is a shortcut implementation
// of Int8EnumConverter based on a function.
type Int8EnumConvert func(int, int8) (int8, error)

// Convert converts an item type of int8 into another item of int8.
// It is suggested to return EndOfInt8Iterator to stop iteration.
func (c Int8EnumConvert) Convert(n int, item int8) (int8, error) { return c(n, item) }

// NoInt8EnumConvert does nothing with item, just returns it as is.
var NoInt8EnumConvert Int8EnumConverter = Int8EnumConvert(
	func(_ int, item int8) (int8, error) { return item, nil })

type enumFromInt8Converter struct {
	Int8Converter
}

func (ch enumFromInt8Converter) Convert(_ int, item int8) (int8, error) {
	return ch.Int8Converter.Convert(item)
}

// EnumFromInt8Converter adapts checker type of Int8Converter
// to the interface Int8EnumConverter.
// If converter is nil it is return based on NoInt8Convert enum checker.
func EnumFromInt8Converter(converter Int8Converter) Int8EnumConverter {
	if converter == nil {
		converter = NoInt8Convert
	}
	return &enumFromInt8Converter{converter}
}

type doubleInt8EnumConverter struct {
	lhs, rhs Int8EnumConverter
}

func (c doubleInt8EnumConverter) Convert(n int, item int8) (int8, error) {
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

// EnumInt8ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumInt8ConverterSeries(converters ...Int8EnumConverter) Int8EnumConverter {
	var series = NoInt8EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt8EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingInt8Iterator does iteration with
// converting by previously set converter.
type EnumConvertingInt8Iterator struct {
	preparedInt8Item
	converter Int8EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingInt8Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt8Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfInt8Iterator(err) {
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

// Int8EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Int8EnumConverting(items Int8Iterator, converters ...Int8EnumConverter) Int8Iterator {
	if items == nil {
		return EmptyInt8Iterator
	}
	return &EnumConvertingInt8Iterator{
		preparedInt8Item{base: items}, EnumInt8ConverterSeries(converters...), 0}
}
