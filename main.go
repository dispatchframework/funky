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
	"strconv"

	"github.com/dispatchframework/funky/pkg/funky"
)

const (
	serversEnvVar   = "SERVERS"
	serverCmdEnvVar = "SERVER_CMD"
	portEnvVar      = "PORT"
)

type funkyHandler struct {
	router funky.Router
}

func (f funkyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body funky.Request
	err := json.NewDecoder(r.Body).Decode(&body)
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

	resp, _ := f.router.Delegate(&body)

	json.NewEncoder(w).Encode(resp)
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
	numServers, err := strconv.Atoi(os.Getenv(serversEnvVar))
	if err != nil {
		log.Fatalf("Unable to parse %s environment variable", serversEnvVar)
	}
	if numServers < 1 {
		numServers = 1
	}

	serverCmd := os.Getenv(serverCmdEnvVar)
	serverFactory, err := funky.NewDefaultServerFactory(serverCmd)
	if err != nil {
		log.Fatal("Too few arguments to server command.")
	}

	router, err := funky.NewRouter(numServers, serverFactory)
	if err != nil {
		log.Fatalf("Failed creating new router: %+v", err)
	}

	handler := funkyHandler{
		router: router,
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
