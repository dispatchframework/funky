///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package funky

import (
	"net/http"
)

// RequestTransformer general interface for transforming http requests into Dispatch requests
type RequestTransformer interface {
	Transform(req *http.Request) (*Request, error)
}

// DefaultRequestTransformer transforms a request into a format suitable for Dispatch language servers
type DefaultRequestTransformer struct {
	rw        HTTPReaderWriter
	injectors []ContextInjector
}

// NewDefaultRequestTransformer constructs a new DefaultRequestTransformer
func NewDefaultRequestTransformer(rw HTTPReaderWriter, injectors []ContextInjector) RequestTransformer {
	return &DefaultRequestTransformer{
		rw:        rw,
		injectors: injectors,
	}
}

// Transform transforms an http request into a Dispatch request.
func (r *DefaultRequestTransformer) Transform(req *http.Request) (*Request, error) {
	var body Request
	body.Context = map[string]interface{}{}

	for _, v := range r.injectors {
		if reqAware, ok := v.(HTTPRequestAware); ok {
			reqAware.SetHTTPRequest(req)
		}

		if err := v.Inject(&body); err != nil {
			return nil, err
		}
	}

	var err error
	// populate payload
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}

	switch contentType {
	case "application/base64":
		var payload []byte
		err = r.rw.Read(&payload, req)
		body.Payload = payload
	case "text/plain":
		var payload string
		err = r.rw.Read(&payload, req)
		body.Payload = payload
	case "application/json", "application/yaml":
		var payload map[string]interface{}
		err = r.rw.Read(&payload, req)
		body.Payload = payload
	default:
		err = UnsupportedMediaTypeError(contentType)
	}

	if err != nil {
		return nil, err
	}

	return &body, nil
}
