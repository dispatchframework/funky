///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestNewRouterSuccess(t *testing.T) {
	server := new(ServerMock)
	server.On("Start").Return(nil)

	serverFactory := new(ServerFactoryMock)
	serverFactory.On("CreateServer", uint16(FirstPort)).Return(server, nil)

	_, err := NewRouter(1, serverFactory)

	if err != nil {
		t.Errorf("Failed to construct DefaultRouter: %+v", err)
	}
}

func TestNewRouterZeroServers(t *testing.T) {
	serverFactory := new(ServerFactoryMock)

	_, err := NewRouter(0, serverFactory)

	if err == nil {
		t.Fatal("NewRouter should fail with IllegalArgumentError when numServers is zero")
	}

	if _, ok := err.(IllegalArgumentError); !ok {
		t.Errorf("NewRouter failed with %s, but expecting IllegalArgumentError", err.Error())
	}
}

func TestNewRouterFailCreatingServer(t *testing.T) {
	serverFactory := new(ServerFactoryMock)
	serverFactory.On("CreateServer", uint16(FirstPort)).Return(nil, IllegalArgumentError("port"))

	_, err := NewRouter(1, serverFactory)

	if err == nil {
		t.Errorf("Should have failed creating new router when server errored")
	}
}

func TestNewRouterFailStartingServer(t *testing.T) {
	server := new(ServerMock)
	server.On("Start").Return(errors.New("Failed to start server"))
	serverFactory := new(ServerFactoryMock)
	serverFactory.On("CreateServer", uint16(FirstPort)).Return(server, nil)

	_, err := NewRouter(1, serverFactory)

	server.AssertCalled(t, "Start")

	if err == nil {
		t.Errorf("Should have failed creating router when server start errored")
	}
}

func TestDelegateSuccess(t *testing.T) {
	server := new(ServerMock)
	server.On("Start").Return(nil)
	server.On("IsIdle").Return(true)
	server.On("SetIdle", mock.AnythingOfType("bool")).Return().Return()
	server.On("Invoke", map[string]interface{}{}).Return(nil, nil)
	server.On("Stdout").Return(&bytes.Buffer{})
	server.On("Stderr").Return(&bytes.Buffer{})

	serverFactory := new(ServerFactoryMock)
	serverFactory.On("CreateServer", uint16(FirstPort)).Return(server, nil)

	router, _ := NewRouter(1, serverFactory)

	_, err := router.Delegate(map[string]interface{}{})

	if err != nil {
		t.Error("Received unexpected error calling Delegate")
	}
}

func TestRouterShutdownSuccess(t *testing.T) {
	server := new(ServerMock)
	server.On("Start").Return(nil)
	server.On("Shutdown").Return(nil)

	serverFactory := new(ServerFactoryMock)
	serverFactory.On("CreateServer", uint16(FirstPort)).Return(server, nil)

	router, _ := NewRouter(1, serverFactory)
	err := router.Shutdown()

	if err != nil {
		t.Errorf("Failed to Shutdown servers with error: %+v", err)
	}
}

func TestRouterShutdownFailure(t *testing.T) {
	server := new(ServerMock)
	server.On("Start").Return(nil)
	server.On("Shutdown").Return(errors.New("failed to shutdown server"))

	serverFactory := new(ServerFactoryMock)
	serverFactory.On("CreateServer", uint16(FirstPort)).Return(server, nil)

	router, _ := NewRouter(1, serverFactory)

	err := router.Shutdown()

	if err == nil {
		t.Error("Should have to shutdown router")
	}
}