// Code generated by mockery v2.30.10. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// NetworkIntf is an autogenerated mock type for the NetworkIntf type
type NetworkIntf struct {
	mock.Mock
}

// FindIpOnSubnet provides a mock function with given fields: subnet
func (_m *NetworkIntf) FindIpOnSubnet(subnet string) (string, error) {
	ret := _m.Called(subnet)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(subnet)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(subnet)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(subnet)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindNetworkAddress provides a mock function with given fields: address
func (_m *NetworkIntf) FindNetworkAddress(address string) (string, error) {
	ret := _m.Called(address)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(address)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(address)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewNetworkIntf creates a new instance of NetworkIntf. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNetworkIntf(t interface {
	mock.TestingT
	Cleanup(func())
}) *NetworkIntf {
	mock := &NetworkIntf{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}