package main

import (
	"fmt"
)

type IllegalArgumentError string

func (e IllegalArgumentError) Error() string {
	return fmt.Sprintf("The argument is illegal or inappropriate: %s", e)
}
