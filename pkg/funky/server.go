package funky

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

// Server an interface for managing function servers
type Server interface {
	IsIdle() bool
	SetIdle(bool)
	GetPort() uint16
	Invoke(input map[string]interface{}) (io.ReadCloser, error)
	Stdout() *bytes.Buffer
	Stderr() *bytes.Buffer
	Start() error
	Shutdown() error
}

// DefaultServer a struct to hold information about running servers
type DefaultServer struct {
	isIdle bool
	port   uint16
	cmd    *exec.Cmd
	client *http.Client
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

// NewServer returns a new DefaultServer with the given port and command
func NewServer(port uint16, cmd *exec.Cmd) (*DefaultServer, error) {
	if port < 1024 {
		return nil, IllegalArgumentError("port")
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port))

	stdoutBuf, stderrBuf := &bytes.Buffer{}, &bytes.Buffer{}

	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	return &DefaultServer{
		isIdle: true,
		port:   port,
		cmd:    cmd,
		client: &http.Client{},
		stdout: stdoutBuf,
		stderr: stderrBuf,
	}, nil
}

type ServerFactory interface {
	CreateServer(port uint16) (Server, error)
}

type DefaultServerFactory struct {
	serverCmd string
}

func NewDefaultServerFactory(serverCmd string) ServerFactory {
	return &DefaultServerFactory{
		serverCmd: serverCmd,
	}
}

func (f *DefaultServerFactory) CreateServer(port uint16) (Server, error) {
	cmds := strings.Split(f.serverCmd, " ")
	cmd := exec.Command(cmds[0], cmds[1:]...)
	return NewServer(port, cmd)
}

// IsIdle indicates whether this server is currently idle or processing a request
func (s *DefaultServer) IsIdle() bool {
	return s.isIdle
}

// SetIdle sets whether this server is currently idle or processing a request
func (s *DefaultServer) SetIdle(idle bool) {
	s.isIdle = idle
}

// GetPort returns the port this server is running on
func (s *DefaultServer) GetPort() uint16 {
	return s.port
}

// Invoke calls the server with the given input to invoke a Dispatch function
func (s *DefaultServer) Invoke(input map[string]interface{}) (io.ReadCloser, error) {
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

	switch resp.StatusCode {
	case 400:
		return nil, BadRequestError("invalid input")
	case 500:
		return nil, InvocationError(resp.StatusCode)
	case 422:
		return nil, InvalidResponsePayloadError("")
	case 502:
		return nil, UnknownSystemError("")
	}

	return resp.Body, nil
}

// Stdout returns the Buffer containing stdout
func (s *DefaultServer) Stdout() *bytes.Buffer {
	return s.stdout
}

// Stderr returns the Buffer containing stderr
func (s *DefaultServer) Stderr() *bytes.Buffer {
	return s.stderr
}

// Start starts the server
func (s *DefaultServer) Start() error {
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
