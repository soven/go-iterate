package iter

import "github.com/pkg/errors"

// Int16Converter is an object converting an item type of int16.
type Int16Converter interface {
	// Convert should convert an item type of int16 into another item of int16.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Convert(int16) (int16, error)
}

// Int16Convert is a shortcut implementation
// of Int16Converter based on a function.
type Int16Convert func(int16) (int16, error)

// Convert converts an item type of int16 into another item of int16.
// It is suggested to return EndOfInt16Iterator to stop iteration.
func (c Int16Convert) Convert(item int16) (int16, error) { return c(item) }

// NoInt16Convert does nothing with item, just returns it as is.
var NoInt16Convert Int16Converter = Int16Convert(
	func(item int16) (int16, error) { return item, nil })

type doubleInt16Converter struct {
	lhs, rhs Int16Converter
}

func (c doubleInt16Converter) Convert(item int16) (int16, error) {
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

// Int16ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Int16ConverterSeries(converters ...Int16Converter) Int16Converter {
	var series = NoInt16Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt16Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingInt16Iterator does iteration with
// converting by previously set converter.
type ConvertingInt16Iterator struct {
	preparedInt16Item
	converter Int16Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Int16Converting(items Int16Iterator, converters ...Int16Converter) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &ConvertingInt16Iterator{
		preparedInt16Item{base: items}, Int16ConverterSeries(converters...)}
}

// Int16EnumConverter is an object converting an item type of int16 and its ordering number.
type Int16EnumConverter interface {
	// Convert should convert an item type of int16 into another item of int16.
	// It is suggested to return EndOfInt16Iterator to stop iteration.
	Convert(n int, val int16) (int16, error)
}

// Int16EnumConvert is a shortcut implementation
// of Int16EnumConverter based on a function.
type Int16EnumConvert func(int, int16) (int16, error)

// Convert converts an item type of int16 into another item of int16.
// It is suggested to return EndOfInt16Iterator to stop iteration.
func (c Int16EnumConvert) Convert(n int, item int16) (int16, error) { return c(n, item) }

// NoInt16EnumConvert does nothing with item, just returns it as is.
var NoInt16EnumConvert Int16EnumConverter = Int16EnumConvert(
	func(_ int, item int16) (int16, error) { return item, nil })

type enumFromInt16Converter struct {
	Int16Converter
}

func (ch enumFromInt16Converter) Convert(_ int, item int16) (int16, error) {
	return ch.Int16Converter.Convert(item)
}

// EnumFromInt16Converter adapts checker type of Int16Converter
// to the interface Int16EnumConverter.
// If converter is nil it is return based on NoInt16Convert enum checker.
func EnumFromInt16Converter(converter Int16Converter) Int16EnumConverter {
	if converter == nil {
		converter = NoInt16Convert
	}
	return &enumFromInt16Converter{converter}
}

type doubleInt16EnumConverter struct {
	lhs, rhs Int16EnumConverter
}

func (c doubleInt16EnumConverter) Convert(n int, item int16) (int16, error) {
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

// EnumInt16ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumInt16ConverterSeries(converters ...Int16EnumConverter) Int16EnumConverter {
	var series = NoInt16EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt16EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingInt16Iterator does iteration with
// converting by previously set converter.
type EnumConvertingInt16Iterator struct {
	preparedInt16Item
	converter Int16EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingInt16Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt16Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfInt16Iterator(err) {
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

// Int16EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Int16EnumConverting(items Int16Iterator, converters ...Int16EnumConverter) Int16Iterator {
	if items == nil {
		return EmptyInt16Iterator
	}
	return &EnumConvertingInt16Iterator{
		preparedInt16Item{base: items}, EnumInt16ConverterSeries(converters...), 0}
}
