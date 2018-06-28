package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Server interface {
	IsIdle() bool
	SetIdle(bool)
	GetPort() uint16
	GetCmd() *exec.Cmd
	StderrPipe() io.ReadCloser
	StdoutPipe() io.ReadCloser
	Start() error
}

type ServerImpl struct {
	isIdle     bool
	port       uint16
	cmd        *exec.Cmd
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

func (s *ServerImpl) GetCmd() *exec.Cmd {
	return s.cmd
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

func (s *ServerImpl) Wait() error {
	return s.cmd.Wait()
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
		stdoutPipe: stdoutPipe,
		stderrPipe: stderrPipe,
	}, nil
}
