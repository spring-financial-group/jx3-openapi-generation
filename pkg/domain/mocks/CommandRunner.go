// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// CommandRunner is an autogenerated mock type for the CommandRunner type
type CommandRunner struct {
	mock.Mock
}

// Execute provides a mock function with given fields: dir, name, args
func (_m *CommandRunner) Execute(dir string, name string, args ...string) (string, error) {
	_va := make([]interface{}, len(args))
	for _i := range args {
		_va[_i] = args[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, dir, name)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, ...string) string); ok {
		r0 = rf(dir, name, args...)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, ...string) error); ok {
		r1 = rf(dir, name, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExecuteAndLog provides a mock function with given fields: dir, name, args
func (_m *CommandRunner) ExecuteAndLog(dir string, name string, args ...string) error {
	_va := make([]interface{}, len(args))
	for _i := range args {
		_va[_i] = args[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, dir, name)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, ...string) error); ok {
		r0 = rf(dir, name, args...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewCommandRunner interface {
	mock.TestingT
	Cleanup(func())
}

// NewCommandRunner creates a new instance of CommandRunner. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewCommandRunner(t mockConstructorTestingTNewCommandRunner) *CommandRunner {
	mock := &CommandRunner{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
