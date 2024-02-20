// Code generated by mockery v2.42.0. DO NOT EDIT.

package mockery

import (
	context "context"

	configuration "github.com/walteh/retab/pkg/configuration"

	mock "github.com/stretchr/testify/mock"
)

// MockProvider_configuration is an autogenerated mock type for the Provider type
type MockProvider_configuration struct {
	mock.Mock
}

type MockProvider_configuration_Expecter struct {
	mock *mock.Mock
}

func (_m *MockProvider_configuration) EXPECT() *MockProvider_configuration_Expecter {
	return &MockProvider_configuration_Expecter{mock: &_m.Mock}
}

// GetConfigurationForFileType provides a mock function with given fields: ctx, filename
func (_m *MockProvider_configuration) GetConfigurationForFileType(ctx context.Context, filename string) (configuration.Configuration, error) {
	ret := _m.Called(ctx, filename)

	if len(ret) == 0 {
		panic("no return value specified for GetConfigurationForFileType")
	}

	var r0 configuration.Configuration
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (configuration.Configuration, error)); ok {
		return rf(ctx, filename)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) configuration.Configuration); ok {
		r0 = rf(ctx, filename)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(configuration.Configuration)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, filename)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProvider_configuration_GetConfigurationForFileType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetConfigurationForFileType'
type MockProvider_configuration_GetConfigurationForFileType_Call struct {
	*mock.Call
}

// GetConfigurationForFileType is a helper method to define mock.On call
//   - ctx context.Context
//   - filename string
func (_e *MockProvider_configuration_Expecter) GetConfigurationForFileType(ctx interface{}, filename interface{}) *MockProvider_configuration_GetConfigurationForFileType_Call {
	return &MockProvider_configuration_GetConfigurationForFileType_Call{Call: _e.mock.On("GetConfigurationForFileType", ctx, filename)}
}

func (_c *MockProvider_configuration_GetConfigurationForFileType_Call) Run(run func(ctx context.Context, filename string)) *MockProvider_configuration_GetConfigurationForFileType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockProvider_configuration_GetConfigurationForFileType_Call) Return(_a0 configuration.Configuration, _a1 error) *MockProvider_configuration_GetConfigurationForFileType_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProvider_configuration_GetConfigurationForFileType_Call) RunAndReturn(run func(context.Context, string) (configuration.Configuration, error)) *MockProvider_configuration_GetConfigurationForFileType_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockProvider_configuration creates a new instance of MockProvider_configuration. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProvider_configuration(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProvider_configuration {
	mock := &MockProvider_configuration{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
