package pkgerrors // import "go.pedge.io/pkg/errors"

import (
	"bytes"
	"fmt"
)

// Error is an error.
type Error struct {
	Fields  map[string]interface{}
	Message string
}

// New creates a new Error with the given args as the message.
func New(args ...interface{}) *Error {
	return &Error{
		nil,
		fmt.Sprint(args...),
	}
}

// Error returns an error string.
func (e *Error) Error() string {
	buffer := bytes.NewBuffer(nil)
	if e.Message != "" {
		_, _ = buffer.WriteString(e.Message)
		if len(e.Fields) > 0 {
			_ = buffer.WriteByte(' ')
		}
	}
	first := false
	for key, value := range e.Fields {
		if !first {
			_ = buffer.WriteByte(' ')
			first = true
		}
		_, _ = buffer.WriteString(key)
		_ = buffer.WriteByte('=')
		_, _ = buffer.WriteString(fmt.Sprintf("%v", value))
	}
	return buffer.String()
}

// With adds the key/value pairs and returns a new Error.
func (e *Error) With(keyValues ...interface{}) *Error {
	if len(keyValues)%2 != 0 {
		keyValues = append(keyValues, "MISSING")
	}
	fields := make(map[string]interface{}, len(keyValues)/2)
	for i := 0; i < len(keyValues); i += 2 {
		fields[fmt.Sprintf("%v", keyValues[i])] = keyValues[i+1]
	}
	for key, value := range e.Fields {
		fields[key] = value
	}
	return &Error{
		fields,
		e.Message,
	}
}
