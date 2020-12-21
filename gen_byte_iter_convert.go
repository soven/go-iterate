package iter

import "github.com/pkg/errors"

// ByteConverter is an object converting an item type of byte.
type ByteConverter interface {
	// Convert should convert an item type of byte into another item of byte.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Convert(byte) (byte, error)
}

// ByteConvert is a shortcut implementation
// of ByteConverter based on a function.
type ByteConvert func(byte) (byte, error)

// Convert converts an item type of byte into another item of byte.
// It is suggested to return EndOfByteIterator to stop iteration.
func (c ByteConvert) Convert(item byte) (byte, error) { return c(item) }

// NoByteConvert does nothing with item, just returns it as is.
var NoByteConvert ByteConverter = ByteConvert(
	func(item byte) (byte, error) { return item, nil })

type doubleByteConverter struct {
	lhs, rhs ByteConverter
}

func (c doubleByteConverter) Convert(item byte) (byte, error) {
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

// ByteConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func ByteConverterSeries(converters ...ByteConverter) ByteConverter {
	var series = NoByteConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleByteConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingByteIterator does iteration with
// converting by previously set converter.
type ConvertingByteIterator struct {
	preparedByteItem
	converter ByteConverter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedByteItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteConverting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func ByteConverting(items ByteIterator, converters ...ByteConverter) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &ConvertingByteIterator{
		preparedByteItem{base: items}, ByteConverterSeries(converters...)}
}

// ByteEnumConverter is an object converting an item type of byte and its ordering number.
type ByteEnumConverter interface {
	// Convert should convert an item type of byte into another item of byte.
	// It is suggested to return EndOfByteIterator to stop iteration.
	Convert(n int, val byte) (byte, error)
}

// ByteEnumConvert is a shortcut implementation
// of ByteEnumConverter based on a function.
type ByteEnumConvert func(int, byte) (byte, error)

// Convert converts an item type of byte into another item of byte.
// It is suggested to return EndOfByteIterator to stop iteration.
func (c ByteEnumConvert) Convert(n int, item byte) (byte, error) { return c(n, item) }

// NoByteEnumConvert does nothing with item, just returns it as is.
var NoByteEnumConvert ByteEnumConverter = ByteEnumConvert(
	func(_ int, item byte) (byte, error) { return item, nil })

type enumFromByteConverter struct {
	ByteConverter
}

func (ch enumFromByteConverter) Convert(_ int, item byte) (byte, error) {
	return ch.ByteConverter.Convert(item)
}

// EnumFromByteConverter adapts checker type of ByteConverter
// to the interface ByteEnumConverter.
// If converter is nil it is return based on NoByteConvert enum checker.
func EnumFromByteConverter(converter ByteConverter) ByteEnumConverter {
	if converter == nil {
		converter = NoByteConvert
	}
	return &enumFromByteConverter{converter}
}

type doubleByteEnumConverter struct {
	lhs, rhs ByteEnumConverter
}

func (c doubleByteEnumConverter) Convert(n int, item byte) (byte, error) {
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

// EnumByteConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumByteConverterSeries(converters ...ByteEnumConverter) ByteEnumConverter {
	var series = NoByteEnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleByteEnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingByteIterator does iteration with
// converting by previously set converter.
type EnumConvertingByteIterator struct {
	preparedByteItem
	converter ByteEnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingByteIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedByteItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfByteIterator(err) {
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

// ByteEnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func ByteEnumConverting(items ByteIterator, converters ...ByteEnumConverter) ByteIterator {
	if items == nil {
		return EmptyByteIterator
	}
	return &EnumConvertingByteIterator{
		preparedByteItem{base: items}, EnumByteConverterSeries(converters...), 0}
}
