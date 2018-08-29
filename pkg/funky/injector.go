///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package funky

import (
	"net/http"
	"os"
)

type ContextInjector interface {
	Inject(req *Request) error
}

type HTTPRequestAware interface {
	SetHTTPRequest(req *http.Request)
}

type EnvVarSecretInjector struct {
	secrets []string
}

func NewEnvVarSecretInjector(secrets ...string) *EnvVarSecretInjector {
	return &EnvVarSecretInjector{
		secrets: secrets,
	}
}

func (i *EnvVarSecretInjector) Inject(req *Request) error {
	if _, ok := req.Context["secrets"]; ok {
		return IllegalArgumentError("Context[\"secrets\"] already exists.")
	}

	secrets := map[string]string{}
	for _, v := range i.secrets {
		secrets[v] = os.Getenv("d_secret_" + v)
	}

	req.Context["secrets"] = secrets

	return nil
}

type TimeoutInjector struct {
	timeout int
}

func NewTimeoutInjector(timeout ...int) *TimeoutInjector {
	var t int
	if len(timeout) == 0 {
		t = 0
	} else {
		t = timeout[0]
	}
	return &TimeoutInjector{
		timeout: t,
	}
}

func (i *TimeoutInjector) Inject(req *Request) error {
	if _, ok := req.Context["timeout"]; ok {
		return IllegalStateError("Context[\"timeout\"] already exists.")
	}

	req.Context["timeout"] = i.timeout

	return nil
}

type RequestMetadataInjector struct {
	r *http.Request
}

func NewRequestMetadataInjector() *RequestMetadataInjector {
	return &RequestMetadataInjector{}
}

func (i *RequestMetadataInjector) Inject(req *Request) error {
	if _, ok := req.Context["request"]; ok {
		return IllegalStateError("Context[\"request\"] already exists.")
	}

	req.Context["request"] = map[string]interface{}{
		"uri":    i.r.RequestURI,
		"method": i.r.Method,
		"header": i.r.Header,
	}

	return nil
}

func (i *RequestMetadataInjector) SetHTTPRequest(req *http.Request) {
	i.r = req
}
