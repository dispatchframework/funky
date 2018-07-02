package main

import (
	"fmt"
)

type IllegalArgumentError string

func (e IllegalArgumentError) Error() string {
	return fmt.Sprintf("The argument is illegal or inappropriate: %s", e)
}

type TimeoutError string

func (e TimeoutError) Error() string {
	return fmt.Sprintf("The invocation exceeded the timeout: %s", e)
}

type BadRequestError string

func (e BadRequestError) Error() string {
	return fmt.Sprintf("The request body is invalid: %+v", e)
}

type InvocationError string

func (e InvocationError) Error() string {
	return fmt.Sprintf("Invocation of the function failed: %s", e)
}
