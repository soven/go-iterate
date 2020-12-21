package resembled_test

import (
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/soven/go-iterate/internal/testground/resembled"
	"github.com/soven/go-iterate/internal/testground/resembled/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_AlwaysPrefixCheckTrue(t *testing.T) {
	tests := []struct {
		name string
		val  resembled.Type
	}{
		{"some val", 5},
		{"zero val", resembled.Zero},
		{"nil val", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := resembled.AlwaysPrefixCheckTrue.Check(tt.val)
			if !assert.NoError(t, err) {
				return
			}
			assert.True(t, check)
		})
	}
}

func Test_AlwaysPrefixCheckFalse(t *testing.T) {
	tests := []struct {
		name string
		val  resembled.Type
	}{
		{"some val", 5},
		{"zero val", resembled.Zero},
		{"nil val", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := resembled.AlwaysPrefixCheckFalse.Check(tt.val)
			if !assert.NoError(t, err) {
				return
			}
			assert.False(t, check)
		})
	}
}

func mockPrefixChecker(val interface{}, ret ...interface{}) resembled.PrefixChecker {
	m := &mocks.PrefixChecker{}
	m.
		On("Check", val).Return(ret...)
	return m
}

func Test_NotPrefix(t *testing.T) {
	tests := []struct {
		name    string
		checker resembled.PrefixChecker
	}{
		{
			name:    "general",
			checker: mockPrefixChecker(resembled.Zero, true, nil),
		},
		{
			name:    "negative",
			checker: mockPrefixChecker(resembled.Zero, false, nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := tt.checker.Check(resembled.Zero)
			if !assert.NoError(t, err) {
				return
			}
			notCheck, err := resembled.NotPrefix(tt.checker).Check(resembled.Zero)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, check, !notCheck)
		})
	}
}

func Test_AllPrefix(t *testing.T) {
	tests := []struct {
		name     string
		checkers []resembled.PrefixChecker
		want     bool
	}{
		{
			name: "FFF = F",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, false, nil),
				mockPrefixChecker(resembled.Zero, false, nil),
				mockPrefixChecker(resembled.Zero, false, nil),
			},
			want: false,
		},
		{
			name: "FFT = F",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, false, nil),
				mockPrefixChecker(resembled.Zero, false, nil),
				mockPrefixChecker(resembled.Zero, true, nil),
			},
			want: false,
		},
		{
			name: "FTF = F",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, false, nil),
				mockPrefixChecker(resembled.Zero, true, nil),
				mockPrefixChecker(resembled.Zero, false, nil),
			},
			want: false,
		},
		{
			name: "FTT = F",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, false, nil),
				mockPrefixChecker(resembled.Zero, true, nil),
				mockPrefixChecker(resembled.Zero, true, nil),
			},
			want: false,
		},
		{
			name: "TFF = F",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, true, nil),
				mockPrefixChecker(resembled.Zero, false, nil),
				mockPrefixChecker(resembled.Zero, false, nil),
			},
			want: false,
		},
		{
			name: "TFT = F",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, true, nil),
				mockPrefixChecker(resembled.Zero, false, nil),
				mockPrefixChecker(resembled.Zero, true, nil),
			},
			want: false,
		},
		{
			name: "TTF = F",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, true, nil),
				mockPrefixChecker(resembled.Zero, true, nil),
				mockPrefixChecker(resembled.Zero, false, nil),
			},
			want: false,
		},
		{
			name: "TTT = T",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, true, nil),
				mockPrefixChecker(resembled.Zero, true, nil),
				mockPrefixChecker(resembled.Zero, true, nil),
			},
			want: true,
		},
		{
			name:     "empty",
			checkers: []resembled.PrefixChecker{},
			want:     true,
		},
		{
			name: "one true",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, true, nil),
			},
			want: true,
		},
		{
			name: "one false",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, false, nil),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := resembled.AllPrefix(tt.checkers...).Check(resembled.Zero)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, actual)
		})
	}
}

func Test_AnyPrefix(t *testing.T) {
	tests := []struct {
		name     string
		checkers []resembled.PrefixChecker
		want     bool
	}{
		{
			name: "FF = F",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, false, nil),
				mockPrefixChecker(resembled.Zero, false, nil),
			},
			want: false,
		},
		{
			name: "FT = T",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, false, nil),
				mockPrefixChecker(resembled.Zero, true, nil),
			},
			want: true,
		},
		{
			name: "TF = T",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, true, nil),
				mockPrefixChecker(resembled.Zero, false, nil),
			},
			want: true,
		},
		{
			name: "TT = T",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, true, nil),
				mockPrefixChecker(resembled.Zero, true, nil),
			},
			want: true,
		},
		{
			name:     "empty",
			checkers: []resembled.PrefixChecker{},
			want:     false,
		},
		{
			name: "one true",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, true, nil),
			},
			want: true,
		},
		{
			name: "one false",
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(resembled.Zero, false, nil),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := resembled.AnyPrefix(tt.checkers...).Check(resembled.Zero)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, actual)
		})
	}
}

