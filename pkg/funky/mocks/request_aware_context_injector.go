package mocks

import (
	http "net/http"

	funky "github.com/dispatchframework/funky/pkg/funky"
	mock "github.com/stretchr/testify/mock"
)

// HTTPRequestAwareContextInjector is mock type for HTTPRequestAware and ContextInjector types
type HTTPRequestAwareContextInjector struct {
	mock.Mock
}

// SetHTTPRequest provides a mock function with given fields: req
func (_m *HTTPRequestAwareContextInjector) SetHTTPRequest(req *http.Request) {
	_m.Called(req)
}

// Inject provides a mock function with given fields: req
func (_m *HTTPRequestAwareContextInjector) Inject(req *funky.Request) error {
	ret := _m.Called(req)

	var r0 error
	if rf, ok := ret.Get(0).(func(*funky.Request) error); ok {
		r0 = rf(req)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
