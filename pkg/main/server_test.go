package main

import (
	"os/exec"
	"testing"
)

func TestNewServerSuccess(t *testing.T) {
	var port uint16 = 9090
	server, err := NewServer(port, exec.Command("echo"))
	if err != nil {
		t.Errorf("Failed to construct server with error %v", err.Error())
	}

	if server.GetPort() != port {
		t.Errorf("Failed to return expected port %v", port)
	}
}

func TestNewServerInvalidPort(t *testing.T) {
	var port uint16 = 80
	_, err := NewServer(port, exec.Command("echo"))
	if _, ok := err.(IllegalArgumentError); !ok {
		t.Errorf("Should have errored when provided a port less than 1024. Port value: %v", port)
	}
}
