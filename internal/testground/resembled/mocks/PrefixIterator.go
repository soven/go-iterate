// Code generated by mockery v2.4.0. DO NOT EDIT.

package mocks

import (
	resembled "github.com/soven/go-iterate/internal/testground/resembled"
	mock "github.com/stretchr/testify/mock"
)

// PrefixIterator is an autogenerated mock type for the PrefixIterator type
type PrefixIterator struct {
	mock.Mock
}

// Err provides a mock function with given fields:
func (_m *PrefixIterator) Err() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// HasNext provides a mock function with given fields:
func (_m *PrefixIterator) HasNext() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Next provides a mock function with given fields:
func (_m *PrefixIterator) Next() resembled.Type {
	ret := _m.Called()

	var r0 resembled.Type
	if rf, ok := ret.Get(0).(func() resembled.Type); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(resembled.Type)
		}
	}

	return r0
}
