package iter

import "github.com/pkg/errors"

// UintConverter is an object converting an item type of uint.
type UintConverter interface {
	// Convert should convert an item type of uint into another item of uint.
	// It is suggested to return EndOfUintIterator to stop iteration.
	Convert(uint) (uint, error)
}

// UintConvert is a shortcut implementation
// of UintConverter based on a function.
type UintConvert func(uint) (uint, error)

// Convert converts an item type of uint into another item of uint.
// It is suggested to return EndOfUintIterator to stop iteration.
func (c UintConvert) Convert(item uint) (uint, error) { return c(item) }

// NoUintConvert does nothing with item, just returns it as is.
var NoUintConvert UintConverter = UintConvert(
	func(item uint) (uint, error) { return item, nil })

type doubleUintConverter struct {
	lhs, rhs UintConverter
}

func (c doubleUintConverter) Convert(item uint) (uint, error) {
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

// UintConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func UintConverterSeries(converters ...UintConverter) UintConverter {
	var series = NoUintConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUintConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingUintIterator does iteration with
// converting by previously set converter.
type ConvertingUintIterator struct {
	preparedUintItem
	converter UintConverter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUintItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfUintIterator(err) {
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

// UintConverting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func UintConverting(items UintIterator, converters ...UintConverter) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	return &ConvertingUintIterator{
		preparedUintItem{base: items}, UintConverterSeries(converters...)}
}

// UintEnumConverter is an object converting an item type of uint and its ordering number.
type UintEnumConverter interface {
	// Convert should convert an item type of uint into another item of uint.
	// It is suggested to return EndOfUintIterator to stop iteration.
	Convert(n int, val uint) (uint, error)
}

// UintEnumConvert is a shortcut implementation
// of UintEnumConverter based on a function.
type UintEnumConvert func(int, uint) (uint, error)

// Convert converts an item type of uint into another item of uint.
// It is suggested to return EndOfUintIterator to stop iteration.
func (c UintEnumConvert) Convert(n int, item uint) (uint, error) { return c(n, item) }

// NoUintEnumConvert does nothing with item, just returns it as is.
var NoUintEnumConvert UintEnumConverter = UintEnumConvert(
	func(_ int, item uint) (uint, error) { return item, nil })

type enumFromUintConverter struct {
	UintConverter
}

func (ch enumFromUintConverter) Convert(_ int, item uint) (uint, error) {
	return ch.UintConverter.Convert(item)
}

// EnumFromUintConverter adapts checker type of UintConverter
// to the interface UintEnumConverter.
// If converter is nil it is return based on NoUintConvert enum checker.
func EnumFromUintConverter(converter UintConverter) UintEnumConverter {
	if converter == nil {
		converter = NoUintConvert
	}
	return &enumFromUintConverter{converter}
}

type doubleUintEnumConverter struct {
	lhs, rhs UintEnumConverter
}

func (c doubleUintEnumConverter) Convert(n int, item uint) (uint, error) {
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

// EnumUintConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumUintConverterSeries(converters ...UintEnumConverter) UintEnumConverter {
	var series = NoUintEnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUintEnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingUintIterator does iteration with
// converting by previously set converter.
type EnumConvertingUintIterator struct {
	preparedUintItem
	converter UintEnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUintItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfUintIterator(err) {
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

// UintEnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func UintEnumConverting(items UintIterator, converters ...UintEnumConverter) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	return &EnumConvertingUintIterator{
		preparedUintItem{base: items}, EnumUintConverterSeries(converters...), 0}
}
