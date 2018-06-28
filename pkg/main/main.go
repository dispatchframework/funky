package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
)

const serversEnvVar = "SERVERS"
const serverCmdEnvVar = "SERVER_CMD"

var servers []Server
var router Router

func handler(w http.ResponseWriter, r *http.Request) {
	reader := r.Body

	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	rawBody := buf.Bytes()
	var body map[string]interface{}
	err := json.Unmarshal(rawBody, &body)
	if err != nil {
		resp := Response{
			Context: Context{
				Error: Error{
					ErrorType: "INPUT_ERROR",
					Message:   "Invalid Input",
				},
			},
		}
		out, _ := json.Marshal(resp)
		fmt.Fprintf(w, string(out))
	}

	resp, _ := router.Delegate(body)

	out, _ := json.Marshal(resp)

	fmt.Fprintf(w, string(out))
}

func main() {
	numServers, err := strconv.Atoi(os.Getenv(serversEnvVar))
	if err != nil || numServers < 1 {
		numServers = 1
	}

	serverCmd := os.Getenv(serverCmdEnvVar)
	servers, err = createServers(numServers, serverCmd)
	router = NewRouter(servers)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		for _, server := range servers {
			server.GetCmd().Process.Kill()
		}
		os.Exit(0)
	}()
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createServers(numServers int, serverCmd string) ([]Server, error) {
	servers := []Server{}
	for i := uint16(0); i < uint16(numServers); i++ {
		cmds := strings.SplitN(serverCmd, " ", 3)
		cmd := exec.Command(cmds[0], cmds[1:]...)
		server, _ := NewServer(9000+i, cmd)
		if err := server.Start(); err != nil {
			return nil, errors.New("blah")
		}
		servers = append(servers, server)
	}

	return servers, nil
}

type Response struct {
	Context Context     `json:"context"`
	Payload interface{} `json:"payload"`
}

type Context struct {
	Error Error `json:"error"`
	Logs  Logs  `json:"logs"`
}

type Error struct {
	ErrorType  string   `json:"error"`
	Message    string   `json:"message"`
	Stacktrace []string `json:"stacktrace"`
}

type Logs struct {
	Stdout []string `json:"stdout"`
	Stderr []string `json:"stderr"`
}
