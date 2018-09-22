///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package funky

import (
	"encoding/json"
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
	// No support for chunked encoding presently
	if req.ContentLength > 0 {
		// populate payload
		contentType := req.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/json"
			req.Header.Add("Content-Type", contentType)
		}

		contentType = getTypeFromFragment(contentType)

		switch contentType {
		case "application/base64":
			var payload []byte
			err = r.rw.Read(&payload, req)
			body.Payload = payload
		case "text/plain":
			var payload string
			err = r.rw.Read(&payload, req)
			body.Payload = payload
		case "application/json", "application/yaml", "json", "yaml":
			var payload interface{}
			err = r.rw.Read(&payload, req)
			body.Payload = payload
		default:
			err = UnsupportedMediaTypeError(contentType)
		}
	}

	if err != nil {
		return nil, err
	}

	// All non POST requests must be transformed as the function interface
	// expects a POST.
	switch req.Method {
	case http.MethodGet, http.MethodOptions:
		payload := make(map[string]interface{})
		values := req.URL.Query()
		for k, v := range values {
			// Although HTTP allows for more than one value for a given key
			// we are simply taking the last.  The functions will expect a map.
			if len(v) > 0 {
				val := v[len(v)-1]
				num := json.Number(val)
				if i, err := num.Int64(); err == nil {
					payload[k] = i
					continue
				}
				if f, err := num.Float64(); err == nil {
					payload[k] = f
					continue
				}
				payload[k] = val
			}
		}
		body.Payload = payload
	}

	return &body, nil
}
