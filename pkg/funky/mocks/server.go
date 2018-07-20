// Code generated by mockery v1.0.0
package mocks

import funky "github.com/dispatchframework/funky/pkg/funky"
import io "io"
import mock "github.com/stretchr/testify/mock"

// Server is an autogenerated mock type for the Server type
type Server struct {
	mock.Mock
}

// GetPort provides a mock function with given fields:
func (_m *Server) GetPort() uint16 {
	ret := _m.Called()

	var r0 uint16
	if rf, ok := ret.Get(0).(func() uint16); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint16)
	}

	return r0
}

// Invoke provides a mock function with given fields: input
func (_m *Server) Invoke(input *funky.Message) (io.ReadCloser, error) {
	ret := _m.Called(input)

	var r0 io.ReadCloser
	if rf, ok := ret.Get(0).(func(*funky.Message) io.ReadCloser); ok {
		r0 = rf(input)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*funky.Message) error); ok {
		r1 = rf(input)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Shutdown provides a mock function with given fields:
func (_m *Server) Shutdown() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Start provides a mock function with given fields:
func (_m *Server) Start() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Stderr provides a mock function with given fields:
func (_m *Server) Stderr() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Stdout provides a mock function with given fields:
func (_m *Server) Stdout() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Terminate provides a mock function with given fields:
func (_m *Server) Terminate() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
