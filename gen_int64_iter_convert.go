package iter

// Code generated by github.com/soven/go-iterate. DO NOT EDIT.
import "github.com/pkg/errors"

type Int64Converter interface {
	// It is suggested to return EndOfInt64Iterator to stop iteration.
	Convert(int64) (int64, error)
}

type Int64Convert func(int64) (int64, error)

func (c Int64Convert) Convert(item int64) (int64, error) { return c(item) }

var NoInt64Convert = Int64Convert(func(item int64) (int64, error) { return item, nil })

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

func Int64ConverterSeries(converters ...Int64Converter) Int64Converter {
	var series Int64Converter = NoInt64Convert
	for i := len(converters) - 1; i >= 0; i-- {
		if converters[i] == nil {
			continue
		}
		series = doubleInt64Converter{lhs: converters[i], rhs: series}
	}

	return series
}

type ConvertingInt64Iterator struct {
	preparedInt64Item
	converter Int64Converter
}

func (it *ConvertingInt64Iterator) HasNext() bool {
	if it.hasNext {
		return true
	}
	if it.preparedInt64Item.HasNext() {
		next := it.base.Next()
		next, err := it.converter.Convert(next)
		if err != nil {
			if !isEndOfInt64Iterator(err) {
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

// Int64Converting sets converter while iterating over items.
// If converters is empty, so all items will not be affected.
func Int64Converting(items Int64Iterator, converters ...Int64Converter) Int64Iterator {
	if items == nil {
		return EmptyInt64Iterator
	}
	return &ConvertingInt64Iterator{preparedInt64Item{base: items}, Int64ConverterSeries(converters...)}
}

func Int64Map(items Int64Iterator, converter ...Int64Converter) error {
	// no error wrapping since no additional context for the error; just return it.
	return Int64Discard(Int64Converting(items, converter...))
}