func mockPrefixIterator(vv ...resembled.Type) resembled.PrefixIterator {
	m := &mocks.PrefixIterator{}
	var sp int
	m.
		On("HasNext").Return(func() bool { return sp < len(vv) }).
		On("Next").Return(
		func() resembled.Type {
			next := vv[sp]
			sp++
			return next
		}).
		On("Err").Return(nil)

	return m
}

func Test_PrefixFiltering(t *testing.T) {
	tests := []struct {
		name     string
		iter     resembled.PrefixIterator
		checkers []resembled.PrefixChecker
		want     []resembled.Type
	}{
		{
			name: "general",
			iter: mockPrefixIterator(1, 2, 3),
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(mock.Anything,
					func(val resembled.Type) bool { return val.(int) >= 2 },
					func(_ resembled.Type) error { return nil },
				),
				mockPrefixChecker(mock.Anything,
					func(val resembled.Type) bool { return val.(int)%2 == 0 },
					func(_ resembled.Type) error { return nil },
				),
			},
			want: []resembled.Type{2},
		},
		{
			name:     "no checkers",
			iter:     mockPrefixIterator(1, 2, 3),
			checkers: nil,
			want:     []resembled.Type{1, 2, 3},
		},
		{
			name:     "no iters",
			iter:     resembled.EmptyPrefixIterator,
			checkers: []resembled.PrefixChecker{mockPrefixChecker(resembled.Zero, true, nil)},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := resembled.PrefixFiltering(tt.iter, tt.checkers...)
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

func Test_PrefixDoingUntil(t *testing.T) {
	tests := []struct {
		name     string
		iter     resembled.PrefixIterator
		checkers []resembled.PrefixChecker
		want     [][]resembled.Type
	}{
		{
			name: "general",
			iter: mockPrefixIterator(1, 2, 3, 5, 8),
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(mock.Anything,
					func(val resembled.Type) bool { return val.(int) == 3 },
					func(_ resembled.Type) error { return nil },
				),
			},
			want: [][]resembled.Type{
				{1, 2, 3},
				{5, 8},
			},
		},
		{
			name:     "no checkers",
			iter:     mockPrefixIterator(1, 2, 3),
			checkers: nil,
			want:     [][]resembled.Type{{1, 2, 3}},
		},
		{
			name:     "no iters",
			iter:     resembled.EmptyPrefixIterator,
			checkers: []resembled.PrefixChecker{mockPrefixChecker(resembled.Zero, true, nil)},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, row := range tt.want {
				items := resembled.PrefixDoingUntil(tt.iter, tt.checkers...)
				for _, item := range row {
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
			}
			if !assert.False(t, tt.iter.HasNext()) {
				return
			}
			assert.NoError(t, tt.iter.Err())
		})
	}
}

func Test_PrefixSkipUntil(t *testing.T) {
	tests := []struct {
		name     string
		iter     resembled.PrefixIterator
		checkers []resembled.PrefixChecker
		want     []resembled.Type
	}{
		{
			name: "general",
			iter: mockPrefixIterator(1, 2, 3, 5, 8),
			checkers: []resembled.PrefixChecker{
				mockPrefixChecker(mock.Anything,
					func(val resembled.Type) bool { return val.(int) == 3 },
					func(_ resembled.Type) error { return nil },
				),
			},
			want: []resembled.Type{5, 8},
		},
		{
			name:     "no checkers",
			iter:     mockPrefixIterator(1, 2, 3),
			checkers: nil,
			want:     nil,
		},
		{
			name:     "no iters",
			iter:     resembled.EmptyPrefixIterator,
			checkers: []resembled.PrefixChecker{mockPrefixChecker(resembled.Zero, true, nil)},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resembled.PrefixSkipUntil(tt.iter, tt.checkers...)
			if !assert.NoError(t, err) {
				return
			}
			for _, item := range tt.want {
				if !assert.True(t, tt.iter.HasNext()) {
					return
				}
				if !assert.Equal(t, item, tt.iter.Next()) {
					return
				}
			}
			if !assert.False(t, tt.iter.HasNext()) {
				return
			}
			assert.NoError(t, tt.iter.Err())
		})
	}
}

func Test_EnumFromPrefixChecker(t *testing.T) {
	tests := []struct {
		name    string
		checker resembled.PrefixChecker
		want    bool
	}{
		{
			name:    "checker true",
			checker: mockPrefixChecker(resembled.Zero, true, nil),
			want:    true,
		},
		{
			name:    "checker false",
			checker: mockPrefixChecker(resembled.Zero, false, nil),
			want:    false,
		},
		{
			name:    "no checker",
			checker: nil,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := resembled.EnumFromPrefixChecker(tt.checker).Check(5, resembled.Zero)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, check)
		})
	}
}

