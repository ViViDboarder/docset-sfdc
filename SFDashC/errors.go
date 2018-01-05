package main

import (
	"errors"
	"fmt"
	"log"
)

// NewCustomError creates a custom error using a string as the message
func NewCustomError(message string) error {
	return errors.New(message)
}

// NewFormatedError creates a new error using Sprintf
func NewFormatedError(format string, a ...interface{}) error {
	return NewCustomError(fmt.Sprintf(format, a...))
}

// NewTypeNotFoundError returns an error for a TOCEntry with an unknown type
func NewTypeNotFoundError(entry TOCEntry) error {
	return NewFormatedError("Type not found: %s %s", entry.Text, entry.ID)
}

// ExitIfError is a helper function for terminating if an error is not nil
func ExitIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// WarnIfError is a helper function for terminating if an error is not nil
func WarnIfError(err error) {
	if err != nil {
		LogWarning(err.Error())
	}
}
