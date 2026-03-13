package main

import (
	"errors"
	"fmt"
)

var Error1 = errors.New("Error1")
var Error2 = errors.New("Error2")

func main() {
	wrapped := fmt.Errorf("%w, %w", Error1, Error2)
	type UnwrapIntoSingle = interface{ Unwrap() error }

	if wrapped, ok := wrapped.(UnwrapIntoSingle); ok && wrapped.Unwrap() != nil {

	}

	fmt.Println(wrapped, errors.Is(wrapped, Error1), errors.Is(wrapped, Error2))
}
