package errors

import "errors"

var (
	NoSuchCommandError  = errors.New("no such command")
	NoPermissionError   = errors.New("insufficient permissions")
	IncorrectUsageError = errors.New("incorrect command usage")
	UnknownError        = errors.New("unknown error")
)
