// Code generated by mockery v2.4.0. DO NOT EDIT.

package mocks

import (
	resembled "github.com/soven/go-iterate/internal/testground/resembled"
	mock "github.com/stretchr/testify/mock"
)

// PrefixEnumChecker is an autogenerated mock type for the PrefixEnumChecker type
type PrefixEnumChecker struct {
	mock.Mock
}

// Check provides a mock function with given fields: _a0, _a1
func (_m *PrefixEnumChecker) Check(_a0 int, _a1 resembled.Type) (bool, error) {
	ret := _m.Called(_a0, _a1)

	var r0 bool
	if rf, ok := ret.Get(0).(func(int, resembled.Type) bool); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int, resembled.Type) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}