///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package funky

import "time"

// Constants indicating the types of errors for Dispatch function invocation
const (
	InputError    = "InputError"
	FunctionError = "FunctionError"
	SystemError   = "SystemError"
)

// Request a struct to hold the request body sent to a Dispatch function
type Request struct {
	Context map[string]interface{} `json:"context"`
	Payload interface{}            `json:"payload"`
}

// Message a struct to hold the response to a Dispatch function invocation
type Message struct {
	Context *Context    `json:"context"`
	Payload interface{} `json:"payload"`
}

// Context a struct to hold the context of a Dispatch function invocation
type Context struct {
	Error    *Error     `json:"error,omitempty"`
	Logs     *Logs      `json:"logs"`
	Deadline *time.Time `json:"deadline,omitempty"`
}

// Error a struct to hold the error status of a Dispatch function invocation
type Error struct {
	ErrorType  string   `json:"type"`
	Message    string   `json:"message"`
	Stacktrace []string `json:"stacktrace"`
}

// Logs a struct to hold the logs of a Dispatch function invocation
type Logs struct {
	Stdout []string `json:"stdout"`
	Stderr []string `json:"stderr"`
}
