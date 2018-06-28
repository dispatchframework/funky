package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
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
	p, err := json.Marshal(input)
	if err != nil {
		// doSomething
	}
	server, err := r.findFreeServer()
	defer r.releaseServer(server)

	var stdoutBuf, stderrBuf bytes.Buffer

	stdoutInterrupt := make(chan bool)
	stdoutPipe := server.StdoutPipe()
	go copy(&stdoutBuf, stdoutPipe, stdoutInterrupt)

	//stderrInterrupt := make(chan bool)
	//stderrPipe := server.StderrPipe()
	//go copy(&stderrBuf, stderrPipe, stderrInterrupt)

	ctx := input["context"].(map[string]interface{})
	timeout := 0.0
	if ctx["timeout"] != nil {
		timeout = ctx["timeout"].(float64)
	}

	var client http.Client
	if timeout != 0 {
		client = http.Client{
			Timeout: time.Duration(timeout) * time.Millisecond,
		}
	} else {
		client = http.Client{}
	}

	resp, err := client.Post(fmt.Sprintf("http://0.0.0.0:%d", server.GetPort()), "application/json", bytes.NewBuffer(p))

	var errors Error

	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout") {
			errors = Error{
				ErrorType: "FUNCTION_EXCEPTION",
				Message:   "Function exceeded timeout",
			}
		}
	} else {
		defer resp.Body.Close()
	}

	stdoutInterrupt <- true
	//stderrInterrupt <- true
	logs := Logs{
		Stdout: splitLogsOnNewline(&stdoutBuf),
		Stderr: splitLogsOnNewline(&stderrBuf),
	}

	context := Context{
		Logs:  logs,
		Error: errors,
	}

	respBuf := new(bytes.Buffer)
	if err == nil {
		respBuf.ReadFrom(resp.Body)
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

func copy(dst io.Writer, src io.Reader, interrupt chan bool) {
	var p = make([]byte, 512)

	for {
		select {
		case <-interrupt:
			return
		default:
			num, err := src.Read(p)

			if num == 0 && err != nil {
				return
			}

			dst.Write(p[:num])
			runtime.Gosched()
		}
		runtime.Gosched()
	}
}
