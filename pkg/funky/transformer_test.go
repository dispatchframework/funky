///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package funky_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dispatchframework/funky/pkg/funky"
	"github.com/dispatchframework/funky/pkg/funky/mocks"
	"github.com/stretchr/testify/mock"
)

func TestTransformSuccess(t *testing.T) {
	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}
	rw := &mocks.HTTPReaderWriter{}
	rw.On("Read", mock.AnythingOfType("*map[string]interface {}"), mock.AnythingOfType("*http.Request")).
		Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*map[string]interface{})
			*arg = payload
		})

	injectors := []funky.ContextInjector{
		funky.NewTimeoutInjector(),
		funky.NewRequestMetadataInjector()}

	transformer := funky.NewDefaultRequestTransformer(rw, injectors)

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(payload)
	req := httptest.NewRequest("GET", "/", &body)
	req.Header.Set("Content-Type", "application/json")
	actual, _ := transformer.Transform(req)

	expected := &funky.Request{
		Context: map[string]interface{}{
			"timeout": 0,
			"request": map[string]interface{}{
				"uri":    req.RequestURI,
				"method": req.Method,
				"header": req.Header,
			},
		},
		Payload: payload,
	}

	if !assert.ObjectsAreEqualValues(expected, actual) {
		t.Errorf("did not get expected result. expected:%+v, actual:%+v", expected, actual)
	}
}

func TestTransformMissingContentType(t *testing.T) {
	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}
	rw := &mocks.HTTPReaderWriter{}
	rw.On("Read", mock.AnythingOfType("*map[string]interface {}"), mock.AnythingOfType("*http.Request")).
		Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*map[string]interface{})
			*arg = payload
		})

	injectors := []funky.ContextInjector{}
	transformer := funky.NewDefaultRequestTransformer(rw, injectors)

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(payload)
	req := httptest.NewRequest("GET", "/", &body)
	actual, _ := transformer.Transform(req)

	expected := &funky.Request{
		Context: map[string]interface{}{},
		Payload: payload,
	}

	if !assert.ObjectsAreEqualValues(expected, actual) {
		t.Errorf("did not get expected result. expected:%+v, actual:%+v", expected, actual)
	}
}

func TestTransformUnsupportedContentType(t *testing.T) {
	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}
	rw := &mocks.HTTPReaderWriter{}
	rw.On("Read", mock.AnythingOfType("*map[string]interface {}"), mock.AnythingOfType("*http.Request")).
		Return(funky.UnsupportedMediaTypeError("application/xml"))

	injectors := []funky.ContextInjector{}
	transformer := funky.NewDefaultRequestTransformer(rw, injectors)

	var body bytes.Buffer
	xml.NewEncoder(&body).Encode(payload)
	req := httptest.NewRequest("GET", "/", &body)
	req.Header.Set("Content-Type", "application/xml")
	actual, err := transformer.Transform(req)

	assert.Nil(t, actual)
	assert.Error(t, err, "Expected Transform to fail with UnsupportedMediaTypeError, instead got %+v", err)
}

func TestTransformNoInjectors(t *testing.T) {
	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}

	rw := &mocks.HTTPReaderWriter{}
	rw.On("Read", mock.AnythingOfType("*map[string]interface {}"), mock.AnythingOfType("*http.Request")).
		Return(nil).
		Run(func(args mock.Arguments) {
			p := args.Get(0).(*map[string]interface{})
			*p = payload
		})

	injectors := []funky.ContextInjector{}
	transformer := funky.NewDefaultRequestTransformer(rw, injectors)

	data := bytes.Buffer{}
	json.NewEncoder(&data).Encode(payload)

	req := httptest.NewRequest("GET", "/", &data)
	req.Header.Set("Content-Type", "application/json")

	actual, err := transformer.Transform(req)

	assert.NoErrorf(t, err, "Expected no error, instead got %+v", err)

	expected := &funky.Request{
		Context: map[string]interface{}{},
		Payload: payload,
	}

	assert.EqualValuesf(t, expected, actual, "Expected: %+v; Actual: %+v", expected, actual)
}

