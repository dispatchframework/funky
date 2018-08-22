package funky_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/dispatchframework/funky/pkg/funky"
	"github.com/ghodss/yaml"
)

func TestReadNoConverters(t *testing.T) {
	rw := funky.NewDefaultHTTPReaderWriter()

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Content-Type", "application/json")

	var out map[string]interface{}
	err := rw.Read(&out, req)

	if err == nil {
		t.Fatal("Should have failed reading with no configured converters")
	}

	if err != funky.UnsupportedMediaTypeError("application/json") {
		t.Errorf("Expected UnsupportedMediaTypeError, got: %+v", err)
	}
}

func TestWriteNoConverters(t *testing.T) {
	rw := funky.NewDefaultHTTPReaderWriter()

	res := httptest.NewRecorder()

	var out map[string]interface{}
	err := rw.Write(&out, "application/json", res)

	if err == nil {
		t.Fatal("Should have failed writing with no configured converters")
	}

	if err != funky.UnsupportedMediaTypeError("application/json") {
		t.Errorf("Expected UnsupportedMediaTypeError, got: %+v", err)
	}
}

func TestReadUnsupportedContentType(t *testing.T) {
	rw := funky.NewDefaultHTTPReaderWriter(funky.NewJSONHTTPMessageConverter(), funky.NewYAMLHTTPMessageConverter(), funky.NewBase64HTTPMessageConverter(), funky.NewPlainTextHTTPMessageConverter())

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Content-Type", "application/xml")

	var out map[string]interface{}
	err := rw.Read(&out, req)

	if err == nil {
		t.Fatal("Should have failed reading Content-Type: application/xml")
	}

	if err != funky.UnsupportedMediaTypeError("application/xml") {
		t.Errorf("Expected UnsupportedMediaTypeError got: %+v", err)
	}
}

func TestWriteUnsupportedContentType(t *testing.T) {
	rw := funky.NewDefaultHTTPReaderWriter(funky.NewJSONHTTPMessageConverter(), funky.NewYAMLHTTPMessageConverter(), funky.NewBase64HTTPMessageConverter(), funky.NewPlainTextHTTPMessageConverter())

	var out map[string]interface{}

	res := httptest.NewRecorder()
	err := rw.Write(&out, "application/xml", res)

	if err == nil {
		t.Fatal("Should have failed writing Content-Type: application/xml")
	}

	if err != funky.UnsupportedMediaTypeError("application/xml") {
		t.Errorf("Expected UnsupportedMediaTypeError got: %+v", err)
	}
}

func TestJsonReader(t *testing.T) {
	rw := funky.NewDefaultHTTPReaderWriter(funky.NewJSONHTTPMessageConverter(), funky.NewYAMLHTTPMessageConverter(), funky.NewBase64HTTPMessageConverter(), funky.NewPlainTextHTTPMessageConverter())

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
	var out map[string]interface{}
	err := rw.Read(&out, req)

	if err != nil {
		t.Errorf("Error extracting request: %s", err)
	}

	if !reflect.DeepEqual(user, out) {
		t.Errorf("The extracted value is: %+v", out)
	}
}

func TestJsonWriter(t *testing.T) {
	rw := funky.NewDefaultHTTPReaderWriter(funky.NewJSONHTTPMessageConverter(), funky.NewYAMLHTTPMessageConverter(), funky.NewBase64HTTPMessageConverter(), funky.NewPlainTextHTTPMessageConverter())

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
	rw := funky.NewDefaultHTTPReaderWriter(funky.NewJSONHTTPMessageConverter(), funky.NewYAMLHTTPMessageConverter(), funky.NewBase64HTTPMessageConverter(), funky.NewPlainTextHTTPMessageConverter())

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
	var out map[string]interface{}
	err = rw.Read(&out, req)

	if err != nil {
		t.Errorf("Error extracting request: %s", err)
	}

	if !reflect.DeepEqual(user, out) {
		t.Errorf("The extracted value is: %+v", out)
	}
}

func TestYamlWriter(t *testing.T) {
	rw := funky.NewDefaultHTTPReaderWriter(funky.NewJSONHTTPMessageConverter(), funky.NewYAMLHTTPMessageConverter(), funky.NewBase64HTTPMessageConverter(), funky.NewPlainTextHTTPMessageConverter())

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
	rw := funky.NewDefaultHTTPReaderWriter(funky.NewJSONHTTPMessageConverter(), funky.NewYAMLHTTPMessageConverter(), funky.NewBase64HTTPMessageConverter(), funky.NewPlainTextHTTPMessageConverter())

	in := "This is my plain text string body"
	buffer := bytes.NewBufferString(in)
	req := httptest.NewRequest("GET", "/", buffer)
	req.Header.Set("Content-Type", "text/plain")

	var out string
	err := rw.Read(&out, req)

	if err != nil {
		t.Errorf("Error extracting request: %s", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Errorf("The extracted value is: %+v", out)
	}
}

func TestPlainTextWriter(t *testing.T) {
	rw := funky.NewDefaultHTTPReaderWriter(funky.NewJSONHTTPMessageConverter(), funky.NewYAMLHTTPMessageConverter(), funky.NewBase64HTTPMessageConverter(), funky.NewPlainTextHTTPMessageConverter())

	body := "This is the plain text to write"

	w := httptest.NewRecorder()

	if err := rw.Write(&body, "text/plain", w); err != nil {
		t.Errorf("Failed encoding body: %s", err)
	}

	if w.Body.String() != body {
		t.Errorf("Encoded body does not match expected. Encoded: %s", w.Body.String())
	}

}

func TestBase64Reader(t *testing.T) {
	rw := funky.NewDefaultHTTPReaderWriter(funky.NewJSONHTTPMessageConverter(), funky.NewYAMLHTTPMessageConverter(), funky.NewBase64HTTPMessageConverter(), funky.NewPlainTextHTTPMessageConverter())

	in := []byte{1, 2, 3, 4, 5, 6, 1, 1}
	inputBuffer := bytes.NewBuffer(in)
	outputBuffer := &bytes.Buffer{}
	writer := base64.NewEncoder(base64.StdEncoding, outputBuffer)
	inputBuffer.WriteTo(writer)
	writer.Close()

	req := httptest.NewRequest("GET", "/", outputBuffer)
	req.Header.Set("Content-Type", "application/base64")

	var out []byte
	err := rw.Read(&out, req)

	if err != nil {
		t.Errorf("Error extracting request: %s", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Errorf("The extracted value is: %+v", out)
	}
}

func TestBase64Writer(t *testing.T) {
	rw := funky.NewDefaultHTTPReaderWriter(funky.NewJSONHTTPMessageConverter(), funky.NewYAMLHTTPMessageConverter(), funky.NewBase64HTTPMessageConverter(), funky.NewPlainTextHTTPMessageConverter())

	body := []byte{0, 1, 2, 3, 4, 5, 6, 7}

	w := httptest.NewRecorder()

	if err := rw.Write(&body, "application/base64", w); err != nil {
		t.Errorf("Failed encoding body: %s", err)
	}

	in := bytes.NewBuffer(body)
	out := bytes.Buffer{}
	in.WriteTo(base64.NewEncoder(base64.StdEncoding, &out))
	if !reflect.DeepEqual(w.Body.Bytes(), out.Bytes()) {
		t.Errorf("Encoded body does not match expected. Encoded: %v", w.Body.Bytes())
	}
}
