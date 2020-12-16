// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

// Uint32Converter is an object converting an item type of uint32.
type Uint32Converter interface {
	// Convert should convert an item type of uint32 into another item of uint32.
	// It is suggested to return EndOfUint32Iterator to stop iteration.
	Convert(uint32) (uint32, error)
}

// Uint32Convert is a shortcut implementation
// of Uint32Converter based on a function.
type Uint32Convert func(uint32) (uint32, error)

// Convert converts an item type of uint32 into another item of uint32.
// It is suggested to return EndOfUint32Iterator to stop iteration.
func (c Uint32Convert) Convert(item uint32) (uint32, error) { return c(item) }

// NoUint32Convert does nothing with item, just returns it as is.
var NoUint32Convert Uint32Converter = Uint32Convert(
	func(item uint32) (uint32, error) { return item, nil })

type doubleUint32Converter struct {
	lhs, rhs Uint32Converter
}

func (c doubleUint32Converter) Convert(item uint32) (uint32, error) {
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

// Uint32ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Uint32ConverterSeries(converters ...Uint32Converter) Uint32Converter {
	var series Uint32Converter = NoUint32Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint32Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingUint32Iterator does iteration with
// converting by previously set converter.
type ConvertingUint32Iterator struct {
	preparedUint32Item
	converter Uint32Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingUint32Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint32Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfUint32Iterator(err) {
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

// Uint32Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Uint32Converting(items Uint32Iterator, converters ...Uint32Converter) Uint32Iterator {
	if items == nil {
		return EmptyUint32Iterator
	}
	return &ConvertingUint32Iterator{
		preparedUint32Item{base: items}, Uint32ConverterSeries(converters...)}
}
