package funky

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/ghodss/yaml"
)

type HttpReaderWriter interface {
	Read(t reflect.Type, req *http.Request) (interface{}, error)
	Write(body interface{}, contentType string, w http.ResponseWriter) error
}

type DefaultHttpReaderWriter struct {
	converters []HttpMessageConverter
}

func NewDefaultHttpReaderWriter(converters ...HttpMessageConverter) *DefaultHttpReaderWriter {
	return &DefaultHttpReaderWriter{
		converters: converters,
	}
}

func (rw *DefaultHttpReaderWriter) Read(t reflect.Type, req *http.Request) (interface{}, error) {
	for _, v := range rw.converters {
		if v.CanRead(t, req.Header.Get("Content-Type")) {
			return v.Read(t, req)
		}
	}

	return nil, errors.New("No converter supports the type and request")
}

func (rw *DefaultHttpReaderWriter) Write(body interface{}, contentType string, w http.ResponseWriter) error {
	for _, v := range rw.converters {
		if v.CanWrite(reflect.TypeOf(body), contentType) {
			return v.Write(body, contentType, w)
		}
	}

	return errors.New("No converter supports the type and body")
}

type HttpMessageConverter interface {
	CanRead(t reflect.Type, mediaType string) bool
	CanWrite(t reflect.Type, mediaType string) bool
	Read(t reflect.Type, req *http.Request) (interface{}, error)
	Write(body interface{}, contentType string, res http.ResponseWriter) error
}

type JsonHttpMessageConverter struct {
	supportedMediaTypes    []string
	supportedMediaTypesMap map[string]bool
}

func NewJsonHttpMessageConverter() *JsonHttpMessageConverter {
	supportedMediaTypes := []string{"application/json"}
	supportedMediaTypesMap := map[string]bool{}
	for _, v := range supportedMediaTypes {
		supportedMediaTypesMap[v] = true
	}
	return &JsonHttpMessageConverter{
		supportedMediaTypes:    supportedMediaTypes,
		supportedMediaTypesMap: supportedMediaTypesMap,
	}
}

func (j *JsonHttpMessageConverter) CanRead(t reflect.Type, mediaType string) bool {
	return j.supportedMediaTypesMap[mediaType]
}

func (j *JsonHttpMessageConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return j.supportedMediaTypesMap[mediaType]
}

func (j *JsonHttpMessageConverter) Read(t reflect.Type, req *http.Request) (interface{}, error) {
	e := reflect.New(t)
	err := json.NewDecoder(req.Body).Decode(e.Interface())
	if err != nil {
		return nil, err
	}

	return e.Interface(), nil
}

func (j *JsonHttpMessageConverter) Write(body interface{}, contentType string, res http.ResponseWriter) error {
	res.Header().Set("Content-Type", contentType)
	return json.NewEncoder(res).Encode(body)
}

type YamlHttpMessageConverter struct {
	supportedMediaTypes    []string
	supportedMediaTypesMap map[string]bool
}

func NewYamlHttpMessageConverter() *YamlHttpMessageConverter {
	supportedMediaTypes := []string{"application/yaml"}
	supportedMediaTypesMap := map[string]bool{}
	for _, v := range supportedMediaTypes {
		supportedMediaTypesMap[v] = true
	}
	return &YamlHttpMessageConverter{
		supportedMediaTypes:    supportedMediaTypes,
		supportedMediaTypesMap: supportedMediaTypesMap,
	}
}

func (y *YamlHttpMessageConverter) CanRead(t reflect.Type, mediaType string) bool {
	return y.supportedMediaTypesMap[mediaType]
}

func (y *YamlHttpMessageConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return y.supportedMediaTypesMap[mediaType]
}

func (y *YamlHttpMessageConverter) Read(t reflect.Type, req *http.Request) (interface{}, error) {
	e := reflect.New(t)

	buffer := bytes.Buffer{}
	if _, err := buffer.ReadFrom(req.Body); err != nil {
		return nil, err
	}

	return e.Interface(), yaml.Unmarshal(buffer.Bytes(), e.Interface())
}

