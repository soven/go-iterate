// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
package iter

import "github.com/pkg/errors"

// Uint64Converter is an object converting an item type of uint64.
type Uint64Converter interface {
	// Convert should convert an item type of uint64 into another item of uint64.
	// It is suggested to return EndOfUint64Iterator to stop iteration.
	Convert(uint64) (uint64, error)
}

// Uint64Convert is a shortcut implementation
// of Uint64Converter based on a function.
type Uint64Convert func(uint64) (uint64, error)

// Convert converts an item type of uint64 into another item of uint64.
// It is suggested to return EndOfUint64Iterator to stop iteration.
func (c Uint64Convert) Convert(item uint64) (uint64, error) { return c(item) }

// NoUint64Convert does nothing with item, just returns it as is.
var NoUint64Convert Uint64Converter = Uint64Convert(
	func(item uint64) (uint64, error) { return item, nil })

type doubleUint64Converter struct {
	lhs, rhs Uint64Converter
}

func (c doubleUint64Converter) Convert(item uint64) (uint64, error) {
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

// Uint64ConverterSeries combines all the given converters to sequenced one
// It returns no converter if the list of converters is empty.
func Uint64ConverterSeries(converters ...Uint64Converter) Uint64Converter {
	var series Uint64Converter = NoUint64Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUint64Converter{lhs: converters[i], rhs: series}
	}

	return series
}

// ConvertingUint64Iterator does iteration with
// converting by previously set converter.
type ConvertingUint64Iterator struct {
	preparedUint64Item
	converter Uint64Converter
}

// HasNext checks if there is the next item
// in the iterator. HasNext is idempotent.
func (it *ConvertingUint64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUint64Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfUint64Iterator(err) {
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

// Uint64Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Uint64Converting(items Uint64Iterator, converters ...Uint64Converter) Uint64Iterator {
	if items == nil {
		return EmptyUint64Iterator
	}
	return &ConvertingUint64Iterator{
		preparedUint64Item{base: items}, Uint64ConverterSeries(converters...)}
}
