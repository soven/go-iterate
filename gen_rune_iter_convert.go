package iter

import "github.com/pkg/errors"

// RuneConverter is an object converting an item type of rune.
type RuneConverter interface {
	// Convert should convert an item type of rune into another item of rune.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Convert(rune) (rune, error)
}

// RuneConvert is a shortcut implementation
// of RuneConverter based on a function.
type RuneConvert func(rune) (rune, error)

// Convert converts an item type of rune into another item of rune.
// It is suggested to return EndOfRuneIterator to stop iteration.
func (c RuneConvert) Convert(item rune) (rune, error) { return c(item) }

// NoRuneConvert does nothing with item, just returns it as is.
var NoRuneConvert RuneConverter = RuneConvert(
	func(item rune) (rune, error) { return item, nil })

type doubleRuneConverter struct {
	lhs, rhs RuneConverter
}

func (c doubleRuneConverter) Convert(item rune) (rune, error) {
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

// RuneConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func RuneConverterSeries(converters ...RuneConverter) RuneConverter {
	var series = NoRuneConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleRuneConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingRuneIterator does iteration with
// converting by previously set converter.
type ConvertingRuneIterator struct {
	preparedRuneItem
	converter RuneConverter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneConverting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func RuneConverting(items RuneIterator, converters ...RuneConverter) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &ConvertingRuneIterator{
		preparedRuneItem{base: items}, RuneConverterSeries(converters...)}
}

// RuneEnumConverter is an object converting an item type of rune and its ordering number.
type RuneEnumConverter interface {
	// Convert should convert an item type of rune into another item of rune.
	// It is suggested to return EndOfRuneIterator to stop iteration.
	Convert(n int, val rune) (rune, error)
}

// RuneEnumConvert is a shortcut implementation
// of RuneEnumConverter based on a function.
type RuneEnumConvert func(int, rune) (rune, error)

// Convert converts an item type of rune into another item of rune.
// It is suggested to return EndOfRuneIterator to stop iteration.
func (c RuneEnumConvert) Convert(n int, item rune) (rune, error) { return c(n, item) }

// NoRuneEnumConvert does nothing with item, just returns it as is.
var NoRuneEnumConvert RuneEnumConverter = RuneEnumConvert(
	func(_ int, item rune) (rune, error) { return item, nil })

type enumFromRuneConverter struct {
	RuneConverter
}

func (ch enumFromRuneConverter) Convert(_ int, item rune) (rune, error) {
	return ch.RuneConverter.Convert(item)
}

// EnumFromRuneConverter adapts checker type of RuneConverter
// to the interface RuneEnumConverter.
// If converter is nil it is return based on NoRuneConvert enum checker.
func EnumFromRuneConverter(converter RuneConverter) RuneEnumConverter {
	if converter == nil {
		converter = NoRuneConvert
	}
	return &enumFromRuneConverter{converter}
}

type doubleRuneEnumConverter struct {
	lhs, rhs RuneEnumConverter
}

func (c doubleRuneEnumConverter) Convert(n int, item rune) (rune, error) {
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

// EnumRuneConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func EnumRuneConverterSeries(converters ...RuneEnumConverter) RuneEnumConverter {
	var series = NoRuneEnumConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleRuneEnumConverter{lhs: converters[i], rhs: series}
	}

	return series
}

// EnumConvertingRuneIterator does iteration with
// converting by previously set converter.
type EnumConvertingRuneIterator struct {
	preparedRuneItem
	converter RuneEnumConverter
	count     int
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *EnumConvertingRuneIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedRuneItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(it.count, next)
		if err != nil {
			if !isEndOfRuneIterator(err) {
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

// RuneEnumConverting sets converter while iterating over items and their ordering numbers.
// If converters is empty, so all items will not be affected.
func RuneEnumConverting(items RuneIterator, converters ...RuneEnumConverter) RuneIterator {
	if items == nil {
		return EmptyRuneIterator
	}
	return &EnumConvertingRuneIterator{
		preparedRuneItem{base: items}, EnumRuneConverterSeries(converters...), 0}
}
