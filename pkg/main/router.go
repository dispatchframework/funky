package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
)

type Router interface {
	Delegate(input map[string]interface{}) (*Response, error)
}

type RouterImpl struct {
	servers []Server
	mutex   *sync.Mutex
}

func NewRouter(servers []Server) *RouterImpl {
	return &RouterImpl{
		servers: servers,
		mutex:   &sync.Mutex{},
	}
}

func (r *RouterImpl) Delegate(input map[string]interface{}) (*Response, error) {
	server, _ := r.findFreeServer()
	defer r.releaseServer(server)

	var stdoutBuf, stderrBuf bytes.Buffer

	stdoutPipe := server.StdoutPipe()
	stdoutCh := copy(stdoutPipe)
	stdoutInterrupt := make(chan []byte)
	write(&stdoutBuf, stdoutCh, stdoutInterrupt)

	stderrPipe := server.StderrPipe()
	stderrCh := copy(stderrPipe)
	stderrInterrupt := make(chan []byte)
	write(&stderrBuf, stderrCh, stderrInterrupt)

	var e Error
	resp, err := server.Invoke(input)
	if resp != nil {
		defer resp.Close()
	}
	if err != nil {
		if _, ok := err.(TimeoutError); ok {
			e = Error{
				ErrorType: FUNCTION_ERROR,
				Message:   "Invocation exceeded the timeout",
			}
		} else if _, ok := err.(BadRequestError); ok {
			e = Error{
				ErrorType: INPUT_ERROR,
				Message:   "Input invalid",
			}
		} else if _, ok := err.(InvocationError); ok {
			e = Error{
				ErrorType: FUNCTION_ERROR,
				Message:   "Failed invoking function.",
			}
		} else {
			e = Error{
				ErrorType: SYSTEM_ERROR,
				Message:   "Something went wrong.",
			}
		}
	}

	close(stdoutInterrupt)
	close(stderrInterrupt)
	close(stdoutCh)
	close(stderrCh)
	logs := Logs{
		Stdout: splitLogsOnNewline(&stdoutBuf),
		Stderr: splitLogsOnNewline(&stderrBuf),
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

func (r *RouterImpl) findFreeServer() (Server, error) {
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

func (r *RouterImpl) releaseServer(server Server) {
	r.mutex.Lock()
	server.SetIdle(true)
	r.mutex.Unlock()
}

func splitLogsOnNewline(stdBuffer *bytes.Buffer) []string {
	return strings.Split(stdBuffer.String(), "\n")
}

func copy(src io.Reader) chan []byte {
	var p = make([]byte, 512)
	var ch = make(chan []byte)
	var num int
	var err error

	go func() {
		for {
			select {
			case <-ch:
				return
			case ch <- p[:num]:
				num, err = src.Read(p)

				if num == 0 && err != nil {
					return
				}
			}
		}
	}()

	return ch
}

func write(dst io.Writer, ch chan []byte, interrupt chan []byte) {

	go func() {
		for {
			select {
			case p := <-ch:
				dst.Write(p)
			case <-interrupt:
				return
			}
		}
	}()
}
