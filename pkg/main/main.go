package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

const (
	serversEnvVar   = "SERVERS"
	serverCmdEnvVar = "SERVER_CMD"
)

// Constants indicating the types of errors for Dispatch function invocation
const (
	InputError    = "InputError"
	FunctionError = "FunctionError"
	SystemError   = "SystemError"
)

// Response a struct to hold the response to a Dispatch function invocation
type Response struct {
	Context Context     `json:"context"`
	Payload interface{} `json:"payload"`
}

// Context a struct to hold the context of a Dispatch function invocation
type Context struct {
	Error Error `json:"error"`
	Logs  Logs  `json:"logs"`
}

// Error a struct to hold the error status of a Dispatch function invocation
type Error struct {
	ErrorType  string   `json:"error"`
	Message    string   `json:"message"`
	Stacktrace []string `json:"stacktrace"`
}

// Logs a struct to hold the logs of a Dispatch function invocation
type Logs struct {
	Stdout []string `json:"stdout"`
	Stderr []string `json:"stderr"`
}

type FunkyHandler struct {
	router Router
}

func (f FunkyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
					ErrorType: InputError,
					Message:   "Invalid Input",
				},
			},
		}
		out, _ := json.Marshal(resp)
		fmt.Fprintf(w, string(out))
		return
	}

	resp, _ := f.router.Delegate(body)

	out, _ := json.Marshal(resp)

	fmt.Fprintf(w, string(out))
}

func main() {
	numServers, err := strconv.Atoi(os.Getenv(serversEnvVar))
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to parse %s environment variable", serversEnvVar))
	}
	if numServers < 1 {
		numServers = 1
	}

	serverCmd := os.Getenv(serverCmdEnvVar)
	router := NewRouter(numServers, serverCmd)
	handler := FunkyHandler{
		router: router,
	}
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		router.Shutdown()
		server.Shutdown(context.TODO())
		os.Exit(0)
	}()

	server.ListenAndServe()
}
