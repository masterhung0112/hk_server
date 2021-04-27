// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `npm run task einterfaces_mocks`.

package mocks

import (
	model "github.com/masterhung0112/hk_server/v5/model"
	mock "github.com/stretchr/testify/mock"
)

// NotificationInterface is an autogenerated mock type for the NotificationInterface type
type NotificationInterface struct {
	mock.Mock
}

// CheckLicense provides a mock function with given fields:
func (_m *NotificationInterface) CheckLicense() *model.AppError {
	ret := _m.Called()

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func() *model.AppError); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// GetNotificationMessage provides a mock function with given fields: ack, userID
func (_m *NotificationInterface) GetNotificationMessage(ack *model.PushNotificationAck, userID string) (*model.PushNotification, *model.AppError) {
	ret := _m.Called(ack, userID)

	var r0 *model.PushNotification
	if rf, ok := ret.Get(0).(func(*model.PushNotificationAck, string) *model.PushNotification); ok {
		r0 = rf(ack, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PushNotification)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*model.PushNotificationAck, string) *model.AppError); ok {
		r1 = rf(ack, userID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}
