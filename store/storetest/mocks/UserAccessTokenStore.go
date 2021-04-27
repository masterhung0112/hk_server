// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `npm run task store_mocks`.

package mocks

import (
	model "github.com/masterhung0112/hk_server/v5/model"
	mock "github.com/stretchr/testify/mock"
)

// UserAccessTokenStore is an autogenerated mock type for the UserAccessTokenStore type
type UserAccessTokenStore struct {
	mock.Mock
}

// Delete provides a mock function with given fields: tokenID
func (_m *UserAccessTokenStore) Delete(tokenID string) error {
	ret := _m.Called(tokenID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(tokenID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteAllForUser provides a mock function with given fields: userID
func (_m *UserAccessTokenStore) DeleteAllForUser(userID string) error {
	ret := _m.Called(userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: tokenID
func (_m *UserAccessTokenStore) Get(tokenID string) (*model.UserAccessToken, error) {
	ret := _m.Called(tokenID)

	var r0 *model.UserAccessToken
	if rf, ok := ret.Get(0).(func(string) *model.UserAccessToken); ok {
		r0 = rf(tokenID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.UserAccessToken)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tokenID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAll provides a mock function with given fields: offset, limit
func (_m *UserAccessTokenStore) GetAll(offset int, limit int) ([]*model.UserAccessToken, error) {
	ret := _m.Called(offset, limit)

	var r0 []*model.UserAccessToken
	if rf, ok := ret.Get(0).(func(int, int) []*model.UserAccessToken); ok {
		r0 = rf(offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.UserAccessToken)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int, int) error); ok {
		r1 = rf(offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByToken provides a mock function with given fields: tokenString
func (_m *UserAccessTokenStore) GetByToken(tokenString string) (*model.UserAccessToken, error) {
	ret := _m.Called(tokenString)

	var r0 *model.UserAccessToken
	if rf, ok := ret.Get(0).(func(string) *model.UserAccessToken); ok {
		r0 = rf(tokenString)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.UserAccessToken)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(tokenString)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByUser provides a mock function with given fields: userID, page, perPage
func (_m *UserAccessTokenStore) GetByUser(userID string, page int, perPage int) ([]*model.UserAccessToken, error) {
	ret := _m.Called(userID, page, perPage)

	var r0 []*model.UserAccessToken
	if rf, ok := ret.Get(0).(func(string, int, int) []*model.UserAccessToken); ok {
		r0 = rf(userID, page, perPage)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.UserAccessToken)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, int, int) error); ok {
		r1 = rf(userID, page, perPage)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: token
func (_m *UserAccessTokenStore) Save(token *model.UserAccessToken) (*model.UserAccessToken, error) {
	ret := _m.Called(token)

	var r0 *model.UserAccessToken
	if rf, ok := ret.Get(0).(func(*model.UserAccessToken) *model.UserAccessToken); ok {
		r0 = rf(token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.UserAccessToken)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.UserAccessToken) error); ok {
		r1 = rf(token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Search provides a mock function with given fields: term
func (_m *UserAccessTokenStore) Search(term string) ([]*model.UserAccessToken, error) {
	ret := _m.Called(term)

	var r0 []*model.UserAccessToken
	if rf, ok := ret.Get(0).(func(string) []*model.UserAccessToken); ok {
		r0 = rf(term)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.UserAccessToken)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(term)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateTokenDisable provides a mock function with given fields: tokenID
func (_m *UserAccessTokenStore) UpdateTokenDisable(tokenID string) error {
	ret := _m.Called(tokenID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(tokenID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateTokenEnable provides a mock function with given fields: tokenID
func (_m *UserAccessTokenStore) UpdateTokenEnable(tokenID string) error {
	ret := _m.Called(tokenID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(tokenID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