func (y *YamlHttpMessageConverter) Write(body interface{}, contentType string, res http.ResponseWriter) error {
	res.Header().Set("Content-Type", contentType)
	data, err := yaml.Marshal(body)
	if err != nil {
		return err
	}

	_, err = res.Write(data)
	return err
}

type Base64HttpMessageConverter struct {
	supportedMediaTypes    []string
	supportedMediaTypesMap map[string]bool
}

func NewBase64HttpMessageConverter() *Base64HttpMessageConverter {
	supportedMediaTypes := []string{"application/base64"}
	supportedMediaTypesMap := map[string]bool{}
	for _, v := range supportedMediaTypes {
		supportedMediaTypesMap[v] = true
	}
	return &Base64HttpMessageConverter{
		supportedMediaTypes:    supportedMediaTypes,
		supportedMediaTypesMap: supportedMediaTypesMap,
	}
}

func (b *Base64HttpMessageConverter) CanRead(t reflect.Type, mediaType string) bool {
	return b.supportedMediaTypesMap[mediaType] && t == reflect.TypeOf([]byte{})
}

func (b *Base64HttpMessageConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return b.supportedMediaTypesMap[mediaType] && t == reflect.TypeOf([]byte{})
}

func (b *Base64HttpMessageConverter) Read(t reflect.Type, req *http.Request) (interface{}, error) {
	if !b.CanRead(t, req.Header.Get("Content-Type")) {
		return nil, errors.New("UnsupportedMediaType")
	}

	buffer := bytes.Buffer{}
	reader := base64.NewDecoder(base64.StdEncoding, req.Body)
	if _, err := buffer.ReadFrom(reader); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (b *Base64HttpMessageConverter) Write(body interface{}, contentType string, res http.ResponseWriter) error {
	if !b.CanWrite(reflect.TypeOf(body), contentType) {
		return errors.New("UnsupportedMediaType")
	}

	writer := base64.NewEncoder(base64.StdEncoding, res)

	_, err := bytes.NewBuffer(body.([]byte)).WriteTo(writer)

	return err
}

type PlainTextHttpMessageConverter struct {
	supportedMediaTypes    []string
	supportedMediaTypesMap map[string]bool
}

func NewPlainTextHttpMessageConverter() *PlainTextHttpMessageConverter {
	supportedMediaTypes := []string{"text/plain"}
	supportedMediaTypesMap := map[string]bool{}
	for _, v := range supportedMediaTypes {
		supportedMediaTypesMap[v] = true
	}
	return &PlainTextHttpMessageConverter{
		supportedMediaTypes:    supportedMediaTypes,
		supportedMediaTypesMap: supportedMediaTypesMap,
	}
}

func (p *PlainTextHttpMessageConverter) CanRead(t reflect.Type, mediaType string) bool {
	return p.supportedMediaTypesMap[mediaType] && t == reflect.TypeOf("")
}

func (p *PlainTextHttpMessageConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return p.supportedMediaTypesMap[mediaType] && t == reflect.TypeOf("")
}

func (p *PlainTextHttpMessageConverter) Read(t reflect.Type, req *http.Request) (interface{}, error) {
	if !p.CanRead(t, req.Header.Get("Content-Type")) {
		return nil, errors.New("UnsupportedMediaType")
	}

	buffer := bytes.Buffer{}
	if _, err := buffer.ReadFrom(req.Body); err != nil {
		return nil, err
	}

	return buffer.String(), nil
}

func (p *PlainTextHttpMessageConverter) Write(body interface{}, contentType string, res http.ResponseWriter) error {
	if !p.CanWrite(reflect.TypeOf(body), contentType) {
		return errors.New("UnsupportedMediaType")
	}

	buffer := bytes.NewBufferString(body.(string))

	_, err := buffer.WriteTo(res)
	return err
}