func Test_AlwaysPrefixEnumCheckTrue(t *testing.T) {
	tests := []struct {
		name string
		n    int
		val  resembled.Type
	}{
		{"some val", 7, 5},
		{"zero val", 0, resembled.Zero},
		{"nil val", 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := resembled.AlwaysPrefixEnumCheckTrue.Check(tt.n, tt.val)
			if !assert.NoError(t, err) {
				return
			}
			assert.True(t, check)
		})
	}
}

func Test_AlwaysPrefixEnumCheckFalse(t *testing.T) {
	tests := []struct {
		name string
		n    int
		val  resembled.Type
	}{
		{"some val", 7, 5},
		{"zero val", 0, resembled.Zero},
		{"nil val", 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := resembled.AlwaysPrefixEnumCheckFalse.Check(tt.n, tt.val)
			if !assert.NoError(t, err) {
				return
			}
			assert.False(t, check)
		})
	}
}

func mockPrefixEnumChecker(n, val interface{}, ret ...interface{}) resembled.PrefixEnumChecker {
	m := &mocks.PrefixEnumChecker{}
	m.
		On("Check", n, val).Return(ret...)
	return m
}

func Test_EnumNotPrefix(t *testing.T) {
	tests := []struct {
		name    string
		checker resembled.PrefixEnumChecker
	}{
		{
			name:    "general",
			checker: mockPrefixEnumChecker(0, resembled.Zero, true, nil),
		},
		{
			name:    "negative",
			checker: mockPrefixEnumChecker(0, resembled.Zero, false, nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := tt.checker.Check(0, resembled.Zero)
			if !assert.NoError(t, err) {
				return
			}
			notCheck, err := resembled.EnumNotPrefix(tt.checker).Check(0, resembled.Zero)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, check, !notCheck)
		})
	}
}

func Test_EnumAllPrefix(t *testing.T) {
	tests := []struct {
		name     string
		checkers []resembled.PrefixEnumChecker
		want     bool
	}{
		{
			name: "FF = F",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, false, nil),
				mockPrefixEnumChecker(0, resembled.Zero, false, nil),
			},
			want: false,
		},
		{
			name: "FT = F",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, false, nil),
				mockPrefixEnumChecker(0, resembled.Zero, true, nil),
			},
			want: false,
		},
		{
			name: "TF = F",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, true, nil),
				mockPrefixEnumChecker(0, resembled.Zero, false, nil),
			},
			want: false,
		},
		{
			name: "TT = T",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, true, nil),
				mockPrefixEnumChecker(0, resembled.Zero, true, nil),
			},
			want: true,
		},
		{
			name:     "empty",
			checkers: []resembled.PrefixEnumChecker{},
			want:     true,
		},
		{
			name: "one true",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, true, nil),
			},
			want: true,
		},
		{
			name: "one false",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, false, nil),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := resembled.EnumAllPrefix(tt.checkers...).Check(0, resembled.Zero)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, actual)
		})
	}
}

func Test_EnumAnyPrefix(t *testing.T) {
	tests := []struct {
		name     string
		checkers []resembled.PrefixEnumChecker
		want     bool
	}{
		{
			name: "FF = F",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, false, nil),
				mockPrefixEnumChecker(0, resembled.Zero, false, nil),
			},
			want: false,
		},
		{
			name: "FT = T",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, false, nil),
				mockPrefixEnumChecker(0, resembled.Zero, true, nil),
			},
			want: true,
		},
		{
			name: "TF = T",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, true, nil),
				mockPrefixEnumChecker(0, resembled.Zero, false, nil),
			},
			want: true,
		},
		{
			name: "TT = T",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, true, nil),
				mockPrefixEnumChecker(0, resembled.Zero, true, nil),
			},
			want: true,
		},
		{
			name:     "empty",
			checkers: []resembled.PrefixEnumChecker{},
			want:     false,
		},
		{
			name: "one true",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, true, nil),
			},
			want: true,
		},
		{
			name: "one false",
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(0, resembled.Zero, false, nil),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := resembled.EnumAnyPrefix(tt.checkers...).Check(0, resembled.Zero)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.want, actual)
		})
	}
}

