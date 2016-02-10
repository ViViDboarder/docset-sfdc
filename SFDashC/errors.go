package main

import (
	"fmt"
	"os"
)

var shouldWarn = true

// Custom errors
type errorString struct {
	message string
}

// Error retrievies the Error message from the error
func (err errorString) Error() string {
	return err.message
}

// NoWarn disables all warning output
func WithoutWarning() {
	shouldWarn = false
}

// NewCustomError creates a custom error using a string as the message
func NewCustomError(message string) error {
	return &errorString{message}
}

// NewFormatedError creates a new error using Sprintf
func NewFormatedError(format string, a ...interface{}) error {
	return NewCustomError(fmt.Sprintf(format, a...))
}

// NewTypeNotFoundError returns an error for a TOCEntry with an unknown type
func NewTypeNotFoundError(entry TOCEntry) error {
	return NewFormatedError("Type not found : %s %s", entry.Text, entry.ID)
}

// ExitIfError is a helper function for terminating if an error is not nil
func ExitIfError(err error) {
	if err != nil {
		fmt.Println("ERROR :", err)
		os.Exit(1)
	}
}

// WarnIfError is a helper function for terminating if an error is not nil
func WarnIfError(err error) {
	if err != nil && shouldWarn {
		fmt.Println("WARNING :", err)
	}
}
