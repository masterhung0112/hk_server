// Code generated by mockery v1.0.0. DO NOT EDIT.

// 'Regenerate

package mocks

import mock "github.com/stretchr/testify/mock"

// dbSelecter is an autogenerated mock type for the dbSelecter type
type dbSelecter struct {
	mock.Mock
}

// Select provides a mock function with given fields: i, query, args
func (_m *dbSelecter) Select(i interface{}, query string, args ...interface{}) ([]interface{}, error) {
	var _ca []interface{}
	_ca = append(_ca, i, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 []interface{}
	if rf, ok := ret.Get(0).(func(interface{}, string, ...interface{}) []interface{}); ok {
		r0 = rf(i, query, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(interface{}, string, ...interface{}) error); ok {
		r1 = rf(i, query, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
