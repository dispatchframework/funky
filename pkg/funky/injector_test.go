///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package funky_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dispatchframework/funky/pkg/funky"
)

func TestEnvVarSecretInjectorInjectSecretsAlreadyExists(t *testing.T) {
	injector := funky.NewEnvVarSecretInjector("username", "password")

	var body funky.Request
	body.Context = map[string]interface{}{}
	body.Context["secrets"] = map[string]string{}

	err := injector.Inject(&body)

	assert.Errorf(t, err, "Injector should have failed with 'Context[\"secrets\"] already exists', instead got %+v", err)
}

func TestTimeoutInjectorTimeoutAlreadyExists(t *testing.T) {
	injector := funky.NewTimeoutInjector()

	body := funky.Request{
		Context: map[string]interface{}{
			"timeout": 0,
		},
	}

	err := injector.Inject(&body)

	assert.Error(t, err)
}

func TestTimeoutInjectorEmptyConstructor(t *testing.T) {
	injector := funky.NewTimeoutInjector()

	body := funky.Request{
		Context: map[string]interface{}{},
	}

	err := injector.Inject(&body)

	expected := funky.Request{
		Context: map[string]interface{}{
			"timeout": 0,
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, body)
}

func TestTimeoutInjectorSingleTimeout(t *testing.T) {
	injector := funky.NewTimeoutInjector(5000)

	body := funky.Request{
		Context: map[string]interface{}{},
	}

	err := injector.Inject(&body)

	expected := funky.Request{
		Context: map[string]interface{}{
			"timeout": 5000,
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, body)
}

func TestTimeoutInjectorMoreThanOneTimeout(t *testing.T) {
	injector := funky.NewTimeoutInjector(5000, 6000, 7000)

	body := funky.Request{
		Context: map[string]interface{}{},
	}

	err := injector.Inject(&body)

	expected := funky.Request{
		Context: map[string]interface{}{
			"timeout": 5000,
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, body)
}

func TestRequestMetadataInjectorRequestAlreadyExists(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	injector := funky.NewRequestMetadataInjector()
	injector.SetHTTPRequest(req)

	body := funky.Request{
		Context: map[string]interface{}{
			"request": map[string]string{},
		},
	}

	err := injector.Inject(&body)

	assert.Errorf(t, err, "Injector should have failed with 'Context[\"request\"] already exists, instead got %+v", err)
}

func TestRequestMetadataInjectorRequestSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	injector := funky.NewRequestMetadataInjector()
	injector.SetHTTPRequest(req)

	body := funky.Request{
		Context: map[string]interface{}{},
	}

	err := injector.Inject(&body)

	expected := funky.Request{
		Context: map[string]interface{}{
			"request": map[string]interface{}{
				"uri":    req.RequestURI,
				"method": req.Method,
				"header": req.Header,
			},
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, body)
}
