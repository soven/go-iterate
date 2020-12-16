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
	var series PrefixConverter = NoPrefixConvert
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

// PrefixConverting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func PrefixConverting(items PrefixIterator, converters ...PrefixConverter) PrefixIterator {
	if items == nil {
		return EmptyPrefixIterator
	}
	return &ConvertingPrefixIterator{
		preparedPrefixItem{base: items}, PrefixConverterSeries(converters...)}
}
