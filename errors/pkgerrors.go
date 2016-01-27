package pkgerrors // import "go.pedge.io/pkg/errors"

import (
	"bytes"
	"errors"
	"fmt"
)

// With adds the key/value pairs and returns a new Builder.
func With(keyValues ...interface{}) *Builder {
	return &Builder{getFields(keyValues, nil)}
}

// New creates a new error.
func New(args ...interface{}) error {
	return errors.New(fmt.Sprint(args...))
}

// Builder is an error builder.
type Builder struct {
	fields map[string]interface{}
}

// With adds the key/value pairs and returns a new Builder.
func (b *Builder) With(keyValues ...interface{}) *Builder {
	return &Builder{getFields(keyValues, b.fields)}
}

// New creates a new error.
func (b *Builder) New(args ...interface{}) error {
	buffer := bytes.NewBuffer(nil)
	var message string
	if len(args) > 0 {
		message = fmt.Sprint(args...)
	}
	if message != "" {
		_, _ = buffer.WriteString(message)
		if len(b.fields) > 0 {
			_ = buffer.WriteByte(' ')
		}
	}
	first := false
	for key, value := range b.fields {
		if !first {
			_ = buffer.WriteByte(' ')
			first = true
		}
		_, _ = buffer.WriteString(key)
		_ = buffer.WriteByte('=')
		_, _ = buffer.WriteString(fmt.Sprintf("%v", value))
	}
	return errors.New(buffer.String())
}

func getFields(keyValues []interface{}, existingFields map[string]interface{}) map[string]interface{} {
	if len(keyValues)%2 != 0 {
		keyValues = append(keyValues, "MISSING")
	}
	fields := make(map[string]interface{}, len(keyValues)/2)
	for i := 0; i < len(keyValues); i += 2 {
		fields[fmt.Sprintf("%v", keyValues[i])] = keyValues[i+1]
	}
	for key, value := range existingFields {
		fields[key] = value
	}
	return fields
}