func TestTransformNonAwareInjectorNoError(t *testing.T) {
	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}

	rw := &mocks.HTTPReaderWriter{}
	rw.On("Read", mock.AnythingOfType("*map[string]interface {}"), mock.AnythingOfType("*http.Request")).
		Return(nil).
		Run(func(args mock.Arguments) {
			p := args.Get(0).(*map[string]interface{})
			*p = payload
		})

	timeoutInjector := mocks.ContextInjector{}
	timeoutInjector.On("Inject", mock.AnythingOfType("*funky.Request")).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*funky.Request)
		arg.Context["timeout"] = 0
	})

	injectors := []funky.ContextInjector{&timeoutInjector}
	transformer := funky.NewDefaultRequestTransformer(rw, injectors)

	data := bytes.Buffer{}
	json.NewEncoder(&data).Encode(payload)

	req := httptest.NewRequest("GET", "/", &data)
	req.Header.Set("Content-Type", "application/json")

	actual, err := transformer.Transform(req)

	assert.NoErrorf(t, err, "Expected no error, instead got %+v", err)

	expected := &funky.Request{
		Context: map[string]interface{}{
			"timeout": 0,
		},
		Payload: payload,
	}

	assert.EqualValuesf(t, expected, actual, "Expected: %+v; Actual: %+v", expected, actual)
}

func TestTransformRequestAwareInjectorNoError(t *testing.T) {
	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}
	data := bytes.Buffer{}
	json.NewEncoder(&data).Encode(payload)

	req := httptest.NewRequest("GET", "/", &data)
	req.Header.Set("Content-Type", "application/json")

	requestCtx := map[string]interface{}{
		"uri":    req.RequestURI,
		"method": req.Method,
		"header": req.Header,
	}

	requestMetadataInjector := mocks.HTTPRequestAwareContextInjector{}
	requestMetadataInjector.On("Inject", mock.AnythingOfType("*funky.Request")).
		Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*funky.Request)
			arg.Context["request"] = requestCtx
		})

	requestMetadataInjector.On("SetHTTPRequest", mock.AnythingOfType("*http.Request")).Return()

	injectors := []funky.ContextInjector{&requestMetadataInjector}

	rw := &mocks.HTTPReaderWriter{}
	rw.On("Read", mock.AnythingOfType("*map[string]interface {}"), mock.AnythingOfType("*http.Request")).
		Return(nil).
		Run(func(args mock.Arguments) {
			p := args.Get(0).(*map[string]interface{})
			*p = payload
		})

	transformer := funky.NewDefaultRequestTransformer(rw, injectors)

	actual, err := transformer.Transform(req)

	requestMetadataInjector.AssertCalled(t, "SetHTTPRequest", mock.AnythingOfType("*http.Request"))
	assert.NoErrorf(t, err, "Expected no error, instead got %+v", err)

	expected := &funky.Request{
		Context: map[string]interface{}{
			"request": requestCtx,
		},
		Payload: payload,
	}

	assert.EqualValuesf(t, expected, actual, "Expected: %+v; Actual: %+v", expected, actual)
}

func TestTransformerInjectorError(t *testing.T) {
	rw := &mocks.HTTPReaderWriter{}

	timeoutInjector := mocks.ContextInjector{}
	timeoutInjector.On("Inject", mock.AnythingOfType("*funky.Request")).
		Return(funky.IllegalStateError("Context[\"timeout\"] already exists"))

	injectors := []funky.ContextInjector{&timeoutInjector}
	transformer := funky.NewDefaultRequestTransformer(rw, injectors)

	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}
	data := bytes.Buffer{}
	json.NewEncoder(&data).Encode(payload)

	req := httptest.NewRequest("GET", "/", &data)
	req.Header.Set("Content-Type", "application/json")

	actual, err := transformer.Transform(req)

	timeoutInjector.AssertCalled(t, "Inject", mock.AnythingOfType("*funky.Request"))
	rw.AssertNotCalled(t, "Read", mock.Anything)

	assert.Nil(t, actual)
	assert.Errorf(t, err, "Expected error, instead got %+v", err)
}
