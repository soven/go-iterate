// Code generated by mockery v2.4.0. DO NOT EDIT.

package mocks

import (
	resembled "github.com/soven/go-iterate/internal/testground/resembled"
	mock "github.com/stretchr/testify/mock"
)

// PrefixHandler is an autogenerated mock type for the PrefixHandler type
type PrefixHandler struct {
	mock.Mock
}

// Handle provides a mock function with given fields: _a0
func (_m *PrefixHandler) Handle(_a0 resembled.Type) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(resembled.Type) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
