package resembled_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/soven/go-iterate/internal/testground/resembled"
	"github.com/soven/go-iterate/internal/testground/resembled/mocks"
)

func Test_PrefixConverterSeries(t *testing.T) {
	tests := []struct {
		name       string
		val        resembled.Type
		converters []resembled.PrefixConverter
		want       resembled.Type
	}{
		{
			name: "general",
			val:  21,
			converters: []resembled.PrefixConverter{
				resembled.PrefixConvert(func(val resembled.Type) (resembled.Type, error) {
					return val.(int) * 2, nil
				}),
				resembled.PrefixConvert(func(val resembled.Type) (resembled.Type, error) {
					return val.(int) / 3, nil
				}),
			},
			want: 14,
		},
		{
			name:       "no converters",
			val:        7,
			converters: nil,
			want:       7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := resembled.PrefixConverterSeries(tt.converters...).Convert(tt.val)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, actual)
		})
	}
}

func mockPrefixConverter(val resembled.Type, ret ...interface{}) resembled.PrefixConverter {
	m := &mocks.PrefixConverter{}
	m.
		On("Convert", val).Return(ret...)
	return m
}

func Test_PrefixConverting(t *testing.T) {
	tests := []struct {
		name       string
		iter       resembled.PrefixIterator
		converters []resembled.PrefixConverter
		want       []resembled.Type
	}{
		{
			name: "general",
			iter: mockPrefixIterator(1, 9, 3),
			converters: []resembled.PrefixConverter{
				resembled.PrefixConvert(func(val resembled.Type) (resembled.Type, error) {
					return val.(int) + 2, nil
				}),
				resembled.PrefixConvert(func(val resembled.Type) (resembled.Type, error) {
					return val.(int) * 3, nil
				}),
			},
			want: []resembled.Type{9, 33, 15},
		},
		{
			name:       "no converters",
			iter:       mockPrefixIterator(1, 2, 3),
			converters: nil,
			want:       []resembled.Type{1, 2, 3},
		},
		{
			name:       "no iters",
			iter:       resembled.EmptyPrefixIterator,
			converters: []resembled.PrefixConverter{mockPrefixConverter(resembled.Zero, true, nil)},
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := resembled.PrefixConverting(tt.iter, tt.converters...)
			for _, item := range tt.want {
				if !assert.True(t, items.HasNext()) {
					return
				}
				if !assert.Equal(t, item, items.Next()) {
					return
				}
			}
			if !assert.False(t, items.HasNext()) {
				return
			}
			assert.NoError(t, items.Err())
		})
	}
}

func Test_EnumFromPrefixConverter(t *testing.T) {
	tests := []struct {
		name      string
		val       resembled.Type
		converter resembled.PrefixConverter
		want      resembled.Type
	}{
		{
			name:      "general",
			val:       7,
			converter: mockPrefixConverter(7, 14, nil),
			want:      14,
		},
		{
			name:      "no converter",
			val:       7,
			converter: nil,
			want:      7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convert, err := resembled.EnumFromPrefixConverter(tt.converter).Convert(5, tt.val)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, convert)
		})
	}
}

func Test_EnumPrefixConverterSeries(t *testing.T) {
	tests := []struct {
		name       string
		n          int
		val        resembled.Type
		converters []resembled.PrefixEnumConverter
		want       resembled.Type
	}{
		{
			name: "general",
			n:    2,
			val:  3,
			converters: []resembled.PrefixEnumConverter{
				resembled.PrefixEnumConvert(func(n int, val resembled.Type) (resembled.Type, error) {
					return val.(int) * n, nil
				}),
				resembled.PrefixEnumConvert(func(n int, val resembled.Type) (resembled.Type, error) {
					return val.(int) + n, nil
				}),
			},
			want: 8,
		},
		{
			name:       "no converters",
			val:        7,
			converters: nil,
			want:       7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := resembled.EnumPrefixConverterSeries(tt.converters...).Convert(tt.n, tt.val)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, actual)
		})
	}
}

func mockPrefixEnumConverter(n int, val resembled.Type, ret ...interface{}) resembled.PrefixEnumConverter {
	m := &mocks.PrefixEnumConverter{}
	m.
		On("Convert", n, val).Return(ret...)
	return m
}

func Test_PrefixEnumConverting(t *testing.T) {
	tests := []struct {
		name       string
		iter       resembled.PrefixIterator
		converters []resembled.PrefixEnumConverter
		want       []resembled.Type
	}{
		{
			name: "general",
			iter: mockPrefixIterator(1, 9, 3),
			converters: []resembled.PrefixEnumConverter{
				resembled.PrefixEnumConvert(func(n int, val resembled.Type) (resembled.Type, error) {
					return val.(int) + n, nil
				}),
				resembled.PrefixEnumConvert(func(n int, val resembled.Type) (resembled.Type, error) {
					return val.(int) * n, nil
				}),
			},
			want: []resembled.Type{0, 10, 10}, // (1+0)*0, (9+1)*1, (3+2)*2
		},
		{
			name:       "no converters",
			iter:       mockPrefixIterator(1, 2, 3),
			converters: nil,
			want:       []resembled.Type{1, 2, 3},
		},
		{
			name:       "no iters",
			iter:       resembled.EmptyPrefixIterator,
			converters: []resembled.PrefixEnumConverter{mockPrefixEnumConverter(0, resembled.Zero, true, nil)},
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := resembled.PrefixEnumConverting(tt.iter, tt.converters...)
			for _, item := range tt.want {
				if !assert.True(t, items.HasNext()) {
					return
				}
				if !assert.Equal(t, item, items.Next()) {
					return
				}
			}
			if !assert.False(t, items.HasNext()) {
				return
			}
			assert.NoError(t, items.Err())
		})
	}
}
