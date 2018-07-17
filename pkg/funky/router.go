///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package funky

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/semaphore"
)

// FirstPort the starting port number for servers created by a Router
const FirstPort uint16 = 9000

// Router an interface for delegating function invocations to idle servers
type Router interface {
	Delegate(input Message) (*Message, error)
	Shutdown() error
}

// DefaultRouter a struct that hold servers that can be delegated to
type DefaultRouter struct {
	servers       []Server
	serverFactory ServerFactory
	mutex         *sync.Mutex
	sem           *semaphore.Weighted
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
		servers:       servers,
		serverFactory: serverFactory,
		mutex:         &sync.Mutex{},
		sem:           semaphore.NewWeighted(int64(numServers)),
	}, nil
}

// Delegate delegates function invocation to an idle server
func (r *DefaultRouter) Delegate(input Message) (*Message, error) {
	server, err := r.findFreeServer()
	if err != nil {
		return nil, err
	}

	defer func() {
		r.releaseServer(server)
	}()

	var e Error
	resp, err := server.Invoke(input)
	if resp != nil {
		defer resp.Close()
	}

	logs := Logs{
		Stdout: server.Stdout(),
		Stderr: server.Stderr(),
	}

	if err != nil {
		switch err.(type) {
		case TimeoutError:
			terminateErr := server.Terminate()
			newServer, serverErr := r.serverFactory.CreateServer(server.GetPort())
			if serverErr != nil || terminateErr != nil {
				Healthy <- false
			} else {
				if newServer.Start() != nil {
					Healthy <- false
				}
				server = newServer
			}
			e = Error{
				ErrorType: FunctionError,
				Message:   err.Error(),
			}
		case InvocationError:
			e = Error{
				ErrorType: FunctionError,
				Message:   err.Error(),
			}
		case BadRequestError, InvalidResponsePayloadError:
			e = Error{
				ErrorType: InputError,
				Message:   err.Error(),
			}
		default:
			e = Error{
				ErrorType: SystemError,
				Message:   err.Error(),
			}
		}
	}

	context := Context{
		Logs:  &logs,
		Error: &e,
	}

	respBuf := new(bytes.Buffer)
	if resp != nil {
		respBuf.ReadFrom(resp)
	}
	respPayload := make(map[string]interface{})
	json.Unmarshal(respBuf.Bytes(), &respPayload)

	response := &Message{
		Context: &context,
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
	if err := r.sem.Acquire(context.TODO(), 1); err != nil {
		return nil, err
	}

	// if we're here, it's guaranteed we have at least one element in servers

	r.mutex.Lock()
	defer r.mutex.Unlock()

	server := r.servers[len(r.servers)-1]
	r.servers = r.servers[:len(r.servers)-1]

	return server, nil
}

func (r *DefaultRouter) releaseServer(server Server) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if server != nil {
		r.servers = append(r.servers, server)
	}
	r.sem.Release(1)
}
