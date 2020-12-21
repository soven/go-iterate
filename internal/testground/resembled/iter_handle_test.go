package resembled_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/soven/go-iterate/internal/testground/resembled"
	"github.com/soven/go-iterate/internal/testground/resembled/mocks"
)

func mockPrefixHandler(val interface{}, ret ...interface{}) resembled.PrefixHandler {
	m := &mocks.PrefixHandler{}
	m.
		On("Handle", val).Return(ret...)
	return m
}

func Test_PrefixHandling(t *testing.T) {
	tests := []struct {
		name        string
		iter        resembled.PrefixIterator
		getHandlers []func(vvRef *[]resembled.Type) resembled.PrefixHandler
		wantIter    []resembled.Type
		wantRes     []resembled.Type
	}{
		{
			name: "general",
			iter: mockPrefixIterator(1, 3, 8),
			getHandlers: []func(*[]resembled.Type) resembled.PrefixHandler{
				func(vvRef *[]resembled.Type) resembled.PrefixHandler {
					return mockPrefixHandler(mock.Anything,
						func(val resembled.Type) error {
							*vvRef = append(*vvRef, val.(int)*2)
							return nil
						},
					)
				},
				func(vvRef *[]resembled.Type) resembled.PrefixHandler {
					return mockPrefixHandler(mock.Anything,
						func(val resembled.Type) error {
							*vvRef = append(*vvRef, val.(int)*3)
							return nil
						},
					)
				},
			},
			wantIter: []resembled.Type{1, 3, 8},
			wantRes:  []resembled.Type{2, 3, 6, 9, 16, 24},
		},
		{
			name:        "no handlers",
			iter:        mockPrefixIterator(1, 2, 3),
			getHandlers: nil,
			wantIter:    []resembled.Type{1, 2, 3},
			wantRes:     nil,
		},
		{
			name: "no iters",
			iter: resembled.EmptyPrefixIterator,
			getHandlers: []func(*[]resembled.Type) resembled.PrefixHandler{
				func(vvRef *[]resembled.Type) resembled.PrefixHandler {
					return mockPrefixHandler(resembled.Zero, nil)
				},
			},
			wantIter: nil,
			wantRes:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual []resembled.Type
			var handlers []resembled.PrefixHandler
			for _, getHandler := range tt.getHandlers {
				handlers = append(handlers, getHandler(&actual))
			}

			items := resembled.PrefixHandling(tt.iter, handlers...)
			for _, item := range tt.wantIter {
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
			if !assert.NoError(t, items.Err()) {
				return
			}

			assert.EqualValues(t, tt.wantRes, actual)
		})
	}
}

func mockPrefixEnumHandler(n, val interface{}, ret ...interface{}) resembled.PrefixEnumHandler {
	m := &mocks.PrefixEnumHandler{}
	m.
		On("Handle", n, val).Return(ret...)
	return m
}

func Test_PrefixEnumHandling(t *testing.T) {
	tests := []struct {
		name        string
		iter        resembled.PrefixIterator
		getHandlers []func(vvRef *[]resembled.Type) resembled.PrefixEnumHandler
		wantIter    []resembled.Type
		wantRes     []resembled.Type
	}{
		{
			name: "general",
			iter: mockPrefixIterator(1, 3, 8),
			getHandlers: []func(*[]resembled.Type) resembled.PrefixEnumHandler{
				func(vvRef *[]resembled.Type) resembled.PrefixEnumHandler {
					return mockPrefixEnumHandler(mock.Anything, mock.Anything,
						func(n int, val resembled.Type) error {
							*vvRef = append(*vvRef, val.(int)*2+n)
							return nil
						},
					)
				},
				func(vvRef *[]resembled.Type) resembled.PrefixEnumHandler {
					return mockPrefixEnumHandler(mock.Anything, mock.Anything,
						func(n int, val resembled.Type) error {
							*vvRef = append(*vvRef, val.(int)*3+n)
							return nil
						},
					)
				},
			},
			wantIter: []resembled.Type{1, 3, 8},
			wantRes:  []resembled.Type{2, 3, 7, 10, 18, 26},
		},
		{
			name:        "no handlers",
			iter:        mockPrefixIterator(1, 2, 3),
			getHandlers: nil,
			wantIter:    []resembled.Type{1, 2, 3},
			wantRes:     nil,
		},
		{
			name: "no iters",
			iter: resembled.EmptyPrefixIterator,
			getHandlers: []func(*[]resembled.Type) resembled.PrefixEnumHandler{
				func(vvRef *[]resembled.Type) resembled.PrefixEnumHandler {
					return mockPrefixEnumHandler(0, resembled.Zero, nil)
				},
			},
			wantIter: nil,
			wantRes:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual []resembled.Type
			var handlers []resembled.PrefixEnumHandler
			for _, getHandler := range tt.getHandlers {
				handlers = append(handlers, getHandler(&actual))
			}

			items := resembled.PrefixEnumHandling(tt.iter, handlers...)
			for _, item := range tt.wantIter {
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
			if !assert.NoError(t, items.Err()) {
				return
			}

			assert.EqualValues(t, tt.wantRes, actual)
		})
	}
}
