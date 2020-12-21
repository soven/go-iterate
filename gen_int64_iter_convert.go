package iter

import "github.com/pkg/errors"

// Int64Converter is an object converting an item type of int64.
type Int64Converter interface {
	// Convert should convert an item type of int64 into another item of int64.
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Convert(int64) (int64, error)
}

// Int64Convert is a shortcut implementation
// of Int64Converter based on a function.
type Int64Convert func(int64) (int64, error)

// Convert converts an item type of int64 into another item of int64.
// It is suggested to return EndOfInt64Iterator to stop iteration.
func (c Int64Convert) Convert(item int64) (int64, error) { return c(item) }

// NoInt64Convert does nothing with item, just returns it as is.
var NoInt64Convert Int64Converter = Int64Convert(
	func(item int64) (int64, error) { return item, nil })

type doubleInt64Converter struct {
	lhs, rhs Int64Converter
}

func (c doubleInt64Converter) Convert(item int64) (int64, error) {
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

// Int64ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Int64ConverterSeries(converters ...Int64Converter) Int64Converter {
	var series = NoInt64Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt64Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingInt64Iterator does iteration with
// converting by previously set converter.
type ConvertingInt64Iterator struct {
	preparedInt64Item
	converter Int64Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Int64Converting(items Int64Iterator, converters ...Int64Converter) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &ConvertingInt64Iterator{
		preparedInt64Item{base: items}, Int64ConverterSeries(converters...)}
}

// Int64EnumConverter is an object converting an item type of int64 and its ordering number.
type Int64EnumConverter interface {
	// Convert should convert an item type of int64 into another item of int64.
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Convert(n int, val int64) (int64, error)
}

// Int64EnumConvert is a shortcut implementation
// of Int64EnumConverter based on a function.
type Int64EnumConvert func(int, int64) (int64, error)

// Convert converts an item type of int64 into another item of int64.
// It is suggested to return EndOfInt64Iterator to stop iteration.
func (c Int64EnumConvert) Convert(n int, item int64) (int64, error) { return c(n, item) }

// NoInt64EnumConvert does nothing with item, just returns it as is.
var NoInt64EnumConvert Int64EnumConverter = Int64EnumConvert(
	func(_ int, item int64) (int64, error) { return item, nil })

type enumFromInt64Converter struct {
	Int64Converter
}

func (ch enumFromInt64Converter) Convert(_ int, item int64) (int64, error) {
	return ch.Int64Converter.Convert(item)
}

// EnumFromInt64Converter adapts checker type of Int64Converter
// to the interface Int64EnumConverter.
// If converter is nil it is return based on NoInt64Convert enum checker.
func EnumFromInt64Converter(converter Int64Converter) Int64EnumConverter {
	if converter == nil {
		converter = NoInt64Convert
	}
	return &enumFromInt64Converter{converter}
}

type doubleInt64EnumConverter struct {
	lhs, rhs Int64EnumConverter
}

func (c doubleInt64EnumConverter) Convert(n int, item int64) (int64, error) {
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

// EnumInt64ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumInt64ConverterSeries(converters ...Int64EnumConverter) Int64EnumConverter {
	var series = NoInt64EnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt64EnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingInt64Iterator does iteration with
// converting by previously set converter.
type EnumConvertingInt64Iterator struct {
	preparedInt64Item
	converter Int64EnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64EnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func Int64EnumConverting(items Int64Iterator, converters ...Int64EnumConverter) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &EnumConvertingInt64Iterator{
		preparedInt64Item{base: items}, EnumInt64ConverterSeries(converters...), 0}
}
