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
)

type funkyHandler struct {
	router funky.Router
}

func (f funkyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body funky.Message
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		resp := funky.Message{
			Context: &funky.Context{
				Error: &funky.Error{
					ErrorType: funky.InputError,
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
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
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
