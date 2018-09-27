///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"

	"github.com/dispatchframework/funky/pkg/funky"
)

const (
	serversEnvVar   = "SERVERS"
	serverCmdEnvVar = "SERVER_CMD"
	portEnvVar      = "PORT"
	timeoutEnvVar   = "TIMEOUT"
	secretsEnvVar   = "SECRETS"
)

type funkyHandler struct {
	router      funky.Router
	transformer funky.RequestTransformer
	rw          funky.HTTPReaderWriter
}

func (f funkyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body *funky.Request
	body, err := f.transformer.Transform(r)
	if err != nil {
		resp := funky.Message{
			Context: &funky.Context{
				Error: &funky.Error{
					ErrorType: funky.InputError,
					Message:   fmt.Sprintf("Invalid Input: %s", err),
				},
			},
		}
		out, _ := json.Marshal(resp)
		fmt.Fprintf(w, string(out))
		return
	}

	resp, _ := f.router.Delegate(body)

	accept := r.Header.Get("Accept")
	matched, err := regexp.Match("(application|\\*)\\/(json|\\*)|^$", []byte(accept))
	if matched {
		accept = "application/json"
	}

	err = f.rw.Write(resp, accept, w)
	if err != nil {
		fmt.Fprintf(w, "Unsupported Accept type: %s", accept)
	}
}

func healthy(c <-chan struct{}) bool {
	select {
	case <-c:
		return false
	default:
		return true
	}
}

func main() {
	numServers, err := envVarGetInt(serversEnvVar, 1)
	if err != nil {
		log.Fatalf("Unable to parse %s environment variable to int", serversEnvVar)
	}
	if numServers < 1 {
		numServers = 1
	}

	serverCmd := os.Getenv(serverCmdEnvVar)
	serverFactory, err := funky.NewDefaultServerFactory(serverCmd)
	if err != nil {
		log.Fatal("Too few arguments to server command.")
	}

	funcTimeout, err := envVarGetInt(timeoutEnvVar, 0)
	if err != nil {
		log.Fatalf("Unable to parse %s environment variable to int", timeoutEnvVar)
	}
	if funcTimeout < 0 {
		funcTimeout = 0
	}

	secrets := strings.Split(os.Getenv(secretsEnvVar), ",")

	rw := funky.NewDefaultHTTPReaderWriter(
		funky.NewJSONHTTPMessageConverter(),
		funky.NewYAMLHTTPMessageConverter(),
		funky.NewBase64HTTPMessageConverter(),
		funky.NewPlainTextHTTPMessageConverter())

	injectors := []funky.ContextInjector{funky.NewTimeoutInjector(funcTimeout),
		funky.NewEnvVarSecretInjector(secrets...), funky.NewRequestMetadataInjector()}
	reqTransformer := funky.NewDefaultRequestTransformer(rw, injectors)

	router, err := funky.NewRouter(numServers, serverFactory)
	if err != nil {
		log.Fatalf("Failed creating new router: %+v", err)
	}

	handler := funkyHandler{
		router:      router,
		transformer: reqTransformer,
		rw:          rw,
	}

	servMux := http.NewServeMux()
	servMux.Handle("/", handler)
	servMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if !healthy(funky.Healthy) {
			w.WriteHeader(500)
		}

		w.Write([]byte("{}"))
	})

	port := os.Getenv(portEnvVar)
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: servMux,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		server.Shutdown(context.TODO())
		router.Shutdown()
		os.Exit(0)
	}()

	server.ListenAndServe()
}

func envVarGetInt(envVar string, orElse int) (int, error) {
	if os.Getenv(envVar) == "" {
		return orElse, nil
	}

	return strconv.Atoi(os.Getenv(envVar))
}
