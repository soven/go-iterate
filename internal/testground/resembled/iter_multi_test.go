package resembled_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/soven/go-iterate/internal/testground/resembled"
	"github.com/soven/go-iterate/internal/testground/resembled/mocks"
)

func Test_SuperPrefixIterator(t *testing.T) {
	tests := []struct {
		name  string
		iters []resembled.PrefixIterator
		want  []resembled.Type
	}{
		{
			name: "general",
			iters: []resembled.PrefixIterator{
				mockPrefixIterator(1, 4, 7),
				mockPrefixIterator(2, 5, 8),
				mockPrefixIterator(3, 6),
			},
			want: []resembled.Type{1, 4, 7, 2, 5, 8, 3, 6},
		},
		{
			name: "some iter empty",
			iters: []resembled.PrefixIterator{
				mockPrefixIterator(1, 4, 7),
				resembled.EmptyPrefixIterator,
				mockPrefixIterator(3, 6),
			},
			want: []resembled.Type{1, 4, 7, 3, 6},
		},
		{
			name:  "iters empty",
			iters: nil,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := resembled.SuperPrefixIterator(tt.iters...)
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

func mockPrefixComparer(lhs, rhs, ret interface{}) resembled.PrefixComparer {
	m := &mocks.PrefixComparer{}
	m.
		On("IsLess", lhs, rhs).Return(ret)
	return m
}

func Test_PriorPrefixIterator(t *testing.T) {
	tests := []struct {
		name     string
		comparer resembled.PrefixComparer
		iters    []resembled.PrefixIterator
		want     []resembled.Type
	}{
		{
			name: "general",
			comparer: mockPrefixComparer(mock.Anything, mock.Anything,
				func(lhs, rhs resembled.Type) bool { return lhs.(int) < rhs.(int) },
			),
			iters: []resembled.PrefixIterator{
				mockPrefixIterator(1, 4, 7),
				mockPrefixIterator(2, 5, 8),
				mockPrefixIterator(3, 6),
			},
			want: []resembled.Type{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name: "unordered",
			comparer: mockPrefixComparer(mock.Anything, mock.Anything,
				func(lhs, rhs resembled.Type) bool { return lhs.(int) < rhs.(int) },
			),
			iters: []resembled.PrefixIterator{
				mockPrefixIterator(16, 4, 10),
				mockPrefixIterator(22, 5, 18),
				mockPrefixIterator(13, 6),
			},
			want: []resembled.Type{13, 6, 16, 4, 10, 22, 5, 18},
		},
		{
			name: "some iter empty",
			comparer: mockPrefixComparer(mock.Anything, mock.Anything,
				func(lhs, rhs resembled.Type) bool { return lhs.(int) < rhs.(int) },
			),
			iters: []resembled.PrefixIterator{
				mockPrefixIterator(1, 4, 7),
				resembled.EmptyPrefixIterator,
				mockPrefixIterator(3, 6),
			},
			want: []resembled.Type{1, 3, 4, 6, 7},
		},
		{
			name:  "iters empty",
			iters: nil,
			want:  nil,
		},
		{
			name:     "comparer empty",
			comparer: nil,
			iters: []resembled.PrefixIterator{
				mockPrefixIterator(1, 4, 7),
				mockPrefixIterator(2, 5, 8),
				mockPrefixIterator(3, 6),
			},
			want: []resembled.Type{1, 4, 7, 2, 5, 8, 3, 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := resembled.PriorPrefixIterator(tt.comparer, tt.iters...)
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
