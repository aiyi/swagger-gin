package errors

import (
	"fmt"
	"net/http"
)

// Validation represents a failure of a precondition
type Validation struct {
	Code    int32
	Name    string
	In      string
	Value   interface{}
	Message string
	Values  []interface{}
}

func (e *Validation) Error() string {
	return e.Message
}

const (
	contentTypeFail    = `unsupported media type %q, only %v are allowed`
	responseFormatFail = `unsupported media type requested, only %v are available`
)

// InvalidContentType error for an invalid content type
func InvalidContentType(value string, allowed []string) *Validation {
	var values []interface{}
	for _, v := range allowed {
		values = append(values, v)
	}
	return &Validation{
		Code:    http.StatusUnsupportedMediaType,
		Name:    "Content-Type",
		In:      "header",
		Value:   value,
		Values:  values,
		Message: fmt.Sprintf(contentTypeFail, value, allowed),
	}
}

// InvalidResponseFormat error for an unacceptable response format request
func InvalidResponseFormat(value string, allowed []string) *Validation {
	var values []interface{}
	for _, v := range allowed {
		values = append(values, v)
	}
	return &Validation{
		Code:    http.StatusNotAcceptable,
		Name:    "Accept",
		In:      "header",
		Value:   value,
		Values:  values,
		Message: fmt.Sprintf(responseFormatFail, allowed),
	}
}
