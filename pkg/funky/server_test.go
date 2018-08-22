///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package funky_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/dispatchframework/funky/pkg/funky"
)

func TestNewServerSuccess(t *testing.T) {
	var port uint16 = 9090
	server, err := funky.NewServer(port, exec.Command("echo"))
	if err != nil {
		t.Fatalf("Failed to construct server with error %v", err.Error())
	}

	if server.GetPort() != port {
		t.Errorf("Failed to return expected port %v", port)
	}
}

func TestNewServerInvalidPort(t *testing.T) {
	var port uint16 = 80
	_, err := funky.NewServer(port, exec.Command("echo"))
	if _, ok := err.(funky.IllegalArgumentError); !ok {
		t.Errorf("Should have errored when provided a port less than 1024. Port value: %v", port)
	}
}

func TestInvokeSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"myField": "Hello, Jon from Winterfell",
		}
		respBytes, _ := json.Marshal(resp)
		fmt.Fprint(w, string(respBytes))
	}))
	defer ts.Close()

	urlParts := strings.Split(ts.URL, ":")
	port, err := strconv.Atoi(urlParts[len(urlParts)-1])
	if err != nil {
		t.Fatalf("Couldn't convert port %s", urlParts[len(urlParts)-1])
	}
	server, err := funky.NewServer(uint16(port), exec.Command("echo"))
	if err != nil {
		t.Fatalf("Failed to create new server: %+v", err)
	}

	req := funky.Request{
		Context: map[string]interface{}{},
	}

	resp, err := server.Invoke(&req)

	if err != nil {
		t.Fatal(err)
	}

	obj, ok := resp.(map[string]interface{})
	if !ok {
		t.Fatal("resp not of type map[string]interface")
	}

	if _, ok := obj["myField"]; !ok {
		t.Fatalf("Was expecting key 'myField' in return map.")
	}
}

func TestInvokeBadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"myField": "Hello, Jon from Winterfell",
		}
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	urlParts := strings.Split(ts.URL, ":")
	port, err := strconv.Atoi(urlParts[len(urlParts)-1])
	if err != nil {
		t.Fatalf("Could not convert port %s", urlParts[len(urlParts)-1])
	}

	server, err := funky.NewServer(uint16(port), exec.Command("echo"))
	if err != nil {
		t.Fatalf("Failed to create new server: %+v", err)
	}

	req := funky.Request{
		Context: map[string]interface{}{},
	}

	_, err = server.Invoke(&req)

	if err == nil {
		t.Fatalf("Expected BadRequest error, instead got no error")
	}

	if _, ok := err.(funky.FunctionServerError); !ok {
		t.Errorf("Expected FunctionServerError got %s", err)
	}
}

func TestInvokeContextPassthrough(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req funky.Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			t.Errorf("Failed parsing json payload: %s", err)
		}

		resp := map[string]interface{}{
			"myField": "Hello, Jon from Winterfell",
			"context": req.Context,
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	urlParts := strings.Split(ts.URL, ":")
	port, err := strconv.Atoi(urlParts[len(urlParts)-1])
	if err != nil {
		t.Fatalf("Could not convert port %s", urlParts[len(urlParts)-1])
	}

	server, err := funky.NewServer(uint16(port), exec.Command("echo"))
	if err != nil {
		t.Fatalf("Failed to create new server: %+v", err)
	}

	context := map[string]interface{}{
		"unknown": "field",
	}

	req := funky.Request{
		Context: context,
	}

	resp, err := server.Invoke(&req)

	if err != nil {
		t.Fatalf("Failed to invoke function. Expected success")
	}

	if result, ok := resp.(map[string]interface{}); ok {
		if obj, ok := result["context"]; ok {
			if ctx, ok := obj.(map[string]interface{}); ok {
				if unknown, ok := ctx["unknown"]; !ok || unknown != "field" {
					t.Errorf("Did not properly pass along context value. Received: %s", unknown)
				}
			} else {
				t.Errorf("Expected context to be a map[string]interface")
			}
		} else {
			t.Errorf("Expected context key not found")
		}
	} else {
		t.Errorf("Result from invoke was not a map[string]interface{} like expected")
	}
}
