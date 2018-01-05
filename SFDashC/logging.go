package main

import (
	"fmt"
	"log"
)

const (
	prefix = "SFDashC"
	// Log Levels
	ERROR   = iota
	WARNING = iota
	INFO    = iota
	DEBUG   = iota
)

var logLevel int

func init() {
	logLevel = INFO
}

func getLevelText() string {
	switch logLevel {
	case ERROR:
		return "ERROR"
	case WARNING:
		return "WARNING"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

func getLogPrefix() string {
	return fmt.Sprintf("%s: %s:", prefix, getLevelText())
}

// SetLogLevel will set the maximum level to print
func SetLogLevel(level int) {
	logLevel = level
}

// Log will print a formatted message with a prefix for a specified level
// If the level is greater than the maximum log level, it will not print
func Log(level int, format string, a ...interface{}) {
	if level <= logLevel {
		message := fmt.Sprintf(format, a...)
		message = fmt.Sprintf("%s: %s: %s", prefix, getLevelText(), message)
		log.Println(message)
	}
}

// LogError will print an error message
// If the level is greater than the maximum log level, it will not print
// It is recommended to use log.Fatal() instead since it will handle exits for you
func LogError(format string, a ...interface{}) {
	Log(ERROR, format, a...)
}

// LogInfo will print a warning message
// If the level is greater than the maximum log level, it will not print
func LogWarning(format string, a ...interface{}) {
	Log(WARNING, format, a...)
}

// LogInfo will print an info message
// If the level is greater than the maximum log level, it will not print
func LogInfo(format string, a ...interface{}) {
	Log(INFO, format, a...)
}

// LogDebug will print an debug message
// If the level is greater than the maximum log level, it will not print
func LogDebug(format string, a ...interface{}) {
	Log(DEBUG, format, a...)
}
