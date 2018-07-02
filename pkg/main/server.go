package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Server interface {
	IsIdle() bool
	SetIdle(bool)
	GetPort() uint16
	Invoke(input map[string]interface{}) (io.ReadCloser, error)
	StderrPipe() io.ReadCloser
	StdoutPipe() io.ReadCloser
	Start() error
	Shutdown() error
}

type ServerImpl struct {
	isIdle     bool
	port       uint16
	cmd        *exec.Cmd
	client     *http.Client
	stdoutPipe io.ReadCloser
	stderrPipe io.ReadCloser
}

func (s *ServerImpl) IsIdle() bool {
	return s.isIdle
}

func (s *ServerImpl) SetIdle(idle bool) {
	s.isIdle = idle
}

func (s *ServerImpl) GetPort() uint16 {
	return s.port
}

func (s *ServerImpl) Invoke(input map[string]interface{}) (io.ReadCloser, error) {
	p, err := json.Marshal(input)

	ctx, ok := input["context"].(map[string]interface{})
	if !ok {
		return nil, BadRequestError(p)
	}
	timeout := 0.0
	if ctx["timeout"] != nil {
		timeout = ctx["timeout"].(float64)
	}

	s.client.Timeout = time.Duration(timeout) * time.Millisecond

	resp, err := s.client.Post(fmt.Sprintf("http://0.0.0.0:%d", s.GetPort()), "application/json", bytes.NewBuffer(p))

	if err != nil {
		println(err.Error())
		if strings.Contains(err.Error(), "Client.Timeout") {
			return nil, TimeoutError(fmt.Sprintf("%d", timeout))
		} else if strings.Contains(err.Error(), "connection refused") {
			return nil, errors.New("connection refused")
		}
	}

	if resp.StatusCode == 400 {
		return nil, BadRequestError("invalid input")
	} else if resp.StatusCode == 500 {
		return nil, InvocationError(resp.StatusCode)
	}

	return resp.Body, nil
}

func (s *ServerImpl) StderrPipe() io.ReadCloser {
	return s.stderrPipe
}

func (s *ServerImpl) StdoutPipe() io.ReadCloser {
	return s.stdoutPipe
}

func (s *ServerImpl) Start() error {
	return s.cmd.Start()
}

func (s *ServerImpl) Shutdown() error {
	err := s.cmd.Wait()
	if err != nil {
		return s.cmd.Process.Kill()
	}

	return err
}

func NewServer(port uint16, cmd *exec.Cmd) (*ServerImpl, error) {
	if port < 1024 {
		return nil, IllegalArgumentError("port")
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port))

	// TODO: proper error handling
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	return &ServerImpl{
		isIdle:     true,
		port:       port,
		cmd:        cmd,
		client:     &http.Client{},
		stdoutPipe: stdoutPipe,
		stderrPipe: stderrPipe,
	}, nil
}
