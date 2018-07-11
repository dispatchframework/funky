///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sync/semaphore"
)

// FirstPort the starting port number for servers created by a Router
const FirstPort uint16 = 9000

// Router an interface for delegating function invocations to idle servers
type Router interface {
	Delegate(input map[string]interface{}) (*Response, error)
	Shutdown() error
}

// DefaultRouter a struct that hold servers that can be delegated to
type DefaultRouter struct {
	servers []Server
	mutex   *sync.Mutex
	sem     *semaphore.Weighted
}

// NewRouter constructor for DefaultRouters
func NewRouter(numServers int, serverFactory ServerFactory) (*DefaultRouter, error) {
	if numServers < 1 {
		return nil, IllegalArgumentError("numServers")
	}

	servers, err := createServers(numServers, serverFactory)
	if err != nil {
		return nil, err
	}

	return &DefaultRouter{
		servers: servers,
		mutex:   &sync.Mutex{},
		sem:     semaphore.NewWeighted(int64(numServers)),
	}, nil
}

// Delegate delegates function invocation to an idle server
func (r *DefaultRouter) Delegate(input map[string]interface{}) (*Response, error) {
	server, err := r.findFreeServer()
	if err != nil {
		return nil, err
	}

	defer r.releaseServer(server)

	var e Error
	resp, err := server.Invoke(input)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		switch err.(type) {
		case TimeoutError:
			e = Error{
				ErrorType: FunctionError,
				Message:   err.Error(),
			}
			break
		case BadRequestError:
			e = Error{
				ErrorType: InputError,
				Message:   err.Error(),
			}
			break
		case InvocationError:
			e = Error{
				ErrorType: FunctionError,
				Message:   err.Error(),
			}
			break
		case InvalidResponsePayloadError:
			e = Error{
				ErrorType: InputError,
				Message:   err.Error(),
			}
			break
		case UnknownSystemError:
			e = Error{
				ErrorType: SystemError,
				Message:   err.Error(),
			}
			break
		default:
			e = Error{
				ErrorType: SystemError,
				Message:   err.Error(),
			}
			break
		}
	}

	logs := Logs{
		Stdout: splitLogsOnNewline(server.Stdout()),
		Stderr: splitLogsOnNewline(server.Stderr()),
	}

	context := Context{
		Logs:  logs,
		Error: e,
	}

	respBuf := new(bytes.Buffer)
	if resp != nil {
		respBuf.ReadFrom(resp)
	}
	respPayload := make(map[string]interface{})
	json.Unmarshal(respBuf.Bytes(), &respPayload)

	response := &Response{
		Context: context,
		Payload: respPayload,
	}

	return response, nil
}

// Shutdown shuts down the servers managed by this router
func (r *DefaultRouter) Shutdown() error {
	var err error
	for _, server := range r.servers {
		err = server.Shutdown()
	}

	if err != nil {
		return errors.New("Failed to shutdown one or more servers")
	}

	return nil
}

func createServers(numServers int, serverFactory ServerFactory) ([]Server, error) {
	servers := []Server{}
	for i := uint16(0); i < uint16(numServers); i++ {
		server, err := serverFactory.CreateServer(FirstPort + i)
		if err != nil {
			return nil, fmt.Errorf("Failed creating server on port %d", FirstPort+i)
		}
		if err := server.Start(); err != nil {
			return nil, fmt.Errorf("Failed to start server %+v", err)
		}
		servers = append(servers, server)
	}

	return servers, nil
}

func (r *DefaultRouter) findFreeServer() (Server, error) {
	var ret Server

	r.sem.Acquire(context.TODO(), 1)
	for _, server := range r.servers {
		r.mutex.Lock()
		if server.IsIdle() {
			server.SetIdle(false)
			ret = server
		}
		r.mutex.Unlock()
		if ret != nil {
			break
		}
	}

	if ret == nil {
		r.sem.Release(1)
		return nil, errors.New("no free server")
	}

	return ret, nil
}

func (r *DefaultRouter) releaseServer(server Server) {
	r.mutex.Lock()
	server.SetIdle(true)
	r.mutex.Unlock()
	r.sem.Release(1)
}

func splitLogsOnNewline(stdBuffer *bytes.Buffer) []string {
	b, _ := stdBuffer.ReadBytes(0)
	b = bytes.TrimRight(b, "\x0000")
	stdBuffer.UnreadByte()
	return strings.Split(string(b), "\n")
}
