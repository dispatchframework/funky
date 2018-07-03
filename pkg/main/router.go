package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"sync"
)

// Router an interface for delegating function invocations to idle servers
type Router interface {
	Delegate(input map[string]interface{}) (*Response, error)
}

// DefaultRouter a struct that hold servers that can be delegated to
type DefaultRouter struct {
	servers []Server
	mutex   *sync.Mutex
}

// NewRouter constructor for DefaultRouters
func NewRouter(servers []Server) *DefaultRouter {
	return &DefaultRouter{
		servers: servers,
		mutex:   &sync.Mutex{},
	}
}

// Delegate delegates function invocation to an idle server
func (r *DefaultRouter) Delegate(input map[string]interface{}) (*Response, error) {
	server, _ := r.findFreeServer()
	defer r.releaseServer(server)

	var e Error
	resp, err := server.Invoke(input)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		if _, ok := err.(TimeoutError); ok {
			e = Error{
				ErrorType: FunctionError,
				Message:   "Invocation exceeded the timeout",
			}
		} else if _, ok := err.(BadRequestError); ok {
			e = Error{
				ErrorType: InputError,
				Message:   "Input invalid",
			}
		} else if _, ok := err.(InvocationError); ok {
			e = Error{
				ErrorType: FunctionError,
				Message:   "Failed invoking function.",
			}
		} else if _, ok := err.(InvalidResponsePayloadError); ok {
			e = Error{
				ErrorType: InputError,
				Message:   "Failed serialzing function return return",
			}
		} else if _, ok := err.(UnknownSystemError); ok {
			e = Error{
				ErrorType: SystemError,
				Message:   "An unknown error occured",
			}
		} else {
			e = Error{
				ErrorType: SystemError,
				Message:   "Something went wrong.",
			}
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

func (r *DefaultRouter) findFreeServer() (Server, error) {
	var ret Server
	for _, server := range servers {
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
		return nil, errors.New("no free server")
	}

	return ret, nil
}

func (r *DefaultRouter) releaseServer(server Server) {
	r.mutex.Lock()
	server.SetIdle(true)
	r.mutex.Unlock()
}

func splitLogsOnNewline(stdBuffer *bytes.Buffer) []string {
	b, _ := stdBuffer.ReadBytes(0)
	b = bytes.TrimRight(b, "\x0000")
	stdBuffer.UnreadByte()
	return strings.Split(string(b), "\n")
}
