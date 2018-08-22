package funky_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dispatchframework/funky/pkg/funky"
	"github.com/dispatchframework/funky/pkg/funky/mocks"
	"github.com/stretchr/testify/mock"
)

func TestTransformSuccess(t *testing.T) {
	client := mocks.SecretInterface{}

	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}
	rw := mocks.HTTPReaderWriter{}
	rw.On("Read", mock.AnythingOfType("*map[string]interface {}"), mock.AnythingOfType("*http.Request")).
		Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*map[string]interface{})
			*arg = payload
		})

	transformer := funky.NewDefaultRequestTransformer(0, []string{}, &client, &rw)

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(payload)
	req := httptest.NewRequest("GET", "/", &body)
	req.Header.Set("Content-Type", "application/json")
	actual, _ := transformer.Transform(req)

	client.AssertNotCalled(t, "List")

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

func TestTransformWithSecrets(t *testing.T) {
	secrets := []string{"postgres_pwd"}

	secretList := v1.SecretList{
		Items: []v1.Secret{
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "postgres_pwd",
				},
				Data: map[string][]byte{
					"username": []byte("white_rabbit"),
					"password": []byte("im_l8_im_l8"),
				},
			},
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "unknown",
				},
				Data: map[string][]byte{
					"username": []byte("Dr. No"),
					"password": []byte("Spectre"),
				},
			},
		},
	}
	client := mocks.SecretInterface{}
	client.On("List", metav1.ListOptions{}).Return(&secretList, nil)

	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}
	rw := mocks.HTTPReaderWriter{}
	rw.On("Read", mock.AnythingOfType("*map[string]interface {}"), mock.AnythingOfType("*http.Request")).
		Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*map[string]interface{})
			*arg = payload
		})
	transformer := funky.NewDefaultRequestTransformer(0, secrets, &client, &rw)

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(payload)
	req := httptest.NewRequest("GET", "/", &body)
	req.Header.Set("Content-Type", "application/json")
	actual, _ := transformer.Transform(req)

	expected := &funky.Request{
		Context: map[string]interface{}{
			"timeout": 0,
			"secrets": map[string]string{
				"username": "white_rabbit",
				"password": "im_l8_im_l8",
			},
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

func TestTransformWithConflictingSecrets(t *testing.T) {
	secrets := []string{"postgres_pwd", "evil_lair"}

	secretList := v1.SecretList{
		Items: []v1.Secret{
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "postgres_pwd",
				},
				Data: map[string][]byte{
					"username": []byte("white_rabbit"),
					"password": []byte("im_l8_im_l8"),
				},
			},
			v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "evil_lair",
				},
				Data: map[string][]byte{
					"username": []byte("Dr. No"),
					"password": []byte("Spectre"),
				},
			},
		},
	}
	client := mocks.SecretInterface{}
	client.On("List", metav1.ListOptions{}).Return(&secretList, nil)

	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}
	rw := mocks.HTTPReaderWriter{}
	rw.On("Read", mock.AnythingOfType("*map[string]interface {}"), mock.AnythingOfType("*http.Request")).
		Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*map[string]interface{})
			*arg = payload
		})
	transformer := funky.NewDefaultRequestTransformer(0, secrets, &client, &rw)

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(payload)
	req := httptest.NewRequest("GET", "/", &body)
	req.Header.Set("Content-Type", "application/json")
	actual, _ := transformer.Transform(req)

	expected := &funky.Request{
		Context: map[string]interface{}{
			"timeout": 0,
			"secrets": map[string]string{
				"username": "Dr. No",
				"password": "Spectre",
			},
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

func TestTransformListSecretError(t *testing.T) {
	secrets := []string{"postgres_pwd"}
	client := mocks.SecretInterface{}
	client.On("List", metav1.ListOptions{}).Return(nil, errors.New("Failed retrieving secret"))

	rw := mocks.HTTPReaderWriter{}
	transformer := funky.NewDefaultRequestTransformer(0, secrets, &client, &rw)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	_, err := transformer.Transform(req)

	assert.Error(t, err, "Expected error on call to Transform")
}

func TestTransformMissingContentType(t *testing.T) {
	client := mocks.SecretInterface{}

	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}
	rw := mocks.HTTPReaderWriter{}
	rw.On("Read", mock.AnythingOfType("*map[string]interface {}"), mock.AnythingOfType("*http.Request")).
		Return(nil).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).(*map[string]interface{})
			*arg = payload
		})

	transformer := funky.NewDefaultRequestTransformer(0, []string{}, &client, &rw)

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(payload)
	req := httptest.NewRequest("GET", "/", &body)
	actual, _ := transformer.Transform(req)

	client.AssertNotCalled(t, "List")

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

func TestTransformUnsupportedContentType(t *testing.T) {
	client := mocks.SecretInterface{}

	payload := map[string]interface{}{
		"name":  "Jon",
		"place": "Winterfell",
	}
	rw := mocks.HTTPReaderWriter{}
	rw.On("Read", mock.AnythingOfType("*map[string]interface {}"), mock.AnythingOfType("*http.Request")).
		Return(funky.UnsupportedMediaTypeError("application/xml"))

	transformer := funky.NewDefaultRequestTransformer(0, []string{}, &client, &rw)

	var body bytes.Buffer
	xml.NewEncoder(&body).Encode(payload)
	req := httptest.NewRequest("GET", "/", &body)
	req.Header.Set("Content-Type", "application/xml")
	_, err := transformer.Transform(req)

	client.AssertNotCalled(t, "List")

	assert.Error(t, err, "Expected Transform to fail with UnsupportedMediaTypeError, instead got %+v", err)
}