func Test_PrefixEnumFiltering(t *testing.T) {
	tests := []struct {
		name     string
		iter     resembled.PrefixIterator
		checkers []resembled.PrefixEnumChecker
		want     []resembled.Type
	}{
		{
			name: "general",
			iter: mockPrefixIterator(1, 2, 3, 5, 8),
			checkers: []resembled.PrefixEnumChecker{
				resembled.EnumAnyPrefix(
					resembled.EnumAllPrefix(
						mockPrefixEnumChecker(mock.Anything, mock.Anything,
							func(_ int, val resembled.Type) bool { return val.(int) >= 3 },
							func(_ int, _ resembled.Type) error { return nil },
						),
						mockPrefixEnumChecker(mock.Anything, mock.Anything,
							func(_ int, val resembled.Type) bool { return val.(int)%2 == 0 },
							func(_ int, _ resembled.Type) error { return nil },
						),
					),
					mockPrefixEnumChecker(mock.Anything, mock.Anything,
						func(n int, _ resembled.Type) bool { return n%2 == 1 },
						func(_ int, _ resembled.Type) error { return nil },
					),
				),
			},
			want: []resembled.Type{2, 5, 8},
		},
		{
			name:     "no checkers",
			iter:     mockPrefixIterator(1, 2, 3),
			checkers: nil,
			want:     []resembled.Type{1, 2, 3},
		},
		{
			name:     "no iters",
			iter:     resembled.EmptyPrefixIterator,
			checkers: []resembled.PrefixEnumChecker{mockPrefixEnumChecker(0, resembled.Zero, true, nil)},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := resembled.PrefixEnumFiltering(tt.iter, tt.checkers...)
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

func Test_PrefixEnumDoingUntil(t *testing.T) {
	tests := []struct {
		name     string
		iter     resembled.PrefixIterator
		checkers []resembled.PrefixEnumChecker
		want     [][]resembled.Type
	}{
		{
			name: "general",
			iter: mockPrefixIterator(1, 2, 3, 5, 8),
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(mock.Anything, mock.Anything,
					func(n int, val resembled.Type) bool { return n == 2 && val.(int) == 3 },
					func(_ int, _ resembled.Type) error { return nil },
				),
			},
			want: [][]resembled.Type{
				{1, 2, 3},
				{5, 8},
			},
		},
		{
			name:     "no checkers",
			iter:     mockPrefixIterator(1, 2, 3),
			checkers: nil,
			want:     [][]resembled.Type{{1, 2, 3}},
		},
		{
			name:     "no iters",
			iter:     resembled.EmptyPrefixIterator,
			checkers: []resembled.PrefixEnumChecker{mockPrefixEnumChecker(0, resembled.Zero, true, nil)},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, row := range tt.want {
				items := resembled.PrefixEnumDoingUntil(tt.iter, tt.checkers...)
				for _, item := range row {
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
			}
			if !assert.False(t, tt.iter.HasNext()) {
				return
			}
			assert.NoError(t, tt.iter.Err())
		})
	}
}

func Test_PrefixEnumSkipUntil(t *testing.T) {
	tests := []struct {
		name     string
		iter     resembled.PrefixIterator
		checkers []resembled.PrefixEnumChecker
		want     []resembled.Type
	}{
		{
			name: "general",
			iter: mockPrefixIterator(1, 2, 3, 5, 8),
			checkers: []resembled.PrefixEnumChecker{
				mockPrefixEnumChecker(mock.Anything, mock.Anything,
					func(n int, val resembled.Type) bool { return n == 2 && val.(int) == 3 },
					func(_ int, _ resembled.Type) error { return nil },
				),
			},
			want: []resembled.Type{5, 8},
		},
		{
			name:     "no checkers",
			iter:     mockPrefixIterator(1, 2, 3),
			checkers: nil,
			want:     nil,
		},
		{
			name:     "no iters",
			iter:     resembled.EmptyPrefixIterator,
			checkers: []resembled.PrefixEnumChecker{mockPrefixEnumChecker(0, resembled.Zero, true, nil)},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resembled.PrefixEnumSkipUntil(tt.iter, tt.checkers...)
			if !assert.NoError(t, err) {
				return
			}
			for _, item := range tt.want {
				if !assert.True(t, tt.iter.HasNext()) {
					return
				}
				if !assert.Equal(t, item, tt.iter.Next()) {
					return
				}
			}
			if !assert.False(t, tt.iter.HasNext()) {
				return
			}
			assert.NoError(t, tt.iter.Err())
		})
	}
}

func Test_PrefixGettingBatch(t *testing.T) {
	tests := []struct {
		name    string
		iter    resembled.PrefixIterator
		batchBy int
		want    [][]resembled.Type
	}{
		{
			name:    "general",
			iter:    mockPrefixIterator(1, 2, 3, 5, 8),
			batchBy: 2,
			want: [][]resembled.Type{
				{1, 2},
				{3, 5},
				{8},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, row := range tt.want {
				items := resembled.PrefixGettingBatch(tt.iter, tt.batchBy)
				for _, item := range row {
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
			}
			if !assert.False(t, tt.iter.HasNext()) {
				return
			}
			assert.NoError(t, tt.iter.Err())
		})
	}
}
