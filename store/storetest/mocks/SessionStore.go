// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import (
	model "github.com/masterhung0112/hk_server/model"
	mock "github.com/stretchr/testify/mock"
)

// SessionStore is an autogenerated mock type for the SessionStore type
type SessionStore struct {
	mock.Mock
}

// Get provides a mock function with given fields: sessionIdOrToken
func (_m *SessionStore) Get(sessionIdOrToken string) (*model.Session, error) {
	ret := _m.Called(sessionIdOrToken)

	var r0 *model.Session
	if rf, ok := ret.Get(0).(func(string) *model.Session); ok {
		r0 = rf(sessionIdOrToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(sessionIdOrToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSessions provides a mock function with given fields: userId
func (_m *SessionStore) GetSessions(userId string) ([]*model.Session, error) {
	ret := _m.Called(userId)

	var r0 []*model.Session
	if rf, ok := ret.Get(0).(func(string) []*model.Session); ok {
		r0 = rf(userId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Remove provides a mock function with given fields: sessionIdOrToken
func (_m *SessionStore) Remove(sessionIdOrToken string) error {
	ret := _m.Called(sessionIdOrToken)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(sessionIdOrToken)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Save provides a mock function with given fields: session
func (_m *SessionStore) Save(session *model.Session) (*model.Session, error) {
	ret := _m.Called(session)

	var r0 *model.Session
	if rf, ok := ret.Get(0).(func(*model.Session) *model.Session); ok {
		r0 = rf(session)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Session) error); ok {
		r1 = rf(session)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateRoles provides a mock function with given fields: userId, roles
func (_m *SessionStore) UpdateRoles(userId string, roles string) (string, error) {
	ret := _m.Called(userId, roles)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(userId, roles)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(userId, roles)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
