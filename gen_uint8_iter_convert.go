// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
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
	var series Uint8Converter = NoUint8Convert
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
				err = errors.Wrap(err, "filtering iterator: check")
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
