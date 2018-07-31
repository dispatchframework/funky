///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package funky

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Server an interface for managing function servers
type Server interface {
	GetPort() uint16
	Invoke(input *Message) (interface{}, error)
	Stdout() []string
	Stderr() []string
	Start() error
	Shutdown() error
	Terminate() error
}

// DefaultServer a struct to hold information about running servers
type DefaultServer struct {
	port   uint16
	cmd    *exec.Cmd
	client *http.Client

	lock   sync.RWMutex
	stdout []string
	stderr []string
}

// NewServer returns a new DefaultServer with the given port and command
func NewServer(port uint16, cmd *exec.Cmd) (*DefaultServer, error) {
	if port < 1024 {
		return nil, IllegalArgumentError("port")
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port))

	return &DefaultServer{
		port:   port,
		cmd:    cmd,
		client: &http.Client{},
	}, nil
}

// ServerFactory an interface for creating new Servers decoupled from the concrete implementation
type ServerFactory interface {
	CreateServer(port uint16) (Server, error)
}

// DefaultServerFactory concrete implementation of ServerFactory.
type DefaultServerFactory struct {
	cmd  string
	args []string
}

// NewDefaultServerFactory a DefaultServerFactory constructor; validates the server command.
func NewDefaultServerFactory(serverCmd string) (ServerFactory, error) {
	cmds := strings.Fields(serverCmd)

	if len(cmds) < 1 {
		return nil, IllegalArgumentError(serverCmd)
	}

	return &DefaultServerFactory{
		cmd:  cmds[0],
		args: cmds[1:],
	}, nil
}

// CreateServer creates a new server by initiating a Command with the given port and preconfigured server command
func (f *DefaultServerFactory) CreateServer(port uint16) (Server, error) {
	cmd := exec.Command(f.cmd, f.args...)
	return NewServer(port, cmd)
}

// GetPort returns the port this server is running on
func (s *DefaultServer) GetPort() uint16 {
	return s.port
}

// Invoke calls the server with the given input to invoke a Dispatch function
func (s *DefaultServer) Invoke(input *Message) (interface{}, error) {
	p, err := json.Marshal(input)

	timeout := time.Duration(0)
	if input.Context.Deadline != nil {
		timeout = time.Until(*input.Context.Deadline)
	}

	if timeout < 0 {
		return nil, TimeoutError("Did not invoke, already exceeded timeout")
	}

	s.client.Timeout = timeout

	s.resetStreams()

	url := fmt.Sprintf("http://127.0.0.1:%d", s.GetPort())
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(p))
	if err == nil {
		defer resp.Body.Close()
	}

	if err != nil {
		if isTimeout(err) {
			return nil, TimeoutError("Function execution exceeded the timeout")
		} else if isConnectionRefused(err) {
			return nil, ConnectionRefusedError(url)
		} else {
			return nil, UnknownSystemError(err.Error())
		}
	}

	if resp.StatusCode >= 400 {
		var e Error
		json.NewDecoder(resp.Body).Decode(&e)
		return nil, FunctionServerError{
			APIError: e,
		}
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, InvalidResponsePayloadError(err.Error())
	}

	return result, nil
}

// Stdout returns the Buffer containing stdout
func (s *DefaultServer) Stdout() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.stdout
}

// Stderr returns the Buffer containing stderr
func (s *DefaultServer) Stderr() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.stderr
}

func (s *DefaultServer) resetStreams() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.stdout = nil
	s.stderr = nil
}

func scanStream(r io.Reader, ss *[]string, l sync.Locker) {
	s := bufio.NewScanner(r)
	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, io.EOF
		}

		for i := 0; i < len(data); i++ {
			if data[i] == '\n' {
				return i + 1, data[:i], nil
			}
		}

		return len(data), data, nil
	})
	for s.Scan() {
		func() {
			l.Lock()
			defer l.Unlock()
			*ss = append(*ss, s.Text())
		}()
	}
}

// Start starts the server
func (s *DefaultServer) Start() error {

	stdout, _ := s.cmd.StdoutPipe()
	stderr, _ := s.cmd.StderrPipe()

	go scanStream(stdout, &s.stdout, &s.lock)
	go scanStream(stderr, &s.stderr, &s.lock)

	return s.cmd.Start()
}

// Shutdown shuts down the server, kills it if necessary
func (s *DefaultServer) Shutdown() error {
	err := s.cmd.Wait()
	if err != nil {
		return s.cmd.Process.Kill()
	}

	return err
}

// Terminate kills the server without waiting for a graceful shutdown.
func (s *DefaultServer) Terminate() error {
	return s.cmd.Process.Kill()
}

func isTimeout(err error) bool {
	type hasTimeout interface {
		Timeout() bool
	}
	e, ok := err.(hasTimeout)
	return ok && e.Timeout()
}

func isConnectionRefused(err error) bool {
	return strings.Contains(err.Error(), "connection refused")
}
