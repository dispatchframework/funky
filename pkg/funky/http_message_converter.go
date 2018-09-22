///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package funky

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
)

func getTypeFromFragment(mediaType string) string {
	fragments := strings.Split(mediaType, "+")
	if len(fragments) == 2 {
		// Use the identifier as the content type
		mediaType = fragments[1]
	}
	return mediaType
}

// HTTPReaderWriter Describes an interface that can read arbitary objects from an http request and write arbitary objects to an http response
type HTTPReaderWriter interface {
	Read(result interface{}, req *http.Request) error
	Write(body interface{}, contentType string, w http.ResponseWriter) error
}

// DefaultHTTPReaderWriter Default implementation for HttpReaderWriter
type DefaultHTTPReaderWriter struct {
	converters []HTTPMessageConverter
}

// NewDefaultHTTPReaderWriter constructs a new DefaultHttpReaderWriter with the given HttpMessageConverters
func NewDefaultHTTPReaderWriter(converters ...HTTPMessageConverter) *DefaultHTTPReaderWriter {
	return &DefaultHTTPReaderWriter{
		converters: converters,
	}
}

// Read Tries to read the request into the object passed as result. result should be a pointer to the data. If no configured HttpMessageConverter can read the request this will return an UnsupportedMediaTypeError
func (rw *DefaultHTTPReaderWriter) Read(result interface{}, req *http.Request) error {
	contentType := req.Header.Get("Content-Type")
	for _, v := range rw.converters {
		if v.CanRead(reflect.TypeOf(result), contentType) {
			return v.Read(result, req)
		}
	}

	return UnsupportedMediaTypeError(contentType)
}

// Write Tries to write the body into the ResponseWrite. If no configured HttpMessageConverter can write the response this will return an UnsupportedMediaTypeError
func (rw *DefaultHTTPReaderWriter) Write(body interface{}, contentType string, w http.ResponseWriter) error {
	for _, v := range rw.converters {
		if v.CanWrite(reflect.TypeOf(body), contentType) {
			return v.Write(body, contentType, w)
		}
	}

	return UnsupportedMediaTypeError(contentType)
}

// HTTPMessageConverter a generic interface for converting an http request into a given type, or writing a given type into an http response
type HTTPMessageConverter interface {
	CanRead(t reflect.Type, mediaType string) bool
	CanWrite(t reflect.Type, mediaType string) bool
	Read(result interface{}, req *http.Request) error
	Write(body interface{}, contentType string, res http.ResponseWriter) error
}

// JSONHTTPMessageConverter an HTTPMessageConverter for reading and writing json encoded data
type JSONHTTPMessageConverter struct {
	supportedMediaTypes    []string
	supportedMediaTypesMap map[string]bool
}

// NewJSONHTTPMessageConverter constructs a JSONHTTPMessageConverter, sets supported media type to "application/json"
func NewJSONHTTPMessageConverter() *JSONHTTPMessageConverter {
	supportedMediaTypes := []string{"*/*", "application/json", "json"}
	supportedMediaTypesMap := map[string]bool{}
	for _, v := range supportedMediaTypes {
		supportedMediaTypesMap[v] = true
	}
	return &JSONHTTPMessageConverter{
		supportedMediaTypes:    supportedMediaTypes,
		supportedMediaTypesMap: supportedMediaTypesMap,
	}
}

// CanRead returns true if this converter can convert input described in the given media type into the given type, false otherwise
func (j *JSONHTTPMessageConverter) CanRead(t reflect.Type, mediaType string) bool {
	mediaType = getTypeFromFragment(mediaType)
	return j.supportedMediaTypesMap[mediaType]
}

// CanWrite returns true if this converter can convert the given type into the given media type, false otherwise
func (j *JSONHTTPMessageConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return j.supportedMediaTypesMap[mediaType]
}

func (j *JSONHTTPMessageConverter) Read(result interface{}, req *http.Request) error {
	err := json.NewDecoder(req.Body).Decode(result)
	if err != nil {
		return err
	}

	return nil
}

func (j *JSONHTTPMessageConverter) Write(body interface{}, contentType string, res http.ResponseWriter) error {
	res.Header().Set("Content-Type", contentType)
	return json.NewEncoder(res).Encode(body)
}

// YAMLHTTPMessageConverter an HTTPMessageConverter for reading and writing yaml encoded data
type YAMLHTTPMessageConverter struct {
	supportedMediaTypes    []string
	supportedMediaTypesMap map[string]bool
}

// NewYAMLHTTPMessageConverter constructs a YAMLHTTPMessageConverter, sets supported media type to "application/yaml"
func NewYAMLHTTPMessageConverter() *YAMLHTTPMessageConverter {
	supportedMediaTypes := []string{"*/*", "application/yaml", "yaml"}
	supportedMediaTypesMap := map[string]bool{}
	for _, v := range supportedMediaTypes {
		supportedMediaTypesMap[v] = true
	}
	return &YAMLHTTPMessageConverter{
		supportedMediaTypes:    supportedMediaTypes,
		supportedMediaTypesMap: supportedMediaTypesMap,
	}
}

// CanRead returns true if this converter can convert input described in the given media type into the given type, false otherwise
func (y *YAMLHTTPMessageConverter) CanRead(t reflect.Type, mediaType string) bool {
	mediaType = getTypeFromFragment(mediaType)
	return y.supportedMediaTypesMap[mediaType]
}

