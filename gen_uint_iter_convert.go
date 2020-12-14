package iter

// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
import "github.com/pkg/errors"

type UintConverter interface {
	// It is suggested to return EndOfUintIterator to stop iteration.
	Convert(uint) (uint, error)
}

type UintConvert func(uint) (uint, error)

func (c UintConvert) Convert(item uint) (uint, error) { return c(item) }

var NoUintConvert = UintConvert(func(item uint) (uint, error) { return item, nil })

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

func UintConverterSeries(converters ...UintConverter) UintConverter {
	var series UintConverter = NoUintConvert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleUintConverter{lhs: converters[i], rhs: series}
	}

	return series
}

type ConvertingUintIterator struct {
	preparedUintItem
	converter UintConverter
}

func (it *ConvertingUintIterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedUintItem.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfUintIterator(err) {
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

// UintConverting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func UintConverting(items UintIterator, converters ...UintConverter) UintIterator {
	if items == nil {
		return EmptyUintIterator
	}
	return &ConvertingUintIterator{preparedUintItem{base: items}, UintConverterSeries(converters...)}
}

func UintMap(items UintIterator, converter ...UintConverter) error {
	// no error wrapping since no additional context for the error; just return it.
	return UintDiscard(UintConverting(items, converter...))
}
