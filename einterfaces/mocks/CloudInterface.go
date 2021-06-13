// Code generated by mockery v1.0.0. DO NOT EDIT.

// 'Regenerate

package mocks

import (
	model "github.com/masterhung0112/hk_server/v5/model"
	mock "github.com/stretchr/testify/mock"
)

// CloudInterface is an autogenerated mock type for the CloudInterface type
type CloudInterface struct {
	mock.Mock
}

// ChangeSubscription provides a mock function with given fields: userID, subscriptionID, subscriptionChange
func (_m *CloudInterface) ChangeSubscription(userID string, subscriptionID string, subscriptionChange *model.SubscriptionChange) (*model.Subscription, error) {
	ret := _m.Called(userID, subscriptionID, subscriptionChange)

	var r0 *model.Subscription
	if rf, ok := ret.Get(0).(func(string, string, *model.SubscriptionChange) *model.Subscription); ok {
		r0 = rf(userID, subscriptionID, subscriptionChange)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Subscription)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, *model.SubscriptionChange) error); ok {
		r1 = rf(userID, subscriptionID, subscriptionChange)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConfirmCustomerPayment provides a mock function with given fields: userID, confirmRequest
func (_m *CloudInterface) ConfirmCustomerPayment(userID string, confirmRequest *model.ConfirmPaymentMethodRequest) error {
	ret := _m.Called(userID, confirmRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *model.ConfirmPaymentMethodRequest) error); ok {
		r0 = rf(userID, confirmRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateCustomerPayment provides a mock function with given fields: userID
func (_m *CloudInterface) CreateCustomerPayment(userID string) (*model.StripeSetupIntent, error) {
	ret := _m.Called(userID)

	var r0 *model.StripeSetupIntent
	if rf, ok := ret.Get(0).(func(string) *model.StripeSetupIntent); ok {
		r0 = rf(userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.StripeSetupIntent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCloudCustomer provides a mock function with given fields: userID
func (_m *CloudInterface) GetCloudCustomer(userID string) (*model.CloudCustomer, error) {
	ret := _m.Called(userID)

	var r0 *model.CloudCustomer
	if rf, ok := ret.Get(0).(func(string) *model.CloudCustomer); ok {
		r0 = rf(userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.CloudCustomer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCloudProducts provides a mock function with given fields: userID
func (_m *CloudInterface) GetCloudProducts(userID string) ([]*model.Product, error) {
	ret := _m.Called(userID)

	var r0 []*model.Product
	if rf, ok := ret.Get(0).(func(string) []*model.Product); ok {
		r0 = rf(userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Product)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetInvoicePDF provides a mock function with given fields: userID, invoiceID
func (_m *CloudInterface) GetInvoicePDF(userID string, invoiceID string) ([]byte, string, error) {
	ret := _m.Called(userID, invoiceID)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, string) []byte); ok {
		r0 = rf(userID, invoiceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 string
	if rf, ok := ret.Get(1).(func(string, string) string); ok {
		r1 = rf(userID, invoiceID)
	} else {
		r1 = ret.Get(1).(string)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string, string) error); ok {
		r2 = rf(userID, invoiceID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetInvoicesForSubscription provides a mock function with given fields: userID
func (_m *CloudInterface) GetInvoicesForSubscription(userID string) ([]*model.Invoice, error) {
	ret := _m.Called(userID)

	var r0 []*model.Invoice
	if rf, ok := ret.Get(0).(func(string) []*model.Invoice); ok {
		r0 = rf(userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Invoice)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSubscription provides a mock function with given fields: userID
func (_m *CloudInterface) GetSubscription(userID string) (*model.Subscription, error) {
	ret := _m.Called(userID)

	var r0 *model.Subscription
	if rf, ok := ret.Get(0).(func(string) *model.Subscription); ok {
		r0 = rf(userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Subscription)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateCloudCustomer provides a mock function with given fields: userID, customerInfo
func (_m *CloudInterface) UpdateCloudCustomer(userID string, customerInfo *model.CloudCustomerInfo) (*model.CloudCustomer, error) {
	ret := _m.Called(userID, customerInfo)

	var r0 *model.CloudCustomer
	if rf, ok := ret.Get(0).(func(string, *model.CloudCustomerInfo) *model.CloudCustomer); ok {
		r0 = rf(userID, customerInfo)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.CloudCustomer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *model.CloudCustomerInfo) error); ok {
		r1 = rf(userID, customerInfo)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateCloudCustomerAddress provides a mock function with given fields: userID, address
func (_m *CloudInterface) UpdateCloudCustomerAddress(userID string, address *model.Address) (*model.CloudCustomer, error) {
	ret := _m.Called(userID, address)

	var r0 *model.CloudCustomer
	if rf, ok := ret.Get(0).(func(string, *model.Address) *model.CloudCustomer); ok {
		r0 = rf(userID, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.CloudCustomer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *model.Address) error); ok {
		r1 = rf(userID, address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