// CanWrite returns true if this converter can convert the given type into the given media type, false otherwise
func (y *YAMLHTTPMessageConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return y.supportedMediaTypesMap[mediaType]
}

func (y *YAMLHTTPMessageConverter) Read(result interface{}, req *http.Request) error {
	buffer := bytes.Buffer{}
	if _, err := buffer.ReadFrom(req.Body); err != nil {
		return err
	}

	return yaml.Unmarshal(buffer.Bytes(), result)
}

func (y *YAMLHTTPMessageConverter) Write(body interface{}, contentType string, res http.ResponseWriter) error {
	res.Header().Set("Content-Type", contentType)
	data, err := yaml.Marshal(body)
	if err != nil {
		return err
	}

	_, err = res.Write(data)
	return err
}

// Base64HTTPMessageConverter an HTTPMessageConverter for reading and writing base64 encoded data
type Base64HTTPMessageConverter struct {
	supportedMediaTypes    []string
	supportedMediaTypesMap map[string]bool
}

// NewBase64HTTPMessageConverter constructs a Base64HTTPMessageConverter, sets supported media type to "application/base64"
func NewBase64HTTPMessageConverter() *Base64HTTPMessageConverter {
	supportedMediaTypes := []string{"*/*", "application/base64"}
	supportedMediaTypesMap := map[string]bool{}
	for _, v := range supportedMediaTypes {
		supportedMediaTypesMap[v] = true
	}
	return &Base64HTTPMessageConverter{
		supportedMediaTypes:    supportedMediaTypes,
		supportedMediaTypesMap: supportedMediaTypesMap,
	}
}

// CanRead returns true if this converter can convert input described in the given media type into the given type, false otherwise
func (b *Base64HTTPMessageConverter) CanRead(t reflect.Type, mediaType string) bool {
	return b.supportedMediaTypesMap[mediaType] && t == reflect.TypeOf(&[]byte{})
}

// CanWrite returns true if this converter can convert the given type into the given media type, false otherwise
func (b *Base64HTTPMessageConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return b.supportedMediaTypesMap[mediaType] && t == reflect.TypeOf(&[]byte{})
}

func (b *Base64HTTPMessageConverter) Read(result interface{}, req *http.Request) error {
	contentType := req.Header.Get("Content-Type")
	if !b.CanRead(reflect.TypeOf(result), contentType) {
		return UnsupportedMediaTypeError(contentType)
	}

	buffer := bytes.Buffer{}
	reader := base64.NewDecoder(base64.StdEncoding, req.Body)
	if _, err := buffer.ReadFrom(reader); err != nil {
		return err
	}

	*result.(*[]byte) = buffer.Bytes()
	return nil
}

func (b *Base64HTTPMessageConverter) Write(body interface{}, contentType string, res http.ResponseWriter) error {
	if !b.CanWrite(reflect.TypeOf(body), contentType) {
		return UnsupportedMediaTypeError(contentType)
	}

	writer := base64.NewEncoder(base64.StdEncoding, res)

	_, err := bytes.NewBuffer(*body.(*[]byte)).WriteTo(writer)

	return err
}

// PlainTextHTTPMessageConverter an HTTPMessageConverter for reading and writing plain text data
type PlainTextHTTPMessageConverter struct {
	supportedMediaTypes    []string
	supportedMediaTypesMap map[string]bool
}

// NewPlainTextHTTPMessageConverter constructs a PlainTextHTTPMessageConverter, sets supported media type to "text/plain"
func NewPlainTextHTTPMessageConverter() *PlainTextHTTPMessageConverter {
	supportedMediaTypes := []string{"*/*", "text/plain"}
	supportedMediaTypesMap := map[string]bool{}
	for _, v := range supportedMediaTypes {
		supportedMediaTypesMap[v] = true
	}
	return &PlainTextHTTPMessageConverter{
		supportedMediaTypes:    supportedMediaTypes,
		supportedMediaTypesMap: supportedMediaTypesMap,
	}
}

// CanRead returns true if this converter can convert input described in the given media type into the given type, false otherwise
func (p *PlainTextHTTPMessageConverter) CanRead(t reflect.Type, mediaType string) bool {
	return p.supportedMediaTypesMap[mediaType] && t == reflect.PtrTo(reflect.TypeOf(""))
}

// CanWrite returns true if this converter can convert the given type into the given media type, false otherwise
func (p *PlainTextHTTPMessageConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return p.supportedMediaTypesMap[mediaType] && t == reflect.PtrTo(reflect.TypeOf(""))
}

func (p *PlainTextHTTPMessageConverter) Read(result interface{}, req *http.Request) error {
	contentType := req.Header.Get("Content-Type")
	if !p.CanRead(reflect.TypeOf(result), contentType) {
		return UnsupportedMediaTypeError(contentType)
	}

	buffer := bytes.Buffer{}
	if _, err := buffer.ReadFrom(req.Body); err != nil {
		return err
	}

	*result.(*string) = buffer.String()
	return nil
}

func (p *PlainTextHTTPMessageConverter) Write(body interface{}, contentType string, res http.ResponseWriter) error {
	if !p.CanWrite(reflect.TypeOf(body), contentType) {
		return UnsupportedMediaTypeError(contentType)
	}

	buffer := bytes.NewBufferString(*body.(*string))

	_, err := buffer.WriteTo(res)
	return err
}
