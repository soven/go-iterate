package iter

import "github.com/pkg/errors"

// Int32Converter is an object converting an item type of int32.
type Int32Converter interface {
	// Convert should convert an item type of int32 into another item of int32.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Convert(int32) (int32, error)
}

// Int32Convert is a shortcut implementation
// of Int32Converter based on a function.
type Int32Convert func(int32) (int32, error)

// Convert converts an item type of int32 into another item of int32.
// It is suggested to return EndOfInt32Iterator to stop iteration.
func (c Int32Convert) Convert(item int32) (int32, error) { return c(item) }

// NoInt32Convert does nothing with item, just returns it as is.
var NoInt32Convert Int32Converter = Int32Convert(
	func(item int32) (int32, error) { return item, nil })

type doubleInt32Converter struct {
	lhs, rhs Int32Converter
}

func (c doubleInt32Converter) Convert(item int32) (int32, error) {
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

// Int32ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Int32ConverterSeries(converters ...Int32Converter) Int32Converter {
	var series = NoInt32Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt32Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingInt32Iterator does iteration with
// converting by previously set converter.
type ConvertingInt32Iterator struct {
	preparedInt32Item
	converter Int32Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
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

// Int32Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Int32Converting(items Int32Iterator, converters ...Int32Converter) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &ConvertingInt32Iterator{
		preparedInt32Item{base: items}, Int32ConverterSeries(converters...)}
}

// Int32EnumConverter is an object converting an item type of int32 and its ordering number.
type Int32EnumConverter interface {
	// Convert should convert an item type of int32 into another item of int32.
	// It is suggested to return EndOfInt32Iterator to stop iteration.
	Convert(n int, val int32) (int32, error)
}

// Int32EnumConvert is a shortcut implementation
// of Int32EnumConverter based on a function.
type Int32EnumConvert func(int, int32) (int32, error)

// Convert converts an item type of int32 into another item of int32.
// It is suggested to return EndOfInt32Iterator to stop iteration.
func (c Int32EnumConvert) Convert(n int, item int32) (int32, error) { return c(n, item) }

// NoInt32EnumConvert does nothing with item, just returns it as is.
var NoInt32EnumConvert Int32EnumConverter = Int32EnumConvert(
	func(_ int, item int32) (int32, error) { return item, nil })

type enumFromInt32Converter struct {
	Int32Converter
}

func (ch enumFromInt32Converter) Convert(_ int, item int32) (int32, error) {
	return ch.Int32Converter.Convert(item)
}

// EnumFromInt32Converter adapts checker type of Int32Converter
// to the interface Int32EnumConverter.
// If converter is nil it is return based on NoInt32Convert enum checker.
func EnumFromInt32Converter(converter Int32Converter) Int32EnumConverter {
	if converter == nil {
		converter = NoInt32Convert
	}
	return &enumFromInt32Converter{converter}
}

type doubleInt32EnumConverter struct {
	lhs, rhs Int32EnumConverter
}

func (c doubleInt32EnumConverter) Convert(n int, item int32) (int32, error) {
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

// EnumInt32ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumInt32ConverterSeries(converters ...Int32EnumConverter) Int32EnumConverter {
	var series = NoInt32EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt32EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingInt32Iterator does iteration with
// converting by previously set converter.
type EnumConvertingInt32Iterator struct {
	preparedInt32Item
	converter Int32EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingInt32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt32Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfInt32Iterator(err) {
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

// Int32EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Int32EnumConverting(items Int32Iterator, converters ...Int32EnumConverter) Int32Iterator {
	if items == nil {
		return EmptyInt32Iterator
	}
	return &EnumConvertingInt32Iterator{
		preparedInt32Item{base: items}, EnumInt32ConverterSeries(converters...), 0}
}
