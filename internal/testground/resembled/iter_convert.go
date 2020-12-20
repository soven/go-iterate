package resembled

import "github.com/pkg/errors"

// PrefixConverter is an object converting an item type of Type.
type PrefixConverter interface {
	// Convert should convert an item type of Type into another item of Type.
	// It is suggested to return EndOfPrefixIterator to stop iteration.
	Convert(Type) (Type, error)
}

// PrefixConvert is a shortcut implementation
// of PrefixConverter based on a function.
type PrefixConvert func(Type) (Type, error)

// Convert converts an item type of Type into another item of Type.
// It is suggested to return EndOfPrefixIterator to stop iteration.
func (c PrefixConvert) Convert(item Type) (Type, error) { return c(item) }

// NoPrefixConvert does nothing with item, just returns it as is.
var NoPrefixConvert PrefixConverter = PrefixConvert(
	func(item Type) (Type, error) { return item, nil })

type doublePrefixConverter struct {
	lhs, rhs PrefixConverter
}

func (c doublePrefixConverter) Convert(item Type) (Type, error) {
	item, err := c.lhs.Convert(item)
	if err != nil {
		return Zero, errors.Wrap(err, "convert lhs")
	}
	item, err = c.rhs.Convert(item)
	if err != nil {
		return Zero, errors.Wrap(err, "convert rhs")
	}
	return item, nil
}

// PrefixConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func PrefixConverterSeries(converters ...PrefixConverter) PrefixConverter {
	var series = NoPrefixConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doublePrefixConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingPrefixIterator does iteration with
// converting by previously set converter.
type ConvertingPrefixIterator struct {
	preparedPrefixItem
	converter PrefixConverter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingPrefixIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedPrefixItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfPrefixIterator(err) {
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

// PrefixConverting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func PrefixConverting(items PrefixIterator, converters ...PrefixConverter) PrefixIterator {
	if items == nil {
		return EmptyPrefixIterator
	}
	return &ConvertingPrefixIterator{
		preparedPrefixItem{base: items}, PrefixConverterSeries(converters...)}
}

// PrefixEnumConverter is an object converting an item type of Type and its ordering number.
type PrefixEnumConverter interface {
	// Convert should convert an item type of Type into another item of Type.
	// It is suggested to return EndOfPrefixIterator to stop iteration.
	Convert(n int, val Type) (Type, error)
}

// PrefixEnumConvert is a shortcut implementation
// of PrefixEnumConverter based on a function.
type PrefixEnumConvert func(int, Type) (Type, error)

// Convert converts an item type of Type into another item of Type.
// It is suggested to return EndOfPrefixIterator to stop iteration.
func (c PrefixEnumConvert) Convert(n int, item Type) (Type, error) { return c(n, item) }

// NoPrefixEnumConvert does nothing with item, just returns it as is.
var NoPrefixEnumConvert PrefixEnumConverter = PrefixEnumConvert(
	func(_ int, item Type) (Type, error) { return item, nil })

type enumFromPrefixConverter struct {
	PrefixConverter
}

func (ch enumFromPrefixConverter) Convert(_ int, item Type) (Type, error) {
	return ch.PrefixConverter.Convert(item)
}

// EnumFromPrefixConverter adapts checker type of PrefixConverter
// to the interface PrefixEnumConverter.
// If converter is nil it is return based on NoPrefixConvert enum checker.
func EnumFromPrefixConverter(converter PrefixConverter) PrefixEnumConverter {
	if converter == nil {
		converter = NoPrefixConvert
	}
	return &enumFromPrefixConverter{converter}
}

type doublePrefixEnumConverter struct {
	lhs, rhs PrefixEnumConverter
}

func (c doublePrefixEnumConverter) Convert(n int, item Type) (Type, error) {
	item, err := c.lhs.Convert(n, item)
	if err != nil {
		return Zero, errors.Wrap(err, "convert lhs")
	}
	item, err = c.rhs.Convert(n, item)
	if err != nil {
		return Zero, errors.Wrap(err, "convert rhs")
	}
	return item, nil
}

// EnumPrefixConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumPrefixConverterSeries(converters ...PrefixEnumConverter) PrefixEnumConverter {
	var series = NoPrefixEnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doublePrefixEnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingPrefixIterator does iteration with
// converting by previously set converter.
type EnumConvertingPrefixIterator struct {
	preparedPrefixItem
	converter PrefixEnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingPrefixIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedPrefixItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfPrefixIterator(err) {
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

// PrefixEnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func PrefixEnumConverting(items PrefixIterator, converters ...PrefixEnumConverter) PrefixIterator {
	if items == nil {
		return EmptyPrefixIterator
	}
	return &EnumConvertingPrefixIterator{
		preparedPrefixItem{base: items}, EnumPrefixConverterSeries(converters...), 0}
}
