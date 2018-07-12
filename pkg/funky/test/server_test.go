///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

	req := funky.Message{
		Context: &funky.Context{
			Timeout: 0,
		},
	}

	resp, err := server.Invoke(req)

	if err != nil {
		t.Fatal(err)
	}

	buff := &bytes.Buffer{}
	buff.ReadFrom(resp)
	obj := make(map[string]interface{})
	err = json.Unmarshal(buff.Bytes(), &obj)
	if err != nil {
		log.Fatal(err)
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
		respBytes, _ := json.Marshal(resp)
		w.WriteHeader(400)
		fmt.Fprint(w, string(respBytes))
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

	req := funky.Message{
		Context: &funky.Context{},
	}

	_, err = server.Invoke(req)

	if err == nil {
		t.Fatalf("Expected BadRequest error, instead got no error")
	}

	if _, ok := err.(funky.BadRequestError); !ok {
		t.Errorf("Expected BadRequestError got %s", err)
	}
}
