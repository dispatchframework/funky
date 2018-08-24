///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package funky

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// RequestTransformer general interface for transforming http requests into Dispatch requests
type RequestTransformer interface {
	Transform(req *http.Request) (*Request, error)
}

// DefaultRequestTransformer transforms a request into a format suitable for Dispatch language servers
type DefaultRequestTransformer struct {
	timeout int
	secrets []string
	rw      HTTPReaderWriter
}

// NewDefaultRequestTransformer constructs a new DefaultRequestTransformer
func NewDefaultRequestTransformer(timeout int, secrets []string, rw HTTPReaderWriter) RequestTransformer {
	return &DefaultRequestTransformer{
		timeout: timeout,
		secrets: secrets,
		rw:      rw,
	}
}

// Transform transforms an http request into a Dispatch request.
func (r *DefaultRequestTransformer) Transform(req *http.Request) (*Request, error) {
	var body Request

	// populate function timeout
	ctx := map[string]interface{}{
		"timeout": r.timeout,
	}

	// populate secrets
	if len(r.secrets) != 0 {

		secrets := map[string]string{}
		for _, secretName := range r.secrets {
			path := fmt.Sprintf("/run/secrets/%s", secretName)

			f, err := os.Open(path)
			if os.IsNotExist(err) {
				return nil, UnknownSystemError("file does not exist")
			}
			var data map[string]string
			json.NewDecoder(f).Decode(&data)

			for key, value := range data {
				secrets[key] = string(value)
			}
		}

		ctx["secrets"] = secrets
	}

	// populate request attributes
	ctx["request"] = map[string]interface{}{
		"uri":    req.RequestURI,
		"method": req.Method,
		"header": req.Header,
	}

	body.Context = ctx

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
	case "application/json":
		fallthrough
	case "application/yaml":
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
