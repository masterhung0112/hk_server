// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import (
	gorp "github.com/mattermost/gorp"
	mock "github.com/stretchr/testify/mock"

	squirrel "github.com/Masterminds/squirrel"

	store "github.com/masterhung0112/go_server/store"
)

// SqlStore is an autogenerated mock type for the SqlStore type
type SqlStore struct {
	mock.Mock
}

// DriverName provides a mock function with given fields:
func (_m *SqlStore) DriverName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetAllConns provides a mock function with given fields:
func (_m *SqlStore) GetAllConns() []*gorp.DbMap {
	ret := _m.Called()

	var r0 []*gorp.DbMap
	if rf, ok := ret.Get(0).(func() []*gorp.DbMap); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gorp.DbMap)
		}
	}

	return r0
}

// GetMaster provides a mock function with given fields:
func (_m *SqlStore) GetMaster() *gorp.DbMap {
	ret := _m.Called()

	var r0 *gorp.DbMap
	if rf, ok := ret.Get(0).(func() *gorp.DbMap); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gorp.DbMap)
		}
	}

	return r0
}

// GetReplica provides a mock function with given fields:
func (_m *SqlStore) GetReplica() *gorp.DbMap {
	ret := _m.Called()

	var r0 *gorp.DbMap
	if rf, ok := ret.Get(0).(func() *gorp.DbMap); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gorp.DbMap)
		}
	}

	return r0
}

// User provides a mock function with given fields:
func (_m *SqlStore) User() store.UserStore {
	ret := _m.Called()

	var r0 store.UserStore
	if rf, ok := ret.Get(0).(func() store.UserStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(store.UserStore)
		}
	}

	return r0
}

// getQueryBuilder provides a mock function with given fields:
func (_m *SqlStore) getQueryBuilder() squirrel.StatementBuilderType {
	ret := _m.Called()

	var r0 squirrel.StatementBuilderType
	if rf, ok := ret.Get(0).(func() squirrel.StatementBuilderType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(squirrel.StatementBuilderType)
	}

	return r0
}
