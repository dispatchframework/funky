package funky

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
)

func TestJsonReader(t *testing.T) {
	rw := DefaultHttpReaderWriter{
		converters: []HttpMessageConverter{NewJsonHttpMessageConverter(), NewYamlHttpMessageConverter(), NewBase64HttpMessageConverter(), NewPlainTextHttpMessageConverter()},
	}

	user := map[string]interface{}{
		"firstName": "Jon",
		"lastName":  "Snow",
		"address": map[string]interface{}{
			"place":  "Winterfell",
			"region": "The North",
		},
	}

	buffer := bytes.Buffer{}
	json.NewEncoder(&buffer).Encode(user)
	req := httptest.NewRequest("GET", "/", &buffer)
	req.Header.Set("Content-Type", "application/json")
	out, err := rw.Read(reflect.TypeOf(map[string]interface{}{}), req)

	if err != nil {
		t.Errorf("Error extracting request: %s", err)
	}

	if !reflect.DeepEqual(user, (*out.(*map[string]interface{}))) {
		t.Errorf("The extracted value is: %+v", out)
	}
}

func TestJsonWriter(t *testing.T) {
	rw := DefaultHttpReaderWriter{
		converters: []HttpMessageConverter{NewJsonHttpMessageConverter(), NewYamlHttpMessageConverter(), NewBase64HttpMessageConverter(), NewPlainTextHttpMessageConverter()},
	}

	user := map[string]interface{}{
		"firstName": "Jon",
		"lastName":  "Snow",
		"address": map[string]interface{}{
			"place":  "Winterfell",
			"region": "The North",
		},
	}

	w := httptest.NewRecorder()

	if err := rw.Write(user, "application/json", w); err != nil {
		t.Errorf("Failed encoding body: %s", err)
	}

	expected := bytes.Buffer{}
	json.NewEncoder(&expected).Encode(user)
	if expected.String() != w.Body.String() {
		t.Errorf("Encoded body does not match expected. Encoded: %s", w.Body.String())
	}
}

func TestYamlReader(t *testing.T) {
	rw := DefaultHttpReaderWriter{
		converters: []HttpMessageConverter{NewJsonHttpMessageConverter(), NewYamlHttpMessageConverter(), NewBase64HttpMessageConverter(), NewPlainTextHttpMessageConverter()},
	}

	user := map[string]interface{}{
		"firstName": "Jon",
		"lastName":  "Snow",
		"address": map[string]interface{}{
			"place":  "Winterfell",
			"region": "The North",
		},
	}

	data, err := yaml.Marshal(user)
	if err != nil {
		t.Errorf("Failed marshalling yaml: %s", err)
	}

	buffer := bytes.NewBuffer(data)
	req := httptest.NewRequest("GET", "/", buffer)
	req.Header.Set("Content-Type", "application/yaml")
	out, err := rw.Read(reflect.TypeOf(map[string]interface{}{}), req)

	if err != nil {
		t.Errorf("Error extracting request: %s", err)
	}

	if !reflect.DeepEqual(user, (*out.(*map[string]interface{}))) {
		t.Errorf("The extracted value is: %+v", out)
	}
}

func TestYamlWriter(t *testing.T) {
	rw := DefaultHttpReaderWriter{
		converters: []HttpMessageConverter{NewJsonHttpMessageConverter(), NewYamlHttpMessageConverter(), NewBase64HttpMessageConverter(), NewPlainTextHttpMessageConverter()},
	}

	user := map[string]interface{}{
		"firstName": "Jon",
		"lastName":  "Snow",
		"address": map[string]interface{}{
			"place":  "Winterfell",
			"region": "The North",
		},
	}

	w := httptest.NewRecorder()

	if err := rw.Write(user, "application/yaml", w); err != nil {
		t.Errorf("Failed encoding body: %s", err)
	}

	data, err := yaml.Marshal(user)
	if err != nil {
		t.Errorf("Failed marshaling user: %s", err)
	}

	expected := bytes.NewBuffer(data)
	if w.Body.String() != expected.String() {
		t.Errorf("Encoded body does not match expected. Encoded: %s", w.Body.String())
	}
}

func TestPlainTextReader(t *testing.T) {
	rw := DefaultHttpReaderWriter{
		converters: []HttpMessageConverter{NewJsonHttpMessageConverter(), NewYamlHttpMessageConverter(), NewBase64HttpMessageConverter(), NewPlainTextHttpMessageConverter()},
	}

	in := "This is my plain text string body"
	buffer := bytes.NewBufferString(in)
	req := httptest.NewRequest("GET", "/", buffer)
	req.Header.Set("Content-Type", "text/plain")

	out, err := rw.Read(reflect.TypeOf(""), req)

	if err != nil {
		t.Errorf("Error extracting request: %s", err)
	}

	if !reflect.DeepEqual(in, out.(string)) {
		t.Errorf("The extracted value is: %+v", out)
	}
}

func TestPlainTextWriter(t *testing.T) {
	rw := DefaultHttpReaderWriter{
		converters: []HttpMessageConverter{NewJsonHttpMessageConverter(), NewYamlHttpMessageConverter(), NewBase64HttpMessageConverter(), NewPlainTextHttpMessageConverter()},
	}

	body := "This is the plain text to write"

	w := httptest.NewRecorder()

	if err := rw.Write(body, "text/plain", w); err != nil {
		t.Errorf("Failed encoding body: %s", err)
	}

	if w.Body.String() != body {
		t.Errorf("Encoded body does not match expected. Encoded: %s", w.Body.String())
	}

}

func TestBase64Reader(t *testing.T) {
	rw := DefaultHttpReaderWriter{
		converters: []HttpMessageConverter{NewJsonHttpMessageConverter(), NewYamlHttpMessageConverter(), NewBase64HttpMessageConverter(), NewPlainTextHttpMessageConverter()},
	}

	in := []byte{1, 2, 3, 4, 5, 6, 1, 1}
	inputBuffer := bytes.NewBuffer(in)
	outputBuffer := &bytes.Buffer{}
	writer := base64.NewEncoder(base64.StdEncoding, outputBuffer)
	inputBuffer.WriteTo(writer)
	writer.Close()

	req := httptest.NewRequest("GET", "/", outputBuffer)
	req.Header.Set("Content-Type", "application/base64")

	out, err := rw.Read(reflect.TypeOf([]byte{}), req)

	if err != nil {
		t.Errorf("Error extracting request: %s", err)
	}

	if !reflect.DeepEqual(in, out.([]byte)) {
		t.Errorf("The extracted value is: %+v", out)
	}
}

func TestBase64Writer(t *testing.T) {
	rw := DefaultHttpReaderWriter{
		converters: []HttpMessageConverter{NewJsonHttpMessageConverter(), NewYamlHttpMessageConverter(), NewBase64HttpMessageConverter(), NewPlainTextHttpMessageConverter()},
	}

	body := []byte{0, 1, 2, 3, 4, 5, 6, 7}

	w := httptest.NewRecorder()

	if err := rw.Write(body, "application/base64", w); err != nil {
		t.Errorf("Failed encoding body: %s", err)
	}

	in := bytes.NewBuffer(body)
	out := bytes.Buffer{}
	in.WriteTo(base64.NewEncoder(base64.StdEncoding, &out))
	if !reflect.DeepEqual(w.Body.Bytes(), out.Bytes()) {
		t.Errorf("Encoded body does not match expected. Encoded: %v", w.Body.Bytes())
	}
}